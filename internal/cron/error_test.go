package cron

import (
	"errors"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestCronErrorMatchesErrCode(t *testing.T) {
	err := WrapCronError(errors.New("bad field"), "invalid pattern")
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("CronError should match ErrCodeInvalidInput: %v", err)
	}
	if !errors.Is(err, err.Unwrap()) {
		t.Fatalf("CronError should keep cause chain: %v", err)
	}
}

func TestCronErrorContract(t *testing.T) {
	err := NewCronError("bad %s", "pattern")

	// CodeOf classifies the error through the CodeCarrier interface.
	if code, ok := knifer.CodeOf(err); !ok || code != knifer.ErrCodeInvalidInput {
		t.Fatalf("CodeOf = %q, %v", code, ok)
	}
	// Is matches both an ErrCode target and another *CronError with the same code.
	if !err.Is(knifer.ErrCodeInvalidInput) {
		t.Fatal("Is(code) should match")
	}
	if !err.Is(&CronError{Code: knifer.ErrCodeInvalidInput}) {
		t.Fatal("Is(*CronError same code) should match")
	}
	if err.Is(knifer.ErrCodeInternal) || err.Is(errors.New("x")) || err.Is(nil) {
		t.Fatal("Is should not match unrelated targets")
	}
	if err.ErrorCode() != knifer.ErrCodeInvalidInput {
		t.Fatalf("ErrorCode = %q", err.ErrorCode())
	}

	// newSchedulerStartedError carries the unsupported code and the sentinel cause.
	se := newSchedulerStartedError()
	if !errors.Is(se, knifer.ErrCodeUnsupported) || !errors.Is(se, ErrSchedulerStarted) {
		t.Fatalf("scheduler started error chain: %v", se)
	}
}

func TestCronErrorNilReceiver(t *testing.T) {
	var e *CronError
	if e.Error() != "" || e.ErrorCode() != "" || e.Unwrap() != nil || e.Is(knifer.ErrCodeInternal) {
		t.Fatal("nil *CronError methods should be zero-valued and safe")
	}
}
