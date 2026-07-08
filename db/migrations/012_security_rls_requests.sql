-- db/migrations/012_security_rls_requests.sql
--
-- Модуль безопасности №2: Row-Level Security (RLS) на заявках.
--
-- GRANT (модуль 1) решает «к каким ТАБЛИЦАМ есть доступ». RLS решает
-- «какие СТРОКИ таблицы видно». Демонстрируем два уровня:
--   * грубый (по роли): весь профильный персонал видит все заявки;
--   * тонкий (по сотруднику): инженер видит только заявки, назначенные ЛИЧНО ему.
--
-- «Кто я» БД узнаёт из сессионной переменной app.current_employee_id,
-- которую в бою выставляло бы приложение на каждый запрос
-- (SET app.current_employee_id = '<uuid из JWT>'), а на защите — psql вручную.
--
-- Важно: владелец таблицы (asu_soit_user) и суперпользователь ОБХОДЯТ RLS.
-- Поэтому Go-приложение (ходит владельцем) не затрагивается, а демонстрация
-- идёт под невладельческими ролями через SET ROLE. В продакшене приложение
-- подключалось бы невладельческой ролью — тогда RLS применялся бы и к нему.

ALTER TABLE repair_requests ENABLE ROW LEVEL SECURITY;

-- Профильный персонал (кроме инженера) видит и меняет все заявки
DROP POLICY IF EXISTS rr_staff_all ON repair_requests;
CREATE POLICY rr_staff_all ON repair_requests
    FOR ALL
    TO role_admin, role_manager, role_accountant, role_storekeeper
    USING (true)
    WITH CHECK (true);

-- Инженер — только свои заявки (assigned_to = его id из сессии).
-- NULLIF(...,'') страхует от пустой строки; при неустановленной переменной
-- current_setting(...,true) вернёт NULL → условие ложно → 0 строк (fail-closed).
DROP POLICY IF EXISTS rr_engineer_own ON repair_requests;
CREATE POLICY rr_engineer_own ON repair_requests
    FOR ALL
    TO role_engineer
    USING (assigned_to = NULLIF(current_setting('app.current_employee_id', true), '')::uuid)
    WITH CHECK (assigned_to = NULLIF(current_setting('app.current_employee_id', true), '')::uuid);
