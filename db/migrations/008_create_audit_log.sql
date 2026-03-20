-- db/migrations/008_create_audit_log.sql

-- Журнал всех значимых действий в системе
CREATE TABLE audit_log (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id uuid        REFERENCES employees(id) ON DELETE SET NULL,
    table_name  varchar(100) NOT NULL,
    action      varchar(10)  NOT NULL CHECK (action IN ('INSERT', 'UPDATE', 'DELETE')),
    record_id   uuid,                    -- id изменённой записи
    old_data    jsonb,                   -- состояние до изменения
    new_data    jsonb,                   -- состояние после изменения
    ip_address  inet,                    -- IP с которого пришёл запрос
    created_at  timestamptz NOT NULL DEFAULT now()
);

-- Индекс по времени — аудит чаще всего читают за период
CREATE INDEX idx_audit_log_created    ON audit_log(created_at DESC);
CREATE INDEX idx_audit_log_employee   ON audit_log(employee_id);
CREATE INDEX idx_audit_log_table      ON audit_log(table_name);
-- GIN-индекс для поиска внутри jsonb
CREATE INDEX idx_audit_log_old_data   ON audit_log USING gin(old_data);
CREATE INDEX idx_audit_log_new_data   ON audit_log USING gin(new_data);
