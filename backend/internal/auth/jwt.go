package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vladislavgnilitskii/asu-soit/internal/domain"
)

// GenerateToken — создаёт JWT токен для сотрудника
// токен живёт 24 часа
func GenerateToken(employee domain.Employee, roleCode string, secret string) (string, error) {
	claims := jwt.MapClaims{
		"employee_id": employee.ID,
		"login":       employee.Login,
		"role_code":   roleCode,
		// exp — время истечения токена
		"exp": time.Now().Add(24 * time.Hour).Unix(),
		// iat — время создания токена
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken — проверяет токен и возвращает данные из него
func ParseToken(tokenStr string, secret string) (*domain.Claims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		// проверяем что алгоритм тот что ожидаем
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неожиданный алгоритм: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("невалидный токен: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("невалидные claims")
	}

	return &domain.Claims{
		EmployeeID: claims["employee_id"].(string),
		Login:      claims["login"].(string),
		RoleCode:   claims["role_code"].(string),
	}, nil
}
