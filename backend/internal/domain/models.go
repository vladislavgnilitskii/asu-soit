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