//go:build !gopostal

package address_parser

import (
	"context"
	"regexp"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/williandandrade/address-validation-service/internal/domain/entity"
	domainerrors "github.com/williandandrade/address-validation-service/internal/domain/errors"
)

var (
	titleCaser = cases.Title(language.AmericanEnglish)
	zipPattern = regexp.MustCompile(`\b(\d{5}(?:-\d{4})?)\b`)
)

// GopostalParser implements ValidateAddressRepository using regex-based parsing.
// This is the fallback parser when gopostal/libpostal is not available.
// In production (Docker), the gopostal build tag enables the real gopostal parser.
type GopostalParser struct{}

// NewGopostalParser creates a new GopostalParser.
func NewGopostalParser() *GopostalParser {
	return &GopostalParser{}
}

// ParseAddress parses a raw address string using regex-based parsing.
func (p *GopostalParser) ParseAddress(
	_ context.Context,
	rawAddress string,
) (primary *entity.Address, candidates []*entity.Address, err error) {
	cleaned := normalizeWhitespace(rawAddress)

	components := p.extractComponents(cleaned)

	addr := &entity.Address{
		StreetAddress:      components["street"],
		City:               components["city"],
		State:              components["state"],
		PostalCode:         components["postal_code"],
		AddressType:        DetectAddressType(rawAddress),
		CorrectionsApplied: TrackCorrections(rawAddress, components),
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

func (p *GopostalParser) extractComponents(address string) map[string]string {
	components := make(map[string]string)
	remaining := address

	// Extract ZIP code
	if match := zipPattern.FindString(remaining); match != "" {
		components["postal_code"] = match
		remaining = strings.Replace(remaining, match, "", 1)
	}

	// Clean up remaining after ZIP removal
	remaining = normalizeWhitespace(remaining)

	// Remove trailing/leading commas and clean up
	remaining = strings.Trim(remaining, ", ")

	// Split by comma or multiple spaces
	parts := splitAddress(remaining)

	if len(parts) == 0 {
		return components
	}

	// Try to identify state (last 2-letter word that matches a state code)
	for i := len(parts) - 1; i >= 0; i-- {
		part := strings.TrimSpace(parts[i])
		words := strings.Fields(part)

		for j := len(words) - 1; j >= 0; j-- {
			word := strings.ToUpper(words[j])
			if len(word) == 2 && entity.ValidUSStates[word] {
				components["state"] = word

				// Remove state from parts
				words = append(words[:j], words[j+1:]...)
				parts[i] = strings.Join(words, " ")
				if strings.TrimSpace(parts[i]) == "" {
					parts = append(parts[:i], parts[i+1:]...)
				}
				goto stateFound
			}
		}

		// Try full state names
		for name, code := range stateNameToCode {
			if strings.EqualFold(strings.TrimSpace(part), name) {
				components["state"] = code
				parts = append(parts[:i], parts[i+1:]...)
				goto stateFound
			}
		}
	}

stateFound:

	// Assign remaining parts
	switch len(parts) {
	case 0:
		// Nothing left
	case 1:
		part := strings.TrimSpace(parts[0])
		if looksLikeStreet(part) {
			components["street"] = titleCaser.String(strings.ToLower(part))
		} else {
			components["city"] = titleCaser.String(strings.ToLower(part))
		}
	case 2:
		components["street"] = titleCaser.String(strings.ToLower(strings.TrimSpace(parts[0])))
		components["city"] = titleCaser.String(strings.ToLower(strings.TrimSpace(parts[1])))
	default:
		// First part is street, second is city, rest might be additional info
		components["street"] = titleCaser.String(strings.ToLower(strings.TrimSpace(parts[0])))
		components["city"] = titleCaser.String(strings.ToLower(strings.TrimSpace(parts[1])))
	}

	return components
}

func splitAddress(s string) []string {
	// Split by comma first
	if strings.Contains(s, ",") {
		parts := strings.Split(s, ",")
		var result []string
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result
	}

	// No commas: try to split intelligently
	// Look for pattern: "street city state zip"
	// The street typically starts with a number
	words := strings.Fields(s)
	if len(words) <= 1 {
		return words
	}

	// Find where the city starts - after street type abbreviation or at a word boundary
	streetEnd := findStreetEnd(words)
	if streetEnd > 0 && streetEnd < len(words) {
		street := strings.Join(words[:streetEnd], " ")
		city := strings.Join(words[streetEnd:], " ")
		return []string{street, city}
	}

	return []string{s}
}

var streetSuffixes = map[string]bool{
	"st": true, "street": true, "ave": true, "avenue": true,
	"blvd": true, "boulevard": true, "dr": true, "drive": true,
	"ln": true, "lane": true, "rd": true, "road": true,
	"ct": true, "court": true, "pl": true, "place": true,
	"way": true, "cir": true, "circle": true, "pkwy": true,
	"parkway": true, "ter": true, "terrace": true, "trl": true,
	"trail": true, "hwy": true, "highway": true,
}

func findStreetEnd(words []string) int {
	for i, word := range words {
		lower := strings.ToLower(strings.TrimRight(word, ".,"))
		if streetSuffixes[lower] {
			return i + 1
		}
	}
	return 0
}

func looksLikeStreet(s string) bool {
	words := strings.Fields(s)
	if len(words) == 0 {
		return false
	}
	// Streets typically start with a number
	if words[0] != "" && words[0][0] >= '0' && words[0][0] <= '9' {
		return true
	}
	return false
}

func normalizeWhitespace(s string) string {
	s = strings.TrimSpace(s)
	// Collapse multiple spaces
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return s
}
