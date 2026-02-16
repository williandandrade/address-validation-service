# Implementation Plan: Address Normalization API

**Branch**: `001-address-normalization` | **Date**: February 15, 2026 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-address-normalization/spec.md`

## Summary

This implementation delivers a REST API endpoint (`/api/v1/validate-address`) that normalizes US addresses from free-form input into structured, validated components. The service uses the `gopostal` library (libpostal Go bindings) to parse address components and extract street, city, state, and postal code elements. Responses include normalization metadata, candidate addresses for ambiguous inputs, and structured error feedback for invalid addresses.

## Technical Context

**Language/Version**: Go 1.25+  
**Primary Dependencies**: 
- GoFr (web framework)
- gopostal (libpostal Go bindings for address parsing)
- testify/assert (testing assertions)

**Storage**: N/A (stateless validation service)  
**Testing**: Go standard `testing` package + testify/assert for assertions  
**Target Platform**: Linux server (Docker containerized)  
**Project Type**: Single service (web)  
**Performance Goals**: <500ms p95 response time for address normalization  
**Constraints**: 
- US addresses only (all 50 states + DC)
- Minimum 2 of 3 components required (street/city/state)
- 100% accuracy on valid addresses
- Consistent response schema across all inputs

**Scale/Scope**: Single service handling address validation requests with proper error classification and candidates for disambiguation

## Constitution Check

✅ **PASS** - Plan aligns with all constitution principles:

1. **Clean Architecture & Layered Design** ✅
   - Delivery layer: HTTP handler receiving POST requests to `/api/v1/validate-address`
   - Usecase layer: `ValidateAddressUsecase` encapsulating normalization business logic
   - Domain layer: `ValidateAddressRepository` interface for address parsing abstraction
   - Infrastructure layer: `gopostal`-based repository implementation
   - Unidirectional dependency chain: HTTP → Usecase → Repository interface → gopostal implementation

2. **Repository Pattern & Interface-Based Dependencies** ✅
   - Address parsing abstracted behind `ValidateAddressRepository` interface
   - `ValidateAddressUsecase` depends only on interface, not concrete implementation
   - Easy to mock for testing and swap implementations (e.g., for different parsing libraries)
   - No address parsing logic in usecase layer

3. **Comprehensive Testing Strategy** ✅
   - Unit tests for usecase logic (validation rules, confidence tracking)
   - Unit tests for DTO validation and response construction
   - Integration tests for handler + usecase integration
   - Mock repository for isolated usecase testing
   - gopostal-backed repository tests with real address parsing

4. **Observable & Maintainable Operations** ✅
   - Structured logging via `slog` at request boundaries (handler layer)
   - Correlation IDs in logs for request tracing
   - Error context preserved through `fmt.Errorf` wrapping
   - Clear error classification (invalid, corrected, unverifiable, valid)

5. **Unambiguous Error Semantics** ✅
   - Domain error types: `ValidationError`, `ParsingError`, `AmbiguousAddressError`
   - HTTP status mapping: 400 for invalid input, 422 for unprocessable address, 500 for internal errors
   - Structured error responses with field-level feedback and suggestions

## Project Structure

### Documentation (this feature)

```text
specs/001-address-normalization/
├── spec.md                      # Feature specification
├── plan.md                      # This file
├── research.md                  # Phase 0 output (research decisions)
├── data-model.md                # Phase 1 output (entity definitions)
├── quickstart.md                # Phase 1 output (implementation guide)
├── contracts/                   # Phase 1 output (API contracts)
│   └── openapi.yaml             # OpenAPI 3.0 specification
├── checklists/
│   └── requirements.md          # Specification quality checklist
└── tasks.md                     # Phase 2 output (implementation tasks)
```

### Source Code (repository root)

```text
cmd/
└── server/
    └── main.go                  # Application entry point

internal/
├── api/
│   ├── dto/
│   │   ├── request.go           # ValidateRequest, request validation
│   │   └── response.go          # ValidateResponse, error response structures
│   └── handler/
│       ├── validate_address.go        # HTTP handler for /api/v1/validate-address
│       └── validate_address_test.go   # Handler tests
├── domain/
│   ├── entity/
│   │   ├── address.go           # Address entity, validation rules
│   │   └── address_test.go      # Entity tests
│   └── errors/
│       └── errors.go            # Domain error types
├── usecase/
│   ├── validate_address.go           # ValidateAddressUsecase implementation
│   ├── validate_address_test.go      # Usecase unit tests
│   ├── validate_address_mock.go      # Mock repository for testing
│   └── repository.go            # ValidateAddressRepository interface
└── infrastructure/
    └── address_parser/
        ├── gopostal_parser.go       # gopostal implementation of ValidateAddressRepository
        └── gopostal_parser_test.go  # Integration tests for gopostal parsing

