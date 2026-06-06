package verr_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	knifer "github.com/imajinyun/go-knifer"
	"github.com/imajinyun/go-knifer/verr"
	"github.com/sirupsen/logrus"
)

func TestCollectorFacade(t *testing.T) {
	want := errors.New("collector failure")
	c := verr.NewCollector().WithContext(context.Background())
	c.GoRun(func() error { return want }, "collector")
	if got := c.Error(); !verr.ErrorIs(got, want) {
		t.Fatalf("Collector.Error() = %v, want %v", got, want)
	}
}

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

func TestWrapperAndStackFacade(t *testing.T) {
	want := errors.New("wrapper failure")
	if got := verr.Wrap(func() error { return want }).WithWarnf("wrapper").Exec(context.TODO()); !verr.ErrorIs(got, want) {
		t.Fatalf("Wrapper.Exec() = %v, want %v", got, want)
	}
	stack := verr.GetStackTrace(0)
	if len(stack) == 0 {
		t.Fatal("GetStackTrace() returned empty stack")
	}
	if formatted := fmt.Sprintf("%+v", stack); !strings.Contains(formatted, "TestWrapperAndStackFacade") {
		t.Fatalf("formatted stack = %q, want current test", formatted)
	}
}

func TestStackTraceWithOptionsFacade(t *testing.T) {
	stack := verr.GetStackTraceWithOptions(verr.WithStackSkip(0), verr.WithStackDepth(4))
	if len(stack) == 0 || len(stack) > 4 {
		t.Fatalf("GetStackTraceWithOptions length = %d, want 1..4", len(stack))
	}
	formatted := fmt.Sprintf("%+v", stack)
	if !strings.Contains(formatted, "TestStackTraceWithOptionsFacade") {
		t.Fatalf("formatted stack = %q, want current test", formatted)
	}
}

func TestInitWithOptionsFacade(t *testing.T) {
	var b strings.Builder
	verr.InitWithOptions(verr.WithLogOutput(&b), verr.WithReportCaller(false))
}

type facadeTimer struct{}

func (facadeTimer) Stop() bool { return true }

func TestCollectorWaitOptionsFacade(t *testing.T) {
	c := verr.NewCollector()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	called := false
	done, err := verr.WaitUntilWithOptions(c, time.Second,
		verr.WithWaitContext(ctx),
		verr.WithWaitTimerFactory(func(duration time.Duration) (<-chan time.Time, verr.Timer) {
			called = true
			if duration != time.Second {
				t.Fatalf("timer duration = %s, want 1s", duration)
			}
			return make(chan time.Time), facadeTimer{}
		}),
	)
	if done || err != nil {
		t.Fatalf("WaitUntilWithOptions() = (%v, %v), want (false, nil)", done, err)
	}
	if !called {
		t.Fatal("facade wait timer factory was not called")
	}
}

func TestMustExitFacade(t *testing.T) {
	verr.MustExit(context.Background(), nil)
	want := errors.New("exit")
	defer func() {
		if got := recover(); got != want {
			t.Fatalf("panic = %v, want original error", got)
		}
	}()
	verr.MustExit(context.Background(), want)
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
