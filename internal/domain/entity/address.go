package entity

import (
	"fmt"
	"regexp"
	"strings"
)

// Address represents a normalized, validated US address.
type Address struct {
	StreetAddress      string      `json:"street_address"`
	City               string      `json:"city"`
	State              string      `json:"state"`
	PostalCode         string      `json:"postal_code"`
	AddressType        string      `json:"address_type"`
	FormattedAddress   string      `json:"formatted_address,omitempty"`
	Confidence         *Confidence `json:"confidence,omitempty"`
	CorrectionsApplied []string    `json:"corrections_applied,omitempty"`
}

// Confidence tracks the source of each address component.
type Confidence struct {
	StateConfidence  string `json:"state_confidence"`
	CityConfidence   string `json:"city_confidence"`
	PostalConfidence string `json:"postal_confidence"`
}

var zipRegex = regexp.MustCompile(`^\d{5}(-\d{4})?$`)

// ValidUSStates contains all valid 2-letter USPS state codes.
var ValidUSStates = map[string]bool{
	"AL": true, "AK": true, "AZ": true, "AR": true, "CA": true,
	"CO": true, "CT": true, "DE": true, "DC": true, "FL": true,
	"GA": true, "HI": true, "ID": true, "IL": true, "IN": true,
	"IA": true, "KS": true, "KY": true, "LA": true, "ME": true,
	"MD": true, "MA": true, "MI": true, "MN": true, "MS": true,
	"MO": true, "MT": true, "NE": true, "NV": true, "NH": true,
	"NJ": true, "NM": true, "NY": true, "NC": true, "ND": true,
	"OH": true, "OK": true, "OR": true, "PA": true, "RI": true,
	"SC": true, "SD": true, "TN": true, "TX": true, "UT": true,
	"VT": true, "VA": true, "WA": true, "WV": true, "WI": true,
	"WY": true,
}

// Validate checks that Address meets minimum requirements.
func (a *Address) Validate() error {
	presentCount := 0
	if a.StreetAddress != "" {
		presentCount++
	}
	if a.City != "" {
		presentCount++
	}
	if a.State != "" {
		presentCount++
	}

	if presentCount < 2 {
		return fmt.Errorf("at least 2 of street_address, city, state must be present")
	}

	if a.State != "" && !ValidUSStates[strings.ToUpper(a.State)] {
		return fmt.Errorf("invalid state code: %s", a.State)
	}

	if a.PostalCode != "" && !zipRegex.MatchString(a.PostalCode) {
		return fmt.Errorf("postal_code must be 5 or 9-digit format")
	}

	return nil
}

// FormatAddress generates a human-readable formatted address string.
func (a *Address) FormatAddress() string {
	if a.FormattedAddress != "" {
		return a.FormattedAddress
	}

	parts := []string{}
	if a.StreetAddress != "" {
		parts = append(parts, a.StreetAddress)
	}
	if a.City != "" {
		parts = append(parts, a.City)
	}

	stateZip := ""
	if a.State != "" {
		stateZip = a.State
	}
	if a.PostalCode != "" {
		if stateZip != "" {
			stateZip += " " + a.PostalCode
		} else {
			stateZip = a.PostalCode
		}
	}
	if stateZip != "" {
		parts = append(parts, stateZip)
	}

	a.FormattedAddress = strings.Join(parts, ", ")
	return a.FormattedAddress
}
