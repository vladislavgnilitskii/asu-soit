// Package dbtx позволяет репозиториям работать как напрямую с пулом, так и
// внутри транзакции запроса (request-scoped transaction), не зная, что именно
// перед ними. Транзакцию в контекст кладёт middleware auth.TxPerRequest —
// вместе с сессионной переменной app.current_employee_id (личность из JWT),
// которую читают политики RLS (модуль 2) и триггеры аудита (модуль 4).
package dbtx

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// DBTX — то, на чём можно выполнять запросы. Интерфейс удовлетворяют и
// *pgxpool.Pool, и pgx.Tx, поэтому репозиторий работает с любым из них.
type DBTX interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Begin(ctx context.Context) (pgx.Tx, error)
}

type ctxKey struct{}

// With кладёт executor (обычно транзакцию запроса) в контекст.
func With(ctx context.Context, q DBTX) context.Context {
	return context.WithValue(ctx, ctxKey{}, q)
}

// From возвращает executor из контекста (транзакцию запроса), а если его нет —
// fallback (пул). Так репозиторий автоматически попадает в транзакцию с
// личностью сотрудника, если она открыта, и работает как раньше, если нет.
func From(ctx context.Context, fallback DBTX) DBTX {
	if q, ok := ctx.Value(ctxKey{}).(DBTX); ok && q != nil {
		return q
	}
	return fallback
}
