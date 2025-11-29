# Market Backend (Auth + Market)# Marketplace Backend - MVP
A Go-based microservices backend consisting of two services:

cd deployments
cd ../service/Auth
go mod download
cd ../Market
go mod download
cd service/Market
docker-compose logs -f market-service
# Marketplace Backend (Auth + Market)

Minimal backend for a marketplace MVP built with two Go services: Auth and Market.

## Overview
Auth service issues JWT access/refresh tokens and manages roles (user, seller, admin). Market service provides products, categories, cart, orders, seller profile and admin moderation.

## Stack
Go (Gin), PostgreSQL, Redis, Prometheus metrics, Docker Compose (dev). JWT (HS256). Migrations via golang-migrate.

## Quick Start (Docker Compose)
```bash
docker compose -f deployments/docker-compose.dev.yml build
docker compose -f deployments/docker-compose.dev.yml up -d
docker logs --tail=100 auth-service-dev
docker logs --tail=100 market-service-dev
```
Services: Auth :8081, Market :8080, Postgres ports 5433/5434, Redis 6380/6381.

## Environment
Copy `.env.example` to `.env`. Use identical `JWT_ACCESS_SECRET` for Auth and Market.

## Basic Flow
1. Register seller: `POST /auth/register` (role seller) → take access_token
2. Create seller profile: `POST /api/seller/register` (shop_name)
3. Create product: `POST /api/seller/products` (title, price, category_id)
4. Buyer adds to cart: `POST /api/cart/items`
5. Create order: `POST /api/user/orders`

## Key Endpoints (Auth)
`POST /auth/register`, `POST /auth/login`, `POST /auth/refresh`, `POST /auth/logout`, `GET /health`

## Key Endpoints (Market)
Public: `GET /health`, `GET /api/products`, `GET /api/categories`
Cart/User: `GET /api/cart`, `POST /api/cart/items`, `POST /api/user/orders`, `GET /api/user/orders`
Seller (role seller): `POST /api/seller/register`, `GET /api/seller/profile`, `POST /api/seller/products`
Admin (role admin): category & product moderation (e.g. `POST /api/admin/categories`, `PUT /api/admin/products/:id/status`)

## Roles
user (default) / seller (seller endpoints) / admin (moderation & management).

## Postman
Import `Market.postman_collection.json`. Fill `access_token` after Auth register/login.

## Tests
Run Auth tests:
```bash
cd service/Auth
go test ./...
```

## Development Without Compose
```bash
# start databases first via compose or locally
go run service/Auth/cmd/main.go
go run service/Market/cmd/main.go
```

## Migrations
```bash
migrate create -ext sql -dir db/auth_migrations -seq <name>
migrate create -ext sql -dir db/market_migrations -seq <name>
```

## Troubleshooting
- 404 seller profile → register seller (`POST /api/seller/register`)
- 400 create product → missing required fields (title, category_id, price)
- Invalid token on Market → secrets mismatch; check `JWT_ACCESS_SECRET`.

## Security Notes
Short-lived access tokens (15m), refresh tokens (7d), bcrypt passwords, basic role-based checks, prepared statements, optional rate limit middleware.

## License
MIT. See `LICENSE`.

## Contributing
Fork, create feature branch, PR with concise description.

## Directory (high level)
```
service/Auth        # auth service
service/Market      # market service
db/                 # migrations
deployments/        # compose and quickstart
Market.postman_collection.json
```

## Future Improvements
- Pagination parameters standardisation
- Product images storage service
- Async order events (outbox)
- Centralized configuration / secrets vault
</div>
