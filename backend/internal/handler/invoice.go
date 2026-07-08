package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vladislavgnilitskii/asu-soit/internal/domain"
	"github.com/vladislavgnilitskii/asu-soit/internal/repository"
)

type InvoiceHandler struct {
	repo *repository.InvoiceRepository
}

func NewInvoiceHandler(repo *repository.InvoiceRepository) *InvoiceHandler {
	return &InvoiceHandler{repo: repo}
}

func (h *InvoiceHandler) GetByID(c *gin.Context) {
	id, err := h.repo.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondInvoiceError(c, "InvoiceHandler.GetByID", err)
		return
	}
	respondOK(c, id)
}

func (h *InvoiceHandler) GetByRequestID(c *gin.Context) {
	request, err := h.repo.GetByRequestID(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondInvoiceError(c, "InvoiceHandler.GetByRequestID", err)
		return
	}
	respondOK(c, request)
}

func (h *InvoiceHandler) CreateForRequest(c *gin.Context) {
	request, err := h.repo.CreateForRequest(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondInvoiceError(c, "InvoiceHandler.CreateForRequest", err)
		return
	}
	respondCreated(c, request)
}

func (h *InvoiceHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var dto domain.UpdateInvoiceStatusDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	if dto.Status != domain.InvoicePaid && dto.Status != domain.InvoiceCancelled {
		respondError(c, http.StatusBadRequest, "статус можно сменить только на paid или cancelled")
		return
	}

	invoice, err := h.repo.UpdateStatus(c.Request.Context(), id, dto.Status)
	if err != nil {
		respondInvoiceError(c, "InvoiceHandler.UpdateStatus", err)
		return
	}
	respondOK(c, invoice)
}

// respondInvoiceError — маппинг доменных ошибок счёта в HTTP-коды (даю готовым)
func respondInvoiceError(c *gin.Context, where string, err error) {
	switch {
	case errors.Is(err, repository.ErrInvoiceNotFound):
		respondError(c, http.StatusNotFound, "счёт не найден")
	case errors.Is(err, repository.ErrRequestNotFound):
		respondError(c, http.StatusNotFound, "заявка не найдена")
	case errors.Is(err, repository.ErrRequestNotClosed):
		respondError(c, http.StatusConflict, "нельзя выставить счёт по незакрытой заявке")
	case errors.Is(err, repository.ErrInvoiceExists):
		respondError(c, http.StatusConflict, "по этой заявке счёт уже выставлен")
	case errors.Is(err, repository.ErrInvoiceNotPending):
		respondError(c, http.StatusConflict, "счёт уже не в статусе pending")
	default:
		respondInternal(c, where, err)
	}
}
