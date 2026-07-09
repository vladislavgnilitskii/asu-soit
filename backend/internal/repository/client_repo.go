package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vladislavgnilitskii/asu-soit/internal/dbtx"
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

// q — executor запроса: транзакция из контекста (с личностью сотрудника) или пул.
func (r *ClientRepository) q(ctx context.Context) dbtx.DBTX {
	return dbtx.From(ctx, r.db)
}

// GetAll — получить всех клиентов
func (r *ClientRepository) GetAll(ctx context.Context) ([]domain.Client, error) {
	rows, err := r.q(ctx).Query(ctx, `
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
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetAll clients rows: %w", err)
	}
	return clients, nil
}

// GetByID — получить одного клиента по id
func (r *ClientRepository) GetByID(ctx context.Context, id string) (*domain.Client, error) {
	var c domain.Client
	err := r.q(ctx).QueryRow(ctx, `
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
	tx, err := r.q(ctx).Begin(ctx)
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
	`, dto.ClientType, dto.Phone, nullifyEmpty(dto.Email),
	).Scan(&c.ID, &c.ClientType, &c.Phone, &c.Email, &c.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("Create insert client: %w", err)
	}

	// дополнительно пишем в таблицу подтипа
	switch dto.ClientType {
	case domain.ClientIndividual:
		_, err = tx.Exec(ctx, `
			INSERT INTO individuals (client_id, last_name, first_name, middle_name)
			VALUES ($1, $2, $3, $4)
		`, c.ID, dto.LastName, dto.FirstName, nullifyEmpty(dto.MiddleName))
		if err != nil {
			return nil, fmt.Errorf("Create insert individual: %w", err)
		}
	case domain.ClientOrganization:
		_, err = tx.Exec(ctx, `
			INSERT INTO organizations (client_id, name, inn, kpp, contact_person)
			VALUES ($1, $2, $3, $4, $5)
		`, c.ID, dto.Name, dto.INN, nullifyEmpty(dto.KPP), nullifyEmpty(dto.ContactPerson))
		if err != nil {
			return nil, fmt.Errorf("Create insert organization: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("Create commit: %w", err)
	}
	return &c, nil
}

// GetIndividual — данные физлица по client_id
func (r *ClientRepository) GetIndividual(ctx context.Context, clientID string) (*domain.Individual, error) {
	var ind domain.Individual
	err := r.q(ctx).QueryRow(ctx, `
		SELECT id, client_id, last_name, first_name, middle_name
		FROM individuals
		WHERE client_id = $1
	`, clientID).Scan(&ind.ID, &ind.ClientID, &ind.LastName, &ind.FirstName, &ind.MiddleName)
	if err != nil {
		return nil, fmt.Errorf("GetIndividual: %w", err)
	}
	return &ind, nil
}

// GetOrganization — данные организации по client_id
func (r *ClientRepository) GetOrganization(ctx context.Context, clientID string) (*domain.Organization, error) {
	var org domain.Organization
	err := r.q(ctx).QueryRow(ctx, `
		SELECT id, client_id, name, inn, kpp, contact_person
		FROM organizations
		WHERE client_id = $1
	`, clientID).Scan(&org.ID, &org.ClientID, &org.Name, &org.INN, &org.KPP, &org.ContactPerson)
	if err != nil {
		return nil, fmt.Errorf("GetOrganization: %w", err)
	}
	return &org, nil
}
