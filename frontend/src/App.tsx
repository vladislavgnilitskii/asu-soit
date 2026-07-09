import { Navigate, Outlet, Route, Routes } from "react-router-dom"
import { useAuth } from "./lib/auth"
import { AppLayout } from "./components/AppLayout"
import { LoginPage } from "./pages/LoginPage"
import { DashboardPage } from "./pages/DashboardPage"
import { RequestsPage } from "./pages/RequestsPage"
import { RequestDetailPage } from "./pages/RequestDetailPage"
import { ClientsPage } from "./pages/ClientsPage"
import { DevicesPage } from "./pages/DevicesPage"
import { WarehousePage } from "./pages/WarehousePage"
import { EmployeesPage } from "./pages/EmployeesPage"

// Пускает дальше только авторизованных, иначе — на страницу входа.
function ProtectedRoute() {
  const { user } = useAuth()
  if (!user) return <Navigate to="/login" replace />
  return <Outlet />
}

function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route element={<ProtectedRoute />}>
        <Route element={<AppLayout />}>
          <Route path="/" element={<DashboardPage />} />
          <Route path="/requests" element={<RequestsPage />} />
          <Route path="/requests/:id" element={<RequestDetailPage />} />
          <Route path="/clients" element={<ClientsPage />} />
          <Route path="/devices" element={<DevicesPage />} />
          <Route path="/warehouse" element={<WarehousePage />} />
          <Route path="/employees" element={<EmployeesPage />} />
        </Route>
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}

export default App
