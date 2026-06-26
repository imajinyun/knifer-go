package conf

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestConfErrorNilReceiver(t *testing.T) {
	var e *ConfError
	if s := e.Error(); s != "" {
		t.Fatalf("nil ConfError.Error() = %q", s)
	}
	if ec := e.ErrorCode(); ec != "" {
		t.Fatalf("nil ConfError.ErrorCode() = %q", ec)
	}
	if cause := e.Unwrap(); cause != nil {
		t.Fatalf("nil ConfError.Unwrap() = %v", cause)
	}
	if e.Is(nil) {
		t.Fatal("nil ConfError.Is(nil) = true")
	}
}

func TestConfErrorContract(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "missing.setting"))
	assertConfCode(t, err, knifer.ErrCodeNotFound)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Load missing file should preserve os.ErrNotExist: %v", err)
	}

	_, err = Parse("invalid-line")
	assertConfCode(t, err, knifer.ErrCodeInvalidInput)

	_, err = Parse("=empty")
	assertConfCode(t, err, knifer.ErrCodeInvalidInput)

	_, err = ParseYAML("invalid-yaml-line")
	assertConfCode(t, err, knifer.ErrCodeInvalidInput)
}
