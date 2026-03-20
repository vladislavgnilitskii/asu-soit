-- db/migrations/002_create_clients.sql

CREATE TABLE clients (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    client_type client_type NOT NULL,
    phone       varchar(20) NOT NULL,
    email       varchar(150),
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

-- Физические лица
-- Паспортные данные — персональные данные по ФЗ-152,
-- хранятся отдельно для разграничения доступа
CREATE TABLE individuals (
    id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id       uuid        NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    last_name       varchar(100) NOT NULL,
    first_name      varchar(100) NOT NULL,
    middle_name     varchar(100),
    -- Паспортные данные — опциональны, нужны для юридических документов
    passport_series varchar(4),
    passport_number varchar(6),
    CONSTRAINT uq_individuals_client UNIQUE (client_id)
);

-- Юридические лица
CREATE TABLE organizations (
    id             uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id      uuid         NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    name           varchar(255) NOT NULL,
    inn            varchar(12)  NOT NULL,
    kpp            varchar(9),
    contact_person varchar(200),          -- ФИО представителя
    CONSTRAINT uq_organizations_client UNIQUE (client_id),
    CONSTRAINT uq_organizations_inn    UNIQUE (inn)
);

-- Индексы для поиска клиентов
CREATE INDEX idx_clients_phone      ON clients(phone);
CREATE INDEX idx_clients_email      ON clients(email);
CREATE INDEX idx_individuals_name   ON individuals(last_name, first_name);
CREATE INDEX idx_organizations_name ON organizations(name);
CREATE INDEX idx_organizations_inn  ON organizations(inn);
