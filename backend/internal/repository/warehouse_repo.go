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

// WarehouseRepository — SQL-слой для запчастей и движений склада.
type WarehouseRepository struct {
	db *pgxpool.Pool
}

func NewWarehouseRepository(db *pgxpool.Pool) *WarehouseRepository {
	return &WarehouseRepository{db: db}
}

// q — executor запроса: транзакция из контекста (с личностью сотрудника) или пул.
func (r *WarehouseRepository) q(ctx context.Context) dbtx.DBTX {
	return dbtx.From(ctx, r.db)
}

// ListCategories — справочник категорий запчастей
func (r *WarehouseRepository) ListCategories(ctx context.Context) ([]domain.PartCategory, error) {
	rows, err := r.q(ctx).Query(ctx, `SELECT id, name FROM part_categories ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("ListCategories: %w", err)
	}
	defer rows.Close()

	var categories []domain.PartCategory
	for rows.Next() {
		var cat domain.PartCategory
		if err := rows.Scan(&cat.ID, &cat.Name); err != nil {
			return nil, fmt.Errorf("ListCategories scan: %w", err)
		}
		categories = append(categories, cat)
	}
	return categories, nil
}

// GetAllParts — все запчасти на складе
func (r *WarehouseRepository) GetAllParts(ctx context.Context) ([]domain.SparePart, error) {
	rows, err := r.q(ctx).Query(ctx, `
		SELECT id, category_id, name, sku, purchase_price, sale_price, quantity_in_stock, created_at
		FROM spare_parts
		ORDER BY name
	`)
	if err != nil {
		return nil, fmt.Errorf("GetAllParts: %w", err)
	}
	defer rows.Close()

	var parts []domain.SparePart
	for rows.Next() {
		var p domain.SparePart
		err := rows.Scan(&p.ID, &p.CategoryID, &p.Name, &p.SKU, &p.PurchasePrice, &p.SalePrice, &p.QuantityInStock, &p.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("GetAllParts scan: %w", err)
		}
		parts = append(parts, p)
	}
	return parts, nil
}

// GetPartByID — одна запчасть по id
func (r *WarehouseRepository) GetPartByID(ctx context.Context, id string) (*domain.SparePart, error) {
	var p domain.SparePart
	err := r.q(ctx).QueryRow(ctx, `
		SELECT id, category_id, name, sku, purchase_price, sale_price, quantity_in_stock, created_at
		FROM spare_parts
		WHERE id = $1
	`, id).Scan(&p.ID, &p.CategoryID, &p.Name, &p.SKU, &p.PurchasePrice, &p.SalePrice, &p.QuantityInStock, &p.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPartNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("GetPartByID: %w", err)
	}
	return &p, nil
}

// CreatePart — завести новую запчасть в каталоге.
// category_id проверяем заранее (а не полагаемся на FK-ошибку), чтобы
// вернуть понятный 400, а не голый сбой БД — тот же приём, что и в
// RequestRepository.UpdateStatus для status_id (см. §5.7 в STATE.md).
func (r *WarehouseRepository) CreatePart(ctx context.Context, dto domain.CreateSparePartDTO) (*domain.SparePart, error) {
	var categoryExists bool
	if err := r.q(ctx).QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM part_categories WHERE id = $1)
	`, dto.CategoryID).Scan(&categoryExists); err != nil {
		return nil, fmt.Errorf("CreatePart check category: %w", err)
	}
	if !categoryExists {
		return nil, ErrCategoryNotFound
	}

	var p domain.SparePart
	err := r.q(ctx).QueryRow(ctx, `
		INSERT INTO spare_parts (category_id, name, sku, purchase_price, sale_price)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, category_id, name, sku, purchase_price, sale_price, quantity_in_stock, created_at
	`, dto.CategoryID, dto.Name, nullifyEmpty(dto.SKU), dto.PurchasePrice, dto.SalePrice,
	).Scan(&p.ID, &p.CategoryID, &p.Name, &p.SKU, &p.PurchasePrice, &p.SalePrice, &p.QuantityInStock, &p.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("CreatePart insert: %w", err)
	}
	return &p, nil
}

