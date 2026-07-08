-- db/migrations/011_drop_repair_requests_client_id.sql
--
-- Нормализация до строгой 3NF.
--
-- repair_requests.client_id транзитивно выводится через device → client
-- (у устройства ровно один владелец: devices.client_id). Хранение client_id
-- в заявке — избыточность (нарушение 3NF) и, сверх того, источник возможной
-- рассогласованности: раньше client_id и device_id приходили в запрос
-- раздельно и ничто не гарантировало, что client_id совпадает с владельцем
-- устройства. Убираем колонку — клиент заявки определяется через её устройство.
--
-- DROP COLUMN автоматически удаляет зависимые объекты: FK
-- repair_requests_client_id_fkey и индекс idx_repair_requests_client.

ALTER TABLE repair_requests DROP COLUMN client_id;
