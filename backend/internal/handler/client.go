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
	clients, err := h.repo.GetAll(c.Request.Context())
	if err != nil {
		respondInternal(c, "ClientHandler.GetAll", err)
		return
	}
	respondOK(c, clients)
}

// GetByID — обработчик GET /api/v1/clients/:id
func (h *ClientHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	client, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		// отличаем «не найдено» от реального сбоя БД
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(c, http.StatusNotFound, "клиент не найден")
			return
		}
		respondInternal(c, "ClientHandler.GetByID", err)
		return
	}
	respondOK(c, client)
}

// Create — обработчик POST /api/v1/clients
func (h *ClientHandler) Create(c *gin.Context) {
	var dto domain.CreateClientRequest
	if err := c.ShouldBindJSON(&dto); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Пока поддерживаются только физлица. Для организаций нужны свои
	// поля (name/inn/kpp) и запись в таблицу organizations — это отдельная
	// задача. Раньше код молча создавал clients-запись без organizations
	// (orphan). Теперь честно отклоняем, не портя данные.
	switch dto.ClientType {
	case domain.ClientIndividual:
		if dto.LastName == "" || dto.FirstName == "" {
			respondError(c, http.StatusBadRequest, "для физлица обязательны last_name и first_name")
			return
		}
	case domain.ClientOrganization:
		respondError(c, http.StatusNotImplemented, "создание клиентов-организаций пока не реализовано")
		return
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
