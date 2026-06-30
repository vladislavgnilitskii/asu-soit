# Проект: АСУ СОИТ ТехноСервис

Автоматизированная система управления для сервисного центра, предоставляющего
услуги ремонта IT-техники физическим и юридическим лицам.

## Стек

- **Backend:** Go (см. версию в backend/go.mod), HTTP-фреймворк gin (gin-gonic/gin),
  драйвер PostgreSQL — pgx/v5 (jackc/pgx), JWT — golang-jwt/jwt/v5, хеширование
  паролей — golang.org/x/crypto/bcrypt
- **Frontend:** JavaScript — пока не реализован. Планируется React. Когда будет
  создана папка /frontend — обновить этот раздел.
- **БД:** PostgreSQL. Схема в 3НФ. Все таблицы используют uuid (gen_random_uuid())
  вместо serial — намеренное решение по безопасности (защита от IDOR/перебора ID).

## Структура репозитория

```
asu-soit/
├── backend/
│   ├── cmd/server/main.go       — точка входа, сборка всех зависимостей
│   ├── internal/
│   │   ├── config/               — загрузка .env / переменных окружения
│   │   ├── db/                   — пул соединений pgxpool с PostgreSQL
│   │   ├── domain/models.go      — все структуры данных и DTO одним файлом
│   │   ├── repository/           — слой SQL-запросов (*_repo.go на сущность)
│   │   ├── handler/               — слой HTTP (gin.Context), без SQL
│   │   ├── router/router.go      — регистрация всех маршрутов
│   │   └── auth/                 — JWT генерация/валидация, middleware
│   └── go.mod / go.sum
├── frontend/                     — пока пусто
├── db/migrations/                — SQL DDL-миграции, применяются вручную psql
│   ├── 001_create_enums_and_references.sql
│   ├── 002_create_clients.sql
│   ├── 003_create_devices.sql
│   ├── 004_create_employees.sql
│   ├── 005_create_repair_requests.sql
│   ├── 006_create_warehouse.sql
│   ├── 007_create_financials.sql
│   └── 008_create_audit_log.sql
├── docs/                          — Obsidian vault, документация и ADR
└── k8s/                           — манифесты Kubernetes (план, пока не начато)
```

## Архитектура backend (важно соблюдать)

Слоистая архитектура, строгое разделение ответственности:

```
HTTP-запрос → router → handler → repository → PostgreSQL
```

- **router** — только регистрация маршрутов (`method + path → handler func`),
  никакой логики
- **handler** — читает HTTP-запрос (`gin.Context`), валидирует через
  `binding:"required"` теги, вызывает repository, формирует HTTP-ответ.
  НЕ содержит SQL. Использует `respondError/respondOK/respondCreated` из
  `handler/response.go` для единообразных ответов
- **repository** — только SQL через `pgxpool.Pool`. НЕ знает про HTTP/JSON/gin.
  Каждая сущность — отдельный файл `*_repo.go`. Конструктор `NewXxxRepository(db)`
  принимает пул снаружи (dependency injection)
- **domain** — структуры данных. Разделение Entity (`Client`) и DTO
  (`CreateClientRequest`) — Entity это то что в БД, DTO — то что приходит от клиента

Сборка зависимостей происходит только в `main.go`:
`pool → repository → handler → router → server`

## Соглашения по коду Go

- Все SQL-запросы — параметризованные (`$1, $2, ...`), никогда не строить SQL
  конкатенацией строк (защита от SQL-инъекций)
- Операции затрагивающие больше одной таблицы — обязательно в транзакции
  (`tx.Begin` + `defer tx.Rollback` + `tx.Commit`)
- Ошибки оборачиваются через `fmt.Errorf("контекст: %w", err)` — сохраняем
  стектрейс и добавляем понимание откуда ошибка
- `context.Context` — всегда первый аргумент в функциях работающих с БД,
  передаётся через `c.Request.Context()` из gin
- В JSON-тегах — snake_case (`json:"client_type"`), не camelCase
- Чувствительные поля (пароли, хеши) — `json:"-"`, никогда не возвращать в ответе
- Опциональные поля которые могут быть NULL в БД — указатели (`*string`, `*float64`)
- Постоянный uuid вместо serial/int для всех первичных ключей

## Соглашения по БД

- Все таблицы: `uuid PRIMARY KEY DEFAULT gen_random_uuid()`
- Naming: snake_case для таблиц и колонок
- Справочники вынесены в отдельные таблицы (`request_statuses`, `roles`,
  `device_types`, `part_categories`) — не enum/строки напрямую (3НФ)
- `created_at`/`updated_at` — `timestamptz NOT NULL DEFAULT now()`
- Внешние ключи — явный `ON DELETE` (`RESTRICT`/`CASCADE`/`SET NULL`),
  выбор обоснован в комментарии к миграции
- Персональные данные физлиц (паспорт, ФИО) изолированы в отдельной таблице
  `individuals`, отдельно от `organizations` — для соответствия ФЗ-152
- `audit_log` — журнал изменений (`jsonb old_data/new_data`) для аудита по ФЗ-149

## Как запускать локально

```bash
cd backend
set -x CGO_ENABLED 0      # fish shell; для bash: export CGO_ENABLED=0
go run cmd/server/main.go
```

PostgreSQL — локальный, через systemd (`sudo systemctl start postgresql`),
БД `asu_soit`, пользователь `asu_soit_user`. Подключение настраивается через
`.env` (см. `.env.example`) или переменные окружения, не через k8s пока.

Применение миграций — вручную, по порядку файлов:
```bash
psql -h localhost -U asu_soit_user -d asu_soit -f db/migrations/001_*.sql
# и так далее по номерам
```

## Тесты

Тесты пока не написаны. Когда будут — `go test ./...` из папки `backend`.

## Текущий статус реализации

Готово: клиенты (CRUD), заявки на ремонт (CRUD + смена статуса с историей),
JWT-авторизация сотрудников (login, middleware RequireAuth/RequireRole).

Не реализовано: устройства (devices) — есть только в БД и тестовых вставках
через psql, нет CRUD; склад (spare_parts, stock_movements); финансы (invoices);
frontend целиком; деплой (Docker/k8s); мониторинг; тесты.

## Важно при генерации кода

- Следуй существующей структуре слоёв буквально — не клади SQL в handler,
  не клади HTTP-логику в repository
- Новую сущность добавляй по образцу `client_repo.go` + `handler/client.go` —
  смотри эти файлы как референс перед тем как писать новый код
- Не меняй уже применённые миграции в `db/migrations/` — новые изменения схемы
  оформляй новым пронумерованным файлом
- Это учебный проект, цель — глубокое понимание автором (Владиславом) каждой
  технологии, не просто рабочий код. При больших изменениях — кратко объясняй
  почему сделано так, а не просто молча генерируй
