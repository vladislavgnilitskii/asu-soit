package domain

import "time"

type ClientType string

const (
	ClientIndividual   ClientType = "individual"
	ClientOrganization ClientType = "organization"
)

type Client struct {
	ID         string     `json:"id"`
	ClientType ClientType `json:"client_type"`
	Phone      string     `json:"phone"`
	Email      *string    `json:"email,omitempty"` // nullable в БД — указатель
	CreatedAt  time.Time  `json:"created_at"`
}

type Individual struct {
	ID         string  `json:"id"`
	ClientID   string  `json:"client_id"`
	LastName   string  `json:"last_name"`
	FirstName  string  `json:"first_name"`
	MiddleName *string `json:"middle_name,omitempty"` // nullable в БД — указатель
}

type Organization struct {
	ID            string  `json:"id"`
	ClientID      string  `json:"client_id"`
	Name          string  `json:"name"`
	INN           string  `json:"inn"`
	KPP           *string `json:"kpp,omitempty"`            // nullable
	ContactPerson *string `json:"contact_person,omitempty"` // nullable
}

// ClientDetails — клиент вместе с данными его подтипа.
// Заполнено ровно одно из полей Individual/Organization — в зависимости
// от client_type. Используется в ответе GET /clients/:id.
type ClientDetails struct {
	Client
	Individual   *Individual   `json:"individual,omitempty"`
	Organization *Organization `json:"organization,omitempty"`
}

type CreateClientRequest struct {
	ClientType ClientType `json:"client_type" binding:"required"`
	Phone      string     `json:"phone"       binding:"required"`
	Email      string     `json:"email"`

	// поля физлица (client_type = individual)
	LastName   string `json:"last_name"`
	FirstName  string `json:"first_name"`
	MiddleName string `json:"middle_name"`

	// поля организации (client_type = organization)
	Name          string `json:"name"`
	INN           string `json:"inn"`
	KPP           string `json:"kpp"`
	ContactPerson string `json:"contact_person"`
}

