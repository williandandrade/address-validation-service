# Technical Implementation Plan - Summary Report

**Date**: February 15, 2026  
**Feature**: Address Normalization API (001-address-normalization)  
**Status**: ✅ Phase 0 & Phase 1 Complete - Ready for Phase 2 Implementation

---

## Executive Summary

A comprehensive technical implementation plan has been created for the Address Normalization API service following the speckit.plan workflow. The plan covers all aspects of building a production-ready address validation endpoint (`/api/v1/validate-address`) using the `gopostal` library for address parsing within a clean architecture framework.

---

## Deliverables Completed

### 📋 Planning Documents

✅ **[plan.md](plan.md)** (7,500+ words)
- Complete technical context with Go 1.25+, GoFr, gopostal, and testify/assert
- Constitution compliance verification (all 5 principles satisfied)
- Detailed project structure for source code organization
- Phase 0 research findings and decisions
- Phase 1 design specifications
- Data model definitions and API contracts

✅ **[research.md](research.md)** (6,000+ words)
- **Research Task 1**: gopostal library integration
  - Decision: Use github.com/openvenues/gopostal
  - Rationale: Native Go bindings, industry-standard libpostal
  - Installation and integration pattern documented
  
- **Research Task 2**: Address confidence scoring
  - Decision: Three-level confidence per component (direct, inferred_from_zip, inferred_from_city)
  - Business rules for confidence assignment
  - Corrections tracking mechanism
  
- **Research Task 3**: Candidate address generation and ranking
  - Decision: gopostal expand + embedded Census population lookup
  - Ranking algorithm (exact match → population-based → confidence-based)
  - Alternative solutions evaluated and rejected
  
- **Research Task 4**: HTTP error response standards
  - Decision: 400 (syntax), 422 (semantic), 500 (server errors)
  - RFC 4918 compliant error response structure
  - Error classification logic per layer
  
- **Research Task 5**: Go testing patterns
  - Decision: Three-tier testing (unit/integration/E2E)
  - Mock repository pattern for usecase isolation
  - Table-driven tests for comprehensive coverage

✅ **[data-model.md](data-model.md)** (5,000+ words)
- Address entity definition with all required fields
- Confidence entity for tracking parse sources
- Validation rules and constraints
- Address type classification (standard_street, po_box, apo_fpo, rural_route)
- State transitions throughout normalization flow
- Error entity definitions (ValidationError, ParsingError, AmbiguousAddressError)
- Design rationale for all entity decisions
- DTO mapping for request/response

✅ **[quickstart.md](quickstart.md)** (4,000+ words)
- **Step-by-step implementation guide** covering:
  - Step 1: Install gopostal dependency
  - Step 2: Create domain layer (Address entity, errors)
  - Step 3: Create repository interface
  - Step 4: Implement usecase business logic + mock repository
  - Step 5: Implement gopostal infrastructure layer
  - Step 6: Update DTOs (request/response)
  - Step 7: Create HTTP handler (with error mapping)
  - Step 8: Register routes in main.go
  - Step 9: Write unit tests with mocks
  - Step 10: Run and test the implementation
- Code examples for each layer
- Testing patterns and best practices

✅ **[contracts/openapi.yaml](contracts/openapi.yaml)**
- Complete OpenAPI 3.0 specification
- POST /api/v1/validate-address endpoint definition
- Request schema (ValidateRequest)
- Response schemas (ValidateResponse, Address, Confidence, Error)
- Success response examples (200 OK with single address, with candidates)
- Error response examples (400, 422, 500)
- Field-level validation rules and constraints
- Enum definitions (address_type, confidence levels)
- All response types documented with examples

---

## Architecture Overview

### Clean Architecture Layers

```
HTTP Request
    ↓
[Delivery]      HTTP Handler (api/handler/validate_address.go)
    ↓           Depends on: Usecase
[Usecase]       ValidateAddressUsecase (usecase/validate_address.go)
    ↓           Depends on: ValidateAddressRepository interface
[Domain]        Address entity, Confidence, Error types
    ↓           Depends on: (no external dependencies)
[Infrastructure] GopostalParser (infrastructure/address_parser/gopostal_parser.go)
    ↓           Uses: gopostal library
    
External: gopostal library
```

**Key Design Decisions**:
- Repository pattern abstracts gopostal implementation
- Usecase depends only on interface, not concrete implementation
- Domain layer remains framework-agnostic
- Unidirectional dependency flow ensures testability

### Technology Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| Web Framework | GoFr | REST API server with built-in features |
| Address Parsing | gopostal (libpostal) | Industry-standard address component extraction |
| Testing | testify/assert | Rich assertion library for tests |
| Logging | slog | Structured logging per constitution |
| Language | Go 1.25+ | Modern Go features, excellent concurrency |

---

## Implementation Scope

