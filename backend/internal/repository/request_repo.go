package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vladislavgnilitskii/asu-soit/internal/domain"
)

type RequestRepository struct {
	db *pgxpool.Pool
}

func NewRequestRepository(db *pgxpool.Pool) *RequestRepository {
	return &RequestRepository{db: db}
}

// GetAll — все заявки
func (r *RequestRepository) GetAll(ctx context.Context) ([]domain.RepairRequest, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, device_id, assigned_to, status_id,
		       problem_description, diagnostic_result,
		       estimated_cost, final_cost,
		       planned_deadline, created_at, closed_at
		FROM repair_requests
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("GetAll requests: %w", err)
	}
	defer rows.Close()

	var requests []domain.RepairRequest
	for rows.Next() {
		var req domain.RepairRequest
		err := rows.Scan(
			&req.ID, &req.DeviceID, &req.AssignedTo,
			&req.StatusID, &req.ProblemDescription, &req.DiagnosticResult,
			&req.EstimatedCost, &req.FinalCost,
			&req.PlannedDeadline, &req.CreatedAt, &req.ClosedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("GetAll scan: %w", err)
		}
		requests = append(requests, req)
	}
	return requests, nil
}

// GetByID — одна заявка по id
func (r *RequestRepository) GetByID(ctx context.Context, id string) (*domain.RepairRequest, error) {
	var req domain.RepairRequest
	err := r.db.QueryRow(ctx, `
		SELECT id, device_id, assigned_to, status_id,
		       problem_description, diagnostic_result,
		       estimated_cost, final_cost,
		       planned_deadline, created_at, closed_at
		FROM repair_requests
		WHERE id = $1
	`, id).Scan(
		&req.ID, &req.DeviceID, &req.AssignedTo,
		&req.StatusID, &req.ProblemDescription, &req.DiagnosticResult,
		&req.EstimatedCost, &req.FinalCost,
		&req.PlannedDeadline, &req.CreatedAt, &req.ClosedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("GetByID request: %w", err)
	}
	return &req, nil
}

// Create — создать заявку
// статус по умолчанию — "new", берём из справочника
func (r *RequestRepository) Create(ctx context.Context, dto domain.CreateRepairRequestDTO) (*domain.RepairRequest, error) {
	var req domain.RepairRequest
	err := r.db.QueryRow(ctx, `
		INSERT INTO repair_requests
		    (device_id, status_id, problem_description, planned_deadline)
		VALUES (
		    $1,
		    (SELECT id FROM request_statuses WHERE code = 'new'),
		    $2, $3
		)
		RETURNING id, device_id, assigned_to, status_id,
		          problem_description, diagnostic_result,
		          estimated_cost, final_cost,
		          planned_deadline, created_at, closed_at
	`, dto.DeviceID, dto.ProblemDescription, dto.PlannedDeadline,
	).Scan(
		&req.ID, &req.DeviceID, &req.AssignedTo,
		&req.StatusID, &req.ProblemDescription, &req.DiagnosticResult,
		&req.EstimatedCost, &req.FinalCost,
		&req.PlannedDeadline, &req.CreatedAt, &req.ClosedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("Create request: %w", err)
	}
	return &req, nil
}

// UpdateStatus — сменить статус заявки и записать в историю
func (r *RequestRepository) UpdateStatus(ctx context.Context, id string, dto domain.UpdateRequestStatusDTO, employeeID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("UpdateStatus begin: %w", err)
	}
	defer tx.Rollback(ctx)

	// валидируем status_id заранее — чтобы вернуть 400, а не 500 по FK
	var statusExists bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM request_statuses WHERE id = $1)
	`, dto.StatusID).Scan(&statusExists); err != nil {
		return fmt.Errorf("UpdateStatus check status: %w", err)
	}
	if !statusExists {
		return ErrStatusNotFound
	}

	// обновляем статус в самой заявке; RowsAffected=0 → заявки нет
	tag, err := tx.Exec(ctx, `
		UPDATE repair_requests SET status_id = $1, updated_at = now()
		WHERE id = $2
	`, dto.StatusID, id)
	if err != nil {
		return fmt.Errorf("UpdateStatus update: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrRequestNotFound
	}

	// пишем в историю — кто, когда, на какой статус
	_, err = tx.Exec(ctx, `
		INSERT INTO request_status_history (request_id, status_id, changed_by, comment)
		VALUES ($1, $2, $3, $4)
	`, id, dto.StatusID, employeeID, nullifyEmpty(dto.Comment))
	if err != nil {
		return fmt.Errorf("UpdateStatus history: %w", err)
	}

	return tx.Commit(ctx)
}

// Assign — назначить исполнителя на заявку
func (r *RequestRepository) Assign(ctx context.Context, id, assignedTo string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE repair_requests SET assigned_to = $1, updated_at = now()
		WHERE id = $2
	`, assignedTo, id)
	if err != nil {
		return fmt.Errorf("Assign: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrRequestNotFound
	}
	return nil
}

