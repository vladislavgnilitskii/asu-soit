# Frontend — АСУ СОИТ «ТехноСервис»

Внутренняя веб-панель оператора сервисного центра. Потребляет REST API Go-бэкенда.

## Стек

- **Vite + React + TypeScript** — SPA.
- **Tailwind CSS v4 + shadcn/ui** — компоненты и стили.
- **TanStack Query** — загрузка и кэш серверных данных.
- **React Router** — маршрутизация, защищённые маршруты.
- JWT хранится в `localStorage`, роль берётся из claims токена.

## Запуск (разработка)

```bash
npm install
npm run dev
```

Dev-сервер поднимается на `http://localhost:5173`. Запросы к `/api/*`
проксируются на Go-бэкенд `http://localhost:8080` (см. `vite.config.ts`) —
бэкенд должен быть запущен.

## Сборка

```bash
npm run build     # tsc + vite build → dist/
npm run preview   # предпросмотр собранной версии
```

## Структура

```
src/
  lib/         — api-клиент, auth-контекст, типы (зеркало DTO бэка), форматтеры
  components/  — AppLayout (оболочка) + ui/ (shadcn-компоненты)
  pages/       — экраны: Login, Dashboard, Requests, Clients
  App.tsx      — маршруты и защита
  main.tsx     — провайдеры (Query, Router, Auth)
```

## Реализовано

Вход, оболочка с ролевым меню, обзор, список заявок (с RLS — инженер видит
только свои), список клиентов. Дальше: устройства, склад, счета, сотрудники,
формы создания/редактирования.
