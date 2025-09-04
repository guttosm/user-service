# ───────────────────────────────────────────────────────────────────────────────
# Project settings
# ───────────────────────────────────────────────────────────────────────────────
SHELL := /usr/bin/env bash
.DEFAULT_GOAL := help

APP_NAME            ?= user-service
GO                  ?= go
GOLANGCI_LINT       ?= golangci-lint
SWAG                ?= github.com/swaggo/swag/cmd/swag
SWAG_BIN            := $(shell $(GO) env GOPATH)/bin/swag

# Exclude packages (regex passed to grep -Ev). Adjust as needed.
# Example excludes: models & DTOs
EXCLUDE_PKGS_REGEX ?= internal/domain/model|internal/dto

# All packages except excluded ones
PKGS := $(shell $(GO) list ./... | grep -Ev '$(EXCLUDE_PKGS_REGEX)')

# Common test flags
TEST_FLAGS       ?= -race -shuffle=on -count=1
COVER_PROFILE    ?= coverage.out
COVER_MODE       ?= atomic

# ───────────────────────────────────────────────────────────────────────────────
# Helpers
# ───────────────────────────────────────────────────────────────────────────────
define print-target
	@printf "  \033[36m%-20s\033[0m %s\n" "$(1)" "$(2)"
endef

help: ## Show this help
	@echo "Make targets for $(APP_NAME):"
	@echo
	$(call print-target,setup,             Create local folders for bind mounts)
	$(call print-target,install,           Download deps & install tools)
	$(call print-target,run,               Run app locally (no Docker))
	$(call print-target,build,             Build binary)
	$(call print-target,fmt,               go fmt)
	$(call print-target,tidy,              go mod tidy)
	$(call print-target,lint,              Run golangci-lint)
	$(call print-target,swagger,           Generate Swagger docs)
	$(call print-target,test,              Run ALL tests (unit + integration))
	$(call print-target,test-unit,         Run unit tests only, exclude $(EXCLUDE_PKGS_REGEX))
	$(call print-target,test-integration,  Run integration tests (tag: integration))
	$(call print-target,coverage,          Print coverage summary)
	$(call print-target,coverage-html,     Open HTML coverage report)
	$(call print-target,migrate,           Run DB migrations locally via Goose container)
	$(call print-target,docker-up,         Docker Compose up (build)
	$(call print-target,docker-down,       Docker Compose down)
	$(call print-target,docker-restart,    Restart Compose stack)
	$(call print-target,clean,             Clean artifacts)
	$(call print-target,vet,                Run go vet static analysis)
	$(call print-target,staticcheck,        Run staticcheck static analysis)
	$(call print-target,analyze,            Run all static analysis tools)
	@echo

# ───────────────────────────────────────────────────────────────────────────────
# Setup / Dev
# ───────────────────────────────────────────────────────────────────────────────
setup: ## Create local folders for bind mounts
	mkdir -p postgres-data

install: ## Download deps & install tools
	$(GO) mod download
	$(GO) install $(SWAG)@latest

run: ## Run locally without Docker
	@echo "Running $(APP_NAME)..."
	$(GO) run ./cmd/main.go

build: ## Build Go binary
	@echo "Building $(APP_NAME)..."
	$(GO) build -o $(APP_NAME) ./cmd/main.go

fmt: ## Format code
	$(GO) fmt ./...

tidy: ## Tidy go.mod/go.sum
	$(GO) mod tidy

lint: ## Run static analysis (golangci-lint)
	$(GOLANGCI_LINT) run

swagger: ## Generate Swagger docs
	$(GO) install $(SWAG)@latest
	$(SWAG_BIN) init --parseDependency --parseInternal

# ───────────────────────────────────────────────────────────────────────────────
# Testing
# ───────────────────────────────────────────────────────────────────────────────

# All tests (unit + integration). Integration tests must be tagged with //go:build integration.
test: ## Run ALL tests (unit + integration)
	@echo "→ Unit tests"
	$(GO) test $(PKGS) $(TEST_FLAGS) -coverprofile=$(COVER_PROFILE) -covermode=$(COVER_MODE)
	@echo "→ Integration tests"
	$(GO) test -tags=integration ./... -count=1

test-unit: ## Run ONLY unit tests, exclude $(EXCLUDE_PKGS_REGEX)
	@echo "→ Unit tests (excluding: $(EXCLUDE_PKGS_REGEX))"
	$(GO) test $(PKGS) $(TEST_FLAGS) -coverprofile=$(COVER_PROFILE) -covermode=$(COVER_MODE)

test-integration: ## Run ONLY integration tests (tag: integration)
	@echo "→ Integration tests"
	$(GO) test -tags=integration ./... -count=1

coverage: ## Show coverage summary
	$(GO) tool cover -func=$(COVER_PROFILE)

coverage-html: ## Open HTML coverage report
	$(GO) tool cover -html=$(COVER_PROFILE)

# ───────────────────────────────────────────────────────────────────────────────
# Migrations & Docker
# ───────────────────────────────────────────────────────────────────────────────
migrate: ## Run DB migrations locally (requires local Postgres up)
	docker run --rm \
		--network host \
		-v $$PWD/db/migrations:/migrations \
		ghcr.io/pressly/goose \
		-dir /migrations \
		postgres "postgres://admin:admin@localhost:5432/user-db?sslmode=disable" up

docker-up: ## Compose up (build)
	docker compose up --build -d

docker-down: ## Compose down
	docker compose down

docker-restart: docker-down docker-up ## Restart Compose stack

# ───────────────────────────────────────────────────────────────────────────────
# Housekeeping
# ───────────────────────────────────────────────────────────────────────────────
clean: ## Clean compiled files and coverage artifacts
	rm -f $(APP_NAME) $(COVER_PROFILE)

vet: ## Run go vet static analysis
	$(GO) vet ./...

staticcheck: ## Run staticcheck static analysis
	@which staticcheck > /dev/null || (echo "Installing staticcheck..." && $(GO) install honnef.co/go/tools/cmd/staticcheck@latest)
	staticcheck ./...

analyze: vet staticcheck lint ## Run all static analysis tools

.PHONY: help setup install run build fmt tidy lint swagger \
        test test-unit test-integration coverage coverage-html \
        migrate docker-up docker-down docker-restart clean vet staticcheck analyze
