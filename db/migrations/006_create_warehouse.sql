-- db/migrations/006_create_warehouse.sql

CREATE TABLE spare_parts (
    id               uuid          PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id      uuid          NOT NULL REFERENCES part_categories(id),
    name             varchar(255)  NOT NULL,
    sku              varchar(100)  UNIQUE,       -- артикул
    purchase_price   numeric(12,2) NOT NULL CHECK (purchase_price >= 0),
    sale_price       numeric(12,2) NOT NULL CHECK (sale_price >= 0),
    quantity_in_stock int          NOT NULL DEFAULT 0 CHECK (quantity_in_stock >= 0),
    created_at       timestamptz   NOT NULL DEFAULT now()
);

-- Все движения по складу: приход от поставщика, выдача в ремонт, списание
CREATE TABLE stock_movements (
    id            uuid          PRIMARY KEY DEFAULT gen_random_uuid(),
    part_id       uuid          NOT NULL REFERENCES spare_parts(id) ON DELETE RESTRICT,
    request_id    uuid          REFERENCES repair_requests(id),  -- NULL если приход от поставщика
    employee_id   uuid          NOT NULL REFERENCES employees(id),
    movement_type movement_type NOT NULL,
    quantity      int           NOT NULL CHECK (quantity > 0),
    unit_price    numeric(12,2) NOT NULL CHECK (unit_price >= 0),
    -- Номер накладной — для прихода от поставщика
    invoice_number varchar(100),
    note          text,
    created_at    timestamptz   NOT NULL DEFAULT now()
);

CREATE INDEX idx_spare_parts_category    ON spare_parts(category_id);
CREATE INDEX idx_spare_parts_sku         ON spare_parts(sku);
CREATE INDEX idx_stock_movements_part    ON stock_movements(part_id);
CREATE INDEX idx_stock_movements_request ON stock_movements(request_id);
CREATE INDEX idx_stock_movements_created ON stock_movements(created_at DESC);
