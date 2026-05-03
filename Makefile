.PHONY: all build run test lint tidy migrate docker-build docker-up docker-down clean

APP_NAME=gopay
MAIN_FILE=cmd/api/main.go
BUILD_DIR=bin

all: tidy build

build:
	@echo "Building..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)

run:
	@go run $(MAIN_FILE) -config configs/config.dev.yaml

dev:
	@air -c .air.toml

test:
	@go test -v -race -coverprofile=coverage.out ./...

test-coverage:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

lint:
	@golangci-lint run ./...

tidy:
	@go mod tidy
	@go mod verify

fmt:
	@go fmt ./...

vet:
	@go vet ./...

wire:
	@cd cmd/api && wire

migrate:
	@echo "Database auto-migrated on startup"

docker-build:
	@docker-compose -f deployments/docker/docker-compose.yml build

docker-up:
	@docker-compose -f deployments/docker/docker-compose.yml up -d

docker-down:
	@docker-compose -f deployments/docker/docker-compose.yml down

docker-logs:
	@docker-compose -f deployments/docker/docker-compose.yml logs -f

clean:
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

.DEFAULT_GOAL := all