// UpdateDetails — частичное обновление диагностики и стоимости.
// COALESCE оставляет прежнее значение, если новое не передано (nil).
func (r *RequestRepository) UpdateDetails(ctx context.Context, id string, dto domain.UpdateRequestDTO) (*domain.RepairRequest, error) {
	var req domain.RepairRequest
	err := r.db.QueryRow(ctx, `
		UPDATE repair_requests SET
		    diagnostic_result = COALESCE($1, diagnostic_result),
		    estimated_cost    = COALESCE($2, estimated_cost),
		    final_cost        = COALESCE($3, final_cost),
		    updated_at        = now()
		WHERE id = $4
		RETURNING id, device_id, assigned_to, status_id,
		          problem_description, diagnostic_result,
		          estimated_cost, final_cost,
		          planned_deadline, created_at, closed_at
	`, dto.DiagnosticResult, dto.EstimatedCost, dto.FinalCost, id,
	).Scan(
		&req.ID, &req.DeviceID, &req.AssignedTo,
		&req.StatusID, &req.ProblemDescription, &req.DiagnosticResult,
		&req.EstimatedCost, &req.FinalCost,
		&req.PlannedDeadline, &req.CreatedAt, &req.ClosedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrRequestNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("UpdateDetails: %w", err)
	}
	return &req, nil
}

// Close — закрыть заявку: статус 'closed', проставить closed_at, записать историю.
func (r *RequestRepository) Close(ctx context.Context, id, employeeID, comment string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("Close begin: %w", err)
	}
	defer tx.Rollback(ctx)

	// id статуса «закрыта» из справочника
	var closedStatusID string
	if err := tx.QueryRow(ctx, `
		SELECT id FROM request_statuses WHERE code = 'closed'
	`).Scan(&closedStatusID); err != nil {
		return fmt.Errorf("Close lookup status: %w", err)
	}

	tag, err := tx.Exec(ctx, `
		UPDATE repair_requests
		SET status_id = $1, closed_at = now(), updated_at = now()
		WHERE id = $2
	`, closedStatusID, id)
	if err != nil {
		return fmt.Errorf("Close update: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrRequestNotFound
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO request_status_history (request_id, status_id, changed_by, comment)
		VALUES ($1, $2, $3, $4)
	`, id, closedStatusID, employeeID, nullifyEmpty(comment))
	if err != nil {
		return fmt.Errorf("Close history: %w", err)
	}

	return tx.Commit(ctx)
}

// GetHistory — история смены статусов заявки (по возрастанию времени),
// обогащённая кодом/названием статуса и ФИО сотрудника.
func (r *RequestRepository) GetHistory(ctx context.Context, requestID string) ([]domain.StatusHistoryEntry, error) {
	rows, err := r.db.Query(ctx, `
		SELECT h.id, h.status_id, s.code, s.name,
		       h.changed_by, e.last_name || ' ' || e.first_name,
		       h.changed_at, h.comment
		FROM request_status_history h
		JOIN request_statuses s ON s.id = h.status_id
		JOIN employees e        ON e.id = h.changed_by
		WHERE h.request_id = $1
		ORDER BY h.changed_at
	`, requestID)
	if err != nil {
		return nil, fmt.Errorf("GetHistory: %w", err)
	}
	defer rows.Close()

	var history []domain.StatusHistoryEntry
	for rows.Next() {
		var e domain.StatusHistoryEntry
		err := rows.Scan(
			&e.ID, &e.StatusID, &e.StatusCode, &e.StatusName,
			&e.ChangedBy, &e.ChangedByName, &e.ChangedAt, &e.Comment,
		)
		if err != nil {
			return nil, fmt.Errorf("GetHistory scan: %w", err)
		}
		history = append(history, e)
	}
	return history, nil
}
