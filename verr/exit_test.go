package verr_test

import (
	"context"
	"errors"
	"testing"

	"github.com/imajinyun/knifer-go/verr"
	"github.com/sirupsen/logrus"
)

type exitContextKey struct{}

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

func TestMustExitWithOptionsFacade(t *testing.T) {
	ctx := context.WithValue(context.Background(), exitContextKey{}, "facade")
	want := errors.New("custom exit")
	var loggedErr error
	var loggedLevel logrus.Level
	var loggedFormat string
	var panickedErr error

	verr.MustExitWithOptions(ctx, nil,
		verr.WithExitLogFunc(func(context.Context, logrus.Level, error, string, string, ...any) {
			t.Fatal("nil error should not be logged")
		}),
		verr.WithExitPanicFunc(func(error) { t.Fatal("nil error should not panic") }),
	)

	verr.MustExitWithOptions(ctx, want,
		verr.WithExitLogFunc(func(gotCtx context.Context, level logrus.Level, err error, stack string, format string, args ...any) {
			if gotCtx != ctx {
				t.Fatal("exit logger received a different context")
			}
			if stack == "" {
				t.Fatal("exit logger stack is empty")
			}
			loggedErr = err
			loggedLevel = level
			loggedFormat = format
		}),
		verr.WithExitPanicFunc(func(err error) { panickedErr = err }),
	)

	if loggedErr != want || panickedErr != want {
		t.Fatalf("logged=%v panicked=%v, want original error", loggedErr, panickedErr)
	}
	if loggedLevel != logrus.ErrorLevel || loggedFormat != "exit with error" {
		t.Fatalf("log level=%s format=%q, want error exit with error", loggedLevel, loggedFormat)
	}
}
