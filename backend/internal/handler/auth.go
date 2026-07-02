package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vladislavgnilitskii/asu-soit/internal/auth"
	"github.com/vladislavgnilitskii/asu-soit/internal/domain"
	"github.com/vladislavgnilitskii/asu-soit/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	repo   *repository.EmployeeRepository
	secret string
}

func NewAuthHandler(repo *repository.EmployeeRepository, secret string) *AuthHandler {
	return &AuthHandler{repo: repo, secret: secret}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	// ищем сотрудника по логину
	employee, roleCode, err := h.repo.GetByLogin(c.Request.Context(), req.Login)
	if err != nil {
		// не говорим что именно не так — логин или пароль
		// это защита от перебора
		respondError(c, http.StatusUnauthorized, "неверный логин или пароль")
		return
	}

	// проверяем что сотрудник активен
	if !employee.IsActive {
		respondError(c, http.StatusUnauthorized, "аккаунт отключён")
		return
	}

	// сравниваем пароль с хешем в БД
	// bcrypt.CompareHashAndPassword сам всё проверяет
	if err := bcrypt.CompareHashAndPassword(
		[]byte(employee.PasswordHash),
		[]byte(req.Password),
	); err != nil {
		respondError(c, http.StatusUnauthorized, "неверный логин или пароль")
		return
	}

	// генерируем токен
	token, err := auth.GenerateToken(*employee, roleCode, h.secret)
	if err != nil {
		respondInternal(c, "AuthHandler.Login generate token", err)
		return
	}

	// убираем хеш пароля из ответа — он там не нужен
	employee.PasswordHash = ""

	respondOK(c, domain.LoginResponse{
		Token:    token,
		Employee: *employee,
	})
}
