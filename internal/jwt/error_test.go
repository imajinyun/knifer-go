package jwt

import (
	"errors"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestJWTErrorMatchesErrCode(t *testing.T) {
	if !errors.Is(NewJWTError("bad token"), knifer.ErrCodeInvalidInput) {
		t.Fatal("NewJWTError should match knifer.ErrCodeInvalidInput")
	}
	if !errors.Is(JWTErrorf("bad %s", "alg"), knifer.ErrCodeInvalidInput) {
		t.Fatal("JWTErrorf should match knifer.ErrCodeInvalidInput")
	}
	if errors.Is(NewJWTError("bad token"), knifer.ErrCodeNotFound) {
		t.Fatal("should not match an unrelated code")
	}
	code, ok := knifer.CodeOf(NewJWTError("bad token"))
	if !ok || code != knifer.ErrCodeInvalidInput {
		t.Fatalf("CodeOf(NewJWTError) = %q, %v; want invalid input", code, ok)
	}
}

func TestJWTWrappedAndUnsupportedErrors(t *testing.T) {
	cause := errors.New("decode failed")
	err := wrapJWTError(cause, "decode header")
	if !errors.Is(err, cause) {
		t.Fatal("wrapJWTError should unwrap cause")
	}
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatal("wrapJWTError should match invalid input")
	}
	code, ok := knifer.CodeOf(err)
	if !ok || code != knifer.ErrCodeInvalidInput {
		t.Fatalf("CodeOf(wrapJWTError) = %q, %v; want invalid input", code, ok)
	}

	unsupported := unsupportedJWTErrorf("unsupported algorithm")
	if !errors.Is(unsupported, knifer.ErrCodeUnsupported) {
		t.Fatal("unsupportedJWTErrorf should match unsupported code")
	}
	code, ok = knifer.CodeOf(unsupported)
	if !ok || code != knifer.ErrCodeUnsupported {
		t.Fatalf("CodeOf(unsupportedJWTErrorf) = %q, %v; want unsupported", code, ok)
	}
}

func TestJWTErrorIsSameTypeAndNil(t *testing.T) {
	err := NewJWTError("bad token") // ErrCodeInvalidInput
	if !err.Is(&JWTError{Code: knifer.ErrCodeInvalidInput}) {
		t.Fatal("Is(*JWTError same code) should match")
	}
	if err.Is(&JWTError{Code: knifer.ErrCodeInternal}) {
		t.Fatal("Is(*JWTError other code) should not match")
	}
	if err.Is(errors.New("x")) || err.Is(nil) {
		t.Fatal("Is should not match unrelated targets")
	}

	var e *JWTError
	if e.Error() != "" || e.Is(knifer.ErrCodeInvalidInput) {
		t.Fatal("nil *JWTError methods should be zero-valued and safe")
	}
}
