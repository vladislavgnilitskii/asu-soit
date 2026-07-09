# АСУ СОИТ — ТехноСервис

Автоматизированная система управления сервисным обслуживанием IT-техники.

## Стек
- Backend: Go (gin, pgx)
- Frontend: React + TypeScript (Vite, Tailwind, shadcn/ui)
- БД: PostgreSQL 18
- Инфраструктура: Docker, Kubernetes

## Структура репозитория
- `backend/` — Go-приложение
- `frontend/` — React-приложение
- `db/migrations/` — SQL-миграции
- `db/seed/` — dev-сид для локального стенда
- `k8s/` — манифесты Kubernetes
- `docs/` — документация (Obsidian vault)

## Быстрый старт (локальный стенд)

Бэкенд + БД одной командой (нужен Docker):

```bash
docker compose up --build
```

Поднимутся:
- **PostgreSQL** на `localhost:5433` — миграции `001…016` применяются автоматически
  при первой инициализации, затем dev-сид создаёт пользователей
  `admin` / `engineer` (пароль у обоих — `admin123`);
- **backend** на `localhost:8080` (ходит в БД непривилегированной ролью
  `app_backend`).

Чистый прогон с нуля: `docker compose down -v && docker compose up --build`.

Фронтенд пока запускается отдельно (его dev-сервер проксирует `/api` на `:8080`):

```bash
cd frontend && npm install && npm run dev   # http://localhost:5173
```

Вход в UI: `admin` / `admin123`.

## Документация
Документация ведётся в Obsidian в папке `docs/`.

## Документация
Документация ведётся в Obsidian в папке `docs/`.
