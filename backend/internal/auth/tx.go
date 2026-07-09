package auth

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vladislavgnilitskii/asu-soit/internal/dbtx"
)

// TxPerRequest оборачивает каждый аутентифицированный запрос в одну транзакцию
// и привязывает к ней личность сотрудника из JWT через сессионную переменную
// app.current_employee_id. Эту переменную читают:
//   - политики RLS (модуль 2) — какие строки видит сотрудник;
//   - триггеры аудита (модуль 4) — кто автор изменения.
//
// Почему транзакция, а не просто SET на соединении: пул раздаёт соединения по
// кругу, и SET «протёк» бы на следующий запрос другого сотрудника. set_config
// с is_local=true действует только в пределах текущей транзакции и очищается
// сам при commit/rollback — утечка личности между запросами невозможна.
// Значение передаётся параметром ($1), а не склейкой строки — без SQL-инъекций.
//
// Ставится ПОСЛЕ RequireAuth (нужен employee_id из токена).
func TxPerRequest(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		tx, err := pool.Begin(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка"})
			c.Abort()
			return
		}

		// привязываем личность к транзакции (пусто не будет — стоим после RequireAuth)
		if empID := c.GetString("employee_id"); empID != "" {
			if _, err := tx.Exec(ctx,
				"SELECT set_config('app.current_employee_id', $1, true)", empID); err != nil {
				_ = tx.Rollback(ctx)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка"})
				c.Abort()
				return
			}
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
