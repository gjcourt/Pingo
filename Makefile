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

# Container image variables
REGISTRY ?= ghcr.io
REGISTRY_USER ?= gjcourt
IMAGE_TAG ?= $(shell date +%Y-%m-%d)
PLATFORM ?= linux/amd64,linux/arm64

.PHONY: image
image:
	REGISTRY=$(REGISTRY) REGISTRY_USER=$(REGISTRY_USER) IMAGE_TAG=$(IMAGE_TAG) PLATFORM=$(PLATFORM) ./scripts/build_and_push_image.sh

.PHONY: list-images
list-images:
	@echo "Fetching images for $(REGISTRY)/$(REGISTRY_USER)/$(APP_NAME)..."
	@gh api \
		-H "Accept: application/vnd.github+json" \
		-H "X-GitHub-Api-Version: 2022-11-28" \
		/users/$(REGISTRY_USER)/packages/container/$(APP_NAME)/versions \
		--jq '.[].metadata.container.tags[]' | sort -r
