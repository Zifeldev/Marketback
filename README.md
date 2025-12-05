# Marketback# Market Backend (Auth + Market)# Marketplace Backend - MVP

A Go-based microservices backend consisting of two services:

Микросервисный бэкенд для маркетплейса на Go.

cd deployments

## Архитектураcd ../service/Auth

go mod download

```cd ../Market

┌─────────────────┐     ┌─────────────────┐go mod download

│   Auth Service  │     │ Market Service  │cd service/Market

│     (8081)      │     │     (8080)      │docker-compose logs -f market-service

└────────┬────────┘     └────────┬────────┘# Marketplace Backend (Auth + Market)

         │                       │

    ┌────┴────┐             ┌────┴────┐Minimal backend for a marketplace MVP built with two Go services: Auth and Market.

    │ Postgres│             │ Postgres│

    │ (5433)  │             │ (5434)  │## Overview

    └─────────┘             └─────────┘Auth service issues JWT access/refresh tokens and manages roles (user, seller, admin). Market service provides products, categories, cart, orders, seller profile and admin moderation.

         │                       │

         └───────────┬───────────┘## Stack

                     │Go (Gin), PostgreSQL, Redis, Prometheus metrics, Docker Compose (dev). JWT (HS256). Migrations via golang-migrate.

               ┌─────┴─────┐

               │   Redis   │## Quick Start (Docker Compose)

               │  (6380)   │```bash

               └───────────┘docker compose -f deployments/docker-compose.dev.yml build

```docker compose -f deployments/docker-compose.dev.yml up -d

docker logs --tail=100 auth-service-dev

## Сервисыdocker logs --tail=100 market-service-dev

```

### Auth Service (порт 8081)Services: Auth :8081, Market :8080, Postgres ports 5433/5434, Redis 6380/6381.

- Регистрация и авторизация пользователей

- JWT токены (access + refresh)## Environment

- Роли: user, seller, adminCopy `.env.example` to `.env`. Use identical `JWT_ACCESS_SECRET` for Auth and Market.

- Blacklist токенов в Redis

## Basic Flow

### Market Service (порт 8080)1. Register seller: `POST /auth/register` (role seller) → take access_token

- Каталог товаров и категорий2. Create seller profile: `POST /api/seller/register` (shop_name)

- Корзина покупок3. Create product: `POST /api/seller/products` (title, price, category_id)

- Заказы4. Buyer adds to cart: `POST /api/cart/items`

- Управление продавцами5. Create order: `POST /api/user/orders`

- Загрузка изображений

## Key Endpoints (Auth)

## Быстрый старт`POST /auth/register`, `POST /auth/login`, `POST /auth/refresh`, `POST /auth/logout`, `GET /health`



### 1. Клонировать репозиторий## Key Endpoints (Market)

```bashPublic: `GET /health`, `GET /api/products`, `GET /api/categories`

git clone https://github.com/Zifeldev/Marketback.gitCart/User: `GET /api/cart`, `POST /api/cart/items`, `POST /api/user/orders`, `GET /api/user/orders`

cd MarketbackSeller (role seller): `POST /api/seller/register`, `GET /api/seller/profile`, `POST /api/seller/products`

```Admin (role admin): category & product moderation (e.g. `POST /api/admin/categories`, `PUT /api/admin/products/:id/status`)



### 2. Настроить переменные окружения## Roles

```bashuser (default) / seller (seller endpoints) / admin (moderation & management).

cd deployments

cp .env.example .env  # или используйте существующий .env## Postman

```Import `Market.postman_collection.json`. Fill `access_token` after Auth register/login.



### 3. Запустить все сервисы## Tests

```bashRun Auth tests:

docker compose up -d --build```bash

```cd service/Auth

go test ./...

### 4. Проверить статус```

```bash

docker compose ps## Development Without Compose

``````bash

# start databases first via compose or locally

## API Документацияgo run service/Auth/cmd/main.go

go run service/Market/cmd/main.go

После запуска доступна Swagger документация:```

- Auth Service: http://localhost:8081/swagger/index.html

- Market Service: http://localhost:8080/swagger/index.html## Migrations

```bash

## Endpointsmigrate create -ext sql -dir db/auth_migrations -seq <name>

migrate create -ext sql -dir db/market_migrations -seq <name>

### Auth Service```

| Метод | Endpoint | Описание |

|-------|----------|----------|## Troubleshooting

| POST | `/auth/register` | Регистрация |### Common Issues

| POST | `/auth/login` | Авторизация |- **404 Seller Profile**: You must register a seller profile (`POST /api/seller/register`) after creating an account and before adding products.

