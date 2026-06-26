package errx

import (
	"errors"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestPanicErrorPreservesErrorValues(t *testing.T) {
	want := errors.New("panic error")
	got := panicError(want)
	if !errors.Is(got, want) {
		t.Fatalf("panicError(error) = %v, want wrapping original error", got)
	}
	var pe *PanicError
	if !errors.As(got, &pe) {
		t.Fatalf("panicError(error) type = %T, want *PanicError", got)
	}
	if pe.Value != want || pe.Cause != want {
		t.Fatalf("PanicError value/cause = (%v, %v), want original error", pe.Value, pe.Cause)
	}
	if pe.Stack() == "" {
		t.Fatal("PanicError should capture a stack")
	}
	if got := panicError("panic string"); got == nil || got.Error() != "panic string" {
		t.Fatalf("panicError(string) = %v, want converted error", got)
	} else if !errors.Is(got, knifer.ErrCodeInternal) {
		t.Fatalf("panicError(string) = %v, want ErrCodeInternal", got)
	}

	coded := panicError(knifer.NewError(knifer.ErrCodeInvalidInput, "bad input"))
	if !errors.Is(coded, knifer.ErrCodeInvalidInput) {
		t.Fatalf("panicError(coded error) = %v, want ErrCodeInvalidInput", coded)
	}
	if code, ok := knifer.CodeOf(coded); !ok || code != knifer.ErrCodeInvalidInput {
		t.Fatalf("CodeOf(panicError(coded error)) = %q, %v; want invalid input", code, ok)
	}
}

func TestPanicErrorNilReceiver(t *testing.T) {
	var pe *PanicError
	if pe.Error() != "<nil>" || pe.Unwrap() != nil || pe.Stack() != "" || pe.ErrorCode() != "" {
		t.Fatalf("nil PanicError methods returned unexpected values")
	}
}
