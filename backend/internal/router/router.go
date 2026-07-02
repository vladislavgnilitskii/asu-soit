package router

import (
	"github.com/gin-gonic/gin"
	"github.com/vladislavgnilitskii/asu-soit/internal/auth"
	"github.com/vladislavgnilitskii/asu-soit/internal/handler"
)

func Setup(
	clientHandler *handler.ClientHandler,
	requestHandler *handler.RequestHandler,
	authHandler *handler.AuthHandler,
	jwtSecret string,
) *gin.Engine {
	r := gin.Default()

	api := r.Group("/api/v1")
	{
		// публичный маршрут — авторизация
		api.POST("/auth/login", authHandler.Login)

		// защищённые маршруты — нужен валидный токен
		protected := api.Group("")
		protected.Use(auth.RequireAuth(jwtSecret))
		{
			clients := protected.Group("/clients")
			{
				clients.GET("", clientHandler.GetAll)
				clients.GET("/:id", clientHandler.GetByID)
				// создавать клиентов могут admin и manager (приёмка)
				clients.POST("", auth.RequireRole("admin", "manager"), clientHandler.Create)
			}

			requests := protected.Group("/requests")
			{
				requests.GET("", requestHandler.GetAll)
				requests.GET("/:id", requestHandler.GetByID)
				// создавать заявки — admin и manager (приёмка)
				requests.POST("", auth.RequireRole("admin", "manager"), requestHandler.Create)
				// менять статус — те, кто ведёт ремонт
				requests.PATCH("/:id/status", auth.RequireRole("admin", "manager", "engineer"), requestHandler.UpdateStatus)
			}
		}
	}

	return r
}
