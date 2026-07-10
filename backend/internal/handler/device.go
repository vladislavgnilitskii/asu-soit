package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/vladislavgnilitskii/asu-soit/internal/domain"
	"github.com/vladislavgnilitskii/asu-soit/internal/repository"
)

type DeviceHandler struct {
	repo *repository.DeviceRepository
}

func NewDeviceHandler(repo *repository.DeviceRepository) *DeviceHandler {
	return &DeviceHandler{repo: repo}
}

// GetAll — GET /api/v1/devices
func (h *DeviceHandler) GetAll(c *gin.Context) {
	p := parsePageParams(c)
	devices, total, err := h.repo.GetAll(c.Request.Context(), p)
	if err != nil {
		respondInternal(c, "DeviceHandler.GetAll", err)
		return
	}
	respondOK(c, domain.NewPage(devices, total, p))
}

// GetByID — GET /api/v1/devices/:id
func (h *DeviceHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	device, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(c, http.StatusNotFound, "устройство не найдено")
			return
		}
		respondInternal(c, "DeviceHandler.GetByID", err)
		return
	}
	respondOK(c, device)
}

// Create — POST /api/v1/devices
func (h *DeviceHandler) Create(c *gin.Context) {
	var dto domain.CreateDeviceDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	device, err := h.repo.Create(c.Request.Context(), dto)
	if err != nil {
		respondInternal(c, "DeviceHandler.Create", err)
		return
	}
	respondCreated(c, device)
}

// ListTypes — GET /api/v1/device-types
func (h *DeviceHandler) ListTypes(c *gin.Context) {
	types, err := h.repo.ListTypes(c.Request.Context())
	if err != nil {
		respondInternal(c, "DeviceHandler.ListTypes", err)
		return
	}
	respondOK(c, types)
}