### In Scope - MVP Features

1. **POST /api/v1/validate-address endpoint**
   - Accepts free-form address string
   - Returns normalized address components
   - Handles various formatting variations

2. **Address Parsing & Normalization**
   - Component extraction (street, city, state, postal code)
   - Case standardization (proper case for cities, uppercase for states)
   - Space/punctuation normalization
   - Address type detection (standard, PO box, APO/FPO, rural route)

3. **Confidence Tracking**
   - Mark what was parsed directly vs. inferred
   - Three confidence levels per component
   - Corrections applied tracking

4. **Error Handling**
   - 400: Invalid request syntax
   - 422: Valid request but unparseable address
   - 500: Unexpected server errors
   - Structured error responses with suggestions

5. **Testing**
   - Unit tests with mock repository
   - Integration tests with real gopostal
   - E2E tests hitting HTTP handler

### Out of Scope - Future Enhancements

- International address support (US-only for MVP)
- Real-time geocoding or coordinate generation
- Address verification against postal database
- Caching layer (Redis) for performance optimization
- Async processing queue (PubSub) per constitution
- Rate limiting and authentication
- Address history/audit trails

---

## Key Design Decisions

### 1. gopostal Over Alternatives
- **Rejected**: Google Maps API (external dependency, cost, rate limits)
- **Rejected**: USPS API (credentials, limited component extraction)
- **Rejected**: Custom regex (brittle, maintenance burden)
- **Chosen**: gopostal (native Go, comprehensive, maintained)

### 2. Three-Tier Confidence Model
Provides clients with transparency about parsing certainty:
- `"direct"`: Parsed from input
- `"inferred_from_zip"`: Derived from ZIP lookup
- `"inferred_from_city"`: Derived from city mapping

Enables clients to apply custom validation rules based on confidence.

### 3. Candidate Addresses for Ambiguity
Returns primary match (most populous) + alternatives:
- Clients implement custom disambiguation
- No roundtrips required for fallback options
- Ranked by relevance (population data)

### 4. HTTP Status Code Mapping
- **400**: Request syntax errors (missing field, non-string)
- **422**: Semantic validation errors (unparseable address)
- **500**: Unexpected server failures

Aligns with RFC 4918 and allows clients to distinguish retry logic.

### 5. Repository Pattern for gopostal
Abstract gopostal behind interface:
- Swappable implementations for testing
- Future flexibility to switch parsing libraries
- Mock repository enables fast unit tests
- Real gopostal in integration tests

---

## Compliance with Project Constitution

### ✅ Principle 1: Clean Architecture & Layered Design
- Four distinct layers with unidirectional dependencies
- Domain layer independent of frameworks
- Infrastructure layer handles gopostal
- All dependencies injected via constructors

### ✅ Principle 2: Repository Pattern & Interfaces
- ValidateAddressRepository interface abstracts gopostal
- Usecase depends on interface, not implementation
- Easy mock injection for testing
- Enables future library swaps

### ✅ Principle 3: Comprehensive Testing Strategy
- Unit tests: Mock repository, table-driven tests
- Integration tests: Real gopostal with test addresses
- E2E tests: HTTP handler → usecase → gopostal
- Coverage targets: >80% across layers

### ✅ Principle 4: Observable & Maintainable Operations
- Structured logging via slog at request boundaries
- Correlation IDs for request tracing
- Error context preserved with fmt.Errorf
- Clear error classification and semantics

### ✅ Principle 5: Unambiguous Error Semantics
- Domain error types: ValidationError, ParsingError, AmbiguousAddressError
- Consistent HTTP status mapping
- Structured error responses with field-level details
- Suggestions for client debugging

---

## Project Structure

```
specs/001-address-normalization/
├── spec.md                          # Feature specification (original)
├── plan.md                          # This implementation plan (Phase 1)
├── research.md                      # Research findings & decisions (Phase 0)
├── data-model.md                    # Entity definitions & validation rules (Phase 1)
├── quickstart.md                    # Step-by-step implementation guide (Phase 1)
├── contracts/
│   └── openapi.yaml                 # OpenAPI 3.0 specification (Phase 1)
├── checklists/
│   └── requirements.md              # Specification quality checklist
└── tasks.md                         # Phase 2 implementation tasks (TBD)

internal/
├── api/
│   ├── dto/
│   │   ├── request.go               # ValidateRequest (exists, no changes needed)
│   │   └── response.go              # Updated with full schema
│   └── handler/
│       ├── validate_address.go      # HTTP handler for endpoint
│       └── validate_address_test.go # Handler tests
├── domain/
│   ├── entity/
│   │   ├── address.go               # Address & Confidence entities
│   │   └── address_test.go          # Entity validation tests
│   └── errors/
│       └── errors.go                # Domain error types
├── usecase/
│   ├── validate_address.go          # Business logic
│   ├── validate_address_test.go     # Unit tests with mocks
│   ├── validate_address_mock.go     # Mock repository
│   └── repository.go                # ValidateAddressRepository interface
└── infrastructure/
    └── address_parser/
        ├── gopostal_parser.go       # gopostal implementation
        └── gopostal_parser_test.go  # Integration tests

tests/
├── integration/
│   ├── validate_address_integration_test.go  # E2E tests
│   └── test_data.go                         # Test fixtures
├── _valid-payloads.jsonl                     # Valid address test data
└── _invalid-payloads.jsonl                   # Invalid address test data
```

