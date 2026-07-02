# STATE.md

Снимок состояния проекта АСУ СОИТ «ТехноСервис» на **2026-07-02**.
Составлен по факту чтения кода в рабочей директории (включая staged-изменения) и запросов к реальной БД через PostgreSQL.

> **Обновление 2026-07-02 — выполнены Этапы 0, 1 и 2.** Закрыты все security-дыры, устранён тех-долг, миграция 009 применена, добавлены модуль устройств, полноценные организации и полный жизненный цикл заявки (назначение, диагностика/стоимость, закрытие, история статусов). Сборка/vet/тесты проходят (Go 1.26.4). Решения — в `DECISIONS.md`, следующая задача — в `CURRENT_TASK.md`.

## Итог Этапа 2 (2026-07-02)

Заявки end-to-end (`handler/request.go`, `repository/request_repo.go`):
- `PATCH /requests/:id/assign` — назначение исполнителя (admin/manager).
- `PATCH /requests/:id` — частичное обновление диагностики/стоимости (admin/manager/engineer).
- `PATCH /requests/:id/close` — закрытие: статус `closed`, `closed_at`, запись в историю (admin/manager).
- `GET /requests/:id/history` — история статусов с кодом/названием статуса и ФИО автора.
- **§5.7 закрыт:** валидация `status_id` (→400) и существования заявки (→404) вместо 500; sentinel-ошибки `repository.ErrRequestNotFound`/`ErrStatusNotFound`.
- Автор всех записей истории — из JWT.

Проверено рантаймом end-to-end (полный цикл заявки + кейсы 400/404), тестовые данные удалены.

## Итог Этапа 1 (2026-07-02)

- **Организации:** `CreateClientRequest` расширен полями `name/inn/kpp/contact_person`; `ClientRepository.Create` пишет в `organizations`; `GET /clients/:id` возвращает `ClientDetails` (клиент + вложенный подтип individual/organization). Отменяет заглушку 501.
- **Модуль Devices:** `domain` (Device, DeviceType, CreateDeviceDTO), `repository/device_repo.go` (GetAll/GetByID/Create/ListTypes), `handler/device.go`, маршруты в роутере, DI в `main.go`.
  - `GET /devices`, `GET /devices/:id`, `GET /device-types` — любой авторизованный; `POST /devices` — admin/manager.
- **БД:** миграция 009 **применена** к живой БД (дубли-индексы удалены, индексы на FK добавлены).
- Проверено рантаймом end-to-end (создание организации + устройства, чтение с подтипом), тестовые данные удалены.

## Итог Этапа 0 (2026-07-02)

Изменённые/новые файлы:
- `handler/request.go` — ✅ автор смены статуса берётся из JWT (`c.GetString("employee_id")`), убран `X-Employee-ID`; ✅ 500 без утечки деталей; ✅ 404 отличается от сбоя БД.
- `handler/client.go` (переименован из `clinet.go`) — ✅ хелперы `respond*`; ✅ 404/500; ✅ валидация ФИО; ✅ организации честно отклоняются (501) — orphan больше не создаётся.
- `handler/auth.go`, `handler/response.go` — ✅ единый `respondInternal` с логированием, обобщённые 500.
- `router/router.go` — ✅ RBAC: clients POST `admin,manager`; requests POST `admin,manager`; PATCH status `admin,manager,engineer`.
- `config/config.go` — ✅ `APP_ENV` + fail-fast на дефолтные `JWT_SECRET`/`DB_PASSWORD` в production.
- `domain/models.go` — ✅ nullable-поля (`Client.Email`, `Individual.MiddleName`) через указатели; `repository/helpers.go` — `nullifyEmpty`.
- `cmd/genhash/main.go` — ✅ утилита перенесена из корня модуля, пароль из аргумента (был захардкожен); корневой `gen_crypto_pass.go` удалён.
- `auth/jwt_test.go`, `auth/middleware_test.go` — ✅ фундамент тестов (round-trip JWT, отклонение none-alg/чужого секрета, 401/403 middleware).
- `db/migrations/009_optimize_indexes.sql` — подготовлена (дубли-индексы + индексы на FK). **К «живой» БД не применялась** — ждёт подтверждения (изменение схемы).

Проверено рантаймом (сервер против живой БД, данные восстановлены после теста): 401 без токена, 403 при нехватке роли, 404 vs 500, 501 на организацию, запись автора истории из JWT.

## Стек (по факту в коде)

- Backend: Go, модуль `github.com/vladislavgnilitskii/asu-soit` (`backend/go.mod`)
- HTTP: Gin (`gin-gonic/gin`)
- БД: PostgreSQL 18, `pgx/v5` + `pgxpool`
- Auth: JWT (`golang-jwt/jwt/v5`, HS256), пароли — `golang.org/x/crypto/bcrypt`
- Конфиг: `.env` через `joho/godotenv`
- Frontend: отсутствует
- Инфраструктура (Docker/K8s/мониторинг): отсутствует

## Состояние БД (проверено запросами)

Схема из 16 таблиц применена и засеяна:

