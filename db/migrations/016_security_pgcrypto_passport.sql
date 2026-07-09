-- db/migrations/016_security_pgcrypto_passport.sql
--
-- Модуль безопасности №6: шифрование паспортных данных (pgcrypto).
--
-- Маскирование (модуль 3) скрывает паспорт от ролей, но в самой таблице он
-- лежал бы открытым текстом — а значит, был бы виден в дампе/бэкапе и любому,
-- кто получил файлы БД или доступ суперпользователя. Шифруем паспорт «в покое»
-- (at rest): в таблице — только шифртекст (bytea). Расшифровать можно лишь
-- ключом, а КЛЮЧ ХРАНИТСЯ В ПРИЛОЖЕНИИ (env), НЕ в базе. Поэтому украденный
-- дамп базы без ключа бесполезен, и даже администратор БД в psql не прочитает
-- паспорт — только работающее приложение, у которого есть ключ.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- v_individuals (модуль 3) ссылается на паспортные столбцы — пересоздаём её
-- после смены типа.
DROP VIEW IF EXISTS v_individuals;

-- Паспорт → bytea (шифртекст pgp_sym_encrypt). Текущие значения NULL, поэтому
-- смена типа данные не теряет (USING NULL). Шифрование/дешифрование делает
-- приложение своим ключом при записи/чтении.
ALTER TABLE individuals
    ALTER COLUMN passport_series TYPE bytea USING NULL::bytea,
    ALTER COLUMN passport_number TYPE bytea USING NULL::bytea;

COMMENT ON COLUMN individuals.passport_series IS
    'Зашифровано pgp_sym_encrypt; ключ у приложения (env CRYPTO_KEY), не в БД';
COMMENT ON COLUMN individuals.passport_number IS
    'Зашифровано pgp_sym_encrypt; ключ у приложения (env CRYPTO_KEY), не в БД';

-- Витрина больше НЕ раскрывает паспорт: в БД он зашифрован, расшифровать может
-- только приложение с ключом. Имя клиента видно, паспорт — маркер «зашифровано».
CREATE VIEW v_individuals AS
SELECT i.id, i.client_id, i.last_name, i.first_name, i.middle_name,
       '(зашифровано)'::text AS passport_series,
       '(зашифровано)'::text AS passport_number
FROM individuals i;

GRANT SELECT ON v_individuals TO role_accountant, role_admin;