tests/
├── integration/
│   ├── validate_address_integration_test.go  # E2E tests: handler + usecase + gopostal
│   └── test_data.go             # Test fixtures and helper functions
└── _valid-payloads.jsonl        # Valid address test data
└── _invalid-payloads.jsonl      # Invalid address test data
```

**Structure Decision**: Single service architecture with clean layering. The address parsing responsibility is isolated in the infrastructure layer (gopostal implementation), making the business logic testable and independent of the parsing library choice. The `/api/v1/validate-address` endpoint is the primary delivery mechanism, implemented as a handler in the delivery layer.

## Complexity Tracking

No constitution violations requiring justification. The design fully aligns with clean architecture principles and repository pattern constraints.

---

## Phase 0: Research & Decisions

### Research Task 1: gopostal library integration
- **Decision**: Use `github.com/openvenues/gopostal` for address parsing
- **Rationale**: Native Go bindings with good community support; implements libpostal library which is industry-standard for address parsing
- **Alternatives considered**: Google Maps API (external dependency, rate-limited, cost), USPS API (US-specific but less detailed component extraction)
- **Implementation approach**: Create `infrastructure/address_parser/gopostal_parser.go` wrapper implementing `ValidateAddressRepository` interface

### Research Task 2: Address confidence scoring
- **Decision**: Implement confidence tracking using three levels: "direct" (parsed from input), "inferred_from_zip" (derived from ZIP code lookup), "inferred_from_state" (derived from state + city)
- **Rationale**: Provides clients transparency about what was parsed vs. inferred; aligns with spec requirements for correction metadata
- **Implementation approach**: Add `Confidence` struct to domain Address entity with field-level confidence scores; corrections tracking via array of strings

### Research Task 3: Candidate address generation and ranking
- **Decision**: Use gopostal's expansion functionality to generate candidates; rank by US Census data popularity (largest cities first for ambiguous city names)
- **Rationale**: Aligns with spec requirement for "most populous match" as primary result
- **Alternatives considered**: Return only highest-confidence match (less client flexibility), return all candidates unsorted (no guidance for client choice)
- **Implementation approach**: Parse address with gopostal to get all possible interpretations; filter candidates to only valid US cities in target state; rank candidates by population

### Research Task 4: HTTP error response standards
- **Decision**: 
  - 400 Bad Request: Malformed input (missing address field, non-string, empty string)
  - 422 Unprocessable Entity: Valid input but unparseable address (no components found, less than 2 of 3 required components)
  - 500 Internal Server Error: Unexpected failures (gopostal crashes, infrastructure failures)
- **Rationale**: RFC 4918 standard for semantic validation failures; 400 for syntax, 422 for semantic
- **Implementation approach**: Create error classification logic in handler to map domain errors to HTTP codes

### Research Task 5: Go testing patterns
- **Decision**: Use table-driven tests for address parsing variations; mock repository pattern for usecase isolation; integration tests with real gopostal
- **Rationale**: Table-driven tests provide comprehensive coverage; mock repository enables fast unit tests; integration tests catch gopostal-specific issues
- **Implementation approach**: Table-driven tests in `validate_address_test.go`; mock repository in `validate_address_mock.go`; integration tests in `gopostal_parser_test.go`

---

## Phase 1: Design & API Contracts

### Data Model: Address Entity

```go
type Address struct {
    StreetAddress    string      `json:"street_address"`
    City             string      `json:"city"`
    State            string      `json:"state"`
    PostalCode       string      `json:"postal_code"`
    AddressType      string      `json:"address_type"`
    FormattedAddress string      `json:"formatted_address,omitempty"`
    Confidence       *Confidence `json:"confidence,omitempty"`
    CorrectionsApplied []string  `json:"corrections_applied,omitempty"`
}

type Confidence struct {
    StateConfidence  string `json:"state_confidence"`
    CityConfidence   string `json:"city_confidence"`
    PostalConfidence string `json:"postal_confidence"`
}
```

### API Specification: POST /api/v1/validate-address

**Request Schema**:
```json
{
  "address": "string (required)"
}
```

**Success Response (200 OK)**:
```json
{
  "success": true,
  "address": {
    "street_address": "123 Main St",
    "city": "New York",
    "state": "NY",
    "postal_code": "10001",
    "address_type": "standard_street",
    "formatted_address": "123 Main St, New York, NY 10001"
  },
  "confidence": {
    "state_confidence": "direct",
    "city_confidence": "direct",
    "postal_confidence": "direct"
  },
  "corrections_applied": ["Standardized capitalization"]
}
```

**Error Response (400/422)**:
```json
{
  "success": false,
  "errors": [
    {
      "field": "address",
      "reason": "Empty or missing required field",
      "suggestion": "Provide a valid US address"
    }
  ]
}
```

### Implementation Structure

**Layer 1 - Delivery (HTTP)**:
- Handler: `internal/api/handler/validate_address.go`
- Request/Response DTOs: `internal/api/dto/request.go`, `response.go`

**Layer 2 - Usecase (Business Logic)**:
- Usecase: `internal/usecase/validate_address.go`
- Repository Interface: `internal/usecase/repository.go`

**Layer 3 - Domain (Entities & Rules)**:
- Address Entity: `internal/domain/entity/address.go`
- Error Types: `internal/domain/errors/errors.go`

**Layer 4 - Infrastructure (External Dependencies)**:
- gopostal Parser: `internal/infrastructure/address_parser/gopostal_parser.go`

---

## Constitution Compliance Verification

✅ **RE-VERIFIED** - Implementation plan maintains full constitution compliance after Phase 1 design.

All design decisions support:
- Clean architecture with four distinct layers and unidirectional dependencies
- Repository pattern enables gopostal as interchangeable implementation
- Comprehensive testing strategy with unit, integration, and E2E test coverage
- Error semantics map to HTTP status codes per constitution
- Structured logging via slog captures request context and errors

Plan ready for Phase 2 task breakdown.
