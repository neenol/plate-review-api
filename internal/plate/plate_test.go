package plate_test

import (
	"errors"
	"testing"

	"github.com/nglambertjr/plate-review-api/internal/domain"
	"github.com/nglambertjr/plate-review-api/internal/plate"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name       string
		state      string
		raw        string
		want       string
		wantErrIs  error
	}{
		{name: "basic uppercase", state: "CA", raw: "abc123", want: "ABC123"},
		{name: "strip hyphens", state: "TX", raw: "ABC-123", want: "ABC123"},
		{name: "strip spaces", state: "NY", raw: "A B C", want: "ABC"},
		{name: "lowercase state", state: "ca", raw: "ABC123", want: "ABC123"},
		{name: "DC is valid", state: "DC", raw: "DC1234", want: "DC1234"},
		{name: "max length 8", state: "FL", raw: "ABCDEFGH", want: "ABCDEFGH"},
		{name: "invalid state", state: "XX", raw: "ABC123", wantErrIs: domain.ErrInvalidInput},
		{name: "empty after strip", state: "CA", raw: "---", wantErrIs: domain.ErrInvalidInput},
		{name: "too long", state: "CA", raw: "ABCDEFGHI", wantErrIs: domain.ErrInvalidInput},
		{name: "empty raw", state: "CA", raw: "", wantErrIs: domain.ErrInvalidInput},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := plate.Normalize(tc.state, tc.raw)
			if tc.wantErrIs != nil {
				if !errors.Is(err, tc.wantErrIs) {
					t.Fatalf("Normalize(%q, %q): want err %v, got %v", tc.state, tc.raw, tc.wantErrIs, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Normalize(%q, %q): unexpected error %v", tc.state, tc.raw, err)
			}
			if got != tc.want {
				t.Errorf("Normalize(%q, %q) = %q, want %q", tc.state, tc.raw, got, tc.want)
			}
		})
	}
}

func TestHash(t *testing.T) {
	h1 := plate.Hash("CA", "ABC123", "pepper1")
	h2 := plate.Hash("CA", "ABC123", "pepper1")
	h3 := plate.Hash("CA", "ABC123", "pepper2")
	h4 := plate.Hash("TX", "ABC123", "pepper1")

	if h1 != h2 {
		t.Error("same inputs must produce same hash")
	}
	if h1 == h3 {
		t.Error("different pepper must produce different hash")
	}
	if h1 == h4 {
		t.Error("different state must produce different hash")
	}
	if len(h1) != 64 {
		t.Errorf("hash length: want 64, got %d", len(h1))
	}
}

func TestNormalizeAndHash(t *testing.T) {
	norm, hash, err := plate.NormalizeAndHash("CA", "abc-123", "p")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if norm != "ABC123" {
		t.Errorf("normalized: want ABC123, got %s", norm)
	}
	expected := plate.Hash("CA", "ABC123", "p")
	if hash != expected {
		t.Errorf("hash mismatch")
	}
}
