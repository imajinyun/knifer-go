package bloomfilter

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	knifer "github.com/imajinyun/go-knifer"
)

func TestBitSetBloomFilterErrorContract(t *testing.T) {
	bf := NewBitSetBloomFilter(10000, 100, 4)
	err := bf.InitFromFile(filepath.Join(t.TempDir(), "missing.txt"))
	assertBloomFilterCode(t, err, knifer.ErrCodeNotFound)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("InitFromFile should preserve os not-exist cause: %v", err)
	}
}

func TestBloomFilterErrorString(t *testing.T) {
	err := &BloomFilterError{Code: knifer.ErrCodeInternal, Msg: "test error", Cause: nil}
	if got := err.Error(); got != "test error" {
		t.Fatalf("BloomFilterError.Error() = %q, want %q", got, "test error")
	}
	errWithCause := &BloomFilterError{Code: knifer.ErrCodeInternal, Msg: "test", Cause: os.ErrNotExist}
	if got := errWithCause.Error(); got != "test: file does not exist" {
		t.Fatalf("BloomFilterError with cause = %q", got)
	}
}

func assertBloomFilterCode(t *testing.T, err error, code knifer.ErrCode) {
	t.Helper()
	if err == nil {
		t.Fatalf("err = nil, want %s", code)
	}
	if !errors.Is(err, code) {
		t.Fatalf("errors.Is(%v, %s) = false", err, code)
	}
	got, ok := knifer.CodeOf(err)
	if !ok || got != code {
		t.Fatalf("CodeOf(%v) = %q, %v; want %q, true", err, got, ok, code)
	}
	var bloomErr *BloomFilterError
	if !errors.As(err, &bloomErr) {
		t.Fatalf("errors.As(err, *BloomFilterError) = false: %v", err)
	}
}
