package date

import (
	"errors"
	"testing"

	knifer "github.com/imajinyun/go-knifer"
)

func TestDateErrorContract(t *testing.T) {
	_, err := ParseDate("")
	assertDateCode(t, err, knifer.ErrCodeInvalidInput)

	_, err = ParseDate("not-a-date")
	assertDateCode(t, err, knifer.ErrCodeInvalidInput)

	_, err = ParseDateLayout("2026-06-05", "bad-layout")
	assertDateCode(t, err, knifer.ErrCodeInvalidInput)
}

func assertDateCode(t *testing.T, err error, code knifer.ErrCode) {
	t.Helper()
	if err == nil {
		t.Fatalf("err = nil, want %s", code)
	}
	if !errors.Is(err, code) {
		t.Fatalf("errors.Is(%v, %s) = false", err, code)
	}
	got, ok := knifer.CodeOf(err)
	if !ok || got != code {
		t.Fatalf("CodeOf(%v) = %q, %v; want %q, true", err, got, ok, code)
	}
	var dateErr *DateError
	if !errors.As(err, &dateErr) {
		t.Fatalf("errors.As(err, *DateError) = false: %v", err)
	}
}

func TestDateErrorMethods(t *testing.T) {
	// nil receiver
	var e *DateError
	if got := e.Error(); got != "" {
		t.Fatalf("nil Error = %q, want empty", got)
	}
	if got := e.ErrorCode(); got != "" {
		t.Fatalf("nil ErrorCode = %q, want empty", got)
	}
	if got := e.Unwrap(); got != nil {
		t.Fatalf("nil Unwrap = %v, want nil", got)
	}
	if e.Is(nil) {
		t.Fatal("nil Is(nil) should be false")
	}

	// without cause
	e1 := invalidDateInputf("bad input")
	if got := e1.Error(); got != "bad input" {
		t.Fatalf("Error without cause = %q", got)
	}

	// with cause
	e2 := wrapDateError(knifer.ErrCodeInvalidInput, "parse error", errors.New("underlying"))
	if got := e2.Error(); got != "parse error: underlying" {
		t.Fatalf("Error with cause = %q", got)
	}
	var de2 *DateError
	if !errors.As(e2, &de2) {
		t.Fatal("wrapDateError returned non-DateError")
	}
	if got := de2.Unwrap(); got == nil || got.Error() != "underlying" {
		t.Fatalf("Unwrap = %v, want underlying error", got)
	}

	// Is matching by ErrCode
	if !errors.Is(e2, knifer.ErrCodeInvalidInput) {
		t.Fatal("errors.Is(err, ErrCodeInvalidInput) should be true")
	}
	if errors.Is(e2, knifer.ErrCodeInternal) {
		t.Fatal("errors.Is(err, ErrCodeInternal) should be false")
	}

	// Is matching by *DateError
	if !errors.Is(e2, &DateError{Code: knifer.ErrCodeInvalidInput}) {
		t.Fatal("errors.Is(err, *DateError with same code) should be true")
	}
	if errors.Is(e2, &DateError{Code: knifer.ErrCodeInternal}) {
		t.Fatal("errors.Is(err, *DateError with different code) should be false")
	}

	// Is with nil target
	targetErr := errors.New("unrelated")
	if errors.Is(e2, targetErr) {
		t.Fatal("errors.Is(err, unrelated) should be false")
	}
}

func TestDateErrorNilUnwrapReturnsNil(t *testing.T) {
	// wrapDateError with nil cause returns nil
	if got := wrapDateError(knifer.ErrCodeInvalidInput, "msg", nil); got != nil {
		t.Fatal("wrapDateError with nil cause should return nil")
	}
}
