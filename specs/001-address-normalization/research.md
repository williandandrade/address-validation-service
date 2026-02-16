# Research Document: Address Normalization API

**Date**: February 15, 2026  
**Feature**: Address Normalization API (001-address-normalization)  
**Status**: Phase 0 Complete - All Unknowns Resolved

---

## Research Findings Summary

This document consolidates research decisions on critical unknowns identified in the Technical Context. All items have been researched and resolved, enabling Phase 1 design to proceed.

---

## Research Task 1: gopostal Library Integration

### Decision
Use `github.com/openvenues/gopostal` for address parsing in the ValidateAddressUsecase.

### Rationale
- **Native Go bindings**: Wraps libpostal, the industry-standard C library for address parsing
- **Comprehensive parsing**: Handles diverse address formats (street addresses, PO boxes, military addresses, rural routes)
- **Component extraction**: Automatically extracts street, city, state, postal code with high accuracy
- **Active maintenance**: Community-supported with regular updates
- **Performance**: C-based library provides fast parsing suitable for <500ms p95 response time requirement
- **US-focused capability**: Excellent support for US address normalization

### Alternatives Considered

| Alternative | Why Not Chosen |
|------------|----------------|
| Google Maps Geocoding API | External dependency with rate limits; adds latency and cost; introduces reliability coupling |
| USPS Address Standardization API | US-specific but requires postal service credentials; less detailed component extraction; adds operational complexity |
| Custom regex-based parser | Brittle for diverse formats; insufficient for PO boxes and special address types; maintenance burden |
| Amazon Location Service Address | Proprietary; vendor lock-in; cost per request; not optimal for this use case |

### Implementation Details

**Installation**:
```bash
go get github.com/openvenues/gopostal
```

**Key Functions**:
- `gopostal.ParseAddress(input string)`: Returns parsed address components
- `gopostal.ExpandAddress(input string)`: Returns all possible interpretations (for candidates)

**Integration Pattern**:
1. Create `internal/infrastructure/address_parser/gopostal_parser.go`
2. Implement `ValidateAddressRepository` interface wrapping gopostal
3. Handle parsing errors gracefully with domain error types
4. Extract components into domain Address entity

**Error Handling**:
- Wrap gopostal errors with domain context via `fmt.Errorf`
- Classify parsing failures as `ParsingError` or `AmbiguousAddressError`
- Return structured error objects to clients with suggestions

---

## Research Task 2: Address Confidence Scoring

### Decision
Implement confidence tracking with three confidence levels per component, using domain logic in ValidateAddressUsecase.

### Confidence Levels

**For State Component**:
- `"direct"`: Parsed directly from input as valid 2-letter USPS code
- `"inferred_from_zip"`: Inferred from ZIP code via lookup (requires ZIP code present)
- `"inferred_from_city"`: Inferred from city name (when city matches known city-state mapping)

**For City Component**:
- `"direct"`: Parsed directly from input
- `"inferred_from_zip"`: Inferred from ZIP code lookup
- `"inferred_from_state"`: Inferred from state + other components (least reliable)

**For Postal Code Component**:
- `"direct"`: Parsed directly from input (5 or 9-digit format)
- `"inferred"`: Derived from city + state lookup (when ZIP not provided)

### Rationale
- **Transparency**: Clients see what was parsed vs. inferred, enabling informed decisions
- **Alignment with spec**: Spec requirement FR-008 mandates correction/inference metadata
- **Progressive normalization**: Allows partial address processing with confidence indicators
- **Client flexibility**: Clients can apply custom validation rules based on confidence levels

### Corrections Tracking

Track all standardizations applied as array of strings:
- Capitalization fixes: `"Standardized capitalization: 'ny' → 'NY'"`
- Spacing normalization: `"Removed extra spaces in street address"`
- Punctuation normalization: `"Standardized direction abbreviations: 'N.' → 'N'"`
- Component reordering: `"Reordered address components to standard format"`

### Implementation Pattern

