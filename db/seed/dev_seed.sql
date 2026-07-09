-- db/seed/dev_seed.sql
--
-- Сид ТОЛЬКО для локального стенда (docker compose). НЕ применять в проде.
-- Миграции сотрудников не создают, поэтому свежую БД нечем логинить —
-- заводим админа и инженера с известным паролём.
--
-- Логины/пароль:  admin / admin123   и   engineer / admin123
-- (bcrypt-хеш ниже сгенерирован backend/cmd/genhash для строки "admin123").
--
-- Применяется автоматически ПОСЛЕ миграций (см. docker-compose.yml:
-- монтируется в /docker-entrypoint-initdb.d с именем, сортирующимся последним).

INSERT INTO employees (role_id, last_name, first_name, login, password_hash, is_active)
SELECT r.id, 'Администратор', 'Демо', 'admin',
       '$2a$10$ct44GDI51tVmZtQ43zTXiOC32EJrjq6NMBqbwn/AdofUpabggtBNy', true
FROM roles r
WHERE r.code = 'admin'
  AND NOT EXISTS (SELECT 1 FROM employees WHERE login = 'admin');

INSERT INTO employees (role_id, last_name, first_name, login, password_hash, is_active)
SELECT r.id, 'Инженеров', 'Демо', 'engineer',
       '$2a$10$ct44GDI51tVmZtQ43zTXiOC32EJrjq6NMBqbwn/AdofUpabggtBNy', true
FROM roles r
WHERE r.code = 'engineer'
  AND NOT EXISTS (SELECT 1 FROM employees WHERE login = 'engineer');
