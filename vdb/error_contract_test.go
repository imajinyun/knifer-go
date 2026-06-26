package vdb

import (
	"errors"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestFacadeDBErrorContract(t *testing.T) {
	_, _, err := NewBuilder().SQL()
	if err == nil {
		t.Fatal("SQL() error = nil, want invalid input error")
	}
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("errors.Is(err, ErrCodeInvalidInput) = false: %v", err)
	}
	code, ok := knifer.CodeOf(err)
	if !ok || code != knifer.ErrCodeInvalidInput {
		t.Fatalf("CodeOf(err) = %q, %v; want invalid input", code, ok)
	}
	var dbErr *DBError
	if !errors.As(err, &dbErr) {
		t.Fatalf("errors.As(err, *DBError) = false: %v", err)
	}
}
