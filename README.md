# Marketplace Backend (Auth + Market)

A clean and unified description of the Marketplace backend written in Go, consisting of two microservices: **Auth** and **Market**.

---

## Overview
- **Auth Service (port 8081):** user registration, authentication, JWT tokens, roles (`user`, `seller`, `admin`).
- **Market Service (port 8080):** products, categories, cart, orders, seller management, image uploads, admin moderation.

---

## Tech Stack
- Go 1.21+ (Gin)
- PostgreSQL 16
- Redis 7
- JWT (HS256)
- Docker Compose
- golang-migrate (DB migrations)
- Swagger (API documentation)
- Prometheus (metrics)

---

## Quick Start

### 1. Clone the repository
```bash
git clone https://github.com/Zifeldev/Marketback.git
cd Marketback
```

### 2. Configure environment variables
```bash
cd deployments
cp .env.example .env
```
> **Important:** `JWT_ACCESS_SECRET` must be identical for both Auth and Market services.

### 3. Start services
```bash
docker compose -f docker-compose.dev.yml up -d --build
```

### 4. Health checks
```bash
curl http://localhost:8081/health # Auth
curl http://localhost:8080/health # Market
```

### 5. Swagger
- Auth: http://localhost:8081/swagger/index.html
- Market: http://localhost:8080/swagger/index.html

---

## Environment Variables
| Variable | Description | Required |
|----------|-------------|----------|
| `JWT_ACCESS_SECRET` | Access token secret | Yes |
| `JWT_REFRESH_SECRET` | Refresh token secret | Yes |
| `DB_PASSWORD` | PostgreSQL user password | Yes |
| `CORS_ALLOWED_ORIGINS` | CORS whitelist | Yes |
| `BASE_URL` | Public base URL for uploads | Yes |

---

## Key Commands
```bash
# Stop services
docker compose -f deployments/docker-compose.dev.yml down

# View logs
docker compose -f deployments/docker-compose.dev.yml logs -f market-service
docker compose -f deployments/docker-compose.dev.yml logs -f auth-service

# Run tests
cd service/Auth && go test ./...
cd service/Market && go test ./...
```

---

## Services

### Auth Service
- User registration & authentication
- JWT access & refresh tokens
- Token blacklist via Redis
- Role-based access control

### Market Service
- Product and category catalog
- Cart management
- Orders (user & admin)
- Seller profile & products
- Image uploads
- Admin moderation

---

## API Endpoints

### Auth Service
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/auth/register` | Register new user |
| POST | `/auth/login` | Login |
| POST | `/auth/refresh` | Refresh access token |
| POST | `/auth/logout` | Logout |
| GET | `/health` | Health check |

### Market Service — Public
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/products` | List products |
| GET | `/api/products/:id` | Get product by ID |
| GET | `/api/categories` | List categories |
| GET | `/health` | Health check |

### Market Service — User
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/cart` | Get user cart |
| POST | `/api/cart/items` | Add item to cart |
| PUT | `/api/cart/items/:id` | Update cart item |
| DELETE | `/api/cart/items/:id` | Remove from cart |
| POST | `/api/user/orders` | Create order |
| GET | `/api/user/orders` | List user orders |

### Market Service — Seller
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/seller/register` | Register seller profile |
| GET | `/api/seller/profile` | Get seller profile |
| PUT | `/api/seller/profile` | Update seller profile |
| POST | `/api/seller/products` | Create product |
| GET | `/api/seller/products` | List seller products |
| PUT | `/api/seller/products/:id` | Update product |
| DELETE | `/api/seller/products/:id` | Delete product |

### Market Service — Admin
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/admin/categories` | Create category |
| PUT | `/api/admin/categories/:id` | Update category |
| DELETE | `/api/admin/categories/:id` | Delete category |
| PUT | `/api/admin/products/:id/status` | Update product status |
| GET | `/api/admin/sellers` | List all sellers |
| PUT | `/api/admin/sellers/:id/status` | Update seller status |
| GET | `/api/admin/orders` | List all orders |
| PUT | `/api/admin/orders/:id/status` | Update order status |

---

## Basic Usage Flow
1. Register as seller: `POST /auth/register` with `"role": "seller"`.
2. Log in: `POST /auth/login` → get `access_token`.
3. Create seller profile: `POST /api/seller/register`.
4. Create product: `POST /api/seller/products`.
5. User adds items to cart.
6. User creates order.

---

## Project Structure
```
marketbackf/
├── db/
│   ├── auth_migrations/
│   └── market_migrations/
├── deployments/
│   ├── docker-compose.yml
│   ├── docker-compose.dev.yml
│   └── .env.example
├── service/
│   ├── Auth/
│   └── Market/
└── Market.postman_collection.json
```

---

## Security
- Short-lived access tokens (15m)
- Refresh tokens (7d)
- bcrypt password hashing
- Role-based access control
- Prepared SQL statements

---

## License
MIT

