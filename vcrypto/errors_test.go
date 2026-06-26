package vcrypto_test

import (
	"errors"
	"testing"

	"github.com/imajinyun/knifer-go"
	"github.com/imajinyun/knifer-go/vcrypto"
)

func TestErrorContract(t *testing.T) {
	if err := vcrypto.ValidateAESKey([]byte("too-short")); !errors.Is(err, knifer.ErrCodeInvalidInput) || !errors.Is(err, vcrypto.ErrInvalidKey) {
		t.Fatalf("ValidateAESKey error should match root code and domain sentinel: %v", err)
	}
	if err := vcrypto.ValidateAESIV([]byte("bad")); !errors.Is(err, knifer.ErrCodeInvalidInput) || !errors.Is(err, vcrypto.ErrInvalidIV) {
		t.Fatalf("ValidateAESIV error should match root code and domain sentinel: %v", err)
	}
	if err := vcrypto.ValidateAESGCMNonce([]byte("bad")); !errors.Is(err, knifer.ErrCodeInvalidInput) || !errors.Is(err, vcrypto.ErrInvalidIV) {
		t.Fatalf("ValidateAESGCMNonce error should match root code and domain sentinel: %v", err)
	}
	if code, ok := knifer.CodeOf(vcrypto.ErrInvalidCipherText); !ok || code != knifer.ErrCodeInvalidInput {
		t.Fatalf("CodeOf(ErrInvalidCipherText) = %q, %v", code, ok)
	}
	if err := vcrypto.ValidateAESKey([]byte("1234567890123456")); err != nil {
		t.Fatalf("ValidateAESKey(valid) = %v", err)
	}
	if err := vcrypto.ValidateAESIV([]byte("1234567890123456")); err != nil {
		t.Fatalf("ValidateAESIV(valid) = %v", err)
	}
	if err := vcrypto.ValidateAESGCMNonce([]byte("123456789012")); err != nil {
		t.Fatalf("ValidateAESGCMNonce(valid) = %v", err)
	}
}
