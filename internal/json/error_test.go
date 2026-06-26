package json

import (
	"errors"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestJSONErrorMethods(t *testing.T) {
	e := NewJSONError("test msg %d", 1)
	if s := e.Error(); s != "test msg 1" {
		t.Fatalf("Error() = %q", s)
	}
	if ec := e.ErrorCode(); ec != knifer.ErrCodeInvalidInput {
		t.Fatalf("ErrorCode() = %q", ec)
	}
	if cause := e.Unwrap(); cause != nil {
		t.Fatalf("Unwrap() = %v", cause)
	}

	wrapped := WrapJSONError(errors.New("inner"), "outer")
	if s := wrapped.Error(); s != "outer: inner" {
		t.Fatalf("wrapped Error() = %q", s)
	}
}

func TestJSONErrorMatchesErrCode(t *testing.T) {
	if !errors.Is(NewJSONError("empty path"), knifer.ErrCodeInvalidInput) {
		t.Fatal("NewJSONError should match knifer.ErrCodeInvalidInput")
	}
	wrapped := WrapJSONError(errors.New("eof"), "parse failed")
	if !errors.Is(wrapped, knifer.ErrCodeInvalidInput) {
		t.Fatal("WrapJSONError should match knifer.ErrCodeInvalidInput")
	}
	if !errors.Is(wrapped, errors.Unwrap(wrapped)) {
		t.Fatal("WrapJSONError should preserve the cause chain")
	}
}

func TestJSONErrorContractAndNil(t *testing.T) {
	err := NewJSONError("bad")

	// CodeOf classifies through CodeCarrier.
	if code, ok := knifer.CodeOf(err); !ok || code != knifer.ErrCodeInvalidInput {
		t.Fatalf("CodeOf = %q, %v", code, ok)
	}
	// Is handles both ErrCode and same-type *JSONError targets.
	if !err.Is(&JSONError{Code: knifer.ErrCodeInvalidInput}) {
		t.Fatal("Is(*JSONError same code) should match")
	}
	if err.Is(knifer.ErrCodeInternal) || err.Is(errors.New("x")) || err.Is(nil) {
		t.Fatal("Is should not match unrelated targets")
	}

	// nil receiver safety.
	var e *JSONError
	if e.Error() != "" || e.ErrorCode() != "" || e.Unwrap() != nil || e.Is(knifer.ErrCodeInternal) {
		t.Fatal("nil *JSONError methods should be zero-valued and safe")
	}
}
