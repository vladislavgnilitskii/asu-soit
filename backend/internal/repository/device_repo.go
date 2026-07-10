package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vladislavgnilitskii/asu-soit/internal/dbtx"
	"github.com/vladislavgnilitskii/asu-soit/internal/domain"
)

// DeviceRepository — SQL-слой для устройств и справочника их типов
type DeviceRepository struct {
	db *pgxpool.Pool
}

func NewDeviceRepository(db *pgxpool.Pool) *DeviceRepository {
	return &DeviceRepository{db: db}
}

// q — executor запроса: транзакция из контекста (с личностью сотрудника) или пул.
func (r *DeviceRepository) q(ctx context.Context) dbtx.DBTX {
	return dbtx.From(ctx, r.db)
}

// GetAll — все устройства
func (r *DeviceRepository) GetAll(ctx context.Context, p domain.PageParams) ([]domain.Device, int, error) {
	var total int
	if err := r.q(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM devices`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("GetAll devices count: %w", err)
	}

	rows, err := r.q(ctx).Query(ctx, `
		SELECT id, client_id, device_type_id, brand, model,
		       serial_number, appearance_note, created_at
		FROM devices
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, p.Limit, p.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("GetAll devices: %w", err)
	}
	defer rows.Close()

	var devices []domain.Device
	for rows.Next() {
		var d domain.Device
		err := rows.Scan(
			&d.ID, &d.ClientID, &d.DeviceTypeID, &d.Brand, &d.Model,
			&d.SerialNumber, &d.AppearanceNote, &d.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("GetAll devices scan: %w", err)
		}
		devices = append(devices, d)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("GetAll devices rows: %w", err)
	}
	return devices, total, nil
}

// GetByID — одно устройство по id
func (r *DeviceRepository) GetByID(ctx context.Context, id string) (*domain.Device, error) {
	var d domain.Device
	err := r.q(ctx).QueryRow(ctx, `
		SELECT id, client_id, device_type_id, brand, model,
		       serial_number, appearance_note, created_at
		FROM devices
		WHERE id = $1
	`, id).Scan(
		&d.ID, &d.ClientID, &d.DeviceTypeID, &d.Brand, &d.Model,
		&d.SerialNumber, &d.AppearanceNote, &d.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("GetByID device: %w", err)
	}
	return &d, nil
}

// Create — зарегистрировать устройство клиента
func (r *DeviceRepository) Create(ctx context.Context, dto domain.CreateDeviceDTO) (*domain.Device, error) {
	var d domain.Device
	err := r.q(ctx).QueryRow(ctx, `
		INSERT INTO devices
		    (client_id, device_type_id, brand, model, serial_number, appearance_note)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, client_id, device_type_id, brand, model,
		          serial_number, appearance_note, created_at
	`, dto.ClientID, dto.DeviceTypeID, dto.Brand, dto.Model,
		nullifyEmpty(dto.SerialNumber), nullifyEmpty(dto.AppearanceNote),
	).Scan(
		&d.ID, &d.ClientID, &d.DeviceTypeID, &d.Brand, &d.Model,
		&d.SerialNumber, &d.AppearanceNote, &d.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("Create device: %w", err)
	}
	return &d, nil
}

// ListTypes — справочник типов устройств
func (r *DeviceRepository) ListTypes(ctx context.Context) ([]domain.DeviceType, error) {
	rows, err := r.q(ctx).Query(ctx, `
		SELECT id, name FROM device_types ORDER BY name
	`)
	if err != nil {
		return nil, fmt.Errorf("ListTypes: %w", err)
	}
	defer rows.Close()

	var types []domain.DeviceType
	for rows.Next() {
		var t domain.DeviceType
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return nil, fmt.Errorf("ListTypes scan: %w", err)
		}
		types = append(types, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ListTypes rows: %w", err)
	}
	return types, nil
}
