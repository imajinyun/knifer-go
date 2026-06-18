package jwt

import (
	"errors"
	"testing"

	knifer "github.com/imajinyun/go-knifer"
)

func TestJWTErrorError(t *testing.T) {
	err := NewJWTError("bad token")
	if err.Error() != "bad token" {
		t.Fatalf("Error() = %q, want %q", err.Error(), "bad token")
	}
}

func TestJWTErrorErrorWithCause(t *testing.T) {
	err := wrapJWTError(errors.New("decode failed"), "parse header")
	if err.Error() != "parse header: decode failed" {
		t.Fatalf("Error() = %q, want %q", err.Error(), "parse header: decode failed")
	}
}

func TestJWTErrorErrorCode(t *testing.T) {
	err := NewJWTError("bad")
	if err.ErrorCode() != knifer.ErrCodeInvalidInput {
		t.Fatalf("ErrorCode = %v, want %v", err.ErrorCode(), knifer.ErrCodeInvalidInput)
	}
}

func TestJWTErrorErrorCodeNil(t *testing.T) {
	var e *JWTError
	if e.ErrorCode() != "" {
		t.Fatal("nil JWTError.ErrorCode should be empty")
	}
}

func TestJWTErrorUnwrap(t *testing.T) {
	cause := errors.New("root cause")
	err := wrapJWTError(cause, "wrapped")
	if !errors.Is(err, cause) {
		t.Fatal("should unwrap to cause")
	}
	if unwrapped := err.Unwrap(); unwrapped != cause {
		t.Fatalf("Unwrap = %v, want %v", unwrapped, cause)
	}
}

func TestJWTErrorUnwrapNil(t *testing.T) {
	var e *JWTError
	if e.Unwrap() != nil {
		t.Fatal("nil Unwrap should return nil")
	}
}

func TestJWTErrorIs(t *testing.T) {
	err := NewJWTError("bad")
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatal("should match ErrCodeInvalidInput")
	}
	if errors.Is(err, knifer.ErrCodeNotFound) {
		t.Fatal("should not match ErrCodeNotFound")
	}
}

func TestJWTErrorIsNil(t *testing.T) {
	var e *JWTError
	if errors.Is(e, knifer.ErrCodeInvalidInput) {
		t.Fatal("nil JWTError should not match any code")
	}
}

func TestJWTErrorIsTargetNil(t *testing.T) {
	err := NewJWTError("bad")
	if errors.Is(err, nil) {
		t.Fatal("JWTError should not match nil target")
	}
}

func TestJWTErrorf(t *testing.T) {
	err := JWTErrorf("bad %s", "algorithm")
	if err.Error() != "bad algorithm" {
		t.Fatalf("JWTErrorf = %q, want %q", err.Error(), "bad algorithm")
	}
	if err.ErrorCode() != knifer.ErrCodeInvalidInput {
		t.Fatalf("ErrorCode = %v, want %v", err.ErrorCode(), knifer.ErrCodeInvalidInput)
	}
}

func TestJWTErrorUnsupported(t *testing.T) {
	err := unsupportedJWTErrorf("unsupported %s", "alg")
	if err.Error() != "unsupported alg" {
		t.Fatalf("unsupportedJWTErrorf = %q, want %q", err.Error(), "unsupported alg")
	}
	if err.ErrorCode() != knifer.ErrCodeUnsupported {
		t.Fatalf("ErrorCode = %v, want %v", err.ErrorCode(), knifer.ErrCodeUnsupported)
	}
}