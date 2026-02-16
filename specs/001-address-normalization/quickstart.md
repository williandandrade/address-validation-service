# Quick Start Implementation Guide

**Date**: February 15, 2026  
**Feature**: Address Normalization API (001-address-normalization)  
**Target**: Implementing `/api/v1/validate-address` endpoint with gopostal integration

---

## Overview

This guide provides step-by-step instructions for implementing the address validation feature using the clean architecture layers defined in the specification and data model.

**Key Endpoint**: `POST /api/v1/validate-address`  
**Request**: `{"address": "user input"}`  
**Response**: Normalized address with confidence metadata or structured errors

---

## Architecture Layers

```
HTTP Request
    ↓
[Delivery Layer] Handler (api/handler/validate_address.go)
    ↓ (inject usecase)
[Usecase Layer] ValidateAddressUsecase (usecase/validate_address.go)
    ↓ (depend on interface)
[Domain Layer] ValidateAddressRepository interface (usecase/repository.go)
    ↓ (implement interface)
[Infrastructure] GopostalParser (infrastructure/address_parser/gopostal_parser.go)
    ↓
gopostal library
```

**Dependency Flow**: Each layer depends only on layers below it. No upward dependencies. This ensures testability and flexibility.

---

## Step 1: Install Dependencies

### Add gopostal to go.mod

```bash
cd /Users/williandrade/projects/interviews/the-guarantors/address-validation-service
go get github.com/openvenues/gopostal@latest
```

Verify installation:
```bash
go mod verify
```

---

## Step 2: Create Domain Layer

### 2.1 Create Address Entity

**File**: `internal/domain/entity/address.go`

```go
package entity

// Address represents a normalized, validated US address.
type Address struct {
    // Core components
    StreetAddress    string       `json:"street_address"`
    City             string       `json:"city"`
    State            string       `json:"state"`
    PostalCode       string       `json:"postal_code"`
    AddressType      string       `json:"address_type"` // standard_street, po_box, apo_fpo, rural_route
    FormattedAddress string       `json:"formatted_address,omitempty"`
    
    // Metadata
    Confidence         *Confidence `json:"confidence,omitempty"`
    CorrectionsApplied []string    `json:"corrections_applied,omitempty"`
}

// Confidence tracks the source of each address component.
type Confidence struct {
    StateConfidence  string `json:"state_confidence"`
    CityConfidence   string `json:"city_confidence"`
    PostalConfidence string `json:"postal_confidence"`
}

// Validate checks that Address meets requirements.
func (a *Address) Validate() error {
    if a.StreetAddress == "" {
        return fmt.Errorf("street_address required")
    }
    if a.City == "" {
        return fmt.Errorf("city required")
    }
    if len(a.State) != 2 {
        return fmt.Errorf("state must be 2-letter USPS code")
    }
    if !isValidZipCode(a.PostalCode) {
        return fmt.Errorf("postal_code must be 5 or 9-digit format")
    }
    return nil
}

func (a *Address) FormatAddress() {
    if a.FormattedAddress == "" {
        a.FormattedAddress = fmt.Sprintf("%s, %s, %s %s",
            a.StreetAddress, a.City, a.State, a.PostalCode)
    }
}

func isValidZipCode(code string) bool {
    matched, _ := regexp.MatchString(`^\d{5}(-\d{4})?$`, code)
    return matched
}
```

### 2.2 Create Domain Errors

**File**: `internal/domain/errors/errors.go`

```go
package errors

import "fmt"

// ValidationError represents request validation failure.
type ValidationError struct {
    Field      string
    Reason     string
    Value      interface{}
    Suggestion string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Reason)
}

// ParsingError represents address parsing failure.
type ParsingError struct {
    Field      string
    Reason     string
    Value      string
    Suggestion string
}

func (e ParsingError) Error() string {
    return fmt.Sprintf("parsing error on field '%s': %s", e.Field, e.Reason)
}

// AmbiguousAddressError represents multiple valid interpretations.
type AmbiguousAddressError struct {
    Message string
}

func (e AmbiguousAddressError) Error() string {
    return e.Message
}
```

