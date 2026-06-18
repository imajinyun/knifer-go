package errx

import (
	"context"
	"errors"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestWrapperWithInfof(t *testing.T) {
	silenceLogrus(t)

	var called bool
	w := Wrap(func() error { return errors.New("fail") }).
		WithLogFunc(func(context.Context, logrus.Level, error, string, string, ...any) { called = true }).
		WithInfof("info message %d", 1)
	if err := w.Exec(context.Background()); err == nil {
		t.Fatal("expected error")
	}
	if !called {
		t.Fatal("WithInfof log func was not called")
	}
}

func TestWrapperWithLogFunc(t *testing.T) {
	silenceLogrus(t)

	var called bool
	w := Wrap(func() error { return errors.New("fail") }).
		WithLogFunc(func(context.Context, logrus.Level, error, string, string, ...any) { called = true }).
		WithErrorf("error message")
	if err := w.Exec(context.Background()); err == nil {
		t.Fatal("expected error")
	}
	if !called {
		t.Fatal("WithLogFunc was not called")
	}
}

func TestWrapperWithStackOptions(t *testing.T) {
	silenceLogrus(t)

	w := Wrap(func() error { return errors.New("fail") }).
		WithStackOptions(WithDebugStackFunc(func() []byte { return []byte("custom") })).
		WithErrorf("stack options")
	if err := w.Exec(context.Background()); err == nil {
		t.Fatal("expected error")
	}
}

func TestWrapperWithLogFuncAndStackOptionsNilSource(t *testing.T) {
	silenceLogrus(t)

	w := Wrap(func() error { return errors.New("fail") }).
		WithStackOptions().
		WithLogFunc(nil).
		WithErrorf("nil log func falls back to default")
	if err := w.Exec(context.Background()); err == nil {
		t.Fatal("expected error")
	}
}
