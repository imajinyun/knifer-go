package verr_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	knifer "github.com/imajinyun/go-knifer"
	"github.com/imajinyun/go-knifer/verr"
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