---

## Step 3: Create Repository Interface

**File**: `internal/usecase/repository.go`

```go
package usecase

import (
    "context"
    "github.com/williandrade/address-validation-service/internal/domain/entity"
)

// ValidateAddressRepository defines the contract for address parsing.
type ValidateAddressRepository interface {
    // ParseAddress parses a raw address string and returns:
    // - primary Address (best match)
    // - candidates []*Address (alternative interpretations, max 10)
    // - error if parsing completely failed
    ParseAddress(ctx context.Context, rawAddress string) (*entity.Address, []*entity.Address, error)
}
```

---

## Step 4: Create Usecase (Business Logic)

**File**: `internal/usecase/validate_address.go`

```go
package usecase

import (
    "context"
    "fmt"
    "strings"
    
    "github.com/williandrade/address-validation-service/internal/domain/entity"
    "github.com/williandrade/address-validation-service/internal/domain/errors"
)

// ValidateAddressUsecase handles address validation business logic.
type ValidateAddressUsecase struct {
    repo ValidateAddressRepository
}

// NewValidateAddressUsecase creates a new usecase instance.
func NewValidateAddressUsecase(repo ValidateAddressRepository) *ValidateAddressUsecase {
    return &ValidateAddressUsecase{repo: repo}
}

// ValidateAddress normalizes and validates a raw address string.
func (uc *ValidateAddressUsecase) ValidateAddress(
    ctx context.Context,
    rawAddress string,
) (*entity.Address, []*entity.Address, error) {
    
    // Step 1: Validate input
    if strings.TrimSpace(rawAddress) == "" {
        return nil, nil, &errors.ValidationError{
            Field:      "address",
            Reason:     "address field is required and cannot be empty",
            Suggestion: "Provide a valid US address",
        }
    }
    
    // Step 2: Parse address via repository (abstracts gopostal)
    addr, candidates, err := uc.repo.ParseAddress(ctx, rawAddress)
    if err != nil {
        return nil, nil, err
    }
    
    // Step 3: Validate parsed address meets requirements
    if err := addr.Validate(); err != nil {
        return nil, nil, &errors.ParsingError{
            Reason:     err.Error(),
            Suggestion: "Ensure address contains street, city, and state",
        }
    }
    
    // Step 4: Assign confidence scores
    uc.assignConfidence(addr)
    
    // Step 5: Generate formatted address
    addr.FormatAddress()
    
    // Step 6: Process candidates if ambiguous
    if len(candidates) > 0 {
        for _, cand := range candidates {
            cand.FormatAddress()
        }
    }
    
    return addr, candidates, nil
}

// assignConfidence assigns confidence levels to address components.
func (uc *ValidateAddressUsecase) assignConfidence(addr *entity.Address) {
    if addr.Confidence == nil {
        addr.Confidence = &entity.Confidence{}
    }
    
    // For MVP: default to "direct" (all components parsed)
    // Future: Implement inference logic for ZIP code lookups, city→state mapping
    addr.Confidence.StateConfidence = "direct"
    addr.Confidence.CityConfidence = "direct"
    addr.Confidence.PostalConfidence = "direct"
}
```

### 4.2 Create Mock Repository for Testing

**File**: `internal/usecase/validate_address_mock.go`

```go
package usecase

import (
    "context"
    "github.com/williandrade/address-validation-service/internal/domain/entity"
)

// MockValidateAddressRepository is a test double for ValidateAddressRepository.
type MockValidateAddressRepository struct {
    ParseAddressFn func(ctx context.Context, raw string) (*entity.Address, []*entity.Address, error)
}

func (m *MockValidateAddressRepository) ParseAddress(
    ctx context.Context,
    rawAddress string,
) (*entity.Address, []*entity.Address, error) {
    if m.ParseAddressFn != nil {
        return m.ParseAddressFn(ctx, rawAddress)
    }
    return nil, nil, nil
}
```

---

## Step 5: Create Infrastructure Layer (gopostal Integration)

