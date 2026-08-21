APP_NAME=omnilogs-api
COMPOSE_CMD=docker compose
COMPOSE_FILE=..\\compose.yml

.PHONY: build test swagger swagger-check api worker migrate seed

build:
	go build ./...

test:
	go test ./...

api:
	go run ./cmd/api

worker:
	go run ./cmd/worker

migrate:
	go run ./cmd/migrate

seed:
	go run ./cmd/seed

dev:
	$(COMPOSE_CMD) -f $(COMPOSE_FILE) up --build

dev-down:
	$(COMPOSE_CMD) -f $(COMPOSE_FILE) down

swagger:
	go run ./cmd/openapi sync

swagger-check:
	go run ./cmd/openapi validate
