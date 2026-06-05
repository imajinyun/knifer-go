package vblf_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	knifer "github.com/imajinyun/go-knifer"
	"github.com/imajinyun/go-knifer/vblf"
)

func TestFacadeBloomFilter(t *testing.T) {
	f := vblf.NewBitSetBloomFilter(1000, 5, 3)
	f.Add("hello")
	f.Add("world")

	if !f.Contains("hello") {
		t.Fatal("expected filter to contain 'hello'")
	}
	if !f.Contains("world") {
		t.Fatal("expected filter to contain 'world'")
	}
	// false positive possible, but unlikely for this small set
}

func TestFacadeFuncFilter(t *testing.T) {
	f := vblf.NewDefaultBloomFilter(1000)
	f.Add("test")
	if !f.Contains("test") {
		t.Fatal("expected func filter to contain 'test'")
	}
}

func TestFacadeOptionsConstructors(t *testing.T) {
	bitset := vblf.NewBitSetBloomFilterWithOptions(
		vblf.WithBitSetCapacity(1000),
		vblf.WithExpectedElements(5),
		vblf.WithHashFunctionNumber(3),
	)
	if !bitset.Add("hello") || !bitset.Contains("hello") {
		t.Fatal("expected options-created bitset filter to contain value")
	}

	bitmap := vblf.NewBitMapBloomFilterWithOptions(
		vblf.WithBitMapSize(5),
		vblf.WithBloomFilters(vblf.NewFNVFilter(1<<20), vblf.NewRSFilter(1<<20)),
	)
	if !bitmap.Add("world") || !bitmap.Contains("world") {
		t.Fatal("expected options-created bitmap filter to contain value")
	}

	fn := vblf.NewFuncFilterWithOptions(
		vblf.WithMaxValue(1000),
		vblf.WithMachineNum(vblf.BloomMachine64),
		vblf.WithHashFunc(func(s string) int64 { return int64(vblf.JavaDefaultHash(s)) }),
	)
	if !fn.Add("test") || !fn.Contains("test") {
		t.Fatal("expected options-created func filter to contain value")
	}
}

func TestFacadeCreateOptionsConstructors(t *testing.T) {
	bitset := vblf.CreateBitSetWithOptions(
		vblf.WithBitSetCapacity(1000),
		vblf.WithExpectedElements(5),
		vblf.WithHashFunctionNumber(3),
	)
	if !bitset.Add("hello") || !bitset.Contains("hello") {
		t.Fatal("expected CreateBitSetWithOptions filter to contain value")
	}

	bitmap := vblf.CreateBitMapWithOptions(
		vblf.WithBitMapSize(5),
		vblf.WithBloomFilters(vblf.NewFNVFilter(1<<20), vblf.NewRSFilter(1<<20)),
	)
	if !bitmap.Add("world") || !bitmap.Contains("world") {
		t.Fatal("expected CreateBitMapWithOptions filter to contain value")
	}

	fn := vblf.NewFuncFilterFromOptions(vblf.WithMaxValue(1000), vblf.WithHashFunc(func(s string) int64 {
		return int64(vblf.JavaDefaultHash(s))
	}))
	if !fn.Add("alias") || !fn.Contains("alias") {
		t.Fatal("expected NewFuncFilterFromOptions filter to contain value")
	}
}

func TestFacadeHashFunctions(t *testing.T) {
	// smoke test: hash functions should return consistent values
	h1 := vblf.BloomRSHash("abc")
	h2 := vblf.BloomRSHash("abc")
	if h1 != h2 {
		t.Fatal("hash function should be deterministic")
	}
}

func TestFacadeBitMap(t *testing.T) {
	bm := vblf.NewIntMap(100)
	bm.Add(42)
	if !bm.Contains(42) {
		t.Fatal("expected bitmap to contain 42")
	}
}

func TestFacadeBloomFilterErrorContract(t *testing.T) {
	bf := vblf.NewBitSetBloomFilter(1000, 5, 3)
	err := bf.InitFromFile(filepath.Join(t.TempDir(), "missing.txt"))
	assertFacadeBloomFilterCode(t, err, knifer.ErrCodeNotFound)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("InitFromFile should preserve os not-exist cause: %v", err)
	}
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
