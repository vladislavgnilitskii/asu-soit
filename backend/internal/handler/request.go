package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/vladislavgnilitskii/asu-soit/internal/domain"
	"github.com/vladislavgnilitskii/asu-soit/internal/repository"
)

type RequestHandler struct {
	repo *repository.RequestRepository
}

func NewRequestHandler(repo *repository.RequestRepository) *RequestHandler {
	return &RequestHandler{repo: repo}
}

func (h *RequestHandler) GetAll(c *gin.Context) {
	requests, err := h.repo.GetAll(c.Request.Context())
	if err != nil {
		respondInternal(c, "RequestHandler.GetAll", err)
		return
	}
	respondOK(c, requests)
}

func (h *RequestHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	req, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		// отличаем «не найдено» от реального сбоя БД
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(c, http.StatusNotFound, "заявка не найдена")
			return
		}
		respondInternal(c, "RequestHandler.GetByID", err)
		return
	}
	respondOK(c, req)
}

func (h *RequestHandler) Create(c *gin.Context) {
	var dto domain.CreateRepairRequestDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	req, err := h.repo.Create(c.Request.Context(), dto)
	if err != nil {
		respondInternal(c, "RequestHandler.Create", err)
		return
	}
	respondCreated(c, req)
}

func (h *RequestHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var dto domain.UpdateRequestStatusDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	// автора смены берём из JWT-контекста (его кладёт RequireAuth),
	// а не из клиентского ввода — иначе можно подделать чужой employee_id
	employeeID := c.GetString("employee_id")
	if employeeID == "" {
		respondError(c, http.StatusUnauthorized, "не удалось определить сотрудника из токена")
		return
	}
	if err := h.repo.UpdateStatus(c.Request.Context(), id, dto, employeeID); err != nil {
		respondInternal(c, "RequestHandler.UpdateStatus", err)
		return
	}
	respondOK(c, gin.H{"message": "статус обновлён"})
}