| POST | `/auth/refresh` | Обновление токена |- **400 Create Product**: Ensure `title`, `price`, and `category_id` are provided. `category_id` must exist.

| POST | `/auth/logout` | Выход |- **Role is 'user' not 'seller'**: Ensure you passed `"role": "seller"` during registration.

- **Invalid Token**: Check that `JWT_ACCESS_SECRET` is identical in both `.env` files (Auth and Market).

### Market Service- **DB Connection Refused**: Check `docker-compose logs`. Ensure ports 5433/5434 are not occupied.

| Метод | Endpoint | Описание |

|-------|----------|----------|### Diagnostics

| GET | `/api/products` | Список товаров |- Check health: `curl http://localhost:8081/health` (Auth), `curl http://localhost:8080/health` (Market).

| GET | `/api/products/:id` | Товар по ID |- Inspect logs: `docker logs -f market-service-dev`.

| GET | `/api/categories` | Список категорий |- Decode JWT: Use [jwt.io](https://jwt.io) to verify payload (`role`, `exp`).

| GET | `/api/cart` | Корзина (auth) |

| POST | `/api/cart/items` | Добавить в корзину |## Security Notes

| POST | `/api/user/orders` | Создать заказ |Short-lived access tokens (15m), refresh tokens (7d), bcrypt passwords, basic role-based checks, prepared statements, optional rate limit middleware.

| GET | `/api/user/orders` | Мои заказы |

## License

### Seller API (требуется роль seller)MIT. See `LICENSE`.

| Метод | Endpoint | Описание |

|-------|----------|----------|## Contributing

| POST | `/api/seller/register` | Регистрация продавца |Fork, create feature branch, PR with concise description.

| POST | `/api/seller/products` | Создать товар |

| GET | `/api/seller/products` | Мои товары |## Directory (high level)

```

### Admin API (требуется роль admin)service/Auth        # auth service

| Метод | Endpoint | Описание |service/Market      # market service

|-------|----------|----------|db/                 # migrations

| POST | `/api/admin/categories` | Создать категорию |deployments/        # compose and quickstart

| GET | `/api/admin/sellers` | Все продавцы |Market.postman_collection.json

| GET | `/api/admin/orders` | Все заказы |```



## Структура проекта## Future Improvements / Roadmap

### Technical Debt & Reliability

```- **Transactions**: Wrap order creation in a single transaction to ensure stock/order consistency.

marketbackf/- **SQL Builder**: Replace manual `fmt.Sprintf` SQL construction with a query builder to prevent errors.

├── db/- **Error Handling**: Implement unified typed errors and middleware for automatic HTTP status mapping.

│   ├── auth_migrations/    # Миграции Auth DB- **Configuration**: Remove hardcodes, add config validation, and improve environment variable handling.

│   └── market_migrations/  # Миграции Market DB- **Redis Resilience**: Add graceful degradation (skip rate limit) and logging when Redis is unavailable.

├── deployments/

│   ├── docker-compose.yml### Testing & Documentation

│   └── .env- **Tests**: Add integration tests (DB, Redis), increase unit coverage, and add E2E scenarios.

└── service/- **API Documentation**: Generate Swagger/OpenAPI specs for the Market service.

    ├── Auth/               # Auth микросервис

    │   ├── cmd/### Features

    │   └── internal/- **Pagination**: Standardize pagination parameters across endpoints.

    └── Market/             # Market микросервис- **Images**: Add a service for product image storage.

        ├── cmd/- **Async Events**: Implement Transactional Outbox for order events.

        └── internal/</div>

```

## Технологии

- **Go 1.21+**
- **Gin** - HTTP фреймворк
- **PostgreSQL 16** - база данных
- **Redis 7** - кэширование и blacklist токенов
- **golang-migrate** - миграции БД
- **Swagger** - API документация
- **Docker Compose** - оркестрация

## Переменные окружения

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `JWT_ACCESS_SECRET` | Секрет для access токенов | - |
| `JWT_REFRESH_SECRET` | Секрет для refresh токенов | - |
| `JWT_ACCESS_EXPIRATION` | Время жизни access токена | 15m |
| `JWT_REFRESH_EXPIRATION` | Время жизни refresh токена | 168h |
| `LOG_LEVEL` | Уровень логирования | info |

## Команды

```bash
# Запуск всех сервисов
docker compose up -d --build

# Остановка
docker compose down

# Логи
docker compose logs -f market-service
docker compose logs -f auth-service

# Пересборка одного сервиса
docker compose up -d --build market-service
```

## Лицензия

MIT
