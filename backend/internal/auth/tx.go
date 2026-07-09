package auth

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vladislavgnilitskii/asu-soit/internal/dbtx"
)

// dbRoleByCode переводит роль сотрудника из JWT в имя роли БД. Карта —
// белый список: имя роли для SET ROLE берётся только отсюда, а не из запроса.
var dbRoleByCode = map[string]string{
	"admin":       "role_admin",
	"manager":     "role_manager",
	"engineer":    "role_engineer",
	"storekeeper": "role_storekeeper",
	"accountant":  "role_accountant",
}

// TxPerRequest оборачивает каждый аутентифицированный запрос в одну транзакцию,
// перевоплощает соединение в роль сотрудника (SET ROLE) и сообщает БД его
// личность (app.current_employee_id). Эти два факта читают:
//   - GRANT ролей (модуль 1) и политики RLS (модуль 2) — что и какие строки
//     доступны сотруднику;
//   - триггеры аудита (модуль 4) — кто автор изменения.
//
// Почему транзакция, а не просто SET на соединении: пул раздаёт соединения по
// кругу, и SET «протёк» бы на следующий запрос другого сотрудника. set_config
// с is_local=true (и роль, и личность) действует только в пределах текущей
// транзакции и очищается сам при commit/rollback — утечки между запросами нет.
// Оба значения передаются параметром ($1), а не склейкой строки — без инъекций.
//
// Ставится ПОСЛЕ RequireAuth (нужны employee_id и role_code из токена).
func TxPerRequest(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// роль сотрудника → роль БД (по белому списку)
		dbRole, ok := dbRoleByCode[c.GetString("role_code")]
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "неизвестная роль"})
			c.Abort()
			return
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка"})
			c.Abort()
			return
		}

		// перевоплощаемся в роль сотрудника и сообщаем его личность.
		// set_config('role', ...) эквивалентен SET ROLE, но параметризуем.
		empID := c.GetString("employee_id")
		if _, err := tx.Exec(ctx, `
			SELECT set_config('role', $1, true),
			       set_config('app.current_employee_id', $2, true)
		`, dbRole, empID); err != nil {
			_ = tx.Rollback(ctx)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка"})
			c.Abort()
			return
		}

		// репозитории возьмут эту транзакцию из контекста через dbtx.From
		c.Request = c.Request.WithContext(dbtx.With(ctx, tx))

		committed := false
		defer func() {
			if !committed {
				_ = tx.Rollback(ctx) // no-op, если уже закрыта
			}
		}()

		c.Next()

		// коммитим только при успешном ответе; на ошибку — откат (defer)
		if len(c.Errors) == 0 && c.Writer.Status() < http.StatusBadRequest {
			if err := tx.Commit(ctx); err != nil {
				log.Printf("commit транзакции запроса: %v", err)
				return
			}
			committed = true
		}
	}
}
