package vconf_test

import (
	"errors"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
	"github.com/imajinyun/knifer-go/vconf"
)

func TestFacadeConfErrorContract(t *testing.T) {
	_, err := vconf.Parse("invalid-line")
	if err == nil {
		t.Fatal("Parse() error = nil, want invalid input")
	}
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("errors.Is(err, ErrCodeInvalidInput) = false: %v", err)
	}
	code, ok := knifer.CodeOf(err)
	if !ok || code != knifer.ErrCodeInvalidInput {
		t.Fatalf("CodeOf(err) = %q, %v; want invalid input", code, ok)
	}
	var confErr *vconf.Error
	if !errors.As(err, &confErr) {
		t.Fatalf("errors.As(err, *vconf.Error) = false: %v", err)
	}
}
