.PHONY: build run test lint docker-build docker-build-gopostal docker-run clean check fmt

# Build variables
BINARY_NAME=address-validation-service
BUILD_DIR=./build
GO_FILES=$(shell find . -name '*.go' -not -path "./vendor/*")

# Build the application
build:
	@echo "Building..."
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server

# Run the application
run:
	@go run ./cmd/server

# Run tests
test:
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linter
lint:
	@golangci-lint run ./...

# Format code
fmt:
	@go fmt ./...

# Run all checks (lint + test)
check: lint test

# Build Docker image (regex parser — lightweight, works everywhere)
docker-build:
	@docker build -t $(BINARY_NAME):latest .

# Build Docker image with gopostal (needs 4GB+ Docker memory)
docker-build-gopostal:
	@docker build --build-arg PARSER=gopostal -t $(BINARY_NAME):latest .

# Run Docker container
docker-run:
	@docker run -p 8080:8080 $(BINARY_NAME):latest

# Clean build artifacts
clean:
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

# Download dependencies
deps:
	@go mod download
	@go mod tidy

# Install development tools
dev-tools:
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install go.uber.org/mock/mockgen@latest
