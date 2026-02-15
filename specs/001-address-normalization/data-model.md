# Data Model: Address Normalization API

**Date**: February 15, 2026  
**Feature**: Address Normalization API (001-address-normalization)  
**Status**: Phase 1 - Design Complete

---

## Entity Definitions

### Address (Core Entity)

**Purpose**: Represents a normalized, validated US address with metadata about parsing and confidence.

**Go Definition**:
```go
package entity

// Address represents a normalized US address with parsing metadata.
type Address struct {
    // Core address components (required)
    StreetAddress string `json:"street_address"` // "123 Main St" or PO Box, military, rural route identifier
    City          string `json:"city"`           // City or town name
    State         string `json:"state"`          // 2-letter USPS abbreviation (e.g., "NY", "CA")
    PostalCode    string `json:"postal_code"`    // 5-digit ZIP or ZIP+4 (e.g., "10001", "10001-1234")
    
    // Address classification (required)
    AddressType string `json:"address_type"` // "standard_street" | "po_box" | "apo_fpo" | "rural_route"
    
    // Formatted display (optional)
    FormattedAddress string `json:"formatted_address,omitempty"` // Human-readable: "123 Main St, New York, NY 10001"
    
    // Metadata about normalization (optional)
    Confidence         *Confidence `json:"confidence,omitempty"`         // What was inferred vs. parsed directly
    CorrectionsApplied []string    `json:"corrections_applied,omitempty"` // What normalizations were applied
}

// Confidence tracks source of each address component.
type Confidence struct {
    StateConfidence  string `json:"state_confidence"`  // "direct" | "inferred_from_zip" | "inferred_from_city"
    CityConfidence   string `json:"city_confidence"`   // "direct" | "inferred_from_zip" | "inferred_from_state"
    PostalConfidence string `json:"postal_confidence"` // "direct" | "inferred"
}
```

**Field Descriptions**:

| Field | Type | Required | Constraints | Examples |
|-------|------|----------|-----------|----------|
| `StreetAddress` | string | Yes | 3-100 chars; cannot be empty after normalization | "123 Main St", "PO Box 456", "APO AE 09012" |
| `City` | string | Yes | 2-50 chars; must match valid US city or town | "New York", "Springfield", "Los Angeles" |
| `State` | string | Yes | 2 uppercase letters; must be valid USPS code | "NY", "CA", "IL", "TX" (includes "DC" for District of Columbia) |
| `PostalCode` | string | Yes | 5 digits or 5+4 format (ZIP or ZIP+4) | "10001", "90210", "10001-1234" |
| `AddressType` | string | Yes | One of 4 enum values | "standard_street" (default), "po_box", "apo_fpo", "rural_route" |
| `FormattedAddress` | string | No | 20-200 chars; provides human-readable display | "123 Main St, New York, NY 10001" |
| `Confidence` | object | No | Populated by usecase during normalization | See Confidence definition |
| `CorrectionsApplied` | []string | No | Array of 0-10 correction descriptions | ["Standardized capitalization: 'ny' → 'NY'"] |

**Validation Rules**:

1. **Required Components**: Minimum 2 of {street_address, city, state} must be present and valid after parsing
   - All three must be non-empty in final Address object
   - PostalCode must be provided or inferrable from city+state

2. **State Validation**:
   - Must be 2-letter USPS abbreviation
   - Must match one of 50 states + DC (AL, AK, AZ, ..., WY, DC)
   - Case-normalized to uppercase

3. **Postal Code Validation**:
   - If provided: Must match 5-digit or 5+4 format (regex: `^\d{5}(-\d{4})?$`)
   - If not provided: Must be inferrable from city + state via lookup table

4. **City Validation**:
   - Must be valid city/town in target state (after state resolved)
   - Case-normalized to proper case (e.g., "New York", not "new york" or "NEW YORK")
   - Must match Census-recognized city or USPS delivery city

5. **Street Address Validation**:
   - Must contain street number and name OR valid alternative (PO Box, APO, rural route)
   - Cannot be purely numeric or purely alphabetic
   - Standardized to proper case and spacing

### Confidence Entity

**Purpose**: Tracks source of each parsed component to indicate certainty level.

**Go Definition** (part of Address):
```go
type Confidence struct {
    StateConfidence  string `json:"state_confidence"`  // One of: "direct", "inferred_from_zip", "inferred_from_city"
    CityConfidence   string `json:"city_confidence"`   // One of: "direct", "inferred_from_zip", "inferred_from_state"
    PostalConfidence string `json:"postal_confidence"` // One of: "direct", "inferred"
}
```

**Confidence Levels**:

