package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vladislavgnilitskii/asu-soit/internal/domain"
)

type EmployeeRepository struct {
	db *pgxpool.Pool
}

func NewEmployeeRepository(db *pgxpool.Pool) *EmployeeRepository {
	return &EmployeeRepository{db: db}
}

// GetByLogin — найти сотрудника по логину для авторизации
func (r *EmployeeRepository) GetByLogin(ctx context.Context, login string) (*domain.Employee, string, error) {
	var emp domain.Employee
	var roleCode string

	err := r.db.QueryRow(ctx, `
		SELECT e.id, e.role_id, e.last_name, e.first_name,
		       e.middle_name, e.login, e.password_hash, e.is_active,
		       ro.code as role_code
		FROM employees e
		JOIN roles ro ON ro.id = e.role_id
		WHERE e.login = $1
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
