# Chess Opponent Analyzer

Сервис для разведки соперника на [lichess.org](https://lichess.org): по нику оппонента
собирает его последние партии через Lichess API и считает агрегированную статистику —
винрейт, любимые и самые результативные дебюты, среднюю точность и «тильт-фактор».

Состоит из двух частей:

- **Go-бэкенд** — REST API с авторизацией (JWT), кэшированием в Redis и хранением
  пользователей в PostgreSQL.
- **Браузерное расширение** (Manifest V3) — кнопка прямо на странице партии lichess,
  которая забирает ник оппонента и показывает анализ во встроенной панели.

---

## Содержание

- [Возможности](#возможности)
- [Стек](#стек)
- [Архитектура](#архитектура)
- [Структура проекта](#структура-проекта)
- [Конфигурация](#конфигурация)
- [Запуск: локальная разработка](#запуск-локальная-разработка)
- [Запуск: production (Docker)](#запуск-production-docker)
- [API](#api)
- [Метрики анализа](#метрики-анализа)
- [Браузерное расширение](#браузерное-расширение)
- [Тесты](#тесты)
- [Roadmap](#roadmap)

---

## Возможности

- 🔐 Регистрация и вход, JWT access/refresh токены с ротацией и привязкой к устройству.
- ♟️ Анализ оппонента по нику и виду контроля времени (`bullet` / `blitz` / `rapid` / `classical`).
- ⚡ Кэширование результатов в Redis (TTL 2 часа) — повторный запрос не дёргает Lichess.
- 🧱 Rate limiting: глобальный лимит + лимит на IP (на базе token bucket).
- 📖 Swagger-документация всех роутов из коробки.
- 🧩 Браузерное расширение для lichess.org.

---

## Стек

| Слой | Технология |
|------|-----------|
| Язык | Go 1.26 |
| HTTP | стандартный `net/http` (роутинг через method-based `ServeMux`) |
| БД | PostgreSQL (`pgx/v5`) |
| Кэш | Redis (`go-redis/v9`) |
| Auth | JWT (`golang-jwt/v5`), пароли — bcrypt поверх sha256 |
| Логи | `zap` |
| Документация | Swagger (`swaggo/swag` + `http-swagger`) |
| Миграции | `golang-migrate` |
| Расширение | Manifest V3 (vanilla JS) |
| Инфраструктура | Docker Compose, Caddy (reverse proxy) |

---

## Архитектура

```
┌─────────────────────────────┐
│  Браузерное расширение (MV3) │
│  lichess.org/*               │
│                              │
│  content script ──┐          │   клик «Анализировать»
│                   ▼          │
│        background service    │
│        worker (fetch)        │
└───────────────────┬──────────┘
                    │  GET /analyze/{username}?speed=  (Bearer JWT)
                    ▼
┌──────────────────────────────────────────────┐
│                 Go API                         │
│                                                │
│  middleware: logging → ratelimit → CORS        │
│       │                                         │
│       ▼                                         │
│  features/auth        features/analysis         │
│   handler              handler                  │
│   service              service ──┐              │
│   repository           │         │              │
│       │                │         │              │
└───────┼────────────────┼─────────┼─────────────┘
        ▼                ▼         ▼
   ┌─────────┐      ┌────────┐  ┌──────────────┐
   │ Postgres│      │ Redis  │  │ Lichess API  │
   │ users / │      │ cache  │  │ (ndjson      │
   │ tokens  │      │ 2h TTL │  │  партий)     │
   └─────────┘      └────────┘  └──────────────┘
```

Бэкенд построен по принципу **vertical slices + чистые слои**:

- `cmd/` — точка входа.
- `internal/app/` — сборка зависимостей, жизненный цикл сервера, фоновые задачи
  (очистка протухших токенов и IP-лимитеров).
- `internal/features/{auth,analysis}/` — фичи, каждая со своими
  `handler` / `service` / `repository` / `dto`.
- `internal/infrastructure/{postgres,redis,lichess}/` — клиенты внешних систем.
- `internal/core/` — сквозная функциональность: `config`, `middleware`,
  `ratelimiter`, `response`, `apperrors`, `logger`.
- `internal/domain/` — доменные сущности.

---

## Структура проекта

```
.
├── cmd/main.go                # точка входа, Swagger general info
├── internal/
│   ├── app/                   # wiring, сервер, фоновые джобы
│   ├── core/                  # config, middleware, ratelimiter, response, apperrors, logger
│   ├── domain/                # User, RefreshToken, Analysis
│   ├── features/
│   │   ├── auth/              # регистрация/логин/токены/пользователи
│   │   └── analysis/          # анализ оппонента
│   └── infrastructure/        # postgres, redis, lichess clients
├── migrations/                # SQL-миграции (golang-migrate)
├── docs/                      # сгенерированный Swagger (swag init)
├── js_extention/              # браузерное расширение (MV3)
├── Dockerfile
├── docker-compose.yml         # db + redis + web + migration + caddy
├── Caddyfile
└── Makefile
```

---

## Конфигурация

Все настройки берутся из переменных окружения. Шаблон — в [`.env.example`](.env.example).

| Переменная | Назначение |
|------------|-----------|
| `PG_USER`, `PG_PASSWORD`, `PG_HOST`, `PG_PORT`, `PG_DATABASE`, `PG_SSL_MODE` | PostgreSQL |
| `REDIS_HOST`, `REDIS_PORT` | Redis |
| `JWT_SECRET` | секрет для подписи JWT |
| `GO_PORT` | порт HTTP-сервера |
| `APP_ENV` | `development` / `production` |
| `CORS_ALLOWED_ORIGINS` | разрешённые Origin (по умолчанию `*`) |
| `GLOBAL_RATE_LIMIT`, `GLOBAL_RATE_BURST` | глобальный rate limit (rps + burst) |
| `IP_RATE_LIMIT`, `IP_RATE_BURST` | rate limit на IP (rps + burst) |
| `LICHESS_GET_GAMES_URL` | базовый URL Lichess API (`https://lichess.org/api/games/user/`) |

> **Два файла окружения:**
> - `.env` — для локальной разработки и для подстановки переменных в `docker-compose.yml`
>   (здесь `PG_HOST=localhost`).
> - `.env.docker` — runtime-окружение внутри контейнера `web` при запуске через Docker.
>   Отличается тем, что хосты указывают на сервисы compose: `PG_HOST=db`, `REDIS_HOST=redis`.

---

## Запуск: локальная разработка

Подходит, когда код гоняется напрямую через `go run`, а Postgres/Redis крутятся локально.

**Требования:** Go 1.26+, PostgreSQL, Redis, [`golang-migrate`](https://github.com/golang-migrate/migrate)
(CLI `migrate`), при правке Swagger — [`swag`](https://github.com/swaggo/swag).

```bash
# 1. Подготовить окружение
cp .env.example .env       # заполнить значения, PG_HOST=localhost / REDIS_HOST=localhost

# 2. Применить миграции
make local-migrate-up

# 3. Запустить сервер (Makefile подхватывает .env через include/export)
make local-dev-run
```

API поднимется на `http://localhost:${GO_PORT}`, Swagger — на
`http://localhost:${GO_PORT}/swagger/index.html`.

Полезные команды:

```bash
make make-migration name=add_something   # создать новую пару миграций
make local-migrate-down                  # откатить миграции
make test                                # прогнать тесты
```

При изменении Swagger-аннотаций пересобрать спеку:

```bash
swag init -g cmd/main.go --parseInternal --parseDependency -o docs
```

---

## Запуск: production (Docker)

Поднимает весь стек: PostgreSQL, Redis, миграции (одноразово), сам сервис и Caddy
как reverse proxy на портах 80/443.

```bash
# 1. Заполнить оба файла окружения
cp .env.example .env                 # значения для подстановки в compose (creds, GO_PORT)
cp .env.docker.example .env.docker   # runtime для контейнера: PG_HOST=db, REDIS_HOST=redis

# 2. Собрать и поднять
make deploy                   # docker compose up --build -d --remove-orphans
```

Управление:

```bash
make logs        # логи всех сервисов
make logs-web    # логи только бэкенда
make restart     # перезапустить web
make down        # остановить всё
```

---

## API

Полный и актуальный референс — в Swagger UI (`/swagger/index.html`).
Кратко:

| Метод | Путь | Auth | Описание |
|-------|------|:----:|----------|
| `POST` | `/auth/register` | — | Регистрация, возвращает пару токенов |
| `POST` | `/auth/login` | — | Вход по логину/паролю |
| `POST` | `/auth/refresh` | — | Обновить пару токенов по refresh-токену |
| `POST` | `/auth/logout` | — | Инвалидировать один refresh-токен |
| `POST` | `/auth/logout-all` | 🔒 | Инвалидировать все токены пользователя |
| `GET` | `/users/{id}` | 🔒 | Данные пользователя |
| `PATCH` | `/users/{id}` | 🔒 | Обновить username / lichess username |
| `GET` | `/analyze/{username}?speed=` | 🔒 | Анализ оппонента |

🔒 — требуется заголовок `Authorization: Bearer <access_token>`.

---

## Метрики анализа

`GET /analyze/{username}?speed=blitz` возвращает (по последним ~100 партиям выбранного контроля):

| Поле | Что это |
|------|---------|
| `winrate` | общий процент побед |
| `winrate_last10_days` | винрейт за последние 10 дней |
| `most_popular_debut_white` / `_black` | самый частый дебют за белых / чёрных |
| `most_winrate_debut_white` / `_black` | самый результативный дебют за белых / чёрных |
| `avg_accuracy` | средняя точность игры |
| `avg_accuracy_last10_days` | средняя точность за последние 10 дней |
| `tilt_factor` | доля поражений сдачей среди всех поражений (склонность «разваливаться») |

Результат кэшируется в Redis на 2 часа по ключу `{username}_{speed}`.

---

## Браузерное расширение

Manifest V3 расширение для Chrome/Chromium/Edge лежит в [`js_extention/`](js_extention/).
По кнопке на странице партии lichess оно забирает ник оппонента, шлёт запрос к API
через background service worker (что обходит CORS) и рисует панель с результатом.

Установка: `chrome://extensions` → Developer mode → **Load unpacked** → папка `js_extention`.

Подробности, авторизация и нюансы — в [`js_extention/README.md`](js_extention/README.md).

---

## Тесты

```bash
make test        # go test ./... -v
```

Покрыты юнит-тестами:

- **`features/auth`** — валидация DTO, хеширование паролей, генерация JWT и refresh-токенов
  (на моках репозитория).
- **`features/analysis`** — агрегация статистики на мок-партиях (Lichess поднимается как
  `httptest`-сервер с ndjson), плюс граничные случаи: нет партий, игрок не найден,
  отсутствие поражений, исключение чужих партий из знаменателя.
- **`core/ratelimiter`** — глобальный и per-IP лимитеры, переиспользование и сброс.
- **`core/middleware`** — auth middleware: валидный токен, протухший, чужой секрет, мусор.

---

## Roadmap

Планы и технический долг — в [`TODO.md`](TODO.md). Среди прочего: поддержка chess.com,
асинхронный анализ 500+ партий, сравнение своей статистики с оппонентом, метрики (Prometheus).
