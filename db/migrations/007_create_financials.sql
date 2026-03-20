-- db/migrations/007_create_financials.sql

-- Счёт формируется после закрытия заявки
CREATE TABLE invoices (
    id           uuid           PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id   uuid           NOT NULL REFERENCES repair_requests(id) ON DELETE RESTRICT,
    total_amount numeric(12,2)  NOT NULL CHECK (total_amount >= 0),
    status       invoice_status NOT NULL DEFAULT 'pending',
    issued_at    timestamptz    NOT NULL DEFAULT now(),
    CONSTRAINT uq_invoices_request UNIQUE (request_id)  -- один счёт на заявку
);

CREATE INDEX idx_invoices_request ON invoices(request_id);
CREATE INDEX idx_invoices_status  ON invoices(status);