| Level | Meaning | Reliability | When Used |
|-------|---------|------------|-----------|
| `"direct"` | Component extracted directly from input | Highest | gopostal recognized valid component in input |
| `"inferred_from_zip"` | Derived from ZIP code lookup | High | ZIP code present; city/state looked up from ZIP database |
| `"inferred_from_city"` (state only) | Derived from city-to-state mapping | Medium | City recognized; state inferred from city location |
| `"inferred_from_state"` (city only) | Derived from state when only state available | Low | Limited input; fallback only (avoid if possible) |
| `"inferred"` (postal code) | Derived from city+state lookup | Medium | No ZIP in input; inferred from city+state mapping |

**Assignment Rules**:

```
StateConfidence:
  IF state_in_input AND valid_usps_code
    → "direct"
  ELSE IF postal_code_in_input
    → "inferred_from_zip" (lookup state from ZIP database)
  ELSE IF city_in_input AND city_matches_unique_state
    → "inferred_from_city" (e.g., "Springfield" → look up all Springfield cities → pick most populous with state)
  ELSE IF state_inferred_from_other_components
    → "inferred_from_city"

CityConfidence:
  IF city_in_input AND city_matches_state
    → "direct"
  ELSE IF postal_code_in_input
    → "inferred_from_zip" (lookup city from ZIP database)
  ELSE IF state_known AND region_inference_possible
    → "inferred_from_state"

PostalConfidence:
  IF postal_code_in_input AND valid_format
    → "direct"
  ELSE IF city_and_state_known
    → "inferred" (lookup ZIP from city+state database)
```

### Error Entity

**Purpose**: Represents validation or parsing failures with structured feedback.

**Go Definition**:
```go
package entity

// ValidationError represents a request validation failure (400 Bad Request).
type ValidationError struct {
    Field      string      `json:"field"`      // Field name that failed (e.g., "address")
    Reason     string      `json:"reason"`     // Why validation failed
    Value      interface{} `json:"value,omitempty"` // The problematic value
    Suggestion string      `json:"suggestion,omitempty"` // How to fix it
}

// ParsingError represents an unparseable address (422 Unprocessable Entity).
type ParsingError struct {
    Field      string `json:"field"`      // Which component couldn't be parsed (e.g., "city", "state")
    Reason     string `json:"reason"`     // Why parsing failed
    Value      string `json:"value,omitempty"` // The problematic value
    Suggestion string `json:"suggestion,omitempty"` // Did you mean...?
}

// AmbiguousAddressError represents an address with multiple valid interpretations.
type AmbiguousAddressError struct {
    PrimaryAddress *Address   `json:"address"`        // Most populous match
    Candidates     []*Address `json:"candidates"`     // Alternative interpretations (up to 5)
    Message        string     `json:"message"`        // "Multiple valid interpretations found"
}
```

**Error Examples**:

```json
ValidationError: {
    "field": "address",
    "reason": "Empty or missing required field",
    "value": "",
    "suggestion": "Provide at least 2 of: street address, city, state"
}

ParsingError: {
    "field": "state",
    "reason": "Unknown state code",
    "value": "XX",
    "suggestion": "Did you mean 'TX' or 'CT'?"
}

ParsingError: {
    "field": "city",
    "reason": "City not found in provided state",
    "value": "Fakeville",
    "suggestion": "Did you mean 'Springfield' in IL?"
}

AmbiguousAddressError: {
    "address": { /* Primary address (most populous) */ },
    "candidates": [
        { /* Alternative 1 */ },
        { /* Alternative 2 */ }
    ],
    "message": "Multiple valid interpretations found; returning most populous match"
}
```

---

## Address Type Classification

**Purpose**: Identifies special address formats requiring different handling.

**Enum Values**:

| Type | Description | Examples | Special Handling |
|------|-------------|----------|------------------|
| `"standard_street"` | Regular street address | "123 Main St, New York, NY 10001" | Standard delivery; default type |
| `"po_box"` | Post Office Box | "PO Box 456, Springfield, IL 62701" | No street delivery; recipient must pick up |
| `"apo_fpo"` | Armed Forces address | "APO AE 09012", "FPO AP 96261-1234" | Military postal service; restricted shipping |
| `"rural_route"` | Rural delivery route | "Route 1 Box 123, Nowhere, KS 67890" | Rural carrier delivery; may be slower |

**Detection Algorithm**:

```
IF address_contains("PO Box", "P.O. Box", "POB", "Box")
    → type = "po_box"
ELSE IF address_contains("APO", "FPO", "Armed Forces")
    → type = "apo_fpo"
ELSE IF address_contains("Route", "RR", "Rural Route")
    → type = "rural_route"
ELSE
    → type = "standard_street" (default)
```

---

## State Transitions

### Address Normalization Flow

