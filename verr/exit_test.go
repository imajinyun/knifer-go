package verr_test

import (
	"context"
	"errors"
	"testing"

	"github.com/imajinyun/go-knifer/verr"
)

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