```go
type Confidence struct {
    StateConfidence  string `json:"state_confidence"`
    CityConfidence   string `json:"city_confidence"`
    PostalConfidence string `json:"postal_confidence"`
}

type Address struct {
    StreetAddress      string
    City               string
    State              string
    PostalCode         string
    AddressType        string
    FormattedAddress   string
    Confidence         *Confidence  // Set by usecase during normalization
    CorrectionsApplied []string     // Populated during parsing and normalization
}
```

### Business Rules for Confidence Assignment

1. **State Confidence**:
   - Direct: gopostal extracts valid 2-letter state code
   - Inferred from ZIP: Lookup state from ZIP code (fallback if state missing)
   - Inferred from city: Match city to known city-state mapping (least reliable)

2. **City Confidence**:
   - Direct: gopostal extracts city name matching known city in target state
   - Inferred from ZIP: Lookup city from ZIP code
   - Inferred from state: Derive city from state when only state provided (not recommended)

3. **Postal Code Confidence**:
   - Direct: gopostal extracts valid ZIP format (5 or 9 digits)
   - Inferred: Lookup ZIP from city + state (when ZIP not in input)

---

## Research Task 3: Candidate Address Generation and Ranking

### Decision
Use gopostal's expand functionality to generate candidates; rank by US Census population data with explicit ranking algorithm.

### Candidate Generation Strategy

**Step 1: Parse Base Address**
- Use `gopostal.ParseAddress()` to extract primary components
- Identify address type (standard_street, po_box, apo_fpo, rural_route)

**Step 2: Generate Ambiguity Alternatives**
- Use `gopostal.ExpandAddress()` to get all possible interpretations
- Apply state-level filtering (only return alternatives in valid US states)
- Limit to top 10 candidates maximum (clients typically only need 1-3 alternatives)

**Step 3: Rank by Relevance**
Use this ranking algorithm:
1. **Exact match**: If parsed city exactly matches target state's largest city → rank 1
2. **Population-based**: Rank remaining candidates by city population (descending)
3. **Confidence score**: Candidates where all components are "direct" rank higher than inferred

**Ranking Example**:
Input: "123 Main St Springfield" (no state provided)
Candidates generated:
1. Springfield, IL (population ~116,000) → **Rank 1** (most populous)
2. Springfield, MO (population ~169,000) → **Rank 2** (next populous, but further west)
3. Springfield, MA (population ~155,000) → **Rank 3**
4. Springfield, OH (population ~60,000) → **Rank 4**

Return first candidate as primary; include top 5 alternatives in `candidates` array.

### Population Data Source

**Decision**: Use embedded lookup table of top 500 US cities by population (2020 Census data).

**Rationale**:
- Fast O(1) lookup without external API calls
- Guaranteed availability (no network dependency)
- Static data (population changes rarely affect address validation)
- Aligns with constitution principle of resilience without external dependencies

**Implementation**:
```go
// In infrastructure/address_parser/population_data.go
var USCitiesByPopulation = map[string]int{
    "new_york_ny":      8336817,
    "los_angeles_ca":   3979576,
    "springfield_il":   116000,
    // ... ~500 top cities
}
```

### Alternatives Considered

| Alternative | Why Not Chosen |
|------------|----------------|
| Return only highest-confidence match | Limits client flexibility for ambiguous cases; reduces feature value |
| Return all candidates unsorted | Clients have no guidance; poor UX; requires client-side processing |
| Real-time Census Bureau API lookup | Adds external dependency; latency impact; availability risk |
| Return candidates ranked by frequency in US Postal database | No reliable data source accessible without external APIs |

---

## Research Task 4: HTTP Error Response Standards

### Decision
Use standard HTTP status codes with structured error response bodies following RFC 4918 conventions.

### Status Code Mapping

| HTTP Code | Scenario | Example |
|-----------|----------|---------|
| **400 Bad Request** | Malformed/invalid request structure | Missing `address` field, non-string value, empty string |
| **422 Unprocessable Entity** | Valid request but unparseable address | Address contains <2 required components; no recognizable address components |
| **500 Internal Server Error** | Unexpected server failures | gopostal crashes, infrastructure failures, panic conditions |

