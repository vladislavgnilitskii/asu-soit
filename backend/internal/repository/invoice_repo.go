package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vladislavgnilitskii/asu-soit/internal/domain"
)

type InvoiceRepository struct {
	db *pgxpool.Pool
}

func NewInvoiceRepository(db *pgxpool.Pool) *InvoiceRepository {
	return &InvoiceRepository{db: db}
}

// GetByID — счёт по id
func (r *InvoiceRepository) GetByID(ctx context.Context, id string) (*domain.Invoice, error) {
	var inv domain.Invoice
	err := r.db.QueryRow(ctx, `
            SELECT id, request_id, total_amount, status, issued_at
            FROM invoices
            WHERE id = $1
      `, id).Scan(&inv.ID, &inv.RequestID, &inv.TotalAmount, &inv.Status, &inv.IssuedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrInvoiceNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("GetByID invoice: %w", err)
	}
	return &inv, nil
}

// GetByRequestID — счёт конкретной заявки (у заявки не более одного счёта:
// в БД стоит UNIQUE(request_id)). Зеркало GetByID, только фильтр по request_id.
func (r *InvoiceRepository) GetByRequestID(ctx context.Context, requestID string) (*domain.Invoice, error) {
	var inv domain.Invoice
	err := r.db.QueryRow(ctx, `
		SELECT id, request_id, total_amount, status, issued_at
		FROM invoices
		WHERE request_id = $1
	`, requestID).Scan(&inv.ID, &inv.RequestID, &inv.TotalAmount, &inv.Status, &inv.IssuedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrInvoiceNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("GetByRequestID invoice: %w", err)
	}
	return &inv, nil
}

// CreateForRequest — выставить счёт по заявке.
// Сумма считается на сервере: стоимость работы (final_cost заявки) плюс
// стоимость всех выданных на заявку деталей (сумма по request_parts).
func (r *InvoiceRepository) CreateForRequest(ctx context.Context, requestID string) (*domain.Invoice, error) {
	// 1. Заявка должна существовать и быть закрытой. Читаем closed_at и
	//    final_cost одним запросом. Оба поля nullable в БД → указатели:
	//    nil-указатель здесь означает SQL NULL.
	var closedAt *time.Time
	var finalCost *float64
	err := r.db.QueryRow(ctx, `
		SELECT closed_at, final_cost FROM repair_requests WHERE id = $1
	`, requestID).Scan(&closedAt, &finalCost)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrRequestNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("CreateForRequest load request: %w", err)
	}
	if closedAt == nil {
		return nil, ErrRequestNotClosed
	}

	// 2. Считаем сумму деталей. SUM по пустому набору строк вернул бы NULL,
	//    поэтому оборачиваем в COALESCE(..., 0) — «если NULL, то 0».
	var partsTotal float64
	err = r.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(quantity * unit_price), 0)
		FROM request_parts
		WHERE request_id = $1
	`, requestID).Scan(&partsTotal)
	if err != nil {
		return nil, fmt.Errorf("CreateForRequest sum parts: %w", err)
	}

	// 3. Итог = работа + детали. final_cost может отсутствовать (nil) → 0.
	total := partsTotal
	if finalCost != nil {
		total += *finalCost
	}

	// 4. Вставляем счёт. status по умолчанию 'pending' (задан в схеме).
	//    Транзакция здесь не нужна: пишем ровно одну строку, а от повторного
	//    счёта по той же заявке защищает UNIQUE(request_id) на уровне БД —
	//    ловим её нарушение (код 23505), как в Этапе 3 с request_parts.
	var inv domain.Invoice
	err = r.db.QueryRow(ctx, `
		INSERT INTO invoices (request_id, total_amount)
		VALUES ($1, $2)
		RETURNING id, request_id, total_amount, status, issued_at
	`, requestID, total,
	).Scan(&inv.ID, &inv.RequestID, &inv.TotalAmount, &inv.Status, &inv.IssuedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrInvoiceExists
		}
		return nil, fmt.Errorf("CreateForRequest insert: %w", err)
	}
	return &inv, nil
}

// UpdateStatus — сменить статус счёта (pending → paid/cancelled).
// Менять можно только счёт в статусе pending: условие status = 'pending'
// прямо в UPDATE. Если строка не обновилась — разбираемся почему:
// счёта нет вовсе или он уже не pending.
func (r *InvoiceRepository) UpdateStatus(ctx context.Context, id, newStatus string) (*domain.Invoice, error) {
	var inv domain.Invoice
	err := r.db.QueryRow(ctx, `
		UPDATE invoices SET status = $1
		WHERE id = $2 AND status = 'pending'
		RETURNING id, request_id, total_amount, status, issued_at
	`, newStatus, id,
	).Scan(&inv.ID, &inv.RequestID, &inv.TotalAmount, &inv.Status, &inv.IssuedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		// строка не обновилась: либо счёта нет, либо он уже не pending.
		// Уточняем обычным чтением, чтобы вернуть точную ошибку.
		if _, getErr := r.GetByID(ctx, id); errors.Is(getErr, ErrInvoiceNotFound) {
			return nil, ErrInvoiceNotFound
		}
		return nil, ErrInvoiceNotPending
	}
	if err != nil {
		return nil, fmt.Errorf("UpdateStatus invoice: %w", err)
	}
	return &inv, nil
}
