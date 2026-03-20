-- db/migrations/004_create_employees.sql

CREATE TABLE employees (
    id            uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id       uuid         NOT NULL REFERENCES roles(id),
    last_name     varchar(100) NOT NULL,
    first_name    varchar(100) NOT NULL,
    middle_name   varchar(100),
    login         varchar(100) NOT NULL UNIQUE,
    password_hash varchar(255) NOT NULL,   -- bcrypt, никогда не plain text
    is_active     boolean      NOT NULL DEFAULT true,
    created_at    timestamptz  NOT NULL DEFAULT now(),
    updated_at    timestamptz  NOT NULL DEFAULT now()
);

CREATE INDEX idx_employees_login   ON employees(login);
CREATE INDEX idx_employees_role_id ON employees(role_id);
