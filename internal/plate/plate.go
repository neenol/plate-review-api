package plate

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"

	"github.com/nglambertjr/plate-review-api/internal/domain"
)

var validStates = map[string]bool{
	"AL": true, "AK": true, "AZ": true, "AR": true, "CA": true,
	"CO": true, "CT": true, "DE": true, "FL": true, "GA": true,
	"HI": true, "ID": true, "IL": true, "IN": true, "IA": true,
	"KS": true, "KY": true, "LA": true, "ME": true, "MD": true,
	"MA": true, "MI": true, "MN": true, "MS": true, "MO": true,
	"MT": true, "NE": true, "NV": true, "NH": true, "NJ": true,
	"NM": true, "NY": true, "NC": true, "ND": true, "OH": true,
	"OK": true, "OR": true, "PA": true, "RI": true, "SC": true,
	"SD": true, "TN": true, "TX": true, "UT": true, "VT": true,
	"VA": true, "WA": true, "WV": true, "WI": true, "WY": true,
	"DC": true,
}

var nonAlphanumeric = regexp.MustCompile(`[^A-Z0-9]`)

// Normalize uppercases and strips non-alphanumeric characters, then validates.
// Returns (normalizedPlate, nil) or ("", domain.ErrInvalidInput).
func Normalize(state, raw string) (string, error) {
	state = strings.ToUpper(strings.TrimSpace(state))
	if !validStates[state] {
		return "", fmt.Errorf("state %q: %w", state, domain.ErrInvalidInput)
	}

	normalized := nonAlphanumeric.ReplaceAllString(strings.ToUpper(raw), "")
	if len(normalized) == 0 || len(normalized) > 8 {
		return "", fmt.Errorf("plate %q: %w", raw, domain.ErrInvalidInput)
	}

	return normalized, nil
}

// Hash returns the sha256 hex of "STATE:NORMALIZED:PEPPER".
func Hash(state, normalized, pepper string) string {
	h := sha256.Sum256([]byte(state + ":" + normalized + ":" + pepper))
	return fmt.Sprintf("%x", h)
}

// NormalizeAndHash normalizes the plate then hashes it.
// Returns (normalized, hash, nil) or ("", "", domain.ErrInvalidInput).
func NormalizeAndHash(state, raw, pepper string) (string, string, error) {
	normalized, err := Normalize(state, raw)
	if err != nil {
		return "", "", err
	}
	return normalized, Hash(state, normalized, pepper), nil
}
