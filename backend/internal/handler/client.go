package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/vladislavgnilitskii/asu-soit/internal/domain"
	"github.com/vladislavgnilitskii/asu-soit/internal/repository"
)

type ClientHandler struct {
	repo *repository.ClientRepository
}

// NewClientHandler — конструктор
func NewClientHandler(repo *repository.ClientRepository) *ClientHandler {
	return &ClientHandler{repo: repo}
}

// GetAll — обработчик GET /api/v1/clients
func (h *ClientHandler) GetAll(c *gin.Context) {
	p := parsePageParams(c)
	clients, total, err := h.repo.GetAll(c.Request.Context(), p)
	if err != nil {
		respondInternal(c, "ClientHandler.GetAll", err)
		return
	}
	respondOK(c, domain.NewPage(clients, total, p))
}

// GetByID — обработчик GET /api/v1/clients/:id
// возвращает клиента вместе с данными его подтипа (individual/organization)
func (h *ClientHandler) GetByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	client, err := h.repo.GetByID(ctx, id)
	if err != nil {
		// отличаем «не найдено» от реального сбоя БД
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(c, http.StatusNotFound, "клиент не найден")
			return
		}
		respondInternal(c, "ClientHandler.GetByID", err)
		return
	}

	details := domain.ClientDetails{Client: *client}
	switch client.ClientType {
	case domain.ClientIndividual:
		ind, err := h.repo.GetIndividual(ctx, id)
		if err != nil {
			respondInternal(c, "ClientHandler.GetByID individual", err)
			return
		}
		details.Individual = ind
	case domain.ClientOrganization:
		org, err := h.repo.GetOrganization(ctx, id)
		if err != nil {
			respondInternal(c, "ClientHandler.GetByID organization", err)
			return
		}
		details.Organization = org
	}

	respondOK(c, details)
}

// Create — обработчик POST /api/v1/clients
func (h *ClientHandler) Create(c *gin.Context) {
	var dto domain.CreateClientRequest
	if err := c.ShouldBindJSON(&dto); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	// поля обязательные для подтипа проверяем вручную: в одном DTO
	// сосуществуют поля физлица и организации, поэтому binding:"required"
	// на них навесить нельзя.
	switch dto.ClientType {
	case domain.ClientIndividual:
		if dto.LastName == "" || dto.FirstName == "" {
			respondError(c, http.StatusBadRequest, "для физлица обязательны last_name и first_name")
			return
		}
	case domain.ClientOrganization:
		if dto.Name == "" || dto.INN == "" {
			respondError(c, http.StatusBadRequest, "для организации обязательны name и inn")
			return
		}
	default:
		respondError(c, http.StatusBadRequest, "неизвестный client_type")
		return
	}

	client, err := h.repo.Create(c.Request.Context(), dto)
	if err != nil {
		respondInternal(c, "ClientHandler.Create", err)
		return
	}
	respondCreated(c, client)
}
