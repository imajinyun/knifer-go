package file

import (
	"errors"
	"os"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestFileErrorContract(t *testing.T) {
	err := &FileError{Code: knifer.ErrCodeNotFound, Msg: "read file x", Cause: os.ErrNotExist}
	if got := err.Error(); got != "read file x: "+os.ErrNotExist.Error() {
		t.Fatalf("Error() = %q", got)
	}
	if got := err.ErrorCode(); got != knifer.ErrCodeNotFound {
		t.Fatalf("ErrorCode() = %q", got)
	}
	if !errors.Is(err, knifer.ErrCodeNotFound) {
		t.Fatalf("errors.Is(%v, %s) = false", err, knifer.ErrCodeNotFound)
	}
	if !errors.Is(err, &FileError{Code: knifer.ErrCodeNotFound}) {
		t.Fatalf("errors.Is(%v, file error not found) = false", err)
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("errors.Is(%v, %v) = false", err, os.ErrNotExist)
	}
	if err.Unwrap() != os.ErrNotExist {
		t.Fatalf("Unwrap() = %v", err.Unwrap())
	}

	var nilErr *FileError
	if got := nilErr.Error(); got != "" {
		t.Fatalf("nil Error() = %q, want empty", got)
	}
	if got := nilErr.ErrorCode(); got != "" {
		t.Fatalf("nil ErrorCode() = %q, want empty", got)
	}
	if got := nilErr.Unwrap(); got != nil {
		t.Fatalf("nil Unwrap() = %v, want nil", got)
	}
	if nilErr.Is(knifer.ErrCodeNotFound) {
		t.Fatal("nil Is() = true, want false")
	}
}

func TestWrapFileIOMapsNotExistToNotFound(t *testing.T) {
	err := wrapFileIO("read file missing.txt", os.ErrNotExist)
	assertFileCode(t, err, knifer.ErrCodeNotFound)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("errors.Is(%v, %v) = false", err, os.ErrNotExist)
	}
	if err := wrapFileIO("ignored", nil); err != nil {
		t.Fatalf("wrapFileIO(nil) = %v, want nil", err)
	}
}
