package vhttp_test

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
	"github.com/imajinyun/knifer-go/vhttp"
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

func TestFacadeErrorCodes(t *testing.T) {
	cause := errors.New("bad request")
	err := vhttp.NewErrorWithCode(knifer.ErrCodeInvalidInput, "invalid request", cause)
	if !errors.Is(err, cause) || !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("NewErrorWithCode does not unwrap cause or code: %v", err)
	}
	if got := vhttp.ErrorfWithCode(knifer.ErrCodeInvalidInput, "status %d", http.StatusBadRequest).Error(); got != "status 400" {
		t.Fatalf("ErrorfWithCode = %q", got)
	}
}

func TestFacadeSafeBoundaryErrorContract(t *testing.T) {
	secret := "sk-test-secret"
	resp := vhttp.GetSafe(
		"http://private.example/config?token="+secret,
		vhttp.WithLookupIP(func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("10.0.0.1")}, nil
		}),
	).Execute()
	err := resp.Err()
	if !errors.Is(err, knifer.ErrCodeUnsafeResource) {
		t.Fatalf("GetSafe error = %v, want ErrCodeUnsafeResource", err)
	}
	if strings.Contains(err.Error(), secret) || strings.Contains(err.Error(), "token=") {
		t.Fatalf("GetSafe error leaked query secret: %v", err)
	}
}
