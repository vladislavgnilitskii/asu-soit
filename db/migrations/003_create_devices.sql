-- db/migrations/003_create_devices.sql

CREATE TABLE devices (
    id             uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id      uuid         NOT NULL REFERENCES clients(id) ON DELETE RESTRICT,
    device_type_id uuid         NOT NULL REFERENCES device_types(id),
    brand          varchar(100) NOT NULL,
    model          varchar(150) NOT NULL,
    serial_number  varchar(100),
    -- Внешний вид при приёме — важно для фиксации состояния
    appearance_note text,
    created_at     timestamptz  NOT NULL DEFAULT now()
);

CREATE INDEX idx_devices_client        ON devices(client_id);
CREATE INDEX idx_devices_serial_number ON devices(serial_number);