**File**: `internal/infrastructure/address_parser/gopostal_parser.go`

```go
package address_parser

import (
    "context"
    "fmt"
    "strings"
    
    "github.com/openvenues/gopostal/parser"
    "github.com/williandrade/address-validation-service/internal/domain/entity"
    "github.com/williandrade/address-validation-service/internal/domain/errors"
)

// GopostalParser implements ValidateAddressRepository using gopostal library.
type GopostalParser struct{}

func NewGopostalParser() *GopostalParser {
    return &GopostalParser{}
}

// ParseAddress parses a raw address string using gopostal.
func (p *GopostalParser) ParseAddress(
    ctx context.Context,
    rawAddress string,
) (*entity.Address, []*entity.Address, error) {
    
    // Parse using gopostal
    parsed := parser.ParseAddress(rawAddress)
    
    // Extract components
    addr := &entity.Address{
        StreetAddress: p.extractStreet(parsed),
        City:          p.extractCity(parsed),
        State:         p.extractState(parsed),
        PostalCode:    p.extractPostalCode(parsed),
        AddressType:   p.detectAddressType(rawAddress),
    }
    
    // Validate we have minimum required components
    if addr.StreetAddress == "" || addr.City == "" || addr.State == "" {
        return nil, nil, &errors.ParsingError{
            Reason:     "Could not extract required address components (street, city, state)",
            Suggestion: "Ensure address contains at least street address, city, and state",
        }
    }
    
    // TODO: Generate candidates via gopostal.ExpandAddress for ambiguous addresses
    // For MVP: return single address with no candidates
    
    return addr, nil, nil
}

func (p *GopostalParser) extractStreet(parsed map[string]interface{}) string {
    // Extract from parsed components
    // gopostal returns map with keys like "house_number", "road", etc.
    street := ""
    if house, ok := parsed["house_number"].(string); ok {
        street += house + " "
    }
    if road, ok := parsed["road"].(string); ok {
        street += road
    }
    return strings.TrimSpace(street)
}

func (p *GopostalParser) extractCity(parsed map[string]interface{}) string {
    if city, ok := parsed["city"].(string); ok {
        return strings.Title(strings.ToLower(city))
    }
    return ""
}

func (p *GopostalParser) extractState(parsed map[string]interface{}) string {
    if state, ok := parsed["state"].(string); ok {
        return strings.ToUpper(state)
    }
    return ""
}

func (p *GopostalParser) extractPostalCode(parsed map[string]interface{}) string {
    if postal, ok := parsed["postcode"].(string); ok {
        return postal
    }
    return ""
}

func (p *GopostalParser) detectAddressType(rawAddress string) string {
    lower := strings.ToLower(rawAddress)
    
    if strings.Contains(lower, "po box") || strings.Contains(lower, "p.o.") {
        return "po_box"
    }
    if strings.Contains(lower, "apo") || strings.Contains(lower, "fpo") {
        return "apo_fpo"
    }
    if strings.Contains(lower, "route") || strings.Contains(lower, "rr") {
        return "rural_route"
    }
    return "standard_street"
}
```

---

## Step 6: Update DTOs

**File**: `internal/api/dto/request.go` (already exists, no changes needed)

**File**: `internal/api/dto/response.go` (update from existing)

```go
package dto

import "time"

// ValidateResponse represents the API response for address validation.
type ValidateResponse struct {
    Success            bool               `json:"success"`
    Address            *AddressDTO        `json:"address,omitempty"`
    Candidates         []*AddressDTO      `json:"candidates,omitempty"`
    Confidence         *ConfidenceDTO     `json:"confidence,omitempty"`
    CorrectionsApplied []string           `json:"corrections_applied,omitempty"`
    Errors             []ErrorFieldDTO    `json:"errors,omitempty"`
    Message            string             `json:"message"`
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

type ErrorFieldDTO struct {
    Field      string      `json:"field"`
    Reason     string      `json:"reason"`
    Value      interface{} `json:"value,omitempty"`
    Suggestion string      `json:"suggestion,omitempty"`
}

// APIErrorResponse represents an API error response.
type APIErrorResponse struct {
    Code              string         `json:"code"`
    Message           string         `json:"message"`
    Details           map[string]any `json:"details,omitempty"`
    RetryAfterSeconds int            `json:"retry_after_seconds,omitempty"`
}

// HealthResponse represents the health check response.
type HealthResponse struct {
    Status    string            `json:"status"`
    Version   string            `json:"version"`
    Timestamp time.Time         `json:"timestamp"`
    Checks    map[string]string `json:"checks,omitempty"`
}

// ReadinessResponse represents the readiness check response.
type ReadinessResponse struct {
    Ready bool `json:"ready"`
}
```

