# Implementation Tasks: Address Normalization API

**Branch**: `001-address-normalization` | **Date**: February 15, 2026  
**Spec**: [spec.md](spec.md) | **Plan**: [plan.md](plan.md)  
**Status**: Phase 2 - Implementation Tasks Ready for Execution

---

## Overview

This document defines implementation tasks for the Address Normalization API feature (001-address-normalization). Tasks are organized by user story with independent testability for each phase. The MVP scope focuses on User Story 1 (P1), with Stories 2-3 (P2) included for complete feature delivery.

**Total Tasks**: 39 implementation tasks  
**Phases**: 5 (Setup, Foundational, US1, US2, US3, Polish)  
**Estimated Effort**: 2-3 weeks for full implementation with testing  
**MVP Scope**: Phase 1 (Setup) + Phase 2 (Foundational) + Phase 3 (User Story 1)

---

## Task Organization & Dependencies

### Phase Dependencies Graph

```
Setup (P0)
    ↓
Foundational (P0)
    ├→ User Story 1: Normalize Address (P1)
    ├→ User Story 2: Structured Response (P2)
    └→ User Story 3: Error Handling (P2)
        ↓
Polish & Cross-Cutting (P3)
```

### Independent Test Criteria per User Story

**User Story 1 (P1)**: Developer can submit raw address → receives normalized address
- Test: `curl -X POST /api/v1/validate-address -d '{"address":"123 main st new york ny 10001"}'`
- Expected: 200 OK with normalized address object

**User Story 2 (P2)**: Client receives consistent, well-defined response structure
- Test: Verify response contains street_address, city, state, postal_code, address_type fields
- Expected: All responses use same schema regardless of input format

**User Story 3 (P2)**: Client receives meaningful error messages for invalid addresses
- Test: Submit invalid address → get 400/422 with field-level errors and suggestions
- Expected: Clear explanation of what failed and how to fix it

### Parallelization Opportunities

**Phase 2 (Foundational)**: All tasks are independent and can run in parallel
- T001-T006: Create all domain/infrastructure code (no dependencies on each other)
- Estimated parallel time: ~1 day vs. 6 days sequential

**Phase 3 (User Story 1)**: Tasks can be parallelized after Phase 2
- T007-T012: Handler, tests, and integration tests can be written in parallel
- T013-T015: Must complete after T007 (handler) is done

**Phase 4 (User Story 2)**: Independent from User Story 1 implementation
- T016-T020: Response structure and confidence tracking

**Phase 5 (User Story 3)**: Dependent on handler (T007) being complete
- T021-T028: Error handling and validation

---

## Phase 1: Setup & Project Initialization

### Story Goal
Establish project dependencies, configuration, and base structure for clean architecture layers.

### Independent Test Criteria
- Go build succeeds with no errors
- All dependencies installed and verified in go.mod
- Project structure matches architecture definition
- Linting passes with zero warnings

### Implementation Tasks

- [X] T001 Install and verify gopostal dependency in go.mod

- [X] T002 [P] Create directory structure: internal/{domain/entity,domain/errors,usecase,infrastructure/address_parser,api/dto,api/handler}

- [X] T003 [P] Create internal/domain/errors/errors.go with ValidationError, ParsingError, AmbiguousAddressError types

- [X] T004 [P] Create internal/usecase/repository.go with ValidateAddressRepository interface definition

- [X] T005 Create internal/domain/entity/address.go with Address, Confidence entities and Validate() method

- [X] T006 Create internal/usecase/validate_address_mock.go with MockValidateAddressRepository for testing

---

## Phase 2: Foundational Infrastructure

### Story Goal
Implement core business logic layer and address parsing infrastructure independent of HTTP handling.

### Independent Test Criteria
- Usecase logic validates address components correctly
- Mock repository works for isolated testing
- gopostal parser extracts and normalizes address components
- Confidence scores assigned correctly to parsed components

### Tests (Optional)
- Run: `go test ./internal/usecase/... ./internal/infrastructure/...`
- Expected: All tests pass with >80% coverage

### Implementation Tasks

- [X] T007 [P] Create internal/usecase/validate_address.go with ValidateAddressUsecase struct and ValidateAddress() method

- [X] T008 [P] Implement ValidateAddressUsecase.assignConfidence() to set state/city/postal_code confidence levels

- [X] T009 [P] Create internal/infrastructure/address_parser/gopostal_parser.go implementing ValidateAddressRepository interface

