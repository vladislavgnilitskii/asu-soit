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
	employeeID, ok := employeeIDFromCtx(c)
	if !ok {
		return
	}
	if err := h.repo.UpdateStatus(c.Request.Context(), id, dto, employeeID); err != nil {
		switch {
		case errors.Is(err, repository.ErrRequestNotFound):
			respondError(c, http.StatusNotFound, "заявка не найдена")
		case errors.Is(err, repository.ErrStatusNotFound):
			respondError(c, http.StatusBadRequest, "передан несуществующий status_id")
		default:
			respondInternal(c, "RequestHandler.UpdateStatus", err)
		}
		return
	}
	respondOK(c, gin.H{"message": "статус обновлён"})
}

// Assign — PATCH /requests/:id/assign — назначить исполнителя
func (h *RequestHandler) Assign(c *gin.Context) {
	id := c.Param("id")
	var dto domain.AssignRequestDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.repo.Assign(c.Request.Context(), id, dto.AssignedTo); err != nil {
		if errors.Is(err, repository.ErrRequestNotFound) {
			respondError(c, http.StatusNotFound, "заявка не найдена")
			return
		}
		respondInternal(c, "RequestHandler.Assign", err)
		return
	}
	respondOK(c, gin.H{"message": "исполнитель назначен"})
}

// UpdateDetails — PATCH /requests/:id — диагностика и стоимость
func (h *RequestHandler) UpdateDetails(c *gin.Context) {
	id := c.Param("id")
	var dto domain.UpdateRequestDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	if dto.IsEmpty() {
		respondError(c, http.StatusBadRequest, "не передано ни одного поля для обновления")
		return
	}
	req, err := h.repo.UpdateDetails(c.Request.Context(), id, dto)
	if err != nil {
		if errors.Is(err, repository.ErrRequestNotFound) {
			respondError(c, http.StatusNotFound, "заявка не найдена")
			return
		}
		respondInternal(c, "RequestHandler.UpdateDetails", err)
		return
	}
	respondOK(c, req)
}

// Close — PATCH /requests/:id/close — закрыть заявку
func (h *RequestHandler) Close(c *gin.Context) {
	id := c.Param("id")
	var dto domain.CloseRequestDTO
	// тело необязательно — пустой запрос допустим, ошибку разбора игнорируем
	_ = c.ShouldBindJSON(&dto)

	employeeID, ok := employeeIDFromCtx(c)
	if !ok {
		return
	}
	if err := h.repo.Close(c.Request.Context(), id, employeeID, dto.Comment); err != nil {
		if errors.Is(err, repository.ErrRequestNotFound) {
			respondError(c, http.StatusNotFound, "заявка не найдена")
			return
		}
		respondInternal(c, "RequestHandler.Close", err)
		return
	}
	respondOK(c, gin.H{"message": "заявка закрыта"})
}

// History — GET /requests/:id/history — история смены статусов
func (h *RequestHandler) History(c *gin.Context) {
	id := c.Param("id")
	history, err := h.repo.GetHistory(c.Request.Context(), id)
	if err != nil {
		respondInternal(c, "RequestHandler.History", err)
		return
	}
	respondOK(c, history)
}

// employeeIDFromCtx достаёт employee_id из JWT-контекста (кладёт RequireAuth).
// При отсутствии сам отвечает 401 и возвращает ok=false.
func employeeIDFromCtx(c *gin.Context) (string, bool) {
	employeeID := c.GetString("employee_id")
	if employeeID == "" {
		respondError(c, http.StatusUnauthorized, "не удалось определить сотрудника из токена")
		return "", false
	}
	return employeeID, true
}
