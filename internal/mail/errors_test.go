package mail

import (
	"errors"
	"fmt"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

// TestSentinelErrorContract verifies that mail sentinel errors participate in
// the knifer-go error contract while preserving sentinel identity.
func TestSentinelErrorContract(t *testing.T) {
	cases := []struct {
		err  error
		code knifer.ErrCode
	}{
		{ErrInvalidAddress, knifer.ErrCodeInvalidInput},
		{ErrInvalidHeader, knifer.ErrCodeInvalidInput},
		{ErrMissingFrom, knifer.ErrCodeInvalidInput},
		{ErrMissingRecipient, knifer.ErrCodeInvalidInput},
		{ErrMissingBody, knifer.ErrCodeInvalidInput},
		{ErrAttachmentTooLarge, knifer.ErrCodeInvalidInput},
		{ErrTLSRequired, knifer.ErrCodeUnsupported},
		{ErrPlainAuth, knifer.ErrCodeUnsupported},
	}
	for _, tt := range cases {
		t.Run(tt.err.Error(), func(t *testing.T) {
			// Sentinel identity is preserved for errors.Is.
			if !errors.Is(tt.err, tt.err) {
				t.Fatal("sentinel should match itself")
			}
			// Wrapped sentinels remain discoverable.
			wrapped := fmt.Errorf("context: %w", tt.err)
			if !errors.Is(wrapped, tt.err) {
				t.Fatal("wrapped sentinel should match via errors.Is")
			}
			// The error code is classified via the CodeCarrier contract.
			if code, ok := knifer.CodeOf(tt.err); !ok || code != tt.code {
				t.Fatalf("CodeOf = %q, %v; want %q", code, ok, tt.code)
			}
			// errors.Is against the bare ErrCode also matches.
			if !errors.Is(tt.err, tt.code) {
				t.Fatalf("errors.Is(%v, %s) = false", tt.err, tt.code)
			}
			if !errors.Is(wrapped, tt.code) {
				t.Fatalf("errors.Is(wrapped, %s) = false", tt.code)
			}
		})
	}
}

func TestSentinelDistinctIdentity(t *testing.T) {
	// Different sentinels with the same code must remain distinct identities.
	if errors.Is(ErrInvalidAddress, ErrInvalidHeader) {
		t.Fatal("distinct sentinels should not match by identity")
	}
	// But both still classify to the same code.
	c1, _ := knifer.CodeOf(ErrInvalidAddress)
	c2, _ := knifer.CodeOf(ErrInvalidHeader)
	if c1 != c2 {
		t.Fatalf("codes differ: %q vs %q", c1, c2)
	}
}
