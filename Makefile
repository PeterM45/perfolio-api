# Makefile for Perfolio API
.PHONY: setup dev build test lint migrate-up migrate-down migrate-create run clean docker-build docker-run

# Environment variables
ENV ?= development
GO_FILES = $(shell find . -name "*.go" -not -path "./vendor/*" -not -path "./test/*")

# Database credentials from environment or defaults
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= postgres
DB_PASSWORD ?= postgres
DB_NAME ?= perfolio
MIGRATION_PATH ?= ./scripts/migrations

# Development setup
setup:
	go mod download
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/air-verse/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	cp configs/config.example.yaml configs/config.yaml

# Run with hot reload
dev:
	air -c .air.toml

# Build the application
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/perfolio-api cmd/api/main.go

# Run the application
run:
	go run cmd/api/main.go

# Run tests
test:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	golangci-lint run

create-db:
	@echo "Creating database $(DB_NAME) if it doesn't exist..."
	@PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -tc "SELECT 1 FROM pg_database WHERE datname = '$(DB_NAME)'" | grep -q 1 || PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -c "CREATE DATABASE $(DB_NAME);"

# Update migrate-up to depend on create-db
migrate-up: create-db
	migrate -path $(MIGRATION_PATH) -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" up


# Database migrations - down one version
migrate-down:
	migrate -path $(MIGRATION_PATH) -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" down 1

# Create a new migration
migrate-create:
	@if [ -z "$(name)" ]; then echo "Usage: make migrate-create name=migration_name"; exit 1; fi
	migrate create -ext sql -dir $(MIGRATION_PATH) -seq $(name)

# Clean build artifacts
clean:
	rm -rf bin/ tmp/ coverage.out coverage.html

# Docker commands
docker-build:
	docker build -t perfolio-api .

docker-run:
	docker run -p 8080:8080 --env-file .env perfolio-api

# Docker Compose
docker-compose-up:
	docker-compose up -d

docker-compose-down:
	docker-compose down

# Start local development database
start-db:
	docker-compose up -d postgres

# Generate mock files for testing
generate-mocks:
	mockgen -destination=test/mocks/user_repository.go -package=mocks github.com/PeterM45/perfolio-api/internal/user/repository UserRepository
	mockgen -destination=test/mocks/user_service.go -package=mocks github.com/PeterM45/perfolio-api/internal/user/service UserService
	mockgen -destination=test/mocks/content_repository.go -package=mocks github.com/PeterM45/perfolio-api/internal/user/repository PostRepository,WidgetRepository
	mockgen -destination=test/mocks/content_service.go -package=mocks github.com/PeterM45/perfolio-api/internal/user/service PostService,WidgetService