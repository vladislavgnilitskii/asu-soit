package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vladislavgnilitskii/asu-soit/internal/domain"
	"github.com/vladislavgnilitskii/asu-soit/internal/repository"
)

type WarehouseHandler struct {
	repo *repository.WarehouseRepository
}

func NewWarehouseHandler(repo *repository.WarehouseRepository) *WarehouseHandler {
	return &WarehouseHandler{repo: repo}
}

// ListCategories — GET /part-categories
func (h *WarehouseHandler) ListCategories(c *gin.Context) {
	categories, err := h.repo.ListCategories(c.Request.Context())
	if err != nil {
		respondInternal(c, "WarehouseHandler.ListCategories", err)
		return
	}
	respondOK(c, categories)
}

// GetAllParts — GET /spare-parts
func (h *WarehouseHandler) GetAllParts(c *gin.Context) {
	parts, err := h.repo.GetAllParts(c.Request.Context())
	if err != nil {
		respondInternal(c, "WarehouseHandler.GetAllParts", err)
		return
	}
	respondOK(c, parts)
}

// GetPartByID — GET /spare-parts/:id
func (h *WarehouseHandler) GetPartByID(c *gin.Context) {
	part, err := h.repo.GetPartByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondWarehouseError(c, "WarehouseHandler.GetPartByID", err)
		return
	}
	respondOK(c, part)
}

// CreatePart — POST /spare-parts
func (h *WarehouseHandler) CreatePart(c *gin.Context) {
	var dto domain.CreateSparePartDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	part, err := h.repo.CreatePart(c.Request.Context(), dto)
	if err != nil {
		respondWarehouseError(c, "WarehouseHandler.CreatePart", err)
		return
	}
	respondCreated(c, part)
}

// Receive — POST /spare-parts/:id/receive — приход от поставщика
func (h *WarehouseHandler) Receive(c *gin.Context) {
	var dto domain.ReceivePartsDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	employeeID, ok := employeeIDFromCtx(c)
	if !ok {
		return
	}
	part, err := h.repo.Receive(c.Request.Context(), c.Param("id"), employeeID, dto)
	if err != nil {
		respondWarehouseError(c, "WarehouseHandler.Receive", err)
		return
	}
	respondOK(c, part)
}

// WriteOff — POST /spare-parts/:id/writeoff — списание (порча/недостача)
func (h *WarehouseHandler) WriteOff(c *gin.Context) {
	var dto domain.WriteOffPartsDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	employeeID, ok := employeeIDFromCtx(c)
	if !ok {
		return
	}
	part, err := h.repo.WriteOff(c.Request.Context(), c.Param("id"), employeeID, dto)
	if err != nil {
		respondWarehouseError(c, "WarehouseHandler.WriteOff", err)
		return
	}
	respondOK(c, part)
}

// IssueToRequest — POST /requests/:id/parts — выдать деталь в ремонт
func (h *WarehouseHandler) IssueToRequest(c *gin.Context) {
	var dto domain.IssuePartToRequestDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	employeeID, ok := employeeIDFromCtx(c)
	if !ok {
		return
	}
	entry, err := h.repo.IssueToRequest(c.Request.Context(), c.Param("id"), employeeID, dto)
	if err != nil {
		respondWarehouseError(c, "WarehouseHandler.IssueToRequest", err)
		return
	}
	respondCreated(c, entry)
}

// GetRequestParts — GET /requests/:id/parts
func (h *WarehouseHandler) GetRequestParts(c *gin.Context) {
	entries, err := h.repo.GetRequestParts(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondInternal(c, "WarehouseHandler.GetRequestParts", err)
		return
	}
	respondOK(c, entries)
}

// respondWarehouseError мапит доменные sentinel-ошибки склада в HTTP-статусы.
// Общая точка для всех методов WarehouseHandler, чтобы не повторять один
// и тот же switch в каждом обработчике.
func respondWarehouseError(c *gin.Context, where string, err error) {
	switch {
	case errors.Is(err, repository.ErrPartNotFound):
		respondError(c, http.StatusNotFound, "запчасть не найдена")
	case errors.Is(err, repository.ErrRequestNotFound):
		respondError(c, http.StatusNotFound, "заявка не найдена")
	case errors.Is(err, repository.ErrCategoryNotFound):
		respondError(c, http.StatusBadRequest, "передан несуществующий category_id")
	case errors.Is(err, repository.ErrInsufficientStock):
		respondError(c, http.StatusConflict, "недостаточно деталей на складе")
	case errors.Is(err, repository.ErrAlreadyIssued):
		respondError(c, http.StatusConflict, "деталь уже списана на эту заявку")
	default:
		respondInternal(c, where, err)
	}
}
