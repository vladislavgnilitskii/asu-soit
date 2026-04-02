package repository

import (
	"context"
	"fmt"

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
		SELECT id, client_id, device_id, assigned_to, status_id,
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
			&req.ID, &req.ClientID, &req.DeviceID, &req.AssignedTo,
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
		SELECT id, client_id, device_id, assigned_to, status_id,
		       problem_description, diagnostic_result,
		       estimated_cost, final_cost,
		       planned_deadline, created_at, closed_at
		FROM repair_requests
		WHERE id = $1
	`, id).Scan(
		&req.ID, &req.ClientID, &req.DeviceID, &req.AssignedTo,
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
		    (client_id, device_id, status_id, problem_description, planned_deadline)
		VALUES (
		    $1, $2,
		    (SELECT id FROM request_statuses WHERE code = 'new'),
		    $3, $4
		)
		RETURNING id, client_id, device_id, assigned_to, status_id,
		          problem_description, diagnostic_result,
		          estimated_cost, final_cost,
		          planned_deadline, created_at, closed_at
	`, dto.ClientID, dto.DeviceID, dto.ProblemDescription, dto.PlannedDeadline,
	).Scan(
		&req.ID, &req.ClientID, &req.DeviceID, &req.AssignedTo,
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

	// обновляем статус в самой заявке
	_, err = tx.Exec(ctx, `
		UPDATE repair_requests SET status_id = $1, updated_at = now()
		WHERE id = $2
	`, dto.StatusID, id)
	if err != nil {
		return fmt.Errorf("UpdateStatus update: %w", err)
	}

	// пишем в историю — кто, когда, на какой статус
	_, err = tx.Exec(ctx, `
		INSERT INTO request_status_history (request_id, status_id, changed_by, comment)
		VALUES ($1, $2, $3, $4)
	`, id, dto.StatusID, employeeID, dto.Comment)
	if err != nil {
		return fmt.Errorf("UpdateStatus history: %w", err)
	}

	return tx.Commit(ctx)
}
