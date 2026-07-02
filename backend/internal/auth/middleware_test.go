package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newRequest прогоняет один запрос через цепочку middleware + финальный
// обработчик и возвращает записанный ответ.
func runChain(header string, handlers ...gin.HandlerFunc) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.GET("/test", handlers...)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	if header != "" {
		req.Header.Set("Authorization", header)
	}
	r.ServeHTTP(w, req)
	return w
}

func TestRequireAuth_NoHeader(t *testing.T) {
	w := runChain("", RequireAuth(testSecret), func(c *gin.Context) { c.Status(http.StatusOK) })
	if w.Code != http.StatusUnauthorized {
		t.Errorf("без заголовка ожидали 401, получили %d", w.Code)
	}
}

func TestRequireAuth_BadFormat(t *testing.T) {
	w := runChain("Token abc", RequireAuth(testSecret), func(c *gin.Context) { c.Status(http.StatusOK) })
	if w.Code != http.StatusUnauthorized {
		t.Errorf("на кривой формат ожидали 401, получили %d", w.Code)
	}
}

func TestRequireAuth_ValidToken(t *testing.T) {
	token, err := GenerateToken(testEmployee(), "engineer", testSecret)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	var gotEmployeeID string
	w := runChain("Bearer "+token, RequireAuth(testSecret), func(c *gin.Context) {
		gotEmployeeID = c.GetString("employee_id")
		c.Status(http.StatusOK)
	})

	if w.Code != http.StatusOK {
		t.Fatalf("с валидным токеном ожидали 200, получили %d", w.Code)
	}
	if gotEmployeeID != testEmployee().ID {
		t.Errorf("employee_id в контексте: получили %q, ожидали %q", gotEmployeeID, testEmployee().ID)
	}
}

func TestRequireRole_Allowed(t *testing.T) {
	setRole := func(c *gin.Context) { c.Set("role_code", "engineer") }
	final := func(c *gin.Context) { c.Status(http.StatusOK) }

	w := runChain("", setRole, RequireRole("admin", "engineer"), final)
	if w.Code != http.StatusOK {
		t.Errorf("роль engineer разрешена, ожидали 200, получили %d", w.Code)
	}
}

func TestRequireRole_Denied(t *testing.T) {
	setRole := func(c *gin.Context) { c.Set("role_code", "storekeeper") }
	final := func(c *gin.Context) { c.Status(http.StatusOK) }

	w := runChain("", setRole, RequireRole("admin", "engineer"), final)
	if w.Code != http.StatusForbidden {
		t.Errorf("роль storekeeper запрещена, ожидали 403, получили %d", w.Code)
	}
}
