# 🔐 user-service

**user-service** is a modular microservice written in Go responsible for handling user authentication, authorization, and identity management. It provides secure user registration, login, and JWT-based access control for distributed systems.

---

[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=guttosm_user-service&metric=coverage)](https://sonarcloud.io/summary/new_code?id=guttosm_user-service)

## 🚀 Features

- ✅ User registration and login
- 🔐 JWT token generation (access tokens)
- 🧂 Password hashing with bcrypt
- 🔑 Role-based authorization support
- 🧱 Clean Architecture with modular domain/service layers
- 🧪 Unit and integration tests using Testcontainers
- 🧼 GitHub Actions CI/CD (lint → build → test → migrate → Docker)
- 📖 Swagger auto-generated API docs

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

```
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

## 📄 License

MIT © 2025 Gustavo Moraes

---
