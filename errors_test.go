package knifer_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/imajinyun/knifer-go"
)

func TestErrorCodeMatching(t *testing.T) {
	err := knifer.NewError(knifer.ErrCodeInvalidInput, "url empty")
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatal("errors.Is should match the error code")
	}
	if errors.Is(err, knifer.ErrCodeNotFound) {
		t.Fatal("errors.Is should not match a different code")
	}
}

func TestErrorWrapPreservesChain(t *testing.T) {
	cause := errors.New("disk full")
	err := knifer.WrapError(knifer.ErrCodeInternal, "write failed", cause)

	if !errors.Is(err, knifer.ErrCodeInternal) {
		t.Fatal("wrapped error should match its code")
	}
	if !errors.Is(err, cause) {
		t.Fatal("wrapped error should preserve the cause chain")
	}
	if got := errors.Unwrap(err); got != cause {
		t.Fatalf("Unwrap = %v, want %v", got, cause)
	}
}

func TestErrorAs(t *testing.T) {
	err := fmt.Errorf("context: %w", knifer.Errorf(knifer.ErrCodeTimeout, "deadline %ds", 3))
	var ke *knifer.Error
	if !errors.As(err, &ke) {
		t.Fatal("errors.As should extract *knifer.Error")
	}
	if ke.Code != knifer.ErrCodeTimeout {
		t.Fatalf("Code = %q, want %q", ke.Code, knifer.ErrCodeTimeout)
	}
}

func TestCodeOf(t *testing.T) {
	if code, ok := knifer.CodeOf(knifer.NewError(knifer.ErrCodeNotFound, "missing")); !ok || code != knifer.ErrCodeNotFound {
		t.Fatalf("CodeOf(NewError) = %q, %v", code, ok)
	}

	wrapped := fmt.Errorf("context: %w", knifer.WrapError(knifer.ErrCodeInternal, "boom", errors.New("root")))
	if code, ok := knifer.CodeOf(wrapped); !ok || code != knifer.ErrCodeInternal {
		t.Fatalf("CodeOf(wrapped) = %q, %v", code, ok)
	}

	if code, ok := knifer.CodeOf(nil); ok || code != "" {
		t.Fatalf("CodeOf(nil) = %q, %v", code, ok)
	}
}

func TestErrorString(t *testing.T) {
	if got := knifer.NewError(knifer.ErrCodeNotFound, "missing").Error(); got != "GK_NOT_FOUND: missing" {
		t.Fatalf("Error() = %q", got)
	}
	wrapped := knifer.WrapError(knifer.ErrCodeInternal, "boom", errors.New("root"))
	if got := wrapped.Error(); got != "GK_INTERNAL: boom: root" {
		t.Fatalf("Error() = %q", got)
	}
}

func TestErrorStringCauseWithoutMessage(t *testing.T) {
	// Cause set but empty message hits the "CODE: cause" branch.
	err := knifer.WrapError(knifer.ErrCodeTimeout, "", errors.New("deadline"))
	if got := err.Error(); got != "GK_TIMEOUT: deadline" {
		t.Fatalf("Error() = %q", got)
	}
}

func TestErrorNilReceiver(t *testing.T) {
	var e *knifer.Error
	if got := e.Error(); got != "" {
		t.Fatalf("nil Error() = %q", got)
	}
	if got := e.Unwrap(); got != nil {
		t.Fatalf("nil Unwrap() = %v", got)
	}
	if got := e.ErrorCode(); got != "" {
		t.Fatalf("nil ErrorCode() = %q", got)
	}
	if e.Is(knifer.ErrCodeInternal) {
		t.Fatal("nil Is() should be false")
	}
}

func TestErrorCodeAndIsBranches(t *testing.T) {
	err := knifer.NewError(knifer.ErrCodeNotFound, "missing")
	if err.ErrorCode() != knifer.ErrCodeNotFound {
		t.Fatalf("ErrorCode() = %q", err.ErrorCode())
	}
	// Is against another *Error with the same code.
	if !err.Is(knifer.NewError(knifer.ErrCodeNotFound, "other")) {
		t.Fatal("Is(*Error same code) should match")
	}
	// Is against a different-code *Error and an unrelated error returns false.
	if err.Is(knifer.NewError(knifer.ErrCodeInternal, "x")) {
		t.Fatal("Is(*Error other code) should not match")
	}
	if err.Is(errors.New("plain")) {
		t.Fatal("Is(plain error) should not match")
	}
	if err.Is(nil) {
		t.Fatal("Is(nil) should not match")
	}
}

// codeCarrierEmpty implements CodeCarrier but returns an empty code, forcing
// CodeOf to fall back to the sentinel ErrCode scan.
type codeCarrierEmpty struct{}

func (codeCarrierEmpty) Error() string             { return "empty" }
func (codeCarrierEmpty) ErrorCode() knifer.ErrCode { return "" }

func TestCodeOfFallbacks(t *testing.T) {
	// Carrier returns empty code -> fall through to sentinel scan, which also
	// fails to match -> ("", false).
	if code, ok := knifer.CodeOf(codeCarrierEmpty{}); ok || code != "" {
		t.Fatalf("CodeOf(empty carrier) = %q, %v", code, ok)
	}

	// A bare ErrCode value matches itself through the sentinel scan.
	if code, ok := knifer.CodeOf(knifer.ErrCodeUnsupported); !ok || code != knifer.ErrCodeUnsupported {
		t.Fatalf("CodeOf(ErrCode) = %q, %v", code, ok)
	}

	// An error with no knifer-go code yields ("", false).
	if code, ok := knifer.CodeOf(errors.New("plain")); ok || code != "" {
		t.Fatalf("CodeOf(plain) = %q, %v", code, ok)
	}
}

func TestUnifiedErrorTaxonomyCodes(t *testing.T) {
	tests := []struct {
		name string
		code knifer.ErrCode
	}{
		{name: "invalid input", code: knifer.ErrCodeInvalidInput},
		{name: "not found", code: knifer.ErrCodeNotFound},
		{name: "unsupported type", code: knifer.ErrCodeUnsupported},
		{name: "unsafe resource", code: knifer.ErrCodeUnsafeResource},
		{name: "timeout", code: knifer.ErrCodeTimeout},
		{name: "provider failure", code: knifer.ErrCodeProviderFailure},
		{name: "internal", code: knifer.ErrCodeInternal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fmt.Errorf("boundary: %w", knifer.WrapError(tt.code, tt.name, errors.New("cause")))
			if !errors.Is(err, tt.code) {
				t.Fatalf("errors.Is(%v, %v) = false", err, tt.code)
			}
			if code, ok := knifer.CodeOf(err); !ok || code != tt.code {
				t.Fatalf("CodeOf(%v) = %q, %v", err, code, ok)
			}
		})
	}
}
