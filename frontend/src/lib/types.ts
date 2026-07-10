// Типы данных API — зеркало DTO Go-бэкенда (backend/internal/domain).
// Держим в синхроне с сервером вручную (фронт и бэк — разные языки).

// Page — обёртка постраничного ответа (зеркало domain.Page[T]).
export interface Page<T> {
  items: T[]
  total: number
  limit: number
  offset: number
}

// RequestStats — счётчики для дашборда (GET /requests/stats).
export interface RequestStats {
  total: number
  open: number
  closed: number
}

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

export interface DeviceType {
  id: string
  name: string
}

export interface Device {
  id: string
  client_id: string
  device_type_id: string
  brand: string
  model: string
  serial_number?: string | null
  appearance_note?: string | null
  created_at: string
}

export interface StatusHistoryEntry {
  id: string
  status_id: string
  status_code: string
  status_name: string
  changed_by: string
  changed_by_name: string
  changed_at: string
  comment?: string | null
}

export interface RequestPartEntry {
  id: string
  part_id: string
  part_name: string
  quantity: number
  unit_price: number
}

export interface Invoice {
  id: string
  request_id: string
  total_amount: number
  status: "pending" | "paid" | "cancelled"
  issued_at: string
}

export interface PartCategory {
  id: string
  name: string
}

export interface SparePart {
  id: string
  category_id: string
  name: string
  sku?: string | null
  purchase_price: number
  sale_price: number
  quantity_in_stock: number
  created_at: string
}

export interface EngineerListItem {
  id: string
  last_name: string
  first_name: string
  middle_name?: string
}

// --- DTO создания (зеркало binding-структур бэка) ---

export interface CreateClientRequest {
  client_type: ClientType
  phone: string
  email?: string
  // физлицо
  last_name?: string
  first_name?: string
  middle_name?: string
  // организация
  name?: string
  inn?: string
  kpp?: string
  contact_person?: string
}

export interface CreateDeviceDTO {
  client_id: string
  device_type_id: string
  brand: string
  model: string
  serial_number?: string
  appearance_note?: string
}

export interface CreateRepairRequestDTO {
  device_id: string
  problem_description: string
  planned_deadline?: string | null
}

export interface UpdateRequestDTO {
  diagnostic_result?: string
  estimated_cost?: number
  final_cost?: number
}

export interface CreateSparePartDTO {
  category_id: string
  name: string
  sku?: string
  purchase_price: number
  sale_price: number
}

export interface ReceivePartsDTO {
  quantity: number
  unit_price: number
  invoice_number?: string
  note?: string
}

export interface WriteOffPartsDTO {
  quantity: number
  note?: string
}

export interface IssuePartToRequestDTO {
  part_id: string
  quantity: number
}

export interface CreateEmployeeDTO {
  role_id: string
  last_name: string
  first_name: string
  middle_name?: string
  login: string
  password: string
}
