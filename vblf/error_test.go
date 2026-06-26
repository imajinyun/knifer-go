package vblf_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
	"github.com/imajinyun/knifer-go/vblf"
)

func TestFacadeBloomFilterErrorContract(t *testing.T) {
	bf := vblf.NewBitSetBloomFilter(1000, 5, 3)
	err := bf.InitFromFile(filepath.Join(t.TempDir(), "missing.txt"))
	assertFacadeBloomFilterCode(t, err, knifer.ErrCodeNotFound)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("InitFromFile should preserve os not-exist cause: %v", err)
	}
}

func TestFacadeBloomFilterConstructorsReturnValidationErrors(t *testing.T) {
	bf, err := vblf.NewBitSetBloomFilterE(0, 1, 1)
	if err == nil || bf != nil {
		t.Fatalf("NewBitSetBloomFilterE = %#v, %v; want nil invalid-input error", bf, err)
	}
	assertFacadeBloomFilterCode(t, err, knifer.ErrCodeInvalidInput)

	ff, err := vblf.NewFuncFilterWithMachineNumE(1024, 16, func(string) int64 { return 0 })
	if err == nil || ff != nil {
		t.Fatalf("NewFuncFilterWithMachineNumE = %#v, %v; want nil invalid-input error", ff, err)
	}
	assertFacadeBloomFilterCode(t, err, knifer.ErrCodeInvalidInput)
}

func assertFacadeBloomFilterCode(t *testing.T, err error, code knifer.ErrCode) {
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
	var bloomErr *vblf.Error
	if !errors.As(err, &bloomErr) {
		t.Fatalf("errors.As(err, *vblf.Error) = false: %v", err)
	}
}
