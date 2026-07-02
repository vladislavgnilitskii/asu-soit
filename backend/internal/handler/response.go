package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// respondError — единый формат ошибки во всех хендлерах
// вместо того чтобы везде писать gin.H{"error": "..."} вручную
func respondError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}

// respondInternal — ответ на внутреннюю ошибку сервера.
// Настоящую ошибку пишем в лог (для разработчика), а клиенту отдаём
// обобщённое сообщение — чтобы не раскрывать детали SQL/инфраструктуры.
func respondInternal(c *gin.Context, where string, err error) {
	log.Printf("%s: %v", where, err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
}

// respondOK — успешный ответ с данными
func respondOK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, data)
}

// respondCreated — ответ при создании ресурса
func respondCreated(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, data)
}
