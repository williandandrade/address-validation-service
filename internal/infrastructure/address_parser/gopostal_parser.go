//go:build cgo && gopostal

package address_parser

import (
	"context"
	"strings"

	"github.com/openvenues/gopostal/parser"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/williandandrade/address-validation-service/internal/domain/entity"
	domainerrors "github.com/williandandrade/address-validation-service/internal/domain/errors"
)

var titleCaser = cases.Title(language.AmericanEnglish)

// GopostalParser implements ValidateAddressRepository using gopostal library.
type GopostalParser struct{}

// NewGopostalParser creates a new GopostalParser.
func NewGopostalParser() *GopostalParser {
	return &GopostalParser{}
}

// ParseAddress parses a raw address string using gopostal and returns a normalized Address.
func (p *GopostalParser) ParseAddress(
	_ context.Context,
	rawAddress string,
) (*entity.Address, []*entity.Address, error) {
	parsed := parser.ParseAddress(rawAddress)

	components := make(map[string]string)
	for _, comp := range parsed {
		components[comp.Label] = comp.Value
	}

	addr := &entity.Address{
		StreetAddress:      p.buildStreet(components),
		City:               p.normalizeCity(components),
		State:              p.normalizeState(components),
		PostalCode:         p.extractPostalCode(components),
		AddressType:        p.detectAddressType(rawAddress),
		CorrectionsApplied: p.trackCorrections(rawAddress, components),
	}

	if addr.StreetAddress == "" && addr.City == "" && addr.State == "" {
		return nil, nil, &domainerrors.ParsingError{
			Field:      "address",
			Reason:     "Could not extract required address components",
			Suggestion: "Ensure address contains street address, city, and state",
		}
	}

	return addr, nil, nil
}

func (p *GopostalParser) buildStreet(components map[string]string) string {
	var parts []string

	if num, ok := components["house_number"]; ok && num != "" {
		parts = append(parts, num)
	}
	if road, ok := components["road"]; ok && road != "" {
		parts = append(parts, titleCaser.String(strings.ToLower(road)))
	}
	if unit, ok := components["unit"]; ok && unit != "" {
		parts = append(parts, unit)
	}

	return strings.TrimSpace(strings.Join(parts, " "))
}

func (p *GopostalParser) normalizeCity(components map[string]string) string {
	if city, ok := components["city"]; ok && city != "" {
		return titleCaser.String(strings.ToLower(city))
	}
	return ""
}

func (p *GopostalParser) normalizeState(components map[string]string) string {
	if state, ok := components["state"]; ok && state != "" {
		upper := strings.ToUpper(strings.TrimSpace(state))
		if entity.ValidUSStates[upper] {
			return upper
		}
		if code, found := stateNameToCode[strings.ToLower(strings.TrimSpace(state))]; found {
			return code
		}
		return upper
	}
	return ""
}

func (p *GopostalParser) extractPostalCode(components map[string]string) string {
	if postal, ok := components["postcode"]; ok && postal != "" {
		return strings.TrimSpace(postal)
	}
	return ""
}

func (p *GopostalParser) detectAddressType(rawAddress string) string {
	return DetectAddressType(rawAddress)
}

func (p *GopostalParser) trackCorrections(rawAddress string, components map[string]string) []string {
	return TrackCorrections(rawAddress, components)
}
