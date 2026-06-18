package errx

import (
	"context"
	"errors"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestMustExitNoopOnNilError(t *testing.T) {
	MustExit(context.Background(), nil)
}

func TestMustExitPanicsOnError(t *testing.T) {
	silenceLogrus(t)

	want := errors.New("fatal")
	defer func() {
		got := recover()
		if got != want {
			t.Fatalf("panic = %v, want original error", got)
		}
	}()
	MustExit(context.Background(), want)
}

func TestWithExitLogFunc(t *testing.T) {
	var called bool
	cfg := applyExitOptions([]ExitOption{
		WithExitLogFunc(func(ctx context.Context, level logrus.Level, err error, stack string, format string, args ...any) {
			called = true
		}),
	})
	if cfg.logFunc == nil {
		t.Fatal("WithExitLogFunc did not set logFunc")
	}
	cfg.logFunc(context.Background(), logrus.ErrorLevel, errors.New("test"), "stack", "format")
	if !called {
		t.Fatal("custom log func was not called")
	}
}

func TestWithExitPanicFunc(t *testing.T) {
	var got error
	cfg := applyExitOptions([]ExitOption{
		WithExitPanicFunc(func(err error) { got = err }),
	})
	if cfg.panicFunc == nil {
		t.Fatal("WithExitPanicFunc did not set panicFunc")
	}
	want := errors.New("panic test")
	cfg.panicFunc(want)
	if got != want {
		t.Fatalf("panicFunc received %v, want %v", got, want)
	}
}

func TestMustExitWithOptionsUsesCustomPanicFunc(t *testing.T) {
	silenceLogrus(t)

	var got error
	want := errors.New("custom panic")
	MustExitWithOptions(context.Background(), want,
		WithExitPanicFunc(func(err error) { got = err }),
		WithExitLogFunc(func(context.Context, logrus.Level, error, string, string, ...any) {}),
	)
	if got != want {
		t.Fatalf("panicFunc received %v, want %v", got, want)
	}
}

func TestApplyExitOptionsNilLogFuncAndPanicFunc(t *testing.T) {
	silenceLogrus(t)

	cfg := applyExitOptions([]ExitOption{
		WithExitLogFunc(nil),
		WithExitPanicFunc(nil),
	})
	if cfg.logFunc == nil {
		t.Fatal("default logFunc should not be nil after nil option")
	}
	if cfg.panicFunc == nil {
		t.Fatal("default panicFunc should not be nil after nil option")
	}
}

func TestApplyExitOptionsNilSlice(t *testing.T) {
	cfg := applyExitOptions(nil)
	if cfg.logFunc == nil {
		t.Fatal("default logFunc should not be nil")
	}
	if cfg.panicFunc == nil {
		t.Fatal("default panicFunc should not be nil")
	}
}
