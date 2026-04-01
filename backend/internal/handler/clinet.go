package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, clients)
}

// GetByID — обработчик GET /api/v1/clients/:id
func (h *ClientHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	client, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "клиент не найден"})
		return
	}
	c.JSON(http.StatusOK, client)
}

// Create — обработчик POST /api/v1/clients
func (h *ClientHandler) Create(c *gin.Context) {
	var dto domain.CreateClientRequest
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client, err := h.repo.Create(c.Request.Context(), dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, client)
}