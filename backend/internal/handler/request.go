package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	respondOK(c, requests)
}

func (h *RequestHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	req, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, http.StatusNotFound, "заявка не найдена")
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
		respondError(c, http.StatusInternalServerError, err.Error())
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
	// TODO: брать employeeID из JWT токена — сделаем когда добавим авторизацию
	employeeID := c.GetHeader("X-Employee-ID")
	if employeeID == "" {
		respondError(c, http.StatusBadRequest, "заголовок X-Employee-ID обязателен")
		return
	}
	if err := h.repo.UpdateStatus(c.Request.Context(), id, dto, employeeID); err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	respondOK(c, gin.H{"message": "статус обновлён"})
}
