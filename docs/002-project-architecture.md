# Address Validation Service Architecture

## Overview

This project follows clean architecture principles with clear separation of concerns across four main layers:

## Architecture Layers

```
┌─────────────────────────────────────┐
│   Delivery Layer (API/HTTP)         │  Request handling, validation, response formatting
├─────────────────────────────────────┤
│   Usecase/Application Layer         │  Business logic orchestration
├─────────────────────────────────────┤
│   Domain Layer                      │  Business entities, interfaces, domain errors
├─────────────────────────────────────┤
│   Infrastructure Layer              │  Database, cache, external service implementations
└─────────────────────────────────────┘
```

## Layer Responsibilities

### Delivery Layer (`internal/api/`)

- Parse HTTP requests
- Validate input data
- Call use cases
- Format and return responses
- Handle HTTP-specific concerns (headers, status codes)
- **No business logic**

### Usecase Layer (`internal/usecase/`)

- Orchestrate business logic
- Implement application workflows
- Call repositories and domain services
- Independent of framework choices

### Domain Layer (`internal/domain/`)

- Define business entities and value objects
- Declare repository interfaces (not implementations)
- Implement domain logic and validation
- Define domain-specific errors
- Framework-agnostic and testable

### Infrastructure Layer (`internal/infra/`)

- Implement repository interfaces from the domain layer
- Handle database queries and transactions
- Manage caching and message queues
- Integrate with external services

## Dependencies Flow

- **Unidirectional**: Inner layers never depend on outer layers
- **Interface-based**: Domain layer defines interfaces; infrastructure implements them
- **Dependency Injection**: Dependencies passed via constructor functions

## Key Design Patterns

### Repository Pattern

Abstracts database operations behind interfaces defined in the domain layer and implemented in the infrastructure layer.

### Constructor Pattern

Dependencies are explicitly passed through constructor functions (e.g., `NewDomainUseCase`).

### Middleware Pattern
HTTP middleware for cross-cutting concerns like logging, recovery, and authentication.

## Testing Strategy

- **Unit Tests**: Test business logic in isolation with mocks
- **Integration Tests**: Test repository implementations with test database
- **E2E Tests**: Test complete HTTP workflows
- **Test Coverage**: Aim for high coverage in domain and usecase layers, with critical paths tested in delivery and infrastructure layers
- **Tools**: Use `testing` package for unit tests, `httptest` for HTTP tests, and `testcontainers` for integration tests with infrastructure dependencies
- **Assertions**: Use `testify/assert` for more readable assertions in tests

## Error Handling

- Errors are wrapped with context using `fmt.Errorf`
- Domain-specific errors are defined in the domain layer
- Errors are logged once at the application boundary (handlers)

## Configuration

- Environment variables for deployment-specific settings
- Type-safe configuration structs
- Configuration loaded at startup

## Observability

- Structured logging using `slog`
- Request logging middleware
- Database connection health checks

## Tooling

- Use Go 1.25 (latest minor version possible)
- Use `go mod` for dependency management
- Use `golangci-lint` for linting
- Use `go test` for testing with coverage reporting
- Use Makefile for common tasks (build, test, lint)
- Use Containerization (Docker) for consistent development and deployment environments
