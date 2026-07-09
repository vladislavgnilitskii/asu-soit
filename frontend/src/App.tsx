import { lazy, Suspense } from "react"
import { Navigate, Outlet, Route, Routes } from "react-router-dom"
import { Loader2 } from "lucide-react"
import { useAuth } from "./lib/auth"
import { AppLayout } from "./components/AppLayout"
import { LoginPage } from "./pages/LoginPage"

// Экраны грузятся по требованию — меньше стартовый бандл (code splitting).
const DashboardPage = lazy(() =>
  import("./pages/DashboardPage").then((m) => ({ default: m.DashboardPage })),
)
const RequestsPage = lazy(() =>
  import("./pages/RequestsPage").then((m) => ({ default: m.RequestsPage })),
)
const RequestDetailPage = lazy(() =>
  import("./pages/RequestDetailPage").then((m) => ({
    default: m.RequestDetailPage,
  })),
)
const ClientsPage = lazy(() =>
  import("./pages/ClientsPage").then((m) => ({ default: m.ClientsPage })),
)
const DevicesPage = lazy(() =>
  import("./pages/DevicesPage").then((m) => ({ default: m.DevicesPage })),
)
const WarehousePage = lazy(() =>
  import("./pages/WarehousePage").then((m) => ({ default: m.WarehousePage })),
)
const EmployeesPage = lazy(() =>
  import("./pages/EmployeesPage").then((m) => ({ default: m.EmployeesPage })),
)

// Пускает дальше только авторизованных, иначе — на страницу входа.
function ProtectedRoute() {
  const { user } = useAuth()
  if (!user) return <Navigate to="/login" replace />
  return <Outlet />
}

// Заглушка на время подгрузки чанка экрана.
function RouteFallback() {
  return (
    <div className="flex items-center justify-center py-20 text-muted-foreground">
      <Loader2 className="size-5 animate-spin" />
    </div>
  )
}

function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route element={<ProtectedRoute />}>
        <Route element={<AppLayout />}>
          <Route
            path="/"
            element={
              <Suspense fallback={<RouteFallback />}>
                <DashboardPage />
              </Suspense>
            }
          />
          <Route
            path="/requests"
            element={
              <Suspense fallback={<RouteFallback />}>
                <RequestsPage />
              </Suspense>
            }
          />
          <Route
            path="/requests/:id"
            element={
              <Suspense fallback={<RouteFallback />}>
                <RequestDetailPage />
              </Suspense>
            }
          />
          <Route
            path="/clients"
            element={
              <Suspense fallback={<RouteFallback />}>
                <ClientsPage />
              </Suspense>
            }
          />
          <Route
            path="/devices"
            element={
              <Suspense fallback={<RouteFallback />}>
                <DevicesPage />
              </Suspense>
            }
          />
          <Route
            path="/warehouse"
            element={
              <Suspense fallback={<RouteFallback />}>
                <WarehousePage />
              </Suspense>
            }
          />
          <Route
            path="/employees"
            element={
              <Suspense fallback={<RouteFallback />}>
                <EmployeesPage />
              </Suspense>
            }
          />
        </Route>
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}

export default App
