package json

import (
	"errors"
	"testing"

	knifer "github.com/imajinyun/go-knifer"
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
