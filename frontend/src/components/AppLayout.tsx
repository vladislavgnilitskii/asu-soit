import { NavLink, Outlet, useLocation } from "react-router-dom"
import { Wrench } from "lucide-react"
import { useAuth } from "@/lib/auth"
import type { RoleCode } from "@/lib/types"
import { Button } from "@/components/ui/button"
import { ThemeToggle } from "@/components/ThemeToggle"

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
  { to: "/warehouse", label: "Склад", roles: ["admin", "storekeeper", "engineer"] },
  { to: "/employees", label: "Сотрудники", roles: ["admin"] },
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
  const location = useLocation()
  const items = NAV.filter((i) => !i.roles || (user && i.roles.includes(user.role)))

  return (
    <div className="flex min-h-screen">
      <aside className="w-60 shrink-0 border-r border-sidebar-border bg-sidebar text-sidebar-foreground flex flex-col">
        <div className="h-14 flex items-center gap-2.5 px-4 border-b border-sidebar-border">
          <span className="flex size-8 items-center justify-center rounded-md bg-sidebar-primary text-sidebar-primary-foreground">
            <Wrench className="size-4.5" />
          </span>
          <div className="flex flex-col leading-tight">
            <span className="text-sm font-semibold">АСУ СОИТ</span>
            <span className="text-[11px] text-sidebar-foreground/60">
              ТехноСервис
            </span>
          </div>
        </div>
        <nav className="flex-1 p-2 space-y-0.5">
          {items.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.to === "/"}
              className={({ isActive }) =>
                `relative block rounded-md px-3 py-2 text-sm transition-colors ${
                  isActive
                    ? "bg-sidebar-accent text-sidebar-accent-foreground font-medium before:absolute before:left-0 before:top-1/2 before:h-4 before:w-0.5 before:-translate-y-1/2 before:rounded-full before:bg-sidebar-primary"
                    : "text-sidebar-foreground/80 hover:bg-sidebar-accent/50 hover:text-sidebar-accent-foreground"
                }`
              }
            >
              {item.label}
            </NavLink>
          ))}
        </nav>
      </aside>

      <div className="flex-1 flex flex-col min-w-0">
        <header className="sticky top-0 z-10 h-14 border-b bg-background/80 backdrop-blur flex items-center justify-between px-6">
          <span className="text-sm text-muted-foreground">
            Сервисный центр «ТехноСервис»
          </span>
          <div className="flex items-center gap-2">
            <span className="text-sm mr-1">
              {user?.login}
              {user && (
                <span className="ml-2 text-muted-foreground">
                  {ROLE_LABEL[user.role]}
                </span>
              )}
            </span>
            <ThemeToggle />
            <Button variant="outline" size="sm" onClick={logout}>
              Выйти
            </Button>
          </div>
        </header>
        <main key={location.pathname} className="flex-1 p-6 bg-background animate-fade-in">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
