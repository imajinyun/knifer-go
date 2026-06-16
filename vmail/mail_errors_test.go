package vmail

import (
	"errors"
	"testing"
)

func TestFacadeExportsSentinelErrors(t *testing.T) {
	_, err := NewMessage(WithTo("to@example.com"), WithText("body"))
	if !errors.Is(err, ErrMissingFrom) {
		t.Fatalf("NewMessage() error = %v, want %v", err, ErrMissingFrom)
	}
}
