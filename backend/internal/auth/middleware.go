package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// RequireAuth — middleware проверки JWT токена
// если токен валидный — кладёт данные сотрудника в контекст gin
// если невалидный — возвращает 401 и останавливает выполнение
func RequireAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// читаем заголовок Authorization: Bearer <token>
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "токен не передан"})
			c.Abort() // останавливаем выполнение — до хендлера не доходим
			return
		}

		// заголовок должен быть "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "неверный формат токена"})
			c.Abort()
			return
		}

		// проверяем токен
		claims, err := ParseToken(parts[1], secret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// кладём данные сотрудника в контекст — хендлеры смогут их прочитать
		c.Set("employee_id", claims.EmployeeID)
		c.Set("login", claims.Login)
		c.Set("role_code", claims.RoleCode)

		// всё хорошо — продолжаем выполнение
		c.Next()
	}
}

// RequireRole — middleware проверки роли
// используется после RequireAuth
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleCode := c.GetString("role_code")
		for _, role := range roles {
			if role == roleCode {
				c.Next()
				return
			}
		}
		c.JSON(http.StatusForbidden, gin.H{"error": "недостаточно прав"})
		c.Abort()
	}
}
