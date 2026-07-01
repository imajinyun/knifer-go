package crypto

import (
	"errors"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
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
		{"ErrInvalidSM2Signature", ErrInvalidSM2Signature, knifer.ErrCodeInvalidInput, "invalid sm2 signature"},
		{"ErrInvalidOTP", ErrInvalidOTP, knifer.ErrCodeInvalidInput, "invalid otp"},
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

func TestValidateSM4Key(t *testing.T) {
	if err := ValidateSM4Key(make([]byte, 16)); err != nil {
		t.Fatalf("ValidateSM4Key(16) = %v", err)
	}
	for _, size := range []int{0, 15, 17} {
		if err := ValidateSM4Key(make([]byte, size)); !errors.Is(err, ErrInvalidKey) || !errors.Is(err, knifer.ErrCodeInvalidInput) {
			t.Fatalf("ValidateSM4Key(%d) = %v, want invalid key/input", size, err)
		}
	}
}

func TestValidateSM4IV(t *testing.T) {
	if err := ValidateSM4IV(make([]byte, 16)); err != nil {
		t.Fatalf("ValidateSM4IV(16) = %v", err)
	}
	if err := ValidateSM4IV(make([]byte, 12)); !errors.Is(err, ErrInvalidIV) || !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("ValidateSM4IV(12) = %v, want invalid iv/input", err)
	}
}

func TestValidateAESKey(t *testing.T) {
	for _, size := range []int{16, 24, 32} {
		if err := ValidateAESKey(make([]byte, size)); err != nil {
			t.Fatalf("ValidateAESKey(%d) = %v", size, err)
		}
	}
	for _, size := range []int{0, 15, 17, 33} {
		if err := ValidateAESKey(make([]byte, size)); !errors.Is(err, ErrInvalidKey) || !errors.Is(err, knifer.ErrCodeInvalidInput) {
			t.Fatalf("ValidateAESKey(%d) = %v, want invalid key/input", size, err)
		}
	}
}

func TestSentinelIsRejectsUnrelatedErrors(t *testing.T) {
	if errors.Is(ErrInvalidKey, errors.New("invalid key")) {
		t.Fatal("ErrInvalidKey should not match unrelated error with same message")
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
