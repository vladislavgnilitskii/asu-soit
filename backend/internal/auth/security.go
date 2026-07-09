package auth

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// SecurityHeaders навешивает заголовки безопасности на все ответы.
// API отдаёт только JSON, поэтому политика строгая: активное содержимое
// и встраивание в чужие фреймы запрещены.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Cross-Origin-Resource-Policy", "same-origin")
		h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		c.Next()
	}
}

// LoginRateLimiter — простой ограничитель попыток входа по IP
// (фиксированное окно). Защищает от перебора паролей. Состояние — в памяти
// процесса; для одного инстанса достаточно, для кластера нужен общий стор.
type LoginRateLimiter struct {
	mu       sync.Mutex
	hits     map[string]*rateWindow
	limit    int
	interval time.Duration
}

type rateWindow struct {
	count int
	reset time.Time
}

// NewLoginRateLimiter создаёт ограничитель: не более limit попыток
// за interval с одного IP.
func NewLoginRateLimiter(limit int, interval time.Duration) *LoginRateLimiter {
	return &LoginRateLimiter{
		hits:     make(map[string]*rateWindow),
		limit:    limit,
		interval: interval,
	}
}

// Middleware возвращает gin-обработчик, отклоняющий запрос с 429,
// если лимит для IP исчерпан.
func (l *LoginRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !l.allow(c.ClientIP()) {
			c.Header("Retry-After", strconv.Itoa(int(l.interval.Seconds())))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "слишком много попыток входа, повторите позже",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

func (l *LoginRateLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	w, ok := l.hits[ip]
	if !ok || now.After(w.reset) {
		l.hits[ip] = &rateWindow{count: 1, reset: now.Add(l.interval)}
		l.sweep(now) // чистим истёкшие окна, чтобы карта не росла бесконечно
		return true
	}
	if w.count >= l.limit {
		return false
	}
	w.count++
	return true
}

func (l *LoginRateLimiter) sweep(now time.Time) {
	for ip, w := range l.hits {
		if now.After(w.reset) {
			delete(l.hits, ip)
		}
	}
}
