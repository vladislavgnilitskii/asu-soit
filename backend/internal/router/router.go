package router

import (
	"github.com/gin-gonic/gin"
	"github.com/vladislavgnilitskii/asu-soit/internal/auth"
	"github.com/vladislavgnilitskii/asu-soit/internal/handler"
)

func Setup(
	clientHandler *handler.ClientHandler,
	requestHandler *handler.RequestHandler,
	deviceHandler *handler.DeviceHandler,
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
				requests.GET("/:id/history", requestHandler.History)
				// создавать заявки — admin и manager (приёмка)
				requests.POST("", auth.RequireRole("admin", "manager"), requestHandler.Create)
				// менять статус и вести диагностику — те, кто ведёт ремонт
				requests.PATCH("/:id/status", auth.RequireRole("admin", "manager", "engineer"), requestHandler.UpdateStatus)
				requests.PATCH("/:id", auth.RequireRole("admin", "manager", "engineer"), requestHandler.UpdateDetails)
				// назначение исполнителя и закрытие — приёмка/менеджмент
				requests.PATCH("/:id/assign", auth.RequireRole("admin", "manager"), requestHandler.Assign)
				requests.PATCH("/:id/close", auth.RequireRole("admin", "manager"), requestHandler.Close)
			}

			devices := protected.Group("/devices")
			{
				devices.GET("", deviceHandler.GetAll)
				devices.GET("/:id", deviceHandler.GetByID)
				// регистрировать устройства — admin и manager (приёмка)
				devices.POST("", auth.RequireRole("admin", "manager"), deviceHandler.Create)
			}

			// справочник типов устройств — нужен для формы создания устройства
			protected.GET("/device-types", deviceHandler.ListTypes)
		}
	}

	return r
}
