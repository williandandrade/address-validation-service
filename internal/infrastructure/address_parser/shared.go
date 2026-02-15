package address_parser

import "strings"

// stateNameToCode maps full state names to their 2-letter codes.
var stateNameToCode = map[string]string{
	"alabama": "AL", "alaska": "AK", "arizona": "AZ", "arkansas": "AR",
	"california": "CA", "colorado": "CO", "connecticut": "CT", "delaware": "DE",
	"district of columbia": "DC", "florida": "FL", "georgia": "GA", "hawaii": "HI",
	"idaho": "ID", "illinois": "IL", "indiana": "IN", "iowa": "IA",
	"kansas": "KS", "kentucky": "KY", "louisiana": "LA", "maine": "ME",
	"maryland": "MD", "massachusetts": "MA", "michigan": "MI", "minnesota": "MN",
	"mississippi": "MS", "missouri": "MO", "montana": "MT", "nebraska": "NE",
	"nevada": "NV", "new hampshire": "NH", "new jersey": "NJ", "new mexico": "NM",
	"new york": "NY", "north carolina": "NC", "north dakota": "ND", "ohio": "OH",
	"oklahoma": "OK", "oregon": "OR", "pennsylvania": "PA", "rhode island": "RI",
	"south carolina": "SC", "south dakota": "SD", "tennessee": "TN", "texas": "TX",
	"utah": "UT", "vermont": "VT", "virginia": "VA", "washington": "WA",
	"west virginia": "WV", "wisconsin": "WI", "wyoming": "WY",
}

// DetectAddressType classifies an address based on its content.
func DetectAddressType(rawAddress string) string {
	lower := strings.ToLower(rawAddress)

	if strings.Contains(lower, "po box") || strings.Contains(lower, "p.o. box") || strings.Contains(lower, "p.o.") {
		return "po_box"
	}
	if strings.Contains(lower, "apo") || strings.Contains(lower, "fpo") {
		return "apo_fpo"
	}
	if strings.Contains(lower, "rural route") || strings.Contains(lower, " rr ") {
		return "rural_route"
	}
	return "standard_street"
}

// TrackCorrections identifies normalization corrections applied to the input.
func TrackCorrections(rawAddress string, _ map[string]string) []string {
	var corrections []string

	if rawAddress != strings.TrimSpace(rawAddress) || strings.Contains(rawAddress, "  ") {
		corrections = append(corrections, "Normalized whitespace")
	}

	// Check if any word (excluding numbers, punctuation, and state codes) starts with lowercase
	// This indicates capitalization normalization was needed
	words := strings.Fields(rawAddress)
	for _, word := range words {
		cleaned := strings.Trim(word, ",.;:")
		if cleaned == "" {
			continue
		}
		// Skip numbers and 2-letter state codes (they get uppercased separately)
		if len(cleaned) <= 2 {
			continue
		}
		firstChar := rune(cleaned[0])
		if firstChar >= 'a' && firstChar <= 'z' {
			corrections = append(corrections, "Standardized capitalization")
			break
		}
	}

	return corrections
}
