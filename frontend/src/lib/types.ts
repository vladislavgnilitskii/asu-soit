// Типы данных API — зеркало DTO Go-бэкенда (backend/internal/domain).
// Держим в синхроне с сервером вручную (фронт и бэк — разные языки).

export type RoleCode =
  | "admin"
  | "manager"
  | "engineer"
  | "storekeeper"
  | "accountant"
  | "sysadmin"

export interface Employee {
  id: string
  role_id: string
  last_name: string
  first_name: string
  middle_name?: string
  login: string
  is_active: boolean
}

export interface LoginResponse {
  token: string
  employee: Employee
}

export type ClientType = "individual" | "organization"

export interface Client {
  id: string
  client_type: ClientType
  phone: string
  email?: string | null
  created_at: string
}

export interface RepairRequest {
  id: string
  device_id: string
  assigned_to?: string | null
  status_id: string
  problem_description: string
  diagnostic_result?: string | null
  estimated_cost?: number | null
  final_cost?: number | null
  planned_deadline?: string | null
  created_at: string
  closed_at?: string | null
}

export interface RequestStatus {
  id: string
  code: string
  name: string
  sort_order: number
}

export interface Role {
  id: string
  code: string
  name: string
}
