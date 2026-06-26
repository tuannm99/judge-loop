.PHONY: help infra infra-down dev dev-down migrate api-server judge-worker local-agent ui ui-build ui-install build test coverage lint fmt mocks mocks-gen

GO := go
GO_CACHE ?= /tmp/judgeloop-gocache
GOLANGCI_LINT_CACHE ?= /tmp/judgeloop-golangci-lint-cache
COVERAGE_MIN ?= 90
DATABASE_URL       ?= postgres://judgeloop:judgeloop@localhost:5432/judgeloop?sslmode=disable
COMPOSE_INFRA_FILE  = deploy/compose/compose-infras.yml
COMPOSE_DEV_FILE    = deploy/compose/compose-dev.yml

LINT_PKGS = $(shell GOCACHE=$(GO_CACHE) $(GO) list -f '{{.Dir}}' ./... | \
	sed '\#/mocks$$#d; \#/node_modules/#d; s#$(CURDIR)#.#')
TEST_PKGS = $(shell GOCACHE=$(GO_CACHE) $(GO) list -f '{{.Dir}}' ./... | \
	sed '\#/node_modules/#d; s#$(CURDIR)#.#')
COVER_PKGS = $(shell GOCACHE=$(GO_CACHE) $(GO) list ./internal/... | \
	sed '\#/mocks$$#d; \#/infrastructure/postgres$$#d; \#/node_modules/#d')

COLOR_RESET = \033[0m
COLOR_GREEN = \033[32m

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

infra: ## Start PostgreSQL via docker-compose
	docker compose -f $(COMPOSE_INFRA_FILE) up -d

infra-down: ## Stop and remove infrastructure containers
	docker compose -f $(COMPOSE_INFRA_FILE) down

dev: ## Start hot-reload app services (api, worker, local-agent, ui)
	docker compose -f $(COMPOSE_DEV_FILE) up --build

dev-down: ## Stop hot-reload app services
	docker compose -f $(COMPOSE_DEV_FILE) down

migrate: ## Run database migrations
	DATABASE_URL=$(DATABASE_URL) $(GO) run ./cmd/migrate

api-server: ## Start the API server
	DATABASE_URL=$(DATABASE_URL) $(GO) run ./cmd/api-server

judge-worker: ## Start the judge worker
	DATABASE_URL=$(DATABASE_URL) $(GO) run ./cmd/judge-worker

local-agent: ## Start the local agent
	$(GO) run ./cmd/local-agent

ui-install: ## Install UI dependencies
	cd ui && npm install

ui: ## Start the UI dev server (Vite)
	cd ui && npm run dev

ui-build: ## Build the UI for production
	cd ui && npm run build

build: ## Build all binaries into ./bin
	$(GO) build -o bin/migrate      ./cmd/migrate
	$(GO) build -o bin/api-server   ./cmd/api-server
	$(GO) build -o bin/judge-worker ./cmd/judge-worker
	$(GO) build -o bin/local-agent  ./cmd/local-agent

test: ## Run unit tests
	@echo "$(COLOR_GREEN)Running tests...$(COLOR_RESET)"
	@GOCACHE=$(GO_CACHE) $(GO) test $(TEST_PKGS) -cover -v

coverage: ## Check Go coverage threshold
	@echo "$(COLOR_GREEN)Checking test coverage...$(COLOR_RESET)"
	@GOCACHE=$(GO_CACHE) $(GO) test $(COVER_PKGS) \
		-covermode=atomic -coverprofile=/tmp/judgeloop-coverage.out
	@coverage=$$(GOCACHE=$(GO_CACHE) $(GO) tool cover \
		-func=/tmp/judgeloop-coverage.out | awk '/^total:/ {gsub("%", "", $$3); print $$3}'); \
	echo "coverage: $${coverage}% (minimum: $(COVERAGE_MIN)%)"; \
	awk -v coverage="$$coverage" -v minimum="$(COVERAGE_MIN)" \
		'BEGIN { if (coverage + 0 < minimum + 0) exit 1 }'

lint: ## Run golangci-lint
	@echo "$(COLOR_GREEN)Running linter...$(COLOR_RESET)"
	@GOCACHE=$(GO_CACHE) GOLANGCI_LINT_CACHE=$(GOLANGCI_LINT_CACHE) \
		$(GO) tool golangci-lint run --timeout=5m $(LINT_PKGS)

fmt: ## Format Go code
	@echo "$(COLOR_GREEN)Formatting Go code...$(COLOR_RESET)"
	@$(GO) tool golines -w --max-len=120 .
	@$(GO) tool gofumpt -w .

mocks-gen: mocks

mocks: ## Generate mocks
	@echo "$(COLOR_GREEN)Generating mocks via mockery...$(COLOR_RESET)"
	@GOCACHE=$(GO_CACHE) $(GO) tool mockery