| Справочник / таблица | Строк |
|---|---|
| roles | 6 (`admin`, `engineer`, `storekeeper`, `accountant`, `manager`, `sysadmin`) |
| request_statuses | 8 (`new`→`in_diagnosis`→`awaiting`→`in_repair`→`testing`→`done`→`closed`, `cancelled`) |
| device_types | 9 |
| part_categories | 10 |
| employees | 1 (`vgnilitskii`, активен) |
| clients | 2 (оба individual) |
| devices | 1 |
| repair_requests | 1 |

Логин и создание заявок работоспособны на реальных данных (есть активный сотрудник, статус `new` присутствует).

---

## 1. Что реализовано (проверено чтением кода)

### Auth
- `POST /api/v1/auth/login` — логин по логину/паролю, bcrypt-сверка, проверка `is_active`, выдача JWT (HS256, TTL 24 ч, claims: `employee_id`, `login`, `role_code`). Единый ответ 401 «неверный логин или пароль» без раскрытия причины.
- `RequireAuth` — проверка Bearer-токена, защита от алгоритмической подмены (проверка `SigningMethodHMAC`), кладёт `employee_id`/`login`/`role_code` в контекст gin.
- `RequireRole` — проверка роли из контекста, fail-closed (по умолчанию 403).

### Клиенты (`/api/v1/clients`, под `RequireAuth`)
- `GET ""` — список клиентов.
- `GET "/:id"` — клиент по id.
- `POST ""` — создание клиента (только роль `admin`); в транзакции пишет в `clients` и, **если individual**, в `individuals`.

### Заявки (`/api/v1/requests`, под `RequireAuth`, без ролей)
- `GET ""`, `GET "/:id"`, `POST ""` (статус по умолчанию `new` из справочника).
- `PATCH "/:id/status"` — смена статуса с записью в `request_status_history` в транзакции.

### БД
- Миграции 001–008 применены (enum-типы, справочники, clients/individuals/organizations, devices, employees, repair_requests + history + parts, warehouse, financials, audit_log).

---

## 2. Что реализовано частично

- ✅ **Клиенты (individual + organization)** — реализованы оба подтипа (Этап 1), чтение подтипа через `GET /clients/:id`. Осталось: `GetAll` возвращает только базовые поля без подтипа; нет обновления/удаления клиента.
- ✅ **RBAC** — развешен по согласованной матрице (Этап 0, см. `DECISIONS.md`).
- ✅ **История статусов** — автор берётся из JWT (Этап 0).
- ✅ **Заявки end-to-end** — назначение исполнителя, диагностика/стоимость, закрытие (`closed_at`), чтение истории статусов реализованы (Этап 2). Осталось: связка с запчастями (`request_parts`) и счётом (`invoices`) — модули склада/финансов.

---

## 3. Что отсутствует

- ✅ **Handler/Repository для Devices** — реализованы в Этапе 1 (CRUD-минимум + справочник типов).
- **Handler/Repository для Warehouse** — `spare_parts`/`part_categories`/`stock_movements`/`request_parts` есть, кода нет. Списание запчастей на заявку не реализовано.
- **Handler/Repository для Finance** — `invoices` есть, выставление счёта не реализовано. Enum `payment_method` создан, но нигде не используется (нет колонки способа оплаты).
- **Employees CRUD / управление сотрудниками** — только чтение по логину для авторизации.
- **Audit log** — таблица `audit_log` есть, запись в неё из кода отсутствует.
- **Тесты** — ни одного `*_test.go`.
- **Frontend** — не начат.
- **Инфраструктура** (Docker/K8s/Prometheus/Grafana) — не начата.
- **Пагинация/фильтрация** списков — `GetAll` возвращает всё без limit/offset.

---

## 4. Технические долги

> Этап 0 закрыл пункты 1, 2, 3, 5, 6 (✅). Пункт 4 оставлен (минорный). Пункты 7–8 оформлены в миграцию 009 (не применена). Пункт 9 (именование констрейнтов) отложен как косметический.

1. ✅ **Расхождение стиля хендлеров.** `handler/clinet.go` отвечает через `c.JSON` напрямую и отдаёт клиенту `err.Error()`, тогда как `request.go` уже использует `respondOK`/`respondError`/`respondCreated`. Нарушение правила слоя Handler из CLAUDE.md.
2. **Опечатка в имени файла** `handler/clinet.go` → должно быть `client.go`.
3. **`backend/gen_crypto_pass.go`** — `package main` в корне `backend/` (не в `cmd/`), с захардкоженным паролем `password123`. Разовая утилита, замусоривает корень модуля; при `go build ./...` собирается лишний бинарь. Перенести в `cmd/` или удалить.
4. **`domain/models.go`** — Entity и DTO смешаны в одном файле (CLAUDE.md допускает не смешивать в одной структуре, но разнести по смыслу желательно).
5. **Утечка внутренних ошибок в HTTP-ответ.** И `clinet.go`, и `request.go` возвращают `err.Error()` из репозитория напрямую (в т.ч. текст SQL-ошибок) — стоит отдавать обобщённое сообщение, детали в лог.
6. **Небезопасные дефолты конфигурации** — `config.go` подставляет `DB_PASSWORD=localdevpassword` и `JWT_SECRET=dev-secret-change-in-production` по умолчанию. Для dev приемлемо, но нет валидации, что в production секрет переопределён.
7. **Дублирующиеся индексы** (по анализу схемы): `idx_employees_login`, `idx_invoices_request`, `idx_spare_parts_sku`, `idx_organizations_inn` дублируют индексы, созданные UNIQUE-ограничениями.
8. **Отсутствие индексов на части FK** — `devices.device_type_id`, `request_parts.part_id`, `request_status_history.changed_by`/`status_id`, `stock_movements.employee_id`.
9. **Непоследовательное именование UNIQUE-констрейнтов** — часть `*_key` (авто), часть `uq_*`.