### Rationale

**400 vs 422 Distinction**:
- **400**: Indicates client must fix request syntax/format before retry
- **422**: Indicates request syntax is valid but semantic content is invalid (RFC 4918 extension)

This distinction allows clients to:
- Catch 400 errors and validate request structure before retrying
- Treat 422 errors as semantic validation failures requiring user input review
- Never retry 4xx errors without changing request

### Error Response Structure

All error responses include structured error array with field-level details:

```json
{
  "success": false,
  "errors": [
    {
      "field": "address",
      "reason": "Empty or missing required field",
      "value": "",
      "suggestion": "Provide at least 2 of: street address, city, state"
    }
  ],
  "message": "Request validation failed"
}
```

### Error Classification Logic

**In Handler (validate_address.go)**:
1. **Request validation** → 400 (missing/invalid field)
2. **Usecase returns ParsingError** → 422 (unparseable)
3. **Usecase returns AmbiguousAddressError** → 200 (valid but ambiguous - return primary + candidates)
4. **Unexpected errors** → 500 (log and return generic error message)

**Error Categories**:
```go
type ValidationError struct {
    Field      string
    Reason     string
    Suggestion string
    Value      any
}

type ParsingError struct {
    Reason     string
    Suggestion string
}

type AmbiguousAddressError struct {
    Candidates []*Address
    Message    string
}
```

---

## Research Task 5: Go Testing Patterns

### Decision
Implement three-tier testing strategy: unit tests with table-driven approach, mock repositories for isolation, and integration tests with real gopostal.

### Testing Architecture

#### Tier 1: Unit Tests (Fast, Isolated)
**Location**: `internal/usecase/validate_address_test.go`  
**Pattern**: Table-driven tests  
**Dependencies**: Mock repository (injected)

```go
var validateAddressTests = []struct {
    name          string
    input         string
    expectedAddr  *Address
    expectedErr   error
    mockBehavior  func(*MockRepository)
}{
    {
        name:  "valid complete address",
        input: "123 Main St New York NY 10001",
        expectedAddr: &Address{
            StreetAddress: "123 Main St",
            City:          "New York",
            State:         "NY",
            PostalCode:    "10001",
        },
        expectedErr: nil,
    },
    // ... more test cases
}

func TestValidateAddress(t *testing.T) {
    for _, tt := range validateAddressTests {
        t.Run(tt.name, func(t *testing.T) {
            repo := &MockRepository{}
            if tt.mockBehavior != nil {
                tt.mockBehavior(repo)
            }
            
            uc := NewValidateAddressUsecase(repo)
            addr, err := uc.ValidateAddress(context.Background(), tt.input)
            
            assert.Equal(t, tt.expectedAddr, addr)
            assert.Equal(t, tt.expectedErr, err)
        })
    }
}
```

**Benefits**:
- Fast execution (milliseconds per test)
- No external dependencies
- Clear documentation of business logic
- Easy to add new test cases

#### Tier 2: Integration Tests (Moderate Speed, Real Dependencies)
**Location**: `internal/infrastructure/address_parser/gopostal_parser_test.go`  
**Pattern**: Real gopostal calling with assertions on gopostal behavior

```go
func TestGopostalParserRealAddresses(t *testing.T) {
    parser := NewGopostalParser()
    
    tests := []struct {
        input    string
        expStreet string
        expCity   string
        expState  string
    }{
        {
            input:     "123 main st new york ny 10001",
            expStreet: "123 Main St",
            expCity:   "New York",
            expState:  "NY",
        },
        // ... real address test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
            addr, _, err := parser.ParseAddress(context.Background(), tt.input)
            
            assert.NoError(t, err)
            assert.Equal(t, tt.expStreet, addr.StreetAddress)
            assert.Equal(t, tt.expCity, addr.City)
            assert.Equal(t, tt.expState, addr.State)
        })
    }
}
```

