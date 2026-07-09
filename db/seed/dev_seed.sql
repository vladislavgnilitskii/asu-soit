-- db/seed/dev_seed.sql
--
-- Сид ТОЛЬКО для локального стенда (docker compose). НЕ применять в проде.
-- Миграции сотрудников не создают, поэтому свежую БД нечем логинить —
-- заводим админа и инженера с известным паролём.
--
-- Логины: admin, engineer, manager, storekeeper, accountant.
-- Пароль у всех: admin123
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

INSERT INTO employees (role_id, last_name, first_name, login, password_hash, is_active)
SELECT r.id, 'Менеджерова', 'Демо', 'manager',
       '$2a$10$ct44GDI51tVmZtQ43zTXiOC32EJrjq6NMBqbwn/AdofUpabggtBNy', true
FROM roles r
WHERE r.code = 'manager'
  AND NOT EXISTS (SELECT 1 FROM employees WHERE login = 'manager');

INSERT INTO employees (role_id, last_name, first_name, login, password_hash, is_active)
SELECT r.id, 'Кладовщиков', 'Демо', 'storekeeper',
       '$2a$10$ct44GDI51tVmZtQ43zTXiOC32EJrjq6NMBqbwn/AdofUpabggtBNy', true
FROM roles r
WHERE r.code = 'storekeeper'
  AND NOT EXISTS (SELECT 1 FROM employees WHERE login = 'storekeeper');

INSERT INTO employees (role_id, last_name, first_name, login, password_hash, is_active)
SELECT r.id, 'Бухгалтерова', 'Демо', 'accountant',
       '$2a$10$ct44GDI51tVmZtQ43zTXiOC32EJrjq6NMBqbwn/AdofUpabggtBNy', true
FROM roles r
WHERE r.code = 'accountant'
  AND NOT EXISTS (SELECT 1 FROM employees WHERE login = 'accountant');
