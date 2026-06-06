package errx

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/go-multierror"
	knifer "github.com/imajinyun/go-knifer"
	"github.com/sirupsen/logrus"
)

type stackError struct{ stack string }

func (e stackError) Error() string { return "stack error" }
func (e stackError) Stack() string { return e.stack }

func TestGetStackUsesAttachedStackWhenAvailable(t *testing.T) {
	const want = "attached stack"
	if got := GetStack(stackError{stack: want}); got != want {
		t.Fatalf("GetStack() = %q, want %q", got, want)
	}
}

func TestGetStackFallsBackToRuntimeStack(t *testing.T) {
	got := GetStack(errors.New("plain"))
	if !strings.Contains(got, "goroutine") || !strings.Contains(got, "TestGetStackFallsBackToRuntimeStack") {
		t.Fatalf("GetStack() fallback does not look like a runtime stack: %q", got)
	}
	if got := GetStack(nil); got != "" {
		t.Fatalf("GetStack(nil) = %q, want empty", got)
	}
}

func TestErrorIsHandlesNestedMultierror(t *testing.T) {
	target := errors.New("target")
	nested := multierror.Append(nil, errors.New("other"), fmt.Errorf("wrapped: %w", target))
	err := multierror.Append(nil, errors.New("top"), nested)

	if !ErrorIs(err, target) {
		t.Fatalf("ErrorIs() = false, want true for nested multierror")
	}
	if ErrorIs(err, errors.New("missing")) {
		t.Fatal("ErrorIs() = true for an unrelated error")
	}
	if !ErrorIs(nil, nil) {
		t.Fatal("ErrorIs(nil, nil) should be true")
	}
	if ErrorIs(err, nil) {
		t.Fatal("ErrorIs(non-nil, nil) should be false")
	}
}

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

func TestDefaultLogFuncCanBeConfiguredAndReset(t *testing.T) {
	ResetDefaultLogFunc()
	t.Cleanup(ResetDefaultLogFunc)

	want := errors.New("configured default logger")
	called := 0
	ConfigureDefaultLogFunc(func(ctx context.Context, level logrus.Level, err error, stack string, format string, args ...any) {
		called++
		if ctx == nil {
			t.Fatal("logger context is nil")
		}
		if level != logrus.ErrorLevel {
			t.Fatalf("logger level = %s, want error", level)
		}
		if !ErrorIs(err, want) {
			t.Fatalf("logger err = %v, want %v", err, want)
		}
		if format != "configured %s" || len(args) != 1 || args[0] != "logger" {
			t.Fatalf("logger format/args = %q/%v", format, args)
		}
	})
	if err := Recover(func() error { return want }, "configured %s", "logger"); !ErrorIs(err, want) {
		t.Fatalf("Recover() = %v, want %v", err, want)
	}
	if called != 1 {
		t.Fatalf("configured logger called %d times, want 1", called)
	}

	ResetDefaultLogFunc()
	called = 0
	if err := Recover(func() error { return want }, "reset logger"); !ErrorIs(err, want) {
		t.Fatalf("Recover() after reset = %v, want %v", err, want)
	}
	if called != 0 {
		t.Fatalf("configured logger called after reset %d times, want 0", called)
	}
}
