package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vladislavgnilitskii/asu-soit/internal/domain"
	"github.com/vladislavgnilitskii/asu-soit/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type EmployeeHandler struct {
	repo *repository.EmployeeRepository
}

func NewEmployeeHandler(repo *repository.EmployeeRepository) *EmployeeHandler {
	return &EmployeeHandler{repo: repo}
}

// ListRoles — GET /roles: справочник ролей для формы создания сотрудника
func (h *EmployeeHandler) ListRoles(c *gin.Context) {
	roles, err := h.repo.ListRoles(c.Request.Context())
	if err != nil {
		respondInternal(c, "EmployeeHandler.ListRoles", err)
		return
	}
	respondOK(c, roles)
}

// Create — POST /employees — завести сотрудника
func (h *EmployeeHandler) Create(c *gin.Context) {
	var dto domain.CreateEmployeeDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	// хешируем пароль перед передачей в репозиторий — в БД уходит только хеш
	hash, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		respondInternal(c, "EmployeeHandler.Create hash", err)
		return
	}

	employee, err := h.repo.Create(c.Request.Context(), dto, string(hash))
	if err != nil {
		respondEmployeeError(c, "EmployeeHandler.Create", err)
		return
	}
	respondCreated(c, employee)
}

// GetAll — GET /employees
func (h *EmployeeHandler) GetAll(c *gin.Context) {
	employees, err := h.repo.GetAll(c.Request.Context())
	if err != nil {
		respondInternal(c, "EmployeeHandler.GetAll", err)
		return
	}
	respondOK(c, employees)
}

// GetByID — GET /employees/:id
func (h *EmployeeHandler) GetByID(c *gin.Context) {
	employee, err := h.repo.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondEmployeeError(c, "EmployeeHandler.GetByID", err)
		return
	}
	respondOK(c, employee)
}

// Update — PATCH /employees/:id — профиль/роль/активность (в т.ч. деактивация)
func (h *EmployeeHandler) Update(c *gin.Context) {
	var dto domain.UpdateEmployeeDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	employee, err := h.repo.Update(c.Request.Context(), c.Param("id"), dto)
	if err != nil {
		respondEmployeeError(c, "EmployeeHandler.Update", err)
		return
	}
	respondOK(c, employee)
}

// respondEmployeeError — маппинг доменных ошибок сотрудников в HTTP-коды
func respondEmployeeError(c *gin.Context, where string, err error) {
	switch {
	case errors.Is(err, repository.ErrEmployeeNotFound):
		respondError(c, http.StatusNotFound, "сотрудник не найден")
	case errors.Is(err, repository.ErrRoleNotFound):
		respondError(c, http.StatusBadRequest, "передан несуществующий role_id")
	case errors.Is(err, repository.ErrLoginTaken):
		respondError(c, http.StatusConflict, "логин уже занят")
	default:
		respondInternal(c, where, err)
	}
}
