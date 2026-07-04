package vfile

import (
	"errors"
	"io"
	"os"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
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

func TestFacadeFileErrorPreservesProviderCause(t *testing.T) {
	cause := os.ErrPermission
	_, err := ReadFileStringWithOptions("secret.txt", WithOpen(func(string) (io.ReadCloser, error) {
		return nil, cause
	}))
	if !errors.Is(err, knifer.ErrCodeInternal) {
		t.Fatalf("ReadFileStringWithOptions error = %v, want ErrCodeInternal", err)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("ReadFileStringWithOptions error = %v, want provider cause", err)
	}
	var fileErr *Error
	if !errors.As(err, &fileErr) {
		t.Fatalf("errors.As(err, *vfile.Error) = false: %v", err)
	}
}
