package http

import (
	"errors"
	"net"
	"os"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestHTTPErrorMessage(t *testing.T) {
	e := NewHTTPError("read failed", errors.New("conn closed"))
	if e.Error() != "read failed: conn closed" {
		t.Fatalf("error: %q", e.Error())
	}
	if e.Unwrap() == nil {
		t.Fatal("unwrap nil")
	}
}

func TestHTTPErrorWithCode(t *testing.T) {
	err := NewHTTPErrorWithCode(knifer.ErrCodeInvalidInput, "bad url", errors.New("parse"))
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("error should match invalid input: %v", err)
	}
	if code, ok := knifer.CodeOf(err); !ok || code != knifer.ErrCodeInvalidInput {
		t.Fatalf("CodeOf = %q %v, want invalid input", code, ok)
	}
}

func TestHTTPErrorClassifiesTimeout(t *testing.T) {
	cause := &net.DNSError{IsTimeout: true}
	err := NewHTTPError("send request failed", cause)
	if !errors.Is(err, knifer.ErrCodeTimeout) {
		t.Fatalf("timeout error should match timeout code: %v", err)
	}
	deadline := NewHTTPError("deadline", os.ErrDeadlineExceeded)
	if !errors.Is(deadline, knifer.ErrCodeTimeout) {
		t.Fatalf("deadline error should match timeout code: %v", deadline)
	}
}

func TestHTTPErrorf(t *testing.T) {
	e := HTTPErrorf("status %d", 500)
	if e.Error() != "status 500" {
		t.Fatalf("error: %q", e.Error())
	}
	if e.Unwrap() != nil {
		t.Fatal("unwrap should be nil")
	}
}

func TestHTTPErrorMatchesErrCode(t *testing.T) {
	if !errors.Is(NewHTTPError("boom", nil), knifer.ErrCodeInternal) {
		t.Fatal("NewHTTPError should match knifer.ErrCodeInternal")
	}
	if !errors.Is(HTTPErrorf("status %d", 500), knifer.ErrCodeInternal) {
		t.Fatal("HTTPErrorf should match knifer.ErrCodeInternal")
	}
	code, ok := knifer.CodeOf(HTTPErrorf("status %d", 500))
	if !ok || code != knifer.ErrCodeInternal {
		t.Fatalf("CodeOf(HTTPErrorf) = %q, %v; want internal", code, ok)
	}
}
