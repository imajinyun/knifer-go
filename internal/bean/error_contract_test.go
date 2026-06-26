package bean

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestBeanErrorContract(t *testing.T) {
	_, err := ToMap(nil)
	assertBeanInvalidInput(t, err)

	err = FillMap(sourceProfile{}, nil)
	assertBeanInvalidInput(t, err)

	var dst targetProfile
	err = CopyProperties(map[string]any{"age": "not-a-number"}, &dst)
	assertBeanInvalidInput(t, err)
	var numErr *strconv.NumError
	if !errors.As(err, &numErr) {
		t.Fatalf("CopyProperties should preserve strconv.NumError cause: %v", err)
	}

	err = CopyProperties(map[string]any{"age": "42"}, &dst, WithWeaklyTyped(false))
	assertBeanInvalidInput(t, err)
}

func TestBeanErrorMethods(t *testing.T) {
	var nilErr *BeanError
	if nilErr.Error() != "" {
		t.Fatalf("nil Error() = %q", nilErr.Error())
	}
	if nilErr.ErrorCode() != "" {
		t.Fatalf("nil ErrorCode() = %q", nilErr.ErrorCode())
	}
	if nilErr.Unwrap() != nil {
		t.Fatalf("nil Unwrap() = %v", nilErr.Unwrap())
	}
	if nilErr.Is(knifer.ErrCodeInvalidInput) || nilErr.Is(nil) {
		t.Fatalf("nil Is() returned true")
	}

	cause := fmt.Errorf("root")
	err := &BeanError{Code: knifer.ErrCodeInvalidInput, Msg: "bean failed", Cause: cause}
	if got := err.Error(); got != "bean failed: root" {
		t.Fatalf("Error() = %q", got)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("errors.Is(err, cause) = false")
	}
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("errors.Is(err, ErrCodeInvalidInput) = false")
	}
	if !errors.Is(err, &BeanError{Code: knifer.ErrCodeInvalidInput}) {
		t.Fatalf("errors.Is(err, same-code BeanError) = false")
	}
	if errors.Is(err, &BeanError{Code: knifer.ErrCodeNotFound}) {
		t.Fatalf("errors.Is(err, different-code BeanError) = true")
	}
	if err.Is(errors.New("other")) {
		t.Fatalf("BeanError.Is(non-code target) = true")
	}

	withoutCause := &BeanError{Code: knifer.ErrCodeInvalidInput, Msg: "plain"}
	if withoutCause.Error() != "plain" {
		t.Fatalf("withoutCause.Error() = %q", withoutCause.Error())
	}
	if wrapBeanError(knifer.ErrCodeInvalidInput, "ignored", nil) != nil {
		t.Fatalf("wrapBeanError(nil) returned non-nil")
	}
}
