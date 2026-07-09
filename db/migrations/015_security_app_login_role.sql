-- db/migrations/015_security_app_login_role.sql
--
-- Модуль безопасности №5: перевод приложения на непривилегированную роль.
--
-- До сих пор приложение ходило суперпользователем (postgres), который обходит
-- и GRANT, и RLS. Заводим технический ЛОГИН-роль app_backend, под которой
-- подключается приложение. Сама по себе она почти без прав (NOINHERIT — не
-- наследует привилегии role_*), но является ЧЛЕНОМ всех бизнес-ролей и потому
-- может на каждый запрос «перевоплощаться» в нужную: SET ROLE role_<роль>.
-- Тогда СУБД применяет к приложению и GRANT (модуль 1), и RLS (модуль 2):
-- инженер через API увидит только свои заявки, а хеши паролей и audit_log
-- база не отдаст никому, кроме admin.

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_backend') THEN
        -- NOINHERIT: без ambient-прав, доступ только через явный SET ROLE.
        -- Пароль для локальной разработки; в проде задаётся из секрета.
        CREATE ROLE app_backend LOGIN NOINHERIT NOSUPERUSER PASSWORD 'app_backend_pw';
    END IF;
END $$;

-- членство в бизнес-ролях — чтобы app_backend мог SET ROLE в каждую из них
GRANT role_admin, role_manager, role_engineer, role_storekeeper, role_accountant
    TO app_backend;

-- app_backend должна видеть схему, чтобы вызвать функцию логина ниже
GRANT USAGE ON SCHEMA public TO app_backend;

-- ── Логин через SECURITY DEFINER-функцию ─────────────────────────────────
-- Проверка пароля идёт ДО SET ROLE (личность ещё неизвестна), поэтому запрос
-- выполняет сама app_backend. Прямого доступа к employees (там password_hash)
-- у неё быть не должно. Даём его только через функцию: она SECURITY DEFINER
-- (выполняется правами владельца-суперпользователя), а app_backend имеет лишь
-- право EXECUTE. Так app_backend не может ни прочитать таблицу employees
-- целиком, ни выгрузить все хеши — только спросить по конкретному логину.
-- SET search_path фиксирован — защита от подмены имён объектов.
CREATE OR REPLACE FUNCTION auth_get_credentials(p_login text)
RETURNS TABLE (
    id           uuid,
    role_id      uuid,
    last_name    text,
    first_name   text,
    middle_name  text,
    login        text,
    password_hash text,
    is_active    boolean,
    role_code    text
)
LANGUAGE sql
SECURITY DEFINER
SET search_path = pg_catalog, public
AS $$
    SELECT e.id, e.role_id, e.last_name, e.first_name,
           e.middle_name, e.login, e.password_hash, e.is_active, ro.code
    FROM employees e
    JOIN roles ro ON ro.id = e.role_id
    WHERE e.login = p_login;
$$;

-- по умолчанию EXECUTE есть у PUBLIC — отзываем и выдаём только app_backend
REVOKE ALL ON FUNCTION auth_get_credentials(text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION auth_get_credentials(text) TO app_backend;
