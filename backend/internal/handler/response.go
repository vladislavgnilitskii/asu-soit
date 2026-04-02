package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// respondError — единый формат ошибки во всех хендлерах
// вместо того чтобы везде писать gin.H{"error": "..."} вручную
func respondError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}

// respondOK — успешный ответ с данными
func respondOK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, data)
}

// respondCreated — ответ при создании ресурса
func respondCreated(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, data)
}
