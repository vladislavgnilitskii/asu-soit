-- db/migrations/005_create_repair_requests.sql

CREATE TABLE repair_requests (
    id                  uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id           uuid         NOT NULL REFERENCES clients(id)  ON DELETE RESTRICT,
    device_id           uuid         NOT NULL REFERENCES devices(id)  ON DELETE RESTRICT,
    assigned_to         uuid         REFERENCES employees(id)         ON DELETE SET NULL,
    status_id           uuid         NOT NULL REFERENCES request_statuses(id),
    problem_description text         NOT NULL,
    diagnostic_result   text,
    -- Стоимость фиксируется на момент закрытия — не вычисляется динамически
    estimated_cost      numeric(12,2),
    final_cost          numeric(12,2),
    planned_deadline    timestamptz,
    created_at          timestamptz  NOT NULL DEFAULT now(),
    updated_at          timestamptz  NOT NULL DEFAULT now(),
    closed_at           timestamptz
);

-- История смены статусов — 3НФ: меняется статус → пишем запись
CREATE TABLE request_status_history (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id  uuid        NOT NULL REFERENCES repair_requests(id) ON DELETE CASCADE,
    status_id   uuid        NOT NULL REFERENCES request_statuses(id),
    changed_by  uuid        NOT NULL REFERENCES employees(id),
    changed_at  timestamptz NOT NULL DEFAULT now(),
    comment     text
);

-- Детали использованные в заявке
-- unit_price фиксируется на момент использования — цена может меняться
CREATE TABLE request_parts (
    id         uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id uuid         NOT NULL REFERENCES repair_requests(id) ON DELETE RESTRICT,
    part_id    uuid         NOT NULL REFERENCES spare_parts(id)     ON DELETE RESTRICT,
    quantity   int          NOT NULL CHECK (quantity > 0),
    unit_price numeric(12,2) NOT NULL CHECK (unit_price >= 0),
    CONSTRAINT uq_request_parts UNIQUE (request_id, part_id)
);

CREATE INDEX idx_repair_requests_client    ON repair_requests(client_id);
CREATE INDEX idx_repair_requests_device    ON repair_requests(device_id);
CREATE INDEX idx_repair_requests_status    ON repair_requests(status_id);
CREATE INDEX idx_repair_requests_assigned  ON repair_requests(assigned_to);
CREATE INDEX idx_repair_requests_created   ON repair_requests(created_at DESC);
CREATE INDEX idx_status_history_request    ON request_status_history(request_id);
