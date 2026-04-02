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
	Email      string     `json:"email,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type Individual struct {
	ID         string `json:"id"`
	ClientID   string `json:"client_id"`
	LastName   string `json:"last_name"`
	FirstName  string `json:"first_name"`
	MiddleName string `json:"middle_name,omitempty"`
}

type CreateClientRequest struct {
	ClientType ClientType `json:"client_type" binding:"required"`
	Phone      string     `json:"phone"       binding:"required"`
	Email      string     `json:"email"`
	LastName   string     `json:"last_name"`
	FirstName  string     `json:"first_name"`
	MiddleName string     `json:"middle_name"`
}

type RepairRequest struct {
	ID                 string     `json:"id"`
	ClientID           string     `json:"client_id"`
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
	ClientID           string     `json:"client_id"           binding:"required"`
	DeviceID           string     `json:"device_id"           binding:"required"`
	ProblemDescription string     `json:"problem_description" binding:"required"`
	PlannedDeadline    *time.Time `json:"planned_deadline"`
}

type UpdateRequestStatusDTO struct {
	StatusID string `json:"status_id" binding:"required"`
	Comment  string `json:"comment"`
}
