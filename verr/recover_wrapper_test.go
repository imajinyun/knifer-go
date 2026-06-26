package verr_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
	"github.com/imajinyun/knifer-go/verr"
	"github.com/sirupsen/logrus"
)

func TestRecoverFacadeConvertsPanic(t *testing.T) {
	got := verr.Recover(func() error {
		panic("facade panic")
	}, "recover")
	if got == nil || !strings.Contains(got.Error(), "facade panic") {
		t.Fatalf("Recover() = %v, want panic value", got)
	}
	var pe *verr.PanicError
	if !errors.As(got, &pe) || pe.Stack() == "" {
		t.Fatalf("Recover() = %T stack=%q, want PanicError with stack", got, pe.Stack())
	}
	if !errors.Is(got, knifer.ErrCodeInternal) {
		t.Fatalf("Recover() = %v, want ErrCodeInternal", got)
	}
	if code, ok := knifer.CodeOf(got); !ok || code != knifer.ErrCodeInternal {
		t.Fatalf("CodeOf(Recover()) = %q, %v; want internal", code, ok)
	}
}

func TestRecoverWithoutErrorFacade(t *testing.T) {
	verr.ResetDefaultLogFunc()
	t.Cleanup(verr.ResetDefaultLogFunc)

	var calls int
	verr.ConfigureDefaultLogFunc(func(ctx context.Context, level logrus.Level, err error, stack string, format string, args ...any) {
		calls++
		if ctx == nil {
			t.Fatal("logger ctx = nil")
		}
		if level != logrus.ErrorLevel {
			t.Fatalf("logger level = %s, want error", level)
		}
		if err == nil || !strings.Contains(err.Error(), "without-error panic") {
			t.Fatalf("logger err = %v, want panic value", err)
		}
		if stack == "" {
			t.Fatal("logger stack is empty")
		}
		if format != "facade without error" {
			t.Fatalf("logger format = %q, want facade without error", format)
		}
	})

	if got := verr.RecoverWithoutError(func() {}, "facade without error"); got != nil {
		t.Fatalf("RecoverWithoutError(no panic) = %v, want nil", got)
	}
	if calls != 0 {
		t.Fatalf("logger called %d times for no-panic path, want 0", calls)
	}

	got := verr.RecoverWithoutError(func() { panic("without-error panic") }, "facade without error")
	if got == nil {
		t.Fatal("RecoverWithoutError(panic) = nil, want error")
	}
	var pe *verr.PanicError
	if !errors.As(got, &pe) || pe.Stack() == "" {
		t.Fatalf("RecoverWithoutError(panic) = %T stack=%q, want PanicError with stack", got, pe.Stack())
	}
	if !errors.Is(got, knifer.ErrCodeInternal) {
		t.Fatalf("RecoverWithoutError(panic) = %v, want ErrCodeInternal", got)
	}
	if calls != 1 {
		t.Fatalf("logger called %d times, want 1", calls)
	}
}

func TestWrapperFacade(t *testing.T) {
	want := errors.New("wrapper failure")
	if got := verr.Wrap(func() error { return want }).WithWarnf("wrapper").Exec(context.TODO()); !verr.ErrorIs(got, want) {
		t.Fatalf("Wrapper.Exec() = %v, want %v", got, want)
	}
}

func TestDefaultLogFuncFacade(t *testing.T) {
	verr.ResetDefaultLogFunc()
	t.Cleanup(verr.ResetDefaultLogFunc)

	want := errors.New("facade default logger")
	called := 0
	verr.ConfigureDefaultLogFunc(func(ctx context.Context, level logrus.Level, err error, stack string, format string, args ...any) {
		called++
		if level != logrus.ErrorLevel {
			t.Fatalf("logger level = %s, want error", level)
		}
		if !verr.ErrorIs(err, want) {
			t.Fatalf("logger err = %v, want %v", err, want)
		}
		if format != "facade logger" {
			t.Fatalf("logger format = %q, want facade logger", format)
		}
	})
	if got := verr.Recover(func() error { return want }, "facade logger"); !verr.ErrorIs(got, want) {
		t.Fatalf("Recover() = %v, want %v", got, want)
	}
	if called != 1 {
		t.Fatalf("configured logger called %d times, want 1", called)
	}
}
