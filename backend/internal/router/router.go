package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vladislavgnilitskii/asu-soit/internal/auth"
	"github.com/vladislavgnilitskii/asu-soit/internal/handler"
)

func Setup(
	clientHandler *handler.ClientHandler,
	requestHandler *handler.RequestHandler,
	deviceHandler *handler.DeviceHandler,
	warehouseHandler *handler.WarehouseHandler,
	invoiceHandler *handler.InvoiceHandler,
	employeeHandler *handler.EmployeeHandler,
	authHandler *handler.AuthHandler,
	jwtSecret string,
	pool *pgxpool.Pool,
) *gin.Engine {
	r := gin.Default()

	// заголовки безопасности на все ответы
	r.Use(auth.SecurityHeaders())
	// ограничение размера тела запроса — 1 МиБ (защита от крупных payload)
	r.Use(auth.BodyLimit(1 << 20))

	// защита входа от перебора: не более 10 попыток в минуту с одного IP
	loginLimiter := auth.NewLoginRateLimiter(10, time.Minute)

	api := r.Group("/api/v1")
	{
		// публичный маршрут — авторизация (под rate-limit)
		api.POST("/auth/login", loginLimiter.Middleware(), authHandler.Login)

		// защищённые маршруты — нужен валидный токен
		protected := api.Group("")
		protected.Use(auth.RequireAuth(jwtSecret))
		// каждый аутентифицированный запрос идёт в транзакции с личностью
		// сотрудника (app.current_employee_id) — для RLS и аудита
		protected.Use(auth.TxPerRequest(pool))
		{
			clients := protected.Group("/clients")
			{
				// клиенты содержат PII (individuals/organizations); доступ —
				// приёмка (admin, manager), у кого есть права на эти таблицы
				clients.GET("", auth.RequireRole("admin", "manager"), clientHandler.GetAll)
				clients.GET("/:id", auth.RequireRole("admin", "manager"), clientHandler.GetByID)
				clients.POST("", auth.RequireRole("admin", "manager"), clientHandler.Create)
			}

			requests := protected.Group("/requests")
			{
				// список/карточку заявок видят все роли; RLS сам покажет
				// инженеру только его заявки (модуль 2)
				requests.GET("", requestHandler.GetAll)
				requests.GET("/:id", requestHandler.GetByID)
				// история статусов — у кого есть SELECT на request_status_history
				requests.GET("/:id/history", auth.RequireRole("admin", "manager", "engineer"), requestHandler.History)
				// создавать заявки — admin и manager (приёмка)
				requests.POST("", auth.RequireRole("admin", "manager"), requestHandler.Create)
				// менять статус и вести диагностику — те, кто ведёт ремонт
				requests.PATCH("/:id/status", auth.RequireRole("admin", "manager", "engineer"), requestHandler.UpdateStatus)
				requests.PATCH("/:id", auth.RequireRole("admin", "manager", "engineer"), requestHandler.UpdateDetails)
				// назначение исполнителя и закрытие — приёмка/менеджмент
				requests.PATCH("/:id/assign", auth.RequireRole("admin", "manager"), requestHandler.Assign)
				requests.PATCH("/:id/close", auth.RequireRole("admin", "manager"), requestHandler.Close)

				// детали, списанные на заявку
				requests.GET("/:id/parts", warehouseHandler.GetRequestParts)
				// выдать деталь в ремонт — кладовщик физически выдаёт,
				// инженер может списать то, что сам поставил в ремонт
				requests.POST("/:id/parts", auth.RequireRole("admin", "storekeeper", "engineer"), warehouseHandler.IssueToRequest)

				// счёт заявки: смотреть — у кого есть SELECT на invoices; выставлять — бухгалтерия
				requests.GET("/:id/invoice", auth.RequireRole("admin", "manager", "accountant"), invoiceHandler.GetByRequestID)
				requests.POST("/:id/invoice", auth.RequireRole("admin", "accountant"), invoiceHandler.CreateForRequest)
			}

			devices := protected.Group("/devices")
			{
				// устройства видят те, у кого есть SELECT на devices
				devices.GET("", auth.RequireRole("admin", "manager", "engineer"), deviceHandler.GetAll)
				devices.GET("/:id", auth.RequireRole("admin", "manager", "engineer"), deviceHandler.GetByID)
				// регистрировать устройства — admin и manager (приёмка)
				devices.POST("", auth.RequireRole("admin", "manager"), deviceHandler.Create)
			}

			// справочник типов устройств — нужен для формы создания устройства
			protected.GET("/device-types", deviceHandler.ListTypes)

			// справочник статусов заявки — нужен для смены статуса (любой авторизованный)
			protected.GET("/request-statuses", requestHandler.ListStatuses)

			// справочник ролей — для формы создания сотрудника (только админ)
			protected.GET("/roles", auth.RequireRole("admin"), employeeHandler.ListRoles)

			// активные инженеры — для назначения на заявку (через витрину v_employees)
			protected.GET("/engineers", auth.RequireRole("admin", "manager"), employeeHandler.ListEngineers)

			spareParts := protected.Group("/spare-parts")
			{
				// склад видят те, у кого есть SELECT на spare_parts
				spareParts.GET("", auth.RequireRole("admin", "storekeeper", "engineer"), warehouseHandler.GetAllParts)
				spareParts.GET("/:id", auth.RequireRole("admin", "storekeeper", "engineer"), warehouseHandler.GetPartByID)
				// заводить новые позиции каталога и вести приход/списание — склад
				spareParts.POST("", auth.RequireRole("admin", "storekeeper"), warehouseHandler.CreatePart)
				spareParts.POST("/:id/receive", auth.RequireRole("admin", "storekeeper"), warehouseHandler.Receive)
				spareParts.POST("/:id/writeoff", auth.RequireRole("admin", "storekeeper"), warehouseHandler.WriteOff)
			}

			// справочник категорий запчастей
			protected.GET("/part-categories", warehouseHandler.ListCategories)

			invoices := protected.Group("/invoices")
			{
				// счёт видят те, у кого есть SELECT на invoices
				invoices.GET("/:id", auth.RequireRole("admin", "manager", "accountant"), invoiceHandler.GetByID)
				// менять статус счёта (оплачен/отменён) — бухгалтерия
				invoices.PATCH("/:id/status", auth.RequireRole("admin", "accountant"), invoiceHandler.UpdateStatus)
			}

			// управление учётными записями сотрудников — только администратор
			employees := protected.Group("/employees")
			employees.Use(auth.RequireRole("admin"))
			{
				employees.GET("", employeeHandler.GetAll)
				employees.GET("/:id", employeeHandler.GetByID)
				employees.POST("", employeeHandler.Create)
				employees.PATCH("/:id", employeeHandler.Update)
			}
		}
	}

	return r
}