type RepairRequest struct {
	ID                 string     `json:"id"`
	DeviceID           string     `json:"device_id"`
	AssignedTo         *string    `json:"assigned_to,omitempty"`
	StatusID           string     `json:"status_id"`
	ProblemDescription string     `json:"problem_description"`
	DiagnosticResult   *string    `json:"diagnostic_result,omitempty"`
	EstimatedCost      *float64   `json:"estimated_cost,omitempty"`
	FinalCost          *float64   `json:"final_cost,omitempty"`
	PlannedDeadline    *time.Time `json:"planned_deadline,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	ClosedAt           *time.Time `json:"closed_at,omitempty"`
}

type CreateRepairRequestDTO struct {
	// клиент не передаётся: он определяется по устройству (device → client)
	DeviceID           string     `json:"device_id"           binding:"required"`
	ProblemDescription string     `json:"problem_description" binding:"required"`
	PlannedDeadline    *time.Time `json:"planned_deadline"`
}

type UpdateRequestStatusDTO struct {
	StatusID string `json:"status_id" binding:"required"`
	Comment  string `json:"comment"`
}

// AssignRequestDTO — назначение исполнителя на заявку
type AssignRequestDTO struct {
	AssignedTo string `json:"assigned_to" binding:"required"`
}

// UpdateRequestDTO — частичное обновление диагностики и стоимости.
// Поля-указатели: nil означает «не менять», непустое — новое значение.
type UpdateRequestDTO struct {
	DiagnosticResult *string  `json:"diagnostic_result"`
	EstimatedCost    *float64 `json:"estimated_cost"`
	FinalCost        *float64 `json:"final_cost"`
}

// IsEmpty — не передано ни одного поля для обновления
func (d UpdateRequestDTO) IsEmpty() bool {
	return d.DiagnosticResult == nil && d.EstimatedCost == nil && d.FinalCost == nil
}

// CloseRequestDTO — закрытие заявки (комментарий необязателен)
type CloseRequestDTO struct {
	Comment string `json:"comment"`
}

// StatusHistoryEntry — запись истории смены статуса (для чтения).
// Обогащена кодом/названием статуса и именем сотрудника.
type StatusHistoryEntry struct {
	ID            string    `json:"id"`
	StatusID      string    `json:"status_id"`
	StatusCode    string    `json:"status_code"`
	StatusName    string    `json:"status_name"`
	ChangedBy     string    `json:"changed_by"`
	ChangedByName string    `json:"changed_by_name"`
	ChangedAt     time.Time `json:"changed_at"`
	Comment       *string   `json:"comment,omitempty"`
}

// Employee — сотрудник
type Employee struct {
	ID           string `json:"id"`
	RoleID       string `json:"role_id"`
	LastName     string `json:"last_name"`
	FirstName    string `json:"first_name"`
	MiddleName   string `json:"middle_name,omitempty"`
	Login        string `json:"login"`
	PasswordHash string `json:"-"` // json:"-" — никогда не отдавать в ответе
	IsActive     bool   `json:"is_active"`
}

// CreateEmployeeDTO — данные для заведения нового сотрудника.
// Password приходит в открытом виде и хешируется в хендлере (bcrypt) —
// в БД уходит только хеш, plain-пароль нигде не хранится.
type CreateEmployeeDTO struct {
	RoleID     string `json:"role_id"     binding:"required"`
	LastName   string `json:"last_name"   binding:"required"`
	FirstName  string `json:"first_name"  binding:"required"`
	MiddleName string `json:"middle_name"`
	Login      string `json:"login"       binding:"required"`
	Password   string `json:"password"    binding:"required,min=6"`
}

// UpdateEmployeeDTO — частичное обновление профиля/роли/активности.
// Указатели: nil = «не менять». Пароль здесь не меняется (отдельная задача).
type UpdateEmployeeDTO struct {
	RoleID     *string `json:"role_id"`
	LastName   *string `json:"last_name"`
	FirstName  *string `json:"first_name"`
	MiddleName *string `json:"middle_name"`
	IsActive   *bool   `json:"is_active"`
}

// LoginRequest — данные для входа
type LoginRequest struct {
	Login    string `json:"login"    binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse — ответ с токеном
type LoginResponse struct {
	Token    string   `json:"token"`
	Employee Employee `json:"employee"`
}

// Claims — данные которые хранятся внутри JWT токена
type Claims struct {
	EmployeeID string `json:"employee_id"`
	Login      string `json:"login"`
	RoleCode   string `json:"role_code"`
}

// DeviceType — справочник типов устройств (ноутбук, телефон, ...)
type DeviceType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Device — устройство клиента, принятое в ремонт
type Device struct {
	ID             string    `json:"id"`
	ClientID       string    `json:"client_id"`
	DeviceTypeID   string    `json:"device_type_id"`
	Brand          string    `json:"brand"`
	Model          string    `json:"model"`
	SerialNumber   *string   `json:"serial_number,omitempty"`   // nullable
	AppearanceNote *string   `json:"appearance_note,omitempty"` // nullable
	CreatedAt      time.Time `json:"created_at"`
}

type CreateDeviceDTO struct {
	ClientID       string `json:"client_id"      binding:"required"`
	DeviceTypeID   string `json:"device_type_id" binding:"required"`
	Brand          string `json:"brand"          binding:"required"`
	Model          string `json:"model"          binding:"required"`
	SerialNumber   string `json:"serial_number"`
	AppearanceNote string `json:"appearance_note"`
}

// PartCategory — справочник категорий запчастей
type PartCategory struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// SparePart — запчасть на складе
type SparePart struct {
	ID              string    `json:"id"`
	CategoryID      string    `json:"category_id"`
	Name            string    `json:"name"`
	SKU             *string   `json:"sku,omitempty"` // nullable, артикул
	PurchasePrice   float64   `json:"purchase_price"`
	SalePrice       float64   `json:"sale_price"`
	QuantityInStock int       `json:"quantity_in_stock"`
	CreatedAt       time.Time `json:"created_at"`
}

type CreateSparePartDTO struct {
	CategoryID    string  `json:"category_id"    binding:"required"`
	Name          string  `json:"name"            binding:"required"`
	SKU           string  `json:"sku"`
	PurchasePrice float64 `json:"purchase_price"  binding:"required"`
	SalePrice     float64 `json:"sale_price"      binding:"required"`
}

// ReceivePartsDTO — приход детали от поставщика (увеличивает остаток)
type ReceivePartsDTO struct {
	Quantity      int     `json:"quantity"       binding:"required,gt=0"`
	UnitPrice     float64 `json:"unit_price"     binding:"required"`
	InvoiceNumber string  `json:"invoice_number"`
	Note          string  `json:"note"`
}

// WriteOffPartsDTO — списание детали (порча/недостача), без привязки к заявке
type WriteOffPartsDTO struct {
	Quantity int    `json:"quantity" binding:"required,gt=0"`
	Note     string `json:"note"`
}

// IssuePartToRequestDTO — выдать деталь со склада в конкретную заявку.
// Цену клиенту (unit_price) сервер берёт сам из SparePart.SalePrice —
// не принимает от вызывающего, чтобы никто не мог занизить цену в счёте.
type IssuePartToRequestDTO struct {
	PartID   string `json:"part_id"  binding:"required"`
	Quantity int    `json:"quantity" binding:"required,gt=0"`
}

// RequestPartEntry — деталь, списанная на заявку (для чтения).
// Обогащена названием детали — аналогично StatusHistoryEntry.
type RequestPartEntry struct {
	ID        string  `json:"id"`
	PartID    string  `json:"part_id"`
	PartName  string  `json:"part_name"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

type Invoice struct {
	ID          string    `json:"id"`
	RequestID   string    `json:"request_id"`
	TotalAmount float64   `json:"total_amount"`
	Status      string    `json:"status"`
	IssuedAt    time.Time `json:"issued_at"`
}

const (
	InvoicePending   = "pending"
	InvoicePaid      = "paid"
	InvoiceCancelled = "cancelled"
)

type UpdateInvoiceStatusDTO struct {
	Status string `json:"status" binding:"required"`
}
