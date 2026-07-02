package auth

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vladislavgnilitskii/asu-soit/internal/domain"
)

const testSecret = "test-secret"

func testEmployee() domain.Employee {
	return domain.Employee{
		ID:    "11111111-1111-1111-1111-111111111111",
		Login: "ivanov",
	}
}

func TestGenerateAndParseToken_RoundTrip(t *testing.T) {
	token, err := GenerateToken(testEmployee(), "engineer", testSecret)
	if err != nil {
		t.Fatalf("GenerateToken вернул ошибку: %v", err)
	}

	claims, err := ParseToken(token, testSecret)
	if err != nil {
		t.Fatalf("ParseToken вернул ошибку: %v", err)
	}

	if claims.EmployeeID != testEmployee().ID {
		t.Errorf("EmployeeID: получили %q, ожидали %q", claims.EmployeeID, testEmployee().ID)
	}
	if claims.Login != "ivanov" {
		t.Errorf("Login: получили %q, ожидали %q", claims.Login, "ivanov")
	}
	if claims.RoleCode != "engineer" {
		t.Errorf("RoleCode: получили %q, ожидали %q", claims.RoleCode, "engineer")
	}
}

func TestParseToken_WrongSecret(t *testing.T) {
	token, err := GenerateToken(testEmployee(), "engineer", testSecret)
	if err != nil {
		t.Fatalf("GenerateToken вернул ошибку: %v", err)
	}

	if _, err := ParseToken(token, "other-secret"); err == nil {
		t.Error("ожидали ошибку при неверном секрете, получили nil")
	}
}

func TestParseToken_Garbage(t *testing.T) {
	if _, err := ParseToken("not.a.valid.token", testSecret); err == nil {
		t.Error("ожидали ошибку на мусорный токен, получили nil")
	}
}

// Защита от алгоритмической подмены: токен без подписи (alg=none)
// должен быть отвергнут, даже если структурно валиден.
func TestParseToken_RejectsNoneAlgorithm(t *testing.T) {
	claims := jwt.MapClaims{
		"employee_id": testEmployee().ID,
		"login":       "ivanov",
		"role_code":   "admin",
	}
	unsigned := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenStr, err := unsigned.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("не удалось создать none-токен: %v", err)
	}

	if _, err := ParseToken(tokenStr, testSecret); err == nil {
		t.Error("ожидали отклонение none-токена, получили nil")
	}
}