**Benefits**:
- Validates gopostal behavior and component extraction
- Catches library-specific parsing issues
- Documents expected parsing results
- Slower but still acceptable (< 1 second per test)

#### Tier 3: E2E Tests (Full Stack)
**Location**: `tests/integration/validate_address_integration_test.go`  
**Pattern**: HTTP request → Handler → Usecase → gopostal → Response

```go
func TestValidateAddressHTTPEndpoint(t *testing.T) {
    // Setup: Start test HTTP server, initialize handler with real gopostal
    handler := NewValidateAddressHandler(
        NewValidateAddressUsecase(
            NewGopostalParser(),
        ),
    )
    
    tests := []struct {
        name           string
        requestBody    string
        expectedStatus int
        expectedAddr   *Address
    }{
        {
            name:           "valid address returns 200",
            requestBody:    `{"address":"123 Main St New York NY 10001"}`,
            expectedStatus: 200,
            expectedAddr: &Address{
                StreetAddress: "123 Main St",
                City:          "New York",
                State:         "NY",
                PostalCode:    "10001",
            },
        },
        // ... more E2E tests
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            resp := handler.ValidateAddress(tt.requestBody)
            
            assert.Equal(t, tt.expectedStatus, resp.Status)
            assert.Equal(t, tt.expectedAddr, resp.Address)
        })
    }
}
```

**Benefits**:
- Validates complete request/response flow
- Catches integration issues between layers
- Provides real E2E coverage
- Slower but essential for confidence

### Mock Repository Pattern

**Location**: `internal/usecase/validate_address_mock.go`

```go
type MockRepository struct {
    ParseAddressFn func(ctx context.Context, raw string) (*Address, []*Address, error)
}

func (m *MockRepository) ParseAddress(ctx context.Context, raw string) (*Address, []*Address, error) {
    return m.ParseAddressFn(ctx, raw)
}
```

**Usage in Tests**:
```go
mock := &MockRepository{
    ParseAddressFn: func(ctx context.Context, raw string) (*Address, []*Address, error) {
        if strings.Contains(raw, "invalid") {
            return nil, nil, ParsingError{Reason: "invalid"}
        }
        return &Address{City: "New York"}, nil, nil
    },
}

uc := NewValidateAddressUsecase(mock)
```

**Benefits**:
- Isolates usecase logic from repository implementation
- Enables testing error conditions without real gopostal
- Fast test execution
- Easy to set up different scenarios

### Test Data Management

**Location**: `tests/_valid-payloads.jsonl`, `tests/_invalid-payloads.jsonl`

Format: One JSON object per line (JSONL)
```jsonl
{"address":"123 Main St New York NY 10001","expected_city":"New York","expected_state":"NY"}
{"address":"PO Box 123 Springfield IL 62701","expected_type":"po_box"}
```

**Benefits**:
- Separates test data from code
- Enables data-driven testing
- Easy to add new test cases
- Can be shared with QA teams

### Coverage Targets

- **Domain layer**: >85% (entity validation, confidence scoring)
- **Usecase layer**: >85% (business logic, error handling)
- **Handler layer**: >80% (HTTP request/response mapping)
- **Repository (gopostal)**: >75% (integration tests catch parsing issues)
- **Overall**: >80% coverage target

---

## Summary: All Unknowns Resolved

| Unknown | Research Task | Decision | Status |
|---------|---------------|----------|--------|
| Address parsing library | Task 1 | Use gopostal library | ✅ Resolved |
| Confidence scoring approach | Task 2 | Three-level confidence per component | ✅ Resolved |
| Candidate ranking algorithm | Task 3 | Population-based ranking with embedded lookup | ✅ Resolved |
| HTTP error codes | Task 4 | 400/422/500 with structured error bodies | ✅ Resolved |
| Testing strategy | Task 5 | Three-tier: unit/integration/E2E with mocks | ✅ Resolved |

All research complete. Technical context fully clarified. Phase 1 design ready to proceed.
