package verr_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/imajinyun/go-knifer/verr"
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