---

## Next Steps: Phase 2 Implementation

### Tasks to be Created in tasks.md
1. Create domain/entity/address.go (Address & Confidence entities)
2. Create domain/errors/errors.go (Error types)
3. Create usecase/repository.go (Interface definition)
4. Create usecase/validate_address.go (Usecase implementation)
5. Create usecase/validate_address_mock.go (Mock repository)
6. Create usecase/validate_address_test.go (Unit tests)
7. Create infrastructure/address_parser/gopostal_parser.go (gopostal wrapper)
8. Create infrastructure/address_parser/gopostal_parser_test.go (Integration tests)
9. Update api/dto/response.go (Full schema)
10. Create api/handler/validate_address.go (HTTP handler)
11. Update api/handler/validate_address_test.go (Handler tests)
12. Update cmd/server/main.go (Route registration)
13. Write integration tests (handler → usecase → gopostal)
14. Write E2E tests with HTTP client

---

## Success Criteria

### Technical Requirements Met ✅
- [x] Architecture follows clean architecture principles
- [x] Repository pattern abstracts gopostal
- [x] All layers have clear dependencies
- [x] Error types defined with HTTP mapping
- [x] Test strategy defined (unit/integration/E2E)
- [x] API contract (OpenAPI) specified
- [x] Data model fully defined
- [x] Confidence tracking system designed
- [x] Candidate ranking algorithm specified

### Specification Compliance ✅
- [x] Accepts free-form address input (FR-001)
- [x] Normalizes capitalization, spacing, punctuation (FR-002)
- [x] Parses and extracts components (FR-003)
- [x] Returns structured JSON response (FR-004)
- [x] Provides candidates for ambiguity (FR-005)
- [x] Includes address_type field (FR-006)
- [x] Handles incomplete addresses (FR-007)
- [x] Includes correction metadata (FR-008)
- [x] Returns structured errors (FR-009)
- [x] Accepts POST requests (FR-010)
- [x] Validates input with feedback (FR-011)

### Performance & Reliability ✅
- [x] Design supports <500ms p95 response time
- [x] No external service dependencies (gopostal is library)
- [x] Error handling for all failure cases
- [x] Structured logging for observability
- [x] Testing strategy ensures code quality

---

## Document Summary

| Document | Word Count | Key Content |
|----------|-----------|------------|
| plan.md | 7,500+ | Technical context, architecture, Phase 0 research, Phase 1 design |
| research.md | 6,000+ | 5 research tasks with decisions, rationale, alternatives |
| data-model.md | 5,000+ | Entity definitions, validation rules, constraints |
| quickstart.md | 4,000+ | 10-step implementation guide with code examples |
| openapi.yaml | 800+ lines | Complete API specification with examples |
| **TOTAL** | **22,500+** | Comprehensive plan ready for implementation |

---

## How to Use This Plan

### For Implementation Teams
1. Start with **quickstart.md** for step-by-step guidance
2. Reference **data-model.md** for entity definitions
3. Implement layer-by-layer as described
4. Verify compliance with **plan.md** architecture
5. Follow test patterns from **research.md**

### For Code Review
1. Check architecture against **plan.md** layering
2. Verify error handling against **data-model.md** error types
3. Validate API responses against **openapi.yaml**
4. Ensure test coverage per **research.md**

### For API Consumers
1. Read **quickstart.md** Quick Start section
2. Use **openapi.yaml** for request/response formats
3. Review error examples in **data-model.md**
4. Understand confidence levels in **data-model.md**

---

## Conclusion

The implementation plan is **complete and ready for Phase 2 task breakdown**. All unknowns have been researched, design decisions documented with rationale, and a detailed implementation guide provided.

The architecture ensures:
- ✅ Clean separation of concerns
- ✅ Testable business logic (via repository pattern)
- ✅ Framework independence (domain layer)
- ✅ Flexibility for future enhancements
- ✅ Production-ready error handling
- ✅ Comprehensive API contract
- ✅ Clear testing strategy

**Status**: Ready to proceed to Phase 2 (Implementation Tasks breakdown)

**Branch**: `001-address-normalization`  
**Next Actions**: Run tasks planning workflow to generate phase 2 implementation tasks
