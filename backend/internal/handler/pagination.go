package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/vladislavgnilitskii/asu-soit/internal/domain"
)

const (
	defaultPageLimit = 20
	maxPageLimit     = 100
)

// parsePageParams читает ?limit=&offset= из запроса с безопасными дефолтами.
// limit ограничен сверху maxPageLimit, чтобы клиент не мог запросить всё разом.
func parsePageParams(c *gin.Context) domain.PageParams {
	limit := defaultPageLimit
	if n, err := strconv.Atoi(c.Query("limit")); err == nil && n > 0 {
		limit = n
	}
	if limit > maxPageLimit {
		limit = maxPageLimit
	}

	offset := 0
	if n, err := strconv.Atoi(c.Query("offset")); err == nil && n > 0 {
		offset = n
	}

	return domain.PageParams{Limit: limit, Offset: offset}
}
