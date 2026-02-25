.PHONY: all build test lint format clean

APP_NAME := pingo
BUILD_DIR := bin

all: format lint test build

build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/ddns

test:
	@echo "Running tests..."
	@go test -v -race -cover ./...

lint:
	@echo "Running golangci-lint..."
	@golangci-lint run ./...

format:
	@echo "Formatting code..."
	@go fmt ./...
	@go run golang.org/x/tools/cmd/goimports@latest -w .

clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)
