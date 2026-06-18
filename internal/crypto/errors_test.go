package crypto

import (
	"errors"
	"testing"

	knifer "github.com/imajinyun/go-knifer"
)

func TestSentinelErrorCode(t *testing.T) {
	if got := ErrInvalidKey.(*sentinel).ErrorCode(); got != knifer.ErrCodeInvalidInput {
		t.Fatalf("ErrorCode = %q", got)
	}
}

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code knifer.ErrCode
		msg  string
	}{
		{"ErrInvalidKey", ErrInvalidKey, knifer.ErrCodeInvalidInput, "invalid key"},
		{"ErrInvalidIV", ErrInvalidIV, knifer.ErrCodeInvalidInput, "invalid iv"},
		{"ErrInvalidCipherText", ErrInvalidCipherText, knifer.ErrCodeInvalidInput, "invalid cipher text"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.msg {
				t.Fatalf("Error() = %q, want %q", got, tt.msg)
			}
			if !errors.Is(tt.err, tt.code) {
				t.Fatalf("errors.Is(%v, %s) = false", tt.err, tt.code)
			}
		})
	}
}

func TestValidateAESIV(t *testing.T) {
	if err := ValidateAESIV(make([]byte, 16)); err != nil {
		t.Fatalf("ValidateAESIV(16) = %v", err)
	}
	if err := ValidateAESIV(make([]byte, 0)); !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("ValidateAESIV(0) = %v", err)
	}
	if err := ValidateAESIV(nil); !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("ValidateAESIV(nil) = %v", err)
	}
}

func TestValidateAESGCMNonce(t *testing.T) {
	if err := ValidateAESGCMNonce(make([]byte, 12)); err != nil {
		t.Fatalf("ValidateAESGCMNonce(12) = %v", err)
	}
	if err := ValidateAESGCMNonce(make([]byte, 0)); !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("ValidateAESGCMNonce(0) = %v", err)
	}
	if err := ValidateAESGCMNonce(nil); !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("ValidateAESGCMNonce(nil) = %v", err)
	}
}
