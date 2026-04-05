.PHONY: help infra infra-down migrate api-server judge-worker local-agent ui ui-build ui-install build test lint

DATABASE_URL       ?= postgres://judgeloop:judgeloop@localhost:5432/judgeloop?sslmode=disable
JUDGE_WORKER_FLAGS ?= -ui -ui-port=8081 -ui-path=/monitoring
REDIS_URL          ?= localhost:6379
COMPOSE_FILE        = deploy/compose/docker-compose.yml

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

infra: ## Start PostgreSQL and Redis via docker-compose
	docker compose -f $(COMPOSE_FILE) up -d

infra-down: ## Stop and remove infrastructure containers
	docker compose -f $(COMPOSE_FILE) down

migrate: ## Run database migrations
	DATABASE_URL=$(DATABASE_URL) go run ./cmd/migrate

api-server: ## Start the API server
	DATABASE_URL=$(DATABASE_URL) REDIS_URL=$(REDIS_URL) go run ./cmd/api-server

judge-worker: ## Start the judge worker
	DATABASE_URL=$(DATABASE_URL) REDIS_URL=$(REDIS_URL) go run ./cmd/judge-worker $(JUDGE_WORKER_FLAGS)

local-agent: ## Start the local agent
	go run ./cmd/local-agent

ui-install: ## Install UI dependencies
	cd ui && npm install

ui: ## Start the UI dev server (Vite)
	cd ui && npm run dev

ui-build: ## Build the UI for production
	cd ui && npm run build

build: ## Build all binaries into ./bin
	go build -o bin/migrate      ./cmd/migrate
	go build -o bin/api-server   ./cmd/api-server
	go build -o bin/judge-worker ./cmd/judge-worker
	go build -o bin/local-agent  ./cmd/local-agent

test: ## Run unit tests
	go test ./... -cover -v

lint: ## Run go vet
	go vet ./...
