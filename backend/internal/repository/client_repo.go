package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vladislavgnilitskii/asu-soit/internal/domain"
)

// ClientRepository — отвечает только за SQL-запросы к таблицам clients и individuals
// не знает про HTTP, не знает про JSON — только БД
type ClientRepository struct {
	db *pgxpool.Pool
}

// NewClientRepository — конструктор, принимает пул соединений снаружи
func NewClientRepository(db *pgxpool.Pool) *ClientRepository {
	return &ClientRepository{db: db}
}

// GetAll — получить всех клиентов
func (r *ClientRepository) GetAll(ctx context.Context) ([]domain.Client, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, client_type, phone, email, created_at
		FROM clients
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("GetAll: %w", err)
	}
	defer rows.Close()

	var clients []domain.Client
	for rows.Next() {
		var c domain.Client
		err := rows.Scan(&c.ID, &c.ClientType, &c.Phone, &c.Email, &c.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("GetAll scan: %w", err)
		}
		clients = append(clients, c)
	}
	return clients, nil
}

// GetByID — получить одного клиента по id
func (r *ClientRepository) GetByID(ctx context.Context, id string) (*domain.Client, error) {
	var c domain.Client
	err := r.db.QueryRow(ctx, `
		SELECT id, client_type, phone, email, created_at
		FROM clients
		WHERE id = $1
	`, id).Scan(&c.ID, &c.ClientType, &c.Phone, &c.Email, &c.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("GetByID: %w", err)
	}
	return &c, nil
}

// Create — создать клиента
// использует транзакцию потому что пишем в две таблицы:
// clients и individuals (если физлицо)
func (r *ClientRepository) Create(ctx context.Context, dto domain.CreateClientRequest) (*domain.Client, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("Create begin tx: %w", err)
	}
	// если до Commit не дойдём — транзакция откатится автоматически
	defer tx.Rollback(ctx)

	var c domain.Client
	err = tx.QueryRow(ctx, `
		INSERT INTO clients (client_type, phone, email)
		VALUES ($1, $2, $3)
		RETURNING id, client_type, phone, email, created_at
	`, dto.ClientType, dto.Phone, dto.Email,
	).Scan(&c.ID, &c.ClientType, &c.Phone, &c.Email, &c.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("Create insert client: %w", err)
	}

	// если физлицо — дополнительно пишем в individuals
	if dto.ClientType == domain.ClientIndividual {
		_, err = tx.Exec(ctx, `
			INSERT INTO individuals (client_id, last_name, first_name, middle_name)
			VALUES ($1, $2, $3, $4)
		`, c.ID, dto.LastName, dto.FirstName, dto.MiddleName)
		if err != nil {
			return nil, fmt.Errorf("Create insert individual: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("Create commit: %w", err)
	}
	return &c, nil
}