```
┌─────────────────────────────────┐
│  Raw Input Address String       │
│  "123 main st new york ny 10001"│
└────────────┬────────────────────┘
             │
             ▼
┌─────────────────────────────────┐
│  Parse with gopostal            │
│  Extract components             │
└────────────┬────────────────────┘
             │
             ▼
┌─────────────────────────────────┐
│  Normalize Components           │
│  • Capitalization               │
│  • Spacing, punctuation         │
│  • Component reordering         │
└────────────┬────────────────────┘
             │
             ▼
┌─────────────────────────────────┐
│  Validate Minimum Components    │
│  At least 2 of 3:               │
│  street/city/state              │
└────────────┬────────────────────┘
             │
      ┌──────┴──────┐
      ▼             ▼
   VALID        INVALID
      │             │
      ▼             ▼
┌──────────┐  ┌────────────┐
│Validate  │  │Return 422  │
│Formats   │  │ParsingError│
└──────┬───┘  └────────────┘
       │
    ┌──┴──┐
    ▼     ▼
  VALID INVALID
    │      │
    ▼      ▼
┌───────┐ ┌────────────┐
│Assign │ │Return 400  │
│Confid │ │Validation  │
│-ence  │ │Error       │
└───┬───┘ └────────────┘
    │
    ▼
┌────────────────────┐
│Detect Address Type │
│(standard/po/apo...)│
└────────┬───────────┘
         │
         ▼
┌────────────────────┐
│Generate Candidates │
│(if ambiguous)      │
└────────┬───────────┘
         │
         ▼
┌────────────────────┐
│Return Address      │
│+ Confidence        │
│+ Corrections       │
└────────────────────┘
```

---

## Storage & Persistence

**Current Design**: Stateless validation service - no persistence layer required.

- No Address entities stored to database
- Each request independently parsed and validated
- Results returned immediately in HTTP response
- Future enhancement: Optional caching layer for frequently requested addresses

---

## Constraints & Invariants

### Invariants (Always True for Valid Address)

1. All three core fields populated: `StreetAddress`, `City`, `State`, `PostalCode`
2. `State` matches one of 50 states or DC (exactly 2 uppercase letters)
3. `PostalCode` matches format: 5 digits or 5+4 format
4. `City` is a valid city in target `State`
5. `AddressType` is one of four enum values
6. `Confidence` object is always populated (all three levels assigned)
7. `FormattedAddress` generated if not provided in input

### Constraints (Rules for Processing)

1. **Minimum Input**: At least 2 of {street, city, state} required
2. **Maximum Length**: Individual components limited (e.g., street ≤100 chars, city ≤50)
3. **Character Set**: Address components limited to ASCII + common diacritics
4. **Case Normalization**: All components standardized to proper case
5. **Whitespace**: Leading/trailing whitespace trimmed; internal extra spaces removed
6. **Ambiguity Limit**: Maximum 10 candidates returned (to prevent response bloat)

---

## Mapping to DTOs

### Request DTO

```go
// ValidateRequest is received from HTTP client.
type ValidateRequest struct {
    Address string `json:"address" binding:"required,min=3,max=500"`
}
```

### Response DTO

```go
// ValidateResponse is sent to HTTP client.
type ValidateResponse struct {
    Success            bool              `json:"success"`
    Address            *AddressDTO       `json:"address,omitempty"`
    Candidates         []*AddressDTO     `json:"candidates,omitempty"`
    Confidence         *ConfidenceDTO    `json:"confidence,omitempty"`
    CorrectionsApplied []string          `json:"corrections_applied,omitempty"`
    Errors             []ErrorDTO        `json:"errors,omitempty"`
    Message            string            `json:"message"`
}

type AddressDTO struct {
    StreetAddress    string `json:"street_address"`
    City             string `json:"city"`
    State            string `json:"state"`
    PostalCode       string `json:"postal_code"`
    AddressType      string `json:"address_type"`
    FormattedAddress string `json:"formatted_address,omitempty"`
}

type ConfidenceDTO struct {
    StateConfidence  string `json:"state_confidence"`
    CityConfidence   string `json:"city_confidence"`
    PostalConfidence string `json:"postal_confidence"`
}

type ErrorDTO struct {
    Field      string      `json:"field"`
    Reason     string      `json:"reason"`
    Value      interface{} `json:"value,omitempty"`
    Suggestion string      `json:"suggestion,omitempty"`
}
```

---

## Design Decisions

### Why Confidence in Response?

Provides clients transparency about parsing certainty, enabling:
- Custom validation rules (e.g., reject if state is "inferred_from_city")
- Better error messages to users (e.g., "City was inferred; please verify")
- Quality metrics (e.g., track how many inferred vs. direct parses)

### Why CorrectionsApplied Array?

Helps clients understand what was normalized:
- Detect if parsing was lossy (e.g., "PO Box corrected to 'PO Box 123'")
- Communicate changes to users
- Audit trail of transformations

### Why AddressType Field?

Enables clients to:
- Apply special handling (e.g., PO boxes cannot ship via standard carriers)
- Validate against business rules (e.g., "reject APO addresses for this product")
- Flag non-standard addresses for manual review

### Why Candidate Addresses?

Handles ambiguity gracefully:
- Client can implement custom disambiguation logic
- Provides fallback options without requiring server roundtrips
- Ranked by relevance (population), enabling good UX defaults
