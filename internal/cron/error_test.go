package cron

import (
	"errors"
	"testing"

	knifer "github.com/imajinyun/go-knifer"
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