---

## Step 7: Create HTTP Handler

**File**: `internal/api/handler/validate_address.go`

```go
package handler

import (
    "context"
    "log/slog"
    "net/http"
    
    "github.com/williandrade/address-validation-service/internal/api/dto"
    "github.com/williandrade/address-validation-service/internal/domain/errors"
    "github.com/williandrade/address-validation-service/internal/usecase"
)

// ValidateAddressHandler handles POST /api/v1/validate-address requests.
type ValidateAddressHandler struct {
    usecase *usecase.ValidateAddressUsecase
    logger  *slog.Logger
}

func NewValidateAddressHandler(
    uc *usecase.ValidateAddressUsecase,
    logger *slog.Logger,
) *ValidateAddressHandler {
    return &ValidateAddressHandler{
        usecase: uc,
        logger:  logger,
    }
}

// ValidateAddress handles POST /api/v1/validate-address.
func (h *ValidateAddressHandler) ValidateAddress(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Parse request
    var req dto.ValidateRequest
    if err := r.ParseForm(); err != nil {
        h.respondError(w, http.StatusBadRequest, "Invalid request format")
        return
    }
    
    // For JSON requests, use a proper JSON decoder
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.respondError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
        return
    }
    
    h.logger.InfoContext(ctx, "validate_address request", slog.String("address", req.Address))
    
    // Call usecase
    addr, candidates, err := h.usecase.ValidateAddress(ctx, req.Address)
    
    // Handle errors
    if err != nil {
        h.handleUsecaseError(w, err)
        return
    }
    
    // Success response
    resp := &dto.ValidateResponse{
        Success: true,
        Address: &dto.AddressDTO{
            StreetAddress:    addr.StreetAddress,
            City:             addr.City,
            State:            addr.State,
            PostalCode:       addr.PostalCode,
            AddressType:      addr.AddressType,
            FormattedAddress: addr.FormattedAddress,
        },
        Message: "Address validated successfully",
    }
    
    // Include confidence if available
    if addr.Confidence != nil {
        resp.Confidence = &dto.ConfidenceDTO{
            StateConfidence:  addr.Confidence.StateConfidence,
            CityConfidence:   addr.Confidence.CityConfidence,
            PostalConfidence: addr.Confidence.PostalConfidence,
        }
    }
    
    // Include corrections if any
    if len(addr.CorrectionsApplied) > 0 {
        resp.CorrectionsApplied = addr.CorrectionsApplied
    }
    
    // Include candidates if ambiguous
    if len(candidates) > 0 {
        for _, cand := range candidates {
            resp.Candidates = append(resp.Candidates, &dto.AddressDTO{
                StreetAddress:    cand.StreetAddress,
                City:             cand.City,
                State:            cand.State,
                PostalCode:       cand.PostalCode,
                AddressType:      cand.AddressType,
                FormattedAddress: cand.FormattedAddress,
            })
        }
        resp.Message = "Multiple valid interpretations found; returning most populous match"
    }
    
    h.respondJSON(w, http.StatusOK, resp)
}

func (h *ValidateAddressHandler) handleUsecaseError(w http.ResponseWriter, err error) {
    switch err.(type) {
    case *errors.ValidationError:
        ve := err.(*errors.ValidationError)
        resp := &dto.ValidateResponse{
            Success: false,
            Errors: []dto.ErrorFieldDTO{
                {
                    Field:      ve.Field,
                    Reason:     ve.Reason,
                    Value:      ve.Value,
                    Suggestion: ve.Suggestion,
                },
            },
            Message: "Request validation failed",
        }
        h.respondJSON(w, http.StatusBadRequest, resp)
        
    case *errors.ParsingError:
        pe := err.(*errors.ParsingError)
        resp := &dto.ValidateResponse{
            Success: false,
            Errors: []dto.ErrorFieldDTO{
                {
                    Field:      pe.Field,
                    Reason:     pe.Reason,
                    Value:      pe.Value,
                    Suggestion: pe.Suggestion,
                },
            },
            Message: "Address could not be normalized",
        }
        h.respondJSON(w, http.StatusUnprocessableEntity, resp)
        
    default:
        h.logger.ErrorContext(context.Background(), "unhandled error", slog.Any("error", err))
        h.respondError(w, http.StatusInternalServerError, "Internal server error")
    }
}

func (h *ValidateAddressHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}

func (h *ValidateAddressHandler) respondError(w http.ResponseWriter, status int, message string) {
    h.respondJSON(w, status, map[string]interface{}{
        "success": false,
        "message": message,
    })
}
```

