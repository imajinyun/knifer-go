package vhttp_test

import (
	"errors"
	"testing"

	knifer "github.com/imajinyun/go-knifer"
	"github.com/imajinyun/go-knifer/vhttp"
)

func TestFacadeErrorNamesWithoutHTTPPrefix(t *testing.T) {
	cause := errors.New("closed")
	err := vhttp.NewError("read failed", cause)
	if !errors.Is(err, cause) {
		t.Fatalf("NewError() does not unwrap cause")
	}
	if !errors.Is(err, knifer.ErrCodeInternal) {
		t.Fatalf("NewError() does not match ErrCodeInternal")
	}
	code, ok := knifer.CodeOf(err)
	if !ok || code != knifer.ErrCodeInternal {
		t.Fatalf("CodeOf(NewError()) = %q, %v; want internal", code, ok)
	}

	formatted := vhttp.Errorf("status %d", 500)
	if got := errorString(formatted); got != "status 500" {
		t.Fatalf("Errorf().Error() = %q, want status 500", got)
	}
}
