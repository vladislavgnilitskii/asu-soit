import { NavLink, Outlet } from "react-router-dom"
import { useAuth } from "@/lib/auth"
import type { RoleCode } from "@/lib/types"
import { Button } from "@/components/ui/button"

interface NavItem {
  to: string
  label: string
  roles?: RoleCode[] // если не задано — доступно всем авторизованным
}

const NAV: NavItem[] = [
  { to: "/", label: "Обзор" },
  { to: "/requests", label: "Заявки" },
  { to: "/clients", label: "Клиенты", roles: ["admin", "manager"] },
  { to: "/devices", label: "Устройства", roles: ["admin", "manager", "engineer"] },
]

const ROLE_LABEL: Record<RoleCode, string> = {
  admin: "Администратор",
  manager: "Руководитель",
  engineer: "Инженер",
  storekeeper: "Кладовщик",
  accountant: "Бухгалтер",
  sysadmin: "Системный администратор",
}

export function AppLayout() {
  const { user, logout } = useAuth()
  const items = NAV.filter((i) => !i.roles || (user && i.roles.includes(user.role)))

  return (
    <div className="flex min-h-screen">
      <aside className="w-60 shrink-0 border-r bg-sidebar text-sidebar-foreground flex flex-col">
        <div className="h-14 flex items-center px-4 border-b font-semibold">
          АСУ СОИТ
        </div>
        <nav className="flex-1 p-2 space-y-1">
          {items.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.to === "/"}
              className={({ isActive }) =>
                `block rounded-md px-3 py-2 text-sm transition-colors ${
                  isActive
                    ? "bg-sidebar-accent text-sidebar-accent-foreground font-medium"
                    : "hover:bg-sidebar-accent/60"
                }`
              }
            >
              {item.label}
            </NavLink>
          ))}
        </nav>
      </aside>

      <div className="flex-1 flex flex-col">
        <header className="h-14 border-b flex items-center justify-between px-6">
          <span className="text-sm text-muted-foreground">
            Сервисный центр «ТехноСервис»
          </span>
          <div className="flex items-center gap-3">
            <span className="text-sm">
              {user?.login}
              {user && (
                <span className="ml-2 text-muted-foreground">
                  {ROLE_LABEL[user.role]}
                </span>
              )}
            </span>
            <Button variant="outline" size="sm" onClick={logout}>
              Выйти
            </Button>
          </div>
        </header>
        <main className="flex-1 p-6 bg-background">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