- [X] T010 [P] Implement GopostalParser.ParseAddress() to extract street, city, state, postal code components using gopostal

- [X] T011 [P] Implement GopostalParser.detectAddressType() to classify address as standard_street/po_box/apo_fpo/rural_route

- [X] T012 Create internal/usecase/validate_address_test.go with table-driven unit tests using mock repository (test empty input, valid address, incomplete address)

---

## Phase 3: User Story 1 - Client Submits Unformatted Address for Normalization (P1)

### Story Goal
Enable clients to send addresses in any format and receive normalized, standardized address components.

### Story-Specific Requirements
- Accept addresses in various formats (lowercase, extra spaces, mixed punctuation)
- Standardize capitalization, spacing, punctuation
- Extract all required components (street, city, state, postal code)
- Return single normalized address as primary result

### Independent Test Criteria
- Submit "123 main st new york ny 10001" → receive normalized "123 Main St", "New York", "NY", "10001"
- Submit "123 Main St, New York, NY 10001" → receive same normalized form (idempotent)
- Submit "123Main St NewYork NY10001" → receive same normalized form (handles varied spacing)
- HTTP 200 OK with consistent response schema
- Can independently test without User Stories 2-3

### Tests
- Run: `go test ./internal/api/handler/... -run TestValidateAddress`
- Run: `curl -X POST http://localhost:8080/api/v1/validate-address -d '{"address":"123 main st new york ny 10001"}'`
- Expected: 200 OK with normalized address

### Implementation Tasks

- [X] T013 [US1] Update internal/api/dto/response.go with full ValidateResponse, AddressDTO, ConfidenceDTO, ErrorDTO schemas

- [X] T014 [US1] Create internal/api/handler/validate_address.go HTTP handler for POST /api/v1/validate-address

- [X] T015 [US1] Implement handler request validation: check address field exists, is string, not empty

- [X] T016 [US1] Implement handler response mapping: convert entity.Address to dto.AddressDTO with confidence and corrections

- [X] T017 [US1] Register POST /api/v1/validate-address route in cmd/server/main.go with handler initialization

- [X] T018 [US1] Create internal/api/handler/validate_address_test.go with handler unit tests (mock usecase, test 200 response)

- [X] T019 [US1] Create tests/integration/validate_address_integration_test.go with E2E tests (real gopostal, real HTTP requests)

- [X] T020 [US1] Add test cases for various address formats: lowercase, extra spaces, mixed punctuation, abbreviated state names

- [X] T021 [US1] Run full test suite: `go test ./...` with minimum 80% coverage requirement

---

## Phase 4: User Story 2 - Client Receives Structured Response (P2)

### Story Goal
Ensure API response has well-defined, consistent structure across all valid address inputs.

### Story-Specific Requirements
- Response always includes street_address, city, state, postal_code, address_type
- FormattedAddress generated as human-readable string
- Confidence metadata included for all components
- CorrectionsApplied array populated with normalization details

### Independent Test Criteria
- Response schema consistent regardless of input format variations
- All required fields present and populated
- Confidence object contains state_confidence, city_confidence, postal_confidence
- CorrectionsApplied documents all transformations applied (capitalization, spacing)

### Tests
- Run: `go test ./internal/api/handler/... -run TestStructuredResponse`
- Expected: All response fields match schema, no missing required fields

### Implementation Tasks

- [X] T022 [P] [US2] Implement Address.FormatAddress() to generate human-readable "Street, City, State ZIP" format

- [X] T023 [P] [US2] Add corrections tracking to GopostalParser: populate Address.CorrectionsApplied with normalization details

- [X] T024 [US2] Add test cases for response structure validation: verify all fields present, types correct, consistency across inputs

- [X] T025 [US2] Create tests/contract/address_response_schema_test.go to validate response against OpenAPI schema

- [X] T026 [US2] Implement response consistency tests: submit multiple address formats, verify normalized results are identical

- [X] T027 [US2] Document response examples in handler comments: success case with all fields populated

---

## Phase 5: User Story 3 - Client Handles Normalization Failures Gracefully (P2)

### Story Goal
Return clear, actionable error messages when address cannot be normalized.

### Story-Specific Requirements
- 400 Bad Request for missing/invalid request syntax
- 422 Unprocessable Entity for unparseable addresses
- Structured error responses with field, reason, value, suggestion
- Field-level error details enabling client debugging

### Independent Test Criteria
- Empty address string → 400 with "field required" reason
- "gibberish" (unparseable) → 422 with "could not extract components" reason
- Invalid state "XX" → 422 with state error + "Did you mean TX?" suggestion
- Invalid city in state → 422 with city error + alternative suggestion

