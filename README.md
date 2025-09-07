[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=guttosm_user-service&metric=coverage)](https://sonarcloud.io/summary/new_code?id=guttosm_user-service)
[![Bugs](https://sonarcloud.io/api/project_badges/measure?project=guttosm_user-service&metric=bugs)](https://sonarcloud.io/summary/new_code?id=guttosm_user-service)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=guttosm_user-service&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=guttosm_user-service)

# 🔐 user-service

**user-service** is a modular microservice written in Go responsible for handling user authentication, authorization, and identity management. It provides secure user registration, login, and JWT-based access control for distributed systems.

---

## 🚀 Features

- ✅ User registration and login
- 🔐 JWT token generation (access tokens)
- 🧂 Password hashing with bcrypt (configurable cost via BCRYPT_COST)
- 🔑 Role-based authorization support
- 🧱 Clean Architecture with modular domain/service layers
- 📨 Correlation ID propagation (X-Request-ID middleware)
- � Context-aware repository & service calls (timeouts)
- 🚦 In-memory IP rate limiting (configurable in code)
- 🩺 Liveness (/healthz) & Readiness (/readyz) endpoints
- �🧪 Unit and integration tests using Testcontainers
- 🧼 GitHub Actions CI/CD (lint → build → test → migrate → Docker)
- 📖 Swagger auto-generated API docs (disabled manually in production recommended)

---

## 🧱 Stack

| Layer       | Tech                          |
|-------------|-------------------------------|
| Language    | Go 1.24                       |
| Web         | Gin                           |
| Database    | PostgreSQL                    |
| Auth        | JWT + bcrypt                  |
| Migrations  | Goose                         |
| CI/CD       | GitHub Actions                |
| Testing     | Go test + Testcontainers      |
| Docs        | Swagger via Swaggo            |

---

## 🧑‍💻 Getting Started

### 📦 Requirements

- Docker + Docker Compose
- Go 1.24+
- Make

### 🛠 Local Setup

```bash
# Create local volume folders
make setup

# Build and run the app and dependencies
make docker-up
```

Swagger docs: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

---

### 🧪 Running Tests

```bash
# Run unit tests
make test

# Run with HTML coverage
make coverage-html

```

---

### 🧹 Development Tasks

```bash

make lint           # Run golangci-lint
make fmt            # Format code
make tidy           # Clean up go.mod/go.sum
make swagger        # Generate Swagger docs
make build          # Compile binary
make clean          # Remove binary + coverage

```

---

### 🐳 Docker Commands

```bash

make docker-up        # Start all services
make docker-down      # Stop all containers
make docker-restart   # Rebuild and restart everything

```

---

### 🗃 Run DB Migrations

```bash

make migrate

```

Or run migrations in CI via GitHub Actions.

---

## 📦 GitHub Actions CI/CD

- Lint
- Build
- Test + coverage
- SonarCloud scan
- Migrations
- Docker image build (triggered on Git tags)

---

## 📁 Project Structure

```text
user-service/
├── cmd/                  # App entrypoint
├── internal/
│   ├── handler/          # HTTP layer (register, login)
│   ├── service/          # Auth logic (JWT, password)
│   ├── repository/       # DB access
│   ├── middleware/       # Auth middleware
│   └── domain/
│       ├── model/        # User model
│       └── dto/          # Request/response types
├── db/migrations/        # Goose SQL migrations
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── README.md
```

---

## 📡 API Endpoints (Core)

| Method | Path        | Description            |
|--------|-------------|------------------------|
| GET    | /healthz    | Liveness probe         |
| GET    | /readyz     | Readiness probe (DB)   |
| POST   | /api/register | User registration    |
| POST   | /api/login    | User authentication  |
| GET    | /swagger/*any | API docs (dev only)  |

All responses include an `X-Request-ID` header for tracing.

Rate limiting returns HTTP 429 when exceeded.

### Authentication

JWT tokens include user id and role claims. Add future audience/issuer validation as needed.

### Environment Variables (Key Ones)

| Name | Purpose | Default |
|------|---------|---------|
| JWT_SECRET | HMAC signing secret | (required) |
| JWT_EXPIRATION | Token TTL minutes | 60 |
| BCRYPT_COST | Password hashing cost | 10 (library default) |
| POSTGRES_DSN | Connection string | see docker-compose |

Unset / mis-set critical vars cause startup failure.

---

## 🧪 Quality Checks

Run static analysis to catch bugs and code smells early:

```bash
make vet           # Run go vet static analysis
make staticcheck   # Run staticcheck (auto-installs if missing)
make lint          # Run golangci-lint
make analyze       # Run all static analysis tools above
```

---

## 📄 License

MIT © 2025 Gustavo Moraes

---
