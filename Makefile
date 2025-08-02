# ─────────── Settings ───────────
APP_NAME := user-service
GO_FILES := $(shell find . -name '*.go' -not -path "./vendor/*")
SWAG ?= github.com/swaggo/swag/cmd/swag
SWAG_BIN := $(shell go env GOPATH)/bin/swag

# ─────────── Commands ───────────

## Setup: Create local folders for bind mounts
setup:
	mkdir -p postgres-data

## Install dependencies and tools
install:
	go mod download
	go install $(SWAG)@latest

## Run locally without Docker
run:
	@echo "🚀 Running $(APP_NAME)..."
	go run ./cmd/main.go

## Build Go binary
build:
	@echo "🔨 Building $(APP_NAME)..."
	go build -o $(APP_NAME) ./cmd/main.go

## Run tests with coverage
test:
	go test ./... -coverprofile=coverage.out

## View coverage report in terminal
coverage:
	go tool cover -func=coverage.out

## View coverage report in browser
coverage-html:
	go tool cover -html=coverage.out

## Format code
fmt:
	go fmt ./...

## Tidy dependencies
tidy:
	go mod tidy

## Run linter (assumes golangci-lint is installed)
lint:
	golangci-lint run

## Generate Swagger docs
swagger:
	go install $(SWAG)@latest
	$(SWAG_BIN) init --parseDependency --parseInternal

## Run DB migrations locally
migrate:
	docker run --rm \
		--network host \
		-v $$PWD/db/migrations:/migrations \
		ghcr.io/pressly/goose \
		-dir /migrations \
		postgres "postgres://admin:admin@localhost:5432/user-db?sslmode=disable" up

## Docker: Build and run containers
docker-up:
	docker-compose up --build

## Docker: Stop and remove containers
docker-down:
	docker-compose down

## Restart Docker containers
docker-restart: docker-down docker-up

## Clean compiled files and coverage
clean:
	rm -f $(APP_NAME) coverage.out

.PHONY: run build test coverage coverage-html fmt tidy lint swagger docker-up docker-down docker-restart clean setup install migrate
