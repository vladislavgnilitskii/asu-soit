-- db/migrations/001_create_enums_and_references.sql

CREATE TYPE client_type AS ENUM ('individual', 'organization');

CREATE TYPE movement_type AS ENUM ('incoming', 'outgoing', 'writeoff');

CREATE TYPE payment_method AS ENUM ('cash', 'card', 'bank_transfer');

CREATE TYPE invoice_status AS ENUM ('pending', 'paid', 'cancelled');

-- Справочник ролей сотрудников
CREATE TABLE roles (
    id   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code varchar(50)  NOT NULL UNIQUE,  -- 'admin', 'engineer', 'storekeeper', 'accountant', 'manager', 'sysadmin'
    name varchar(100) NOT NULL
);

INSERT INTO roles (code, name) VALUES
    ('admin',       'Администратор'),
    ('engineer',    'Инженер по ремонту'),
    ('storekeeper', 'Кладовщик'),
    ('accountant',  'Бухгалтер'),
    ('manager',     'Руководитель'),
    ('sysadmin',    'Системный администратор');

-- Справочник статусов заявок
CREATE TABLE request_statuses (
    id   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code varchar(50)  NOT NULL UNIQUE,
    name varchar(100) NOT NULL,
    sort_order smallint NOT NULL DEFAULT 0  -- для сортировки в UI
);

INSERT INTO request_statuses (code, name, sort_order) VALUES
    ('new',          'Новая',                    1),
    ('in_diagnosis', 'На диагностике',           2),
    ('awaiting',     'Ожидает согласования',     3),
    ('in_repair',    'В ремонте',                4),
    ('testing',      'Тестирование',             5),
    ('done',         'Готово к выдаче',          6),
    ('closed',       'Выдано клиенту',           7),
    ('cancelled',    'Отменено',                 8);

-- Справочник типов устройств
CREATE TABLE device_types (
    id   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar(100) NOT NULL UNIQUE
);

INSERT INTO device_types (name) VALUES
    ('Ноутбук'),
    ('Настольный ПК'),
    ('Моноблок'),
    ('Сервер'),
    ('Сетевое оборудование'),
    ('Принтер / МФУ'),
    ('Планшет'),
    ('Смартфон'),
    ('Прочее');

-- Справочник категорий запчастей
CREATE TABLE part_categories (
    id   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar(100) NOT NULL UNIQUE
);

INSERT INTO part_categories (name) VALUES
    ('Оперативная память'),
    ('Накопители (HDD/SSD)'),
    ('Процессоры'),
    ('Материнские платы'),
    ('Блоки питания'),
    ('Системы охлаждения'),
    ('Дисплеи и матрицы'),
    ('Сетевые карты'),
    ('Расходные материалы'),
    ('Прочее');
