package shared

import (
	"errors"
	"net"
	"os"
	"strings"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestHTTPErrorAndClassification(t *testing.T) {
	cause := os.ErrDeadlineExceeded
	err := NewHTTPError("request failed", cause)
	if err.ErrorCode() != knifer.ErrCodeTimeout {
		t.Fatalf("NewHTTPError code = %q", err.ErrorCode())
	}
	if !strings.Contains(err.Error(), "request failed") || !errors.Is(err, cause) {
		t.Fatalf("HTTPError error=%q cause=%v", err.Error(), err.Unwrap())
	}
	if !errors.Is(err, knifer.ErrCodeTimeout) {
		t.Fatalf("errors.Is(timeout code) = false: %v", err)
	}
	if got := (*HTTPError)(nil).ErrorCode(); got != "" {
		t.Fatalf("nil HTTPError ErrorCode = %q", got)
	}
	if got := (*HTTPError)(nil).Unwrap(); got != nil {
		t.Fatalf("nil HTTPError Unwrap = %v", got)
	}
	if (*HTTPError)(nil).Is(knifer.ErrCodeTimeout) {
		t.Fatal("nil HTTPError Is returned true")
	}
	if got := HTTPErrorf("bad %s", "request"); got.ErrorCode() != knifer.ErrCodeInternal || got.Error() != "bad request" {
		t.Fatalf("HTTPErrorf = %#v", got)
	}
	if got := HTTPErrorfWithCode(knifer.ErrCodeInvalidInput, "bad %s", "url"); got.ErrorCode() != knifer.ErrCodeInvalidInput {
		t.Fatalf("HTTPErrorfWithCode code = %q", got.ErrorCode())
	}
	if got := NewHTTPErrorWithCode("", "fallback", nil).ErrorCode(); got != knifer.ErrCodeInternal {
		t.Fatalf("NewHTTPErrorWithCode empty code = %q", got)
	}
	if got := ClassifyHTTPErrorCode(timeoutNetError{}, knifer.ErrCodeInternal); got != knifer.ErrCodeTimeout {
		t.Fatalf("ClassifyHTTPErrorCode net timeout = %q", got)
	}
	if got := ClassifyHTTPErrorCode(knifer.NewError(knifer.ErrCodeUnsupported, "unsupported"), knifer.ErrCodeInternal); got != knifer.ErrCodeUnsupported {
		t.Fatalf("ClassifyHTTPErrorCode carrier = %q", got)
	}
	if got := ClassifyHTTPErrorCode(nil, ""); got != knifer.ErrCodeInternal {
		t.Fatalf("ClassifyHTTPErrorCode nil empty fallback = %q", got)
	}
}

func TestHTTPErrorIsSameTypeAndNilError(t *testing.T) {
	err := HTTPErrorfWithCode(knifer.ErrCodeInvalidInput, "bad url")
	if !err.Is(&HTTPError{Code: knifer.ErrCodeInvalidInput}) {
		t.Fatal("Is(*HTTPError same code) should match")
	}
	if err.Is(&HTTPError{Code: knifer.ErrCodeTimeout}) {
		t.Fatal("Is(*HTTPError other code) should not match")
	}
	if err.Is(errors.New("x")) || err.Is(nil) {
		t.Fatal("Is should not match unrelated targets")
	}
	if got := (*HTTPError)(nil).Error(); got != "" {
		t.Fatalf("nil HTTPError Error = %q", got)
	}
}

type timeoutNetError struct{}

func (timeoutNetError) Error() string   { return "timeout" }
func (timeoutNetError) Timeout() bool   { return true }
func (timeoutNetError) Temporary() bool { return false }

var _ net.Error = timeoutNetError{}
