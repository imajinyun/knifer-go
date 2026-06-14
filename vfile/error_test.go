package vfile

import (
	"errors"
	"testing"

	knifer "github.com/imajinyun/go-knifer"
)

func TestFacadeFileErrorContract(t *testing.T) {
	err := CopyFile(t.TempDir()+"/missing.txt", t.TempDir()+"/out.txt")
	if err == nil {
		t.Fatal("CopyFile() error = nil, want invalid input")
	}
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("errors.Is(err, ErrCodeInvalidInput) = false: %v", err)
	}
	code, ok := knifer.CodeOf(err)
	if !ok || code != knifer.ErrCodeInvalidInput {
		t.Fatalf("CodeOf(err) = %q, %v; want invalid input", code, ok)
	}
	var fileErr *Error
	if !errors.As(err, &fileErr) {
		t.Fatalf("errors.As(err, *vfile.Error) = false: %v", err)
	}
}