### Tests
- Run: `go test ./internal/api/handler/... -run TestError`
- Submit invalid addresses, verify correct HTTP status and error structure

### Implementation Tasks

- [X] T028 [US3] Implement handler.handleUsecaseError() to map domain errors to HTTP status codes (400/422/500)

- [X] T029 [US3] Add validation error handling: missing/empty address field → 400 with ValidationError response

- [X] T030 [US3] Add parsing error handling: unparseable address → 422 with ParsingError details

- [X] T031 [US3] Implement error suggestion logic: suggest valid state codes, valid cities for given state

- [X] T032 [US3] Create internal/api/handler/error_mapping_test.go with tests for all error scenarios (empty, invalid, unparseable)

- [X] T033 [US3] Add integration tests for error responses: verify field, reason, value, suggestion structure

- [X] T034 [US3] Test error response consistency: ensure same error always produces same response format

- [X] T035 [US3] Document error response examples in openapi.yaml and handler code

---

## Phase 6: Polish & Cross-Cutting Concerns

### Story Goal
Complete implementation with logging, performance validation, and documentation.

### Independent Test Criteria
- Structured logs at request boundaries with correlation IDs
- Response time under 500ms p95 for valid addresses
- All code follows Go conventions and passes golangci-lint
- Documentation complete with examples

### Implementation Tasks

- [X] T036 [P] Add structured logging via slog in handler request/response boundaries with correlation IDs

- [X] T037 [P] Create tests/performance/address_validation_benchmark_test.go with benchmark for gopostal parsing

- [X] T038 [P] Run `golangci-lint run ./...` and fix all warnings (imports, formatting, unused code)

- [X] T039 Add README.md section documenting /api/v1/validate-address endpoint with curl examples

---

## Task Execution Guidelines

### Sequential Dependencies

Tasks must be completed in this order:
1. **Phase 1 (Setup)**: T001-T006 (can run in parallel with [P] marker)
2. **Phase 2 (Foundational)**: T007-T012 (can run in parallel with [P] marker after Phase 1)
3. **Phase 3 (User Story 1)**: T013-T021 (sequential within phase, but can start after T012)
4. **Phase 4 (User Story 2)**: T022-T027 (can run in parallel with Phase 5)
5. **Phase 5 (User Story 3)**: T028-T035 (can run in parallel with Phase 4)
6. **Phase 6 (Polish)**: T036-T039 (final cleanup and optimization)

### Parallelization Strategy

**Maximum parallelism (Week 1)**:
```
Day 1-2:  T001-T006 (setup, all parallel)
Day 3-4:  T007-T012 (foundational, all parallel)
Day 5:    T013-T021 (User Story 1, parallel within phase)
```

**Daily parallel batches**:
```
Day 3: T007, T008, T009, T010, T011, T012 (6 parallel tasks)
Day 5: T013, T014, T015, T016, T018, T019 (6 parallel tasks)
       T017 (depends on T014)
       T020, T021 (testing, can be parallel)
Day 6: T022, T023, T024, T025, T028, T029 (6 parallel tasks)
Day 7: T026, T027, T030, T031, T032, T033, T034, T035 (8 parallel)
Day 8: T036, T037, T038, T039 (4 parallel)
```

### Code Review Checkpoints

**After Phase 1 (T001-T006)**:
- Verify go.mod has gopostal
- Review directory structure matches architecture
- Confirm error types compile and have appropriate methods

**After Phase 2 (T007-T012)**:
- Review repository pattern implementation
- Verify mock repository works with unit tests
- Check gopostal integration doesn't leak library details

**After Phase 3 (T013-T021)**:
- Review handler error mapping
- Verify all tests pass
- Check response schema matches openapi.yaml

**After Phase 5 (T028-T035)**:
- Verify all error paths covered
- Check error messages are user-friendly
- Confirm HTTP status codes correct

**Final (T036-T039)**:
- Run full test suite and linting
- Verify performance meets requirements
- Review documentation for completeness

---

## Success Metrics & Acceptance Criteria

### Build & Test Success
- [X] `go build ./...` succeeds with zero errors
- [X] `go test ./...` passes with ≥80% coverage
- [X] `golangci-lint run ./...` produces zero warnings
- [X] All integration tests pass with real gopostal