// Receive — приход детали от поставщика: увеличивает остаток и пишет
// движение по складу. Блокировка строки не нужна — приход не может уйти
// в отрицательный остаток, гонка тут не создаёт некорректного состояния.
func (r *WarehouseRepository) Receive(ctx context.Context, partID, employeeID string, dto domain.ReceivePartsDTO) (*domain.SparePart, error) {
	tx, err := r.q(ctx).Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("Receive begin: %w", err)
	}
	defer tx.Rollback(ctx)

	var p domain.SparePart
	err = tx.QueryRow(ctx, `
		UPDATE spare_parts SET quantity_in_stock = quantity_in_stock + $1
		WHERE id = $2
		RETURNING id, category_id, name, sku, purchase_price, sale_price, quantity_in_stock, created_at
	`, dto.Quantity, partID,
	).Scan(&p.ID, &p.CategoryID, &p.Name, &p.SKU, &p.PurchasePrice, &p.SalePrice, &p.QuantityInStock, &p.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPartNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("Receive update stock: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO stock_movements (part_id, employee_id, movement_type, quantity, unit_price, invoice_number, note)
		VALUES ($1, $2, 'incoming', $3, $4, $5, $6)
	`, partID, employeeID, dto.Quantity, dto.UnitPrice, nullifyEmpty(dto.InvoiceNumber), nullifyEmpty(dto.Note))
	if err != nil {
		return nil, fmt.Errorf("Receive insert movement: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("Receive commit: %w", err)
	}
	return &p, nil
}

// WriteOff — списание детали (порча/недостача), без привязки к заявке.
// Блокируем строку (FOR UPDATE) на время транзакции: без этого два
// параллельных списания могут оба прочитать один и тот же остаток и
// оба решить, что деталей достаточно, — уйдём в минус в обход проверки.
func (r *WarehouseRepository) WriteOff(ctx context.Context, partID, employeeID string, dto domain.WriteOffPartsDTO) (*domain.SparePart, error) {
	tx, err := r.q(ctx).Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("WriteOff begin: %w", err)
	}
	defer tx.Rollback(ctx)

	var currentQty int
	var purchasePrice float64
	err = tx.QueryRow(ctx, `
		SELECT quantity_in_stock, purchase_price FROM spare_parts
		WHERE id = $1
		FOR UPDATE
	`, partID).Scan(&currentQty, &purchasePrice)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPartNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("WriteOff lock part: %w", err)
	}
	if currentQty < dto.Quantity {
		return nil, ErrInsufficientStock
	}

	var p domain.SparePart
	err = tx.QueryRow(ctx, `
		UPDATE spare_parts SET quantity_in_stock = quantity_in_stock - $1
		WHERE id = $2
		RETURNING id, category_id, name, sku, purchase_price, sale_price, quantity_in_stock, created_at
	`, dto.Quantity, partID,
	).Scan(&p.ID, &p.CategoryID, &p.Name, &p.SKU, &p.PurchasePrice, &p.SalePrice, &p.QuantityInStock, &p.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("WriteOff update stock: %w", err)
	}

	// цена списания — закупочная: это внутренняя потеря, а не счёт клиенту
	_, err = tx.Exec(ctx, `
		INSERT INTO stock_movements (part_id, employee_id, movement_type, quantity, unit_price, note)
		VALUES ($1, $2, 'writeoff', $3, $4, $5)
	`, partID, employeeID, dto.Quantity, purchasePrice, nullifyEmpty(dto.Note))
	if err != nil {
		return nil, fmt.Errorf("WriteOff insert movement: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("WriteOff commit: %w", err)
	}
	return &p, nil
}

// IssueToRequest — выдать деталь со склада в ремонт по конкретной заявке:
// списывает остаток, пишет движение по складу и добавляет строку в
// request_parts (будущий счёт клиенту) по цене sale_price на момент выдачи.
func (r *WarehouseRepository) IssueToRequest(ctx context.Context, requestID, employeeID string, dto domain.IssuePartToRequestDTO) (*domain.RequestPartEntry, error) {
	tx, err := r.q(ctx).Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("IssueToRequest begin: %w", err)
	}
	defer tx.Rollback(ctx)

	var requestExists bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM repair_requests WHERE id = $1)
	`, requestID).Scan(&requestExists); err != nil {
		return nil, fmt.Errorf("IssueToRequest check request: %w", err)
	}
	if !requestExists {
		return nil, ErrRequestNotFound
	}

	// блокируем строку детали — та же гонка, что и при списании
	var currentQty int
	var salePrice float64
	var partName string
	err = tx.QueryRow(ctx, `
		SELECT quantity_in_stock, sale_price, name FROM spare_parts
		WHERE id = $1
		FOR UPDATE
	`, dto.PartID).Scan(&currentQty, &salePrice, &partName)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPartNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("IssueToRequest lock part: %w", err)
	}
	if currentQty < dto.Quantity {
		return nil, ErrInsufficientStock
	}

	if _, err := tx.Exec(ctx, `
		UPDATE spare_parts SET quantity_in_stock = quantity_in_stock - $1 WHERE id = $2
	`, dto.Quantity, dto.PartID); err != nil {
		return nil, fmt.Errorf("IssueToRequest update stock: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO stock_movements (part_id, request_id, employee_id, movement_type, quantity, unit_price)
		VALUES ($1, $2, $3, 'outgoing', $4, $5)
	`, dto.PartID, requestID, employeeID, dto.Quantity, salePrice); err != nil {
		return nil, fmt.Errorf("IssueToRequest insert movement: %w", err)
	}

	var entry domain.RequestPartEntry
	err = tx.QueryRow(ctx, `
		INSERT INTO request_parts (request_id, part_id, quantity, unit_price)
		VALUES ($1, $2, $3, $4)
		RETURNING id, part_id, quantity, unit_price
	`, requestID, dto.PartID, dto.Quantity, salePrice,
	).Scan(&entry.ID, &entry.PartID, &entry.Quantity, &entry.UnitPrice)
	if err != nil {
		// uq_request_parts(request_id, part_id): эта деталь на эту заявку уже
		// выдавалась — ловим нарушение ограничения и отдаём понятную ошибку,
		// а не голый 500. Предварительный SELECT-проверка тут не подходит:
		// между проверкой и вставкой возможна гонка, констрейнт в БД — источник истины.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrAlreadyIssued
		}
		return nil, fmt.Errorf("IssueToRequest insert request_part: %w", err)
	}
	entry.PartName = partName

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("IssueToRequest commit: %w", err)
	}
	return &entry, nil
}

// GetRequestParts — детали, списанные на заявку (для просмотра/будущего счёта)
func (r *WarehouseRepository) GetRequestParts(ctx context.Context, requestID string) ([]domain.RequestPartEntry, error) {
	rows, err := r.q(ctx).Query(ctx, `
		SELECT rp.id, rp.part_id, sp.name, rp.quantity, rp.unit_price
		FROM request_parts rp
		JOIN spare_parts sp ON sp.id = rp.part_id
		WHERE rp.request_id = $1
		ORDER BY sp.name
	`, requestID)
	if err != nil {
		return nil, fmt.Errorf("GetRequestParts: %w", err)
	}
	defer rows.Close()

	var entries []domain.RequestPartEntry
	for rows.Next() {
		var e domain.RequestPartEntry
		if err := rows.Scan(&e.ID, &e.PartID, &e.PartName, &e.Quantity, &e.UnitPrice); err != nil {
			return nil, fmt.Errorf("GetRequestParts scan: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, nil
}
