// Авторизация: контекст с текущим пользователем (из JWT), вход и выход.
// Личность и роль берём из самого токена (claims), а не из тела ответа —
// роль (role_code) там уже есть и именно она определяет доступ.

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react"
import { api, getToken, setToken, setUnauthorizedHandler } from "./api"
import type { LoginResponse, RoleCode } from "./types"

export interface AuthUser {
  employeeId: string
  login: string
  role: RoleCode
}

interface JwtClaims {
  employee_id: string
  login: string
  role_code: RoleCode
  exp: number
}

// Разбор payload JWT (без проверки подписи — её делает бэкенд;
// фронту нужны только claims для отрисовки и срока годности).
function decodeToken(token: string): AuthUser | null {
  try {
    const payload = token.split(".")[1]
    const json = atob(payload.replace(/-/g, "+").replace(/_/g, "/"))
    const claims = JSON.parse(json) as JwtClaims
    if (claims.exp * 1000 < Date.now()) return null // истёк
    return {
      employeeId: claims.employee_id,
      login: claims.login,
      role: claims.role_code,
    }
  } catch {
    return null
  }
}

interface AuthContextValue {
  user: AuthUser | null
  login: (login: string, password: string) => Promise<void>
  logout: () => void
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(() => {
    const t = getToken()
    return t ? decodeToken(t) : null
  })

  const logout = useCallback(() => {
    setToken(null)
    setUser(null)
  }, [])

  // на 401 из любого запроса — разлогиниваем
  useEffect(() => {
    setUnauthorizedHandler(() => setUser(null))
    return () => setUnauthorizedHandler(null)
  }, [])

  const login = useCallback(async (loginName: string, password: string) => {
    const res = await api.post<LoginResponse>("/auth/login", {
      login: loginName,
      password,
    })
    setToken(res.token)
    const u = decodeToken(res.token)
    if (!u) throw new Error("получен некорректный токен")
    setUser(u)
  }, [])

  const value = useMemo(() => ({ user, login, logout }), [user, login, logout])
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

// eslint-disable-next-line react-refresh/only-export-components
export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error("useAuth должен использоваться внутри AuthProvider")
  return ctx
}
