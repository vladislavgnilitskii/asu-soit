-- db/migrations/017_grant_admin_v_employees.sql
--
-- role_admin не имел SELECT на витрину v_employees: GRANT ALL ON ALL TABLES
-- (миграция 010) выполнялся ДО создания витрины (миграция 013) и не
-- распространяется на объекты, созданные позже, а в 013 admin в список
-- грантов не попал. Из-за этого GET /engineers под админом возвращал пусто
-- (ошибка доступа пряталась в rows.Err() — см. фикс в репозитории).

GRANT SELECT ON v_employees TO role_admin;