> Долги 7–9 требуют изменения схемы → только через **новую** миграцию, существующие не редактировать. Правки требуют согласования (изменение схемы БД).

---

## 5. Потенциальные ошибки

> Пункты 1–6 закрыты в Этапе 0 (2026-07-02) — оставлены для истории с пометкой ✅.

1. ✅ **Спуфинг автора истории статусов (приоритет — высокий).** `request.go:60` берёт `employeeID` из заголовка `X-Employee-ID`, а не из JWT-claims. Любой авторизованный пользователь может записать в `request_status_history` чужой `employee_id`. Плюс: если заголовок не UUID существующего сотрудника — FK-нарушение отдаст клиенту 500 с текстом SQL-ошибки. **Правильно: брать `employee_id` из контекста (`c.GetString("employee_id")`).**
2. ✅ **Broken Access Control на заявках.** Развешен `RequireRole` по согласованной матрице (см. `DECISIONS.md`).
3. ✅ **Создание клиента-организации ломает целостность.** Теперь `client.go` отклоняет `organization` с 501 до записи в БД — orphan невозможен. Полноценная поддержка организаций — отдельная будущая задача.
4. ✅ **Nullable-поля как `string` вместо указателя.** `Client.Email` и `Individual.MiddleName` переведены на `*string`; пустой ввод пишется как NULL (`nullifyEmpty`).
5. ✅ **Нет валидации ФИО для individual.** Добавлена проверка `last_name`/`first_name` в `client.go`.
6. ✅ **`GetByID` не отличает «не найдено» от ошибки БД.** Оба хендлера проверяют `errors.Is(err, pgx.ErrNoRows)` → 404, иначе 500.
7. ✅ **Валидация смены статуса** — существование `status_id` (→400) и заявки (→404) проверяются до вставки (Этап 2).

---

## 6. Какие модули реализовывать дальше (приоритет)

1. **Исправление безопасности (быстрые правки, без новых модулей):** брать `employee_id` из JWT в `UpdateStatus`; развесить `RequireRole` на заявки; унифицировать `clinet.go` на хелперы и перестать отдавать `err.Error()`.
2. **Клиенты — организации + чтение подтипа:** дополнить DTO/репозиторий, чинить orphan-баг.
3. **Devices** (handler+repository) — простой CRUD, разблокирует полноценное создание заявки (сейчас `device_id` приходит «извне»).
4. **Warehouse** — запчасти + движение склада + привязка к заявке (`request_parts`, `stock_movements`).
5. **Finance** — выставление счёта по заявке (`invoices`).
6. **Тесты** — хотя бы репозитории и auth.

---

## 7. Roadmap до MVP

MVP = сотрудник может провести заявку по полному циклу: клиент → устройство → заявка → работа/запчасти → закрытие → счёт, с корректным RBAC и аудитом действий.

- **Этап 0 — Стабилизация (0.5–1 день).** Установить Go-тулчейн, прогнать `go build/vet`. Починить безопасность (JWT-автор, RBAC на заявки, унификация ответов). Убрать/перенести `gen_crypto_pass.go`.
- **Этап 1 — Клиенты и устройства.** Полный клиент (individual + organization, чтение подтипа), Devices CRUD.
- **Этап 2 — Заявки end-to-end.** Назначение инженера (`assigned_to`), диагностика/стоимость (`estimated_cost`/`final_cost`), закрытие (`closed_at`), просмотр истории статусов.
- **Этап 3 — Склад.** Запчасти, приход/списание, привязка запчастей к заявке.
- **Этап 4 — Финансы.** Счёт по закрытой заявке, статусы оплаты.
- **Этап 5 — Аудит и качество.** Запись в `audit_log`, тесты на критичные пути, пагинация списков.
- **Этап 6 — Инфраструктура/Frontend (post-MVP).** Docker, затем React — по отдельному согласованию.

---

## Не проверено (честная пометка)

- **`go build`/`go vet`/`go test`** — ✅ теперь проходят (Go 1.26.4 установлен). Тесты пока только для пакета `auth`; repository/handler покрытия нет (нужна тестовая БД).
- **Миграция 009** — ✅ применена к живой БД (2026-07-02).
- **`.env`** (реальный) — в `.gitignore`, не читался.
- **Причина staged-удаления** `.geminiignore` и `GEMINI.md` — вне контекста задачи.