---

## Step 8: Register Route in main.go

**File**: `cmd/server/main.go` (update existing)

```go
func main() {
    // ... existing setup ...
    
    // Initialize logger
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
    
    // Initialize gopostal parser (infrastructure layer)
    parser := address_parser.NewGopostalParser()
    
    // Initialize usecase
    validateUC := usecase.NewValidateAddressUsecase(parser)
    
    // Initialize handler
    validateHandler := handler.NewValidateAddressHandler(validateUC, logger)
    
    // Register routes
    router.POST("/api/v1/validate-address", validateHandler.ValidateAddress)
    
    // ... start server ...
}
```

---

## Step 9: Write Tests

### 9.1 Usecase Unit Tests

**File**: `internal/usecase/validate_address_test.go`

```go
package usecase

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/williandrade/address-validation-service/internal/domain/entity"
)

func TestValidateAddressWithMock(t *testing.T) {
    tests := []struct {
        name          string
        input         string
        mockAddr      *entity.Address
        mockErr       error
        expectErr     bool
    }{
        {
            name:  "valid address",
            input: "123 Main St New York NY 10001",
            mockAddr: &entity.Address{
                StreetAddress: "123 Main St",
                City:          "New York",
                State:         "NY",
                PostalCode:    "10001",
                AddressType:   "standard_street",
            },
            expectErr: false,
        },
        {
            name:      "empty address",
            input:     "",
            expectErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mock := &MockValidateAddressRepository{
                ParseAddressFn: func(ctx context.Context, raw string) (*entity.Address, []*entity.Address, error) {
                    return tt.mockAddr, nil, tt.mockErr
                },
            }
            
            uc := NewValidateAddressUsecase(mock)
            addr, _, err := uc.ValidateAddress(context.Background(), tt.input)
            
            if tt.expectErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.mockAddr.City, addr.City)
            }
        })
    }
}
```

---

## Step 10: Run and Test

```bash
# Build
make build

# Run tests
go test ./...

# Start server
./build/address-validation-service

# Test endpoint
curl -X POST http://localhost:8080/api/v1/validate-address \
  -H "Content-Type: application/json" \
  -d '{"address":"123 Main St New York NY 10001"}'
```

---

## Summary

You've now implemented:

1. ✅ Domain layer (Address entity, validation rules, error types)
2. ✅ Repository interface (abstraction for address parsing)
3. ✅ Usecase layer (business logic, confidence scoring)
4. ✅ Infrastructure layer (gopostal implementation)
5. ✅ HTTP handler (request/response mapping, error handling)
6. ✅ DTOs (request/response structures)
7. ✅ Route registration (in main.go)
8. ✅ Test coverage (unit tests with mocks)

**Next Steps**: Write integration tests with real gopostal, add E2E tests, implement candidate generation and ranking logic for ambiguous addresses.
