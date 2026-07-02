-- db/migrations/009_optimize_indexes.sql
--
-- Оптимизация индексов (без изменения модели данных):
--   1. Удаляем индексы, дублирующие уже существующие UNIQUE-ограничения
--      (UNIQUE само создаёт индекс — второй btree-индекс на ту же колонку
--       лишний, только замедляет запись).
--   2. Добавляем индексы на FK-колонки, где их не было — PostgreSQL не
--      создаёт их автоматически, а JOIN/фильтрация по ним нужны в отчётах.

-- 1. Удаление дублирующих индексов
DROP INDEX IF EXISTS idx_employees_login;      -- дублирует employees_login_key
DROP INDEX IF EXISTS idx_invoices_request;     -- дублирует uq_invoices_request
DROP INDEX IF EXISTS idx_spare_parts_sku;      -- дублирует spare_parts_sku_key
DROP INDEX IF EXISTS idx_organizations_inn;    -- дублирует uq_organizations_inn

-- 2. Недостающие индексы на внешние ключи
CREATE INDEX IF NOT EXISTS idx_devices_device_type       ON devices(device_type_id);
CREATE INDEX IF NOT EXISTS idx_request_parts_part        ON request_parts(part_id);
CREATE INDEX IF NOT EXISTS idx_status_history_changed_by ON request_status_history(changed_by);
CREATE INDEX IF NOT EXISTS idx_status_history_status     ON request_status_history(status_id);
CREATE INDEX IF NOT EXISTS idx_stock_movements_employee  ON stock_movements(employee_id);
