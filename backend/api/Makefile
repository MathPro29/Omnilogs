APP_NAME=omnilogs-api

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
	docker compose -f ..\docker-compose.yml up --build

swagger:
	go run ./cmd/openapi sync

swagger-check:
	go run ./cmd/openapi validate
