-- db/migrations/014_security_audit_triggers.sql
--
-- Модуль безопасности №4: триггеры → audit_log (автоматический аудит).
--
-- Таблица audit_log существует с самого начала, но до сих пор её заполняло бы
-- само приложение — а значит, аудит можно «забыть» или намеренно обойти.
-- Триггер переносит журналирование на уровень СУБД: база пишет запись САМА на
-- каждый INSERT/UPDATE/DELETE, независимо от того, кто и как сделал изменение.
-- В отличие от RLS, триггеры срабатывают и для владельца, и для суперпользователя.

-- ── Функция аудита ────────────────────────────────────────────────────────
-- SECURITY DEFINER: функция выполняется правами своего владельца (суперпользователь),
-- поэтому запись в audit_log проходит, даже если изменение сделала роль без прав
-- на audit_log (напр. инженер правит свою заявку). При этом прямого INSERT в
-- audit_log у ролей нет — журнал пополняется ТОЛЬКО через триггер.
-- SET search_path фиксирован — защита от перехвата имён объектов (см. модуль 5).
-- Из снимка вырезается password_hash, чтобы хеши паролей не утекали в журнал.
CREATE OR REPLACE FUNCTION fn_audit() RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, public
AS $$
DECLARE
    v_emp uuid := NULLIF(current_setting('app.current_employee_id', true), '')::uuid;
    v_old jsonb;
    v_new jsonb;
    v_rec uuid;
BEGIN
    IF (TG_OP = 'DELETE') THEN
        v_old := to_jsonb(OLD) - 'password_hash';
        v_rec := OLD.id;
    ELSIF (TG_OP = 'INSERT') THEN
        v_new := to_jsonb(NEW) - 'password_hash';
        v_rec := NEW.id;
    ELSE -- UPDATE
        v_old := to_jsonb(OLD) - 'password_hash';
        v_new := to_jsonb(NEW) - 'password_hash';
        v_rec := NEW.id;
    END IF;

    INSERT INTO audit_log(employee_id, table_name, action, record_id, old_data, new_data)
    VALUES (v_emp, TG_TABLE_NAME, TG_OP, v_rec, v_old, v_new);

    RETURN NULL; -- AFTER-триггер, возвращаемое значение игнорируется
END;
$$;

-- ── Навешиваем аудит на критичные таблицы ─────────────────────────────────
-- Заявки (рабочий процесс), счета (финансы), сотрудники (учётные записи/роли).
DROP TRIGGER IF EXISTS trg_audit_repair_requests ON repair_requests;
CREATE TRIGGER trg_audit_repair_requests
    AFTER INSERT OR UPDATE OR DELETE ON repair_requests
    FOR EACH ROW EXECUTE FUNCTION fn_audit();

DROP TRIGGER IF EXISTS trg_audit_invoices ON invoices;
CREATE TRIGGER trg_audit_invoices
    AFTER INSERT OR UPDATE OR DELETE ON invoices
    FOR EACH ROW EXECUTE FUNCTION fn_audit();

DROP TRIGGER IF EXISTS trg_audit_employees ON employees;
CREATE TRIGGER trg_audit_employees
    AFTER INSERT OR UPDATE OR DELETE ON employees
    FOR EACH ROW EXECUTE FUNCTION fn_audit();

-- ── Журнал только для добавления (append-only) ────────────────────────────
-- Запрещаем изменение и удаление строк audit_log, чтобы историю нельзя было
-- переписать задним числом. INSERT (через триггер аудита) остаётся разрешён.
CREATE OR REPLACE FUNCTION fn_audit_guard() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'audit_log — журнал только для добавления: UPDATE/DELETE запрещены';
END;
$$;

DROP TRIGGER IF EXISTS trg_audit_log_immutable ON audit_log;
CREATE TRIGGER trg_audit_log_immutable
    BEFORE UPDATE OR DELETE ON audit_log
    FOR EACH ROW EXECUTE FUNCTION fn_audit_guard();
