package router

import (
	"github.com/gin-gonic/gin"
	"github.com/vladislavgnilitskii/asu-soit/internal/handler"
)

func Setup(
	clientHandler  *handler.ClientHandler,
	requestHandler *handler.RequestHandler,
) *gin.Engine {
	r := gin.Default()

	api := r.Group("/api/v1")
	{
		clients := api.Group("/clients")
		{
			clients.GET("",     clientHandler.GetAll)
			clients.GET("/:id", clientHandler.GetByID)
			clients.POST("",    clientHandler.Create)
		}

		requests := api.Group("/requests")
		{
			requests.GET("",           requestHandler.GetAll)
			requests.GET("/:id",       requestHandler.GetByID)
			requests.POST("",          requestHandler.Create)
			requests.PATCH("/:id/status", requestHandler.UpdateStatus)
		}
	}

	return r
}