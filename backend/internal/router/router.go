package router

import (
	"github.com/gin-gonic/gin"
	"github.com/vladislavgnilitskii/asu-soit/internal/handler"
)

func Setup(clientHandler *handler.ClientHandler) *gin.Engine {
	r := gin.Default()

	// все маршруты API под префиксом /api/v1
	api := r.Group("/api/v1")
	{
		clients := api.Group("/clients")
		{
			clients.GET("",      clientHandler.GetAll)
			clients.GET("/:id",  clientHandler.GetByID)
			clients.POST("",     clientHandler.Create)
		}
	}

	return r
}