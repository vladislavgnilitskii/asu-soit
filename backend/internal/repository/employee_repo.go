package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vladislavgnilitskii/asu-soit/internal/dbtx"
	"github.com/vladislavgnilitskii/asu-soit/internal/domain"
)

type EmployeeRepository struct {
	db *pgxpool.Pool
}

func NewEmployeeRepository(db *pgxpool.Pool) *EmployeeRepository {
	return &EmployeeRepository{db: db}
}

// q — executor запроса: транзакция из контекста (с личностью сотрудника) или пул.
func (r *EmployeeRepository) q(ctx context.Context) dbtx.DBTX {
	return dbtx.From(ctx, r.db)
}

// ListRoles — справочник ролей (для формы создания сотрудника)
func (r *EmployeeRepository) ListRoles(ctx context.Context) ([]domain.Role, error) {
	rows, err := r.q(ctx).Query(ctx, `SELECT id, code, name FROM roles ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("ListRoles: %w", err)
	}
	defer rows.Close()

	var roles []domain.Role
	for rows.Next() {
		var role domain.Role
		if err := rows.Scan(&role.ID, &role.Code, &role.Name); err != nil {
			return nil, fmt.Errorf("ListRoles scan: %w", err)
		}
		roles = append(roles, role)
	}
	return roles, nil
}

// GetByLogin — найти сотрудника по логину для авторизации
func (r *EmployeeRepository) GetByLogin(ctx context.Context, login string) (*domain.Employee, string, error) {
	var emp domain.Employee
	var roleCode string

	// Логин идёт до SET ROLE, поэтому запрос выполняет техническая роль
	// app_backend, у которой нет прямого доступа к employees (там хеши паролей).
	// Данные для проверки пароля отдаёт SECURITY DEFINER-функция (миграция 015).
	err := r.q(ctx).QueryRow(ctx, `
		SELECT id, role_id, last_name, first_name,
		       middle_name, login, password_hash, is_active, role_code
		FROM auth_get_credentials($1)
	`, login).Scan(
		&emp.ID, &emp.RoleID, &emp.LastName, &emp.FirstName,
		&emp.MiddleName, &emp.Login, &emp.PasswordHash, &emp.IsActive,
		&roleCode,
	)
	if err != nil {
		return nil, "", fmt.Errorf("GetByLogin: %w", err)
	}
	return &emp, roleCode, nil
}

// employeeColumns — общий список колонок для чтения сотрудника.
// password_hash сюда НЕ входит: секрет не грузим без надобности.
// COALESCE(middle_name,”) страхует от NULL при сканировании в string.
const employeeColumns = `
	id, role_id, last_name, first_name,
	COALESCE(middle_name, ''), login, is_active
`

func scanEmployee(row pgx.Row) (*domain.Employee, error) {
	var e domain.Employee
	err := row.Scan(&e.ID, &e.RoleID, &e.LastName, &e.FirstName,
		&e.MiddleName, &e.Login, &e.IsActive)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// Create — завести сотрудника. passwordHash уже посчитан в хендлере (bcrypt).
func (r *EmployeeRepository) Create(ctx context.Context, dto domain.CreateEmployeeDTO, passwordHash string) (*domain.Employee, error) {
	// role_id проверяем заранее → понятный 400 вместо сырого FK-сбоя
	var roleExists bool
	if err := r.q(ctx).QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM roles WHERE id = $1)
	`, dto.RoleID).Scan(&roleExists); err != nil {
		return nil, fmt.Errorf("Create check role: %w", err)
	}
	if !roleExists {
		return nil, ErrRoleNotFound
	}

	var e domain.Employee
	err := r.q(ctx).QueryRow(ctx, `
		INSERT INTO employees (role_id, last_name, first_name, middle_name, login, password_hash)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, role_id, last_name, first_name, COALESCE(middle_name, ''), login, is_active
	`, dto.RoleID, dto.LastName, dto.FirstName, dto.MiddleName, dto.Login, passwordHash,
	).Scan(&e.ID, &e.RoleID, &e.LastName, &e.FirstName, &e.MiddleName, &e.Login, &e.IsActive)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrLoginTaken
		}
		return nil, fmt.Errorf("Create employee: %w", err)
	}
	return &e, nil
}

// GetAll — список сотрудников (без секретов)
func (r *EmployeeRepository) GetAll(ctx context.Context) ([]domain.Employee, error) {
	rows, err := r.q(ctx).Query(ctx, `SELECT `+employeeColumns+` FROM employees ORDER BY last_name, first_name`)
	if err != nil {
		return nil, fmt.Errorf("GetAll employees: %w", err)
	}
	defer rows.Close()

	var employees []domain.Employee
	for rows.Next() {
		e, err := scanEmployee(rows)
		if err != nil {
			return nil, fmt.Errorf("GetAll employees scan: %w", err)
		}
		employees = append(employees, *e)
	}
	return employees, nil
}

// GetByID — сотрудник по id (без секретов)
func (r *EmployeeRepository) GetByID(ctx context.Context, id string) (*domain.Employee, error) {
	e, err := scanEmployee(r.q(ctx).QueryRow(ctx, `SELECT `+employeeColumns+` FROM employees WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrEmployeeNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("GetByID employee: %w", err)
	}
	return e, nil
}

// Update — частичное обновление профиля/роли/активности (COALESCE: nil = не менять)
func (r *EmployeeRepository) Update(ctx context.Context, id string, dto domain.UpdateEmployeeDTO) (*domain.Employee, error) {
	// если меняют роль — проверяем, что новая существует
	if dto.RoleID != nil {
		var roleExists bool
		if err := r.q(ctx).QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM roles WHERE id = $1)
		`, *dto.RoleID).Scan(&roleExists); err != nil {
			return nil, fmt.Errorf("Update check role: %w", err)
		}
		if !roleExists {
			return nil, ErrRoleNotFound
		}
	}

	var e domain.Employee
	err := r.q(ctx).QueryRow(ctx, `
		UPDATE employees SET
		    role_id     = COALESCE($1, role_id),
		    last_name   = COALESCE($2, last_name),
		    first_name  = COALESCE($3, first_name),
		    middle_name = COALESCE($4, middle_name),
		    is_active   = COALESCE($5, is_active),
		    updated_at  = now()
		WHERE id = $6
		RETURNING id, role_id, last_name, first_name, COALESCE(middle_name, ''), login, is_active
	`, dto.RoleID, dto.LastName, dto.FirstName, dto.MiddleName, dto.IsActive, id,
	).Scan(&e.ID, &e.RoleID, &e.LastName, &e.FirstName, &e.MiddleName, &e.Login, &e.IsActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrEmployeeNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("Update employee: %w", err)
	}
	return &e, nil
}
