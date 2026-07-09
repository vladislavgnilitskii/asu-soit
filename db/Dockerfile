# Образ Postgres для локального стенда: миграции и dev-сид «запечены» внутрь.
# initdb выполнит их при первой инициализации по порядку имён файлов.
# Так надёжнее, чем монтировать каталог + вложенный файл томами.
FROM postgres:18-alpine

# миграции 001…016 (применяются по порядку имён как superuser)
COPY migrations/*.sql /docker-entrypoint-initdb.d/

# dev-сид (admin/engineer) — имя сортируется после всех миграций
COPY seed/dev_seed.sql /docker-entrypoint-initdb.d/zzz_dev_seed.sql
