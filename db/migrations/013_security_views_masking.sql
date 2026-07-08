-- db/migrations/013_security_views_masking.sql
--
-- Модуль безопасности №3: представления как граница доступа и маскирование PII.
--
-- Идея: вместо прямого доступа к таблице с чувствительными данными даём роли
-- доступ к ПРЕДСТАВЛЕНИЮ, которое:
--   * вообще не содержит секретный столбец (password_hash), или
--   * маскирует персональные данные (паспорт, ИНН) для непривилегированных ролей.
-- Представление читает базовую таблицу правами своего владельца, поэтому роль
-- может видеть безопасную «витрину», не имея доступа к самой таблице.
--
-- Ключевой момент: маскирование в витрине имеет смысл ТОЛЬКО если у роли нет
-- прямого доступа к таблице (иначе маску можно обойти, прочитав таблицу
-- напрямую). Поэтому у бухгалтера ниже отзываем прямой SELECT на PII-таблицы,
-- выданный в модуле 1, и оставляем единственный путь — маскирующую витрину.

-- ── Представление 1: сотрудники БЕЗ password_hash ────────────────────────
-- Роли, кроме admin, не имеют доступа к таблице employees вовсе (модуль 1),
-- но им нужно видеть, кто есть кто (напр. кто назначен на заявку).
-- Отдаём безопасную витрину: имена/логин/роль, но НЕ хеш пароля —
-- столбца password_hash в представлении просто нет.
CREATE OR REPLACE VIEW v_employees AS
SELECT e.id, e.last_name, e.first_name, e.middle_name,
       e.login, e.is_active, r.code AS role_code
FROM employees e
JOIN roles r ON r.id = e.role_id;

GRANT SELECT ON v_employees TO role_manager, role_engineer, role_storekeeper, role_accountant;

-- ── Представление 2: физлица с маскированием паспорта ─────────────────────
-- Полный паспорт видит только роль admin; всем остальным — маска.
-- current_user внутри представления = роль, выполняющая запрос.
CREATE OR REPLACE VIEW v_individuals AS
SELECT i.id, i.client_id, i.last_name, i.first_name, i.middle_name,
       CASE WHEN current_user = 'role_admin' THEN i.passport_series ELSE '**'     END AS passport_series,
       CASE WHEN current_user = 'role_admin' THEN i.passport_number ELSE '******' END AS passport_number
FROM individuals i;

-- ── Представление 3: организации с маскированием ИНН ──────────────────────
CREATE OR REPLACE VIEW v_organizations AS
SELECT o.id, o.client_id, o.name, o.contact_person,
       CASE WHEN current_user = 'role_admin' THEN o.inn ELSE '****' END AS inn,
       CASE WHEN current_user = 'role_admin' THEN o.kpp ELSE '****' END AS kpp
FROM organizations o;

-- Бухгалтеру персональные данные клиента нужны только чтобы опознать клиента
-- при выставлении счёта — паспорт и ИНН ему видеть незачем. Отзываем прямой
-- доступ к PII-таблицам (выдан в модуле 1) и оставляем маскирующие витрины.
REVOKE SELECT ON individuals, organizations FROM role_accountant;
GRANT  SELECT ON v_individuals, v_organizations TO role_accountant, role_admin;
