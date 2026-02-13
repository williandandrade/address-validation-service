<!-- Sync Impact Report -->
<!-- Version: 1.0.0 (Initial)
   - Created initial constitution based on project thought process and architecture
   - Templates requiring updates: spec-template.md, plan-template.md, tasks-template.md
   - Status: All templates reviewed and compatible with initial principles
-->

# Address Validation Service Constitution

## Core Principles

### I. Clean Architecture & Layered Design
The service enforces strict separation of concerns across four layers: Delivery (HTTP), Usecase (business logic), Domain (entities and interfaces), and Infrastructure (implementations and external dependencies).

All code must respect unidirectional dependencies: outer layers depend on inner layers, never vice versa. The domain layer remains framework-agnostic and acts as the source of truth for business rules. This ensures testability, maintainability, and independence from framework choices.

### II. Asynchronous-First Processing with Cache-Backed Resilience
The service prioritizes resilience and scalability through asynchronous processing. Address validation requests are published to Google PubSub, processed by workers, and results cached in Redis before response to clients.

This approach decouples the API from external service failures, enables independent scaling of compute resources, and provides eventual consistency guarantees. Cache hits optimize response latency; timeouts gracefully degrade to explicit waiting indicators. External service failures do not block API availability.

### III. Repository Pattern & Interface-Based Dependencies
All infrastructure concerns (Redis, PubSub, external APIs) are abstracted behind domain-layer interfaces implemented in the infrastructure layer.

Dependencies are explicitly injected via constructor functions, making test doubles and mocks straightforward to provide. No infrastructure details leak into domain or usecase layers. This ensures domain logic remains portable and testable without external service calls.

### IV. Comprehensive Testing Strategy
Testing is tiered: unit tests verify domain and usecase logic in isolation; integration tests validate repository implementations and external service interactions; E2E tests confirm complete HTTP workflows.

Target high coverage in domain and usecase layers (>80%); critical paths in delivery and infrastructure layers must be tested. Use Go's standard `testing` package and `testify/assert` for assertions. Integration tests use `testcontainers` to isolate external dependencies. All tests must run deterministically offline.

### V. Observable & Maintainable Operations
Structured logging via `slog` is mandatory at request boundaries and critical operations. Every log entry must capture correlation IDs, operation context, and error details sufficient for debugging.

Configuration is type-safe via environment variables and loaded at startup. Health checks confirm cache and message queue availability. Error handling logs once at application boundaries; errors are wrapped with `fmt.Errorf` to preserve context across layers.

### VI. Unambiguous Error Semantics
Domain errors are explicitly defined (e.g., `invalid`, `corrected`, `unverifiable`, `valid`) and mapped consistently to HTTP status codes.

Partial addresses, typos, and non-existent addresses must fail gracefully with clear status indicators. External service failures (timeouts, downtimes) return `unverifiable` rather than causing 5xx responses. All error paths are logged and traced.

## Technical Constraints

- **Language & Runtime**: Go 1.25 (latest minor version)
- **Web Framework**: GoFr (https://gofr.dev)
- **Cache**: Redis (eventual consistency model)
- **Message Queue**: Google PubSub (exactly-once semantics preferred)
- **External Validation**: Google Address Validation API
- **Address Parsing**: libpostal (gopostal)
- **Tooling**: `go mod`, `golangci-lint`, Docker, Makefile

All dependencies must be vendored or locked in `go.mod`. No unstable or unmaintained libraries are permitted. Breaking dependency updates require a full test suite pass and documentation of migration steps.

## Development Workflow

1. **Code Style & Linting**: Run `golangci-lint` before commits; all warnings must be resolved or explicitly documented.
2. **Testing**: Every feature or bug fix requires unit tests; integration tests for new repository implementations or contract changes; E2E tests for new HTTP endpoints.
3. **Commits**: Use clear, imperative messages (e.g., `feat: add address validation endpoint`). Reference issue numbers where applicable.
4. **Pull Requests**: All PRs must pass linting, tests, and code review. One approval minimum before merge.

## Governance

This constitution is the source of truth for development practices. All code, configuration, and decisions must align with these principles.

**Amendment Process**: Material changes (new principles, removals, or reinterpretations) require explicit approval and version increment. Minor clarifications use PATCH versioning; new principles use MINOR; principle removals or redefinitions use MAJOR.

**Compliance Verification**: Code review must validate adherence to these principles. Deviations require documented justification in PRs and issues.

**Version Policy**: Constitution uses semantic versioning (MAJOR.MINOR.PATCH). Version changes and rationale are documented in the Sync Impact Report at the top of this file.

**Version**: 1.0.0 | **Ratified**: 2026-02-13 | **Last Amended**: 2026-02-13
