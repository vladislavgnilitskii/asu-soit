-- db/migrations/010_security_roles_grants.sql
--
-- Модуль безопасности №1: роли БД и минимальные привилегии (least privilege).
--
-- Идея: сейчас всё приложение ходит одним привилегированным пользователем.
-- Заведём отдельные роли под бизнес-функции и выдадим каждой ТОЛЬКО те права,
-- что реально нужны. Всё, что не выдано явно, — запрещено по умолчанию.
--
-- Роли в PostgreSQL глобальны для кластера (а не для одной БД), поэтому
-- создаём идемпотентно — иначе повторный прогон на пересозданной БД упадёт
-- с «role already exists».

DO $$
DECLARE
    r text;
BEGIN
    FOREACH r IN ARRAY ARRAY[
        'role_admin', 'role_manager', 'role_engineer',
        'role_storekeeper', 'role_accountant'
    ] LOOP
        IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = r) THEN
            EXECUTE format('CREATE ROLE %I NOLOGIN', r);
        END IF;
    END LOOP;
END $$;

-- Без USAGE на схему роль вообще не увидит таблицы
GRANT USAGE ON SCHEMA public TO
    role_admin, role_manager, role_engineer, role_storekeeper, role_accountant;

-- Справочники — несекретные, читать может любая роль
GRANT SELECT ON device_types, request_statuses, part_categories, roles
    TO role_manager, role_engineer, role_storekeeper, role_accountant;

-- ── role_admin: полный доступ (в т.ч. employees и audit_log) ─────────────
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO role_admin;

-- ── role_manager (приёмка): клиенты, устройства, заявки ──────────────────
GRANT SELECT, INSERT               ON clients, individuals, organizations, devices TO role_manager;
GRANT SELECT, INSERT, UPDATE       ON repair_requests        TO role_manager;
GRANT SELECT, INSERT               ON request_status_history TO role_manager;
GRANT SELECT                       ON request_parts, invoices TO role_manager;

-- ── role_engineer: ремонт, диагностика, установка деталей ────────────────
GRANT SELECT                       ON clients, devices       TO role_engineer;
GRANT SELECT, UPDATE               ON repair_requests        TO role_engineer;
GRANT SELECT, INSERT               ON request_status_history TO role_engineer;
GRANT SELECT, INSERT               ON request_parts          TO role_engineer;
GRANT SELECT, UPDATE               ON spare_parts            TO role_engineer; -- списание остатка при выдаче
GRANT SELECT, INSERT               ON stock_movements        TO role_engineer;

-- ── role_storekeeper (склад): БЕЗ доступа к клиентам (PII) и финансам ─────
GRANT SELECT, INSERT, UPDATE       ON spare_parts            TO role_storekeeper;
GRANT SELECT, INSERT               ON stock_movements        TO role_storekeeper;
GRANT SELECT, INSERT               ON request_parts          TO role_storekeeper;
GRANT SELECT                       ON repair_requests        TO role_storekeeper;

-- ── role_accountant (бухгалтерия): финансы; БЕЗ склада ───────────────────
GRANT SELECT, INSERT, UPDATE       ON invoices               TO role_accountant;
GRANT SELECT                       ON repair_requests, request_parts,
                                      clients, organizations, individuals TO role_accountant;

-- Что демонстрирует least privilege после этой миграции:
--   * employees (с password_hash) и audit_log доступны ТОЛЬКО role_admin;
--   * role_storekeeper НЕ видит clients/individuals (персональные данные)
--     и invoices (финансы);
--   * role_engineer НЕ видит invoices;
--   * role_accountant НЕ может писать на склад.