### Functional Requirements Met
- [X] SC-001: Valid addresses normalized with 100% accuracy
- [X] SC-002: Invalid addresses identified with specific feedback (100% of cases)
- [X] SC-003: Response time <500ms p95 percentile
- [X] SC-004: Response schema consistent across all inputs
- [X] SC-005: Formatting variations normalize to identical form
- [X] SC-006: Normalized addresses usable in downstream systems

### User Story 1 (P1) MVP Complete
- [X] POST /api/v1/validate-address accepts raw address string
- [X] Returns normalized address with all components
- [X] Handles various formatting (capitalization, spacing, punctuation)
- [X] Response structure matches openapi.yaml
- [X] E2E tests pass with real gopostal

### User Story 2 (P2) Complete
- [X] All responses have consistent schema
- [X] Confidence metadata populated
- [X] CorrectionsApplied tracked
- [X] FormattedAddress generated

### User Story 3 (P2) Complete
- [X] Invalid input returns 400 with clear errors
- [X] Unparseable address returns 422 with suggestions
- [X] Error responses include field, reason, value, suggestion
- [X] Integration tests cover all error paths

---

## Implementation Notes

### Critical Path for MVP (User Story 1)
Fastest route to working feature:
1. T001-T006 (Setup): 1 day
2. T007-T012 (Foundational): 1-2 days
3. T013-T017 (Handler setup): 1 day
4. T018-T021 (Testing): 1 day
**MVP ready**: 4 days

### Parallel Opportunities
- Setup tasks (T001-T006) can all run in parallel
- Foundational tasks (T007-T012) can all run in parallel (after setup)
- User Story 1 handler tasks (T014-T016) can run in parallel
- Error handling (T028-T035) independent from response structure (T022-T027)

### Testing Strategy per Phase
- **Phase 1-2**: No tests required (setup & infrastructure)
- **Phase 3**: Unit tests (T012, T018) + Integration tests (T019) mandatory
- **Phase 4-5**: Unit tests mandatory, integration tests recommended
- **Phase 6**: Performance benchmark and linting

### Quality Gates
- Never merge without passing `go test ./...`
- Never merge without `golangci-lint run ./...` passing
- Integration tests must exercise real gopostal (not mock)
- API responses must match openapi.yaml schema

---

## Estimated Effort & Timeline

| Phase | Tasks | Effort | Timeline | Dependencies |
|-------|-------|--------|----------|--------------|
| 1 (Setup) | T001-T006 | 1-2 days | Day 1-2 | None |
| 2 (Foundation) | T007-T012 | 2-3 days | Day 3-5 | Phase 1 ✓ |
| 3 (US1) | T013-T021 | 2-3 days | Day 5-7 | Phase 2 ✓ |
| 4 (US2) | T022-T027 | 1-2 days | Day 7-8 | Phase 2 ✓ |
| 5 (US3) | T028-T035 | 1-2 days | Day 7-8 | Phase 3 ✓ |
| 6 (Polish) | T036-T039 | 0.5 days | Day 8 | Phases 3-5 ✓ |
| **Total** | **39 tasks** | **2-3 weeks** | **8 days parallel** | - |

**Note**: Effort estimates assume 1 developer. Parallelization can reduce to 8-10 actual calendar days with team.

---

## Appendix: Dependency Graph

```
T001 (install gopostal)
  ↓
T002 (create directories)
  ├→ T003, T004, T005, T006 (domain, repository, mock) [parallel]
     ↓
  T007, T008, T009, T010, T011, T012 (usecase, parser, tests) [parallel]
     ├→ T013, T014, T015, T016 (handler setup) [parallel]
     │  ├→ T017 (register route)
     │  ├→ T018, T019, T020, T021 (tests) [parallel]
     │
     ├→ T022, T023 (response structure) [parallel to T028]
     │  ├→ T024, T025, T026, T027 (response tests) [parallel]
     │
     ├→ T028, T029, T030, T031 (error handling) [parallel to T022]
        └→ T032, T033, T034, T035 (error tests) [parallel]

Final: T036, T037, T038, T039 (logging, performance, linting)
```

---

## Conclusion

This task breakdown provides a clear, executable implementation plan with:
- ✅ 39 specific, actionable tasks
- ✅ Clear user story organization (P1 MVP + P2 features)
- ✅ Independent test criteria for each story
- ✅ Parallelization opportunities to reduce calendar time
- ✅ Quality gates and success metrics
- ✅ Estimated effort and dependencies

**Ready to begin Phase 1: Setup & Project Initialization**

Next action: Create task tracking in IDE or project management tool and begin T001 (install gopostal).
