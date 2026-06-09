package bloomfilter

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	knifer "github.com/imajinyun/go-knifer"
)

func TestBitSetBloomFilter_AddAndContains(t *testing.T) {
	bf := NewBitSetBloomFilter(10000, 100, 4)
	if !bf.Add("hello") {
		t.Fatal("first add should return true")
	}
	if bf.Add("hello") {
		t.Fatal("repeat add should return false")
	}
	if !bf.Contains("hello") {
		t.Fatal("should contain hello")
	}
	if bf.Contains("absent-token-xyz") {
		t.Fatal("should not contain absent token")
	}
	if p := bf.FalsePositiveProbability(); p < 0 || p > 1 {
		t.Fatalf("invalid probability: %v", p)
	}
}

func TestBitSetBloomFilterWithOptions(t *testing.T) {
	bf := NewBitSetBloomFilterWithOptions(
		WithBitSetCapacity(10000),
		WithExpectedElements(100),
		WithHashFunctionNumber(4),
	)
	if bf.bitSetSize != 40000 {
		t.Fatalf("bitSetSize = %d, want 40000", bf.bitSetSize)
	}
	if bf.hashFunctionNumber != 4 {
		t.Fatalf("hashFunctionNumber = %d, want 4", bf.hashFunctionNumber)
	}
	if !bf.Add("hello") || !bf.Contains("hello") {
		t.Fatal("options-created bitset filter should add and contain value")
	}
}

func TestBitSetBloomFilterFileOptions(t *testing.T) {
	bf := NewBitSetBloomFilter(1000, 5, 3)
	openedPath := ""
	err := bf.InitFromFileWithOptions("virtual.txt", WithOpenFile(func(path string) (io.ReadCloser, error) {
		openedPath = path
		return io.NopCloser(strings.NewReader("alpha\nbeta\n")), nil
	}))
	if err != nil {
		t.Fatalf("InitFromFileWithOptions: %v", err)
	}
	if openedPath != "virtual.txt" || !bf.Contains("alpha") || !bf.Contains("beta") {
		t.Fatalf("custom open not applied path=%q alpha=%v beta=%v", openedPath, bf.Contains("alpha"), bf.Contains("beta"))
	}

	bf = NewBitSetBloomFilter(1000, 5, 3)
	if err := bf.InitFromReader(strings.NewReader("reader\n")); err != nil {
		t.Fatalf("InitFromReader: %v", err)
	}
	if !bf.Contains("reader") {
		t.Fatal("InitFromReader should add reader line")
	}
}

func TestBitSetBloomFilter_InvalidParamsReturnError(t *testing.T) {
	cases := []struct {
		name string
		c    int
		n    int
		k    int
	}{
		{name: "zero capacity", c: 0, n: 1, k: 1},
		{name: "zero expected", c: 1, n: 0, k: 1},
		{name: "zero hash functions", c: 1, n: 1, k: 0},
		{name: "too many hash functions", c: 1, n: 1, k: 9},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			bf, err := NewBitSetBloomFilterE(tt.c, tt.n, tt.k)
			if err == nil || bf != nil {
				t.Fatalf("NewBitSetBloomFilterE() = %#v, %v; want nil invalid-input error", bf, err)
			}
			assertBloomFilterCode(t, err, knifer.ErrCodeInvalidInput)
			if got := NewBitSetBloomFilter(tt.c, tt.n, tt.k); got != nil {
				t.Fatalf("panic-compatible constructor should return nil on invalid input, got %#v", got)
			}
		})
	}
}

func TestBitMapBloomFilter(t *testing.T) {
	bf := NewBitMapBloomFilter(5)
	if !bf.Add("foo") {
		t.Fatal("add foo should return true")
	}
	if !bf.Contains("foo") {
		t.Fatal("should contain foo")
	}
	if bf.Add("foo") {
		t.Fatal("repeat add foo should return false")
	}
	if bf.Contains("not-in-filter-12345") {
		t.Fatal("should not contain unknown token")
	}
}

func TestBitMapBloomFilter_CustomFilters(t *testing.T) {
	bf := NewBitMapBloomFilterWithFilters(
		5,
		NewFNVFilter(1<<20),
		NewRSFilter(1<<20),
	)
	if !bf.Add("bar") {
		t.Fatal()
	}
	if !bf.Contains("bar") {
		t.Fatal()
	}
}

func TestBitMapBloomFilterWithOptions(t *testing.T) {
	bf := NewBitMapBloomFilterWithOptions(WithBitMapSize(5))
	if len(bf.filters) != 5 {
		t.Fatalf("default filter count = %d, want 5", len(bf.filters))
	}
	if !bf.Add("foo") || !bf.Contains("foo") {
		t.Fatal("options-created bitmap filter should add and contain value")
	}

	custom := NewBitMapBloomFilterWithOptions(
		WithBitMapSize(5),
		WithBloomFilters(NewFNVFilter(1<<20), NewRSFilter(1<<20)),
	)
	if len(custom.filters) != 2 {
		t.Fatalf("custom filter count = %d, want 2", len(custom.filters))
	}
	if !custom.Add("bar") || !custom.Contains("bar") {
		t.Fatal("custom options-created bitmap filter should add and contain value")
	}
}

func TestFuncFilter_MachineNum(t *testing.T) {
	f := NewFuncFilterWithMachineNum(1024, Machine64,
		func(s string) int64 { return int64(JavaDefaultHash(s)) })
	if !f.Add("x") {
		t.Fatal()
	}
	if !f.Contains("x") {
		t.Fatal()
	}
}

func TestFuncFilterWithOptions(t *testing.T) {
	f := NewFuncFilterWithOptions(
		WithMaxValue(1024),
		WithMachineNum(Machine64),
		WithHashFunc(func(s string) int64 { return int64(JavaDefaultHash(s)) }),
	)
	if _, ok := f.bm.(*LongMap); !ok {
		t.Fatalf("backing bitmap = %T, want *LongMap", f.bm)
	}
	if !f.Add("x") || !f.Contains("x") {
		t.Fatal("options-created func filter should add and contain value")
	}
}

func TestFuncFilterWithOptionsUsesDefaultHash(t *testing.T) {
	f := NewFuncFilterWithOptions(WithMaxValue(1024))
	if !f.Add("default") || !f.Contains("default") {
		t.Fatal("options-created func filter should use default hash function")
	}
}

func TestFuncFilter_InvalidOptionsReturnError(t *testing.T) {
	cases := []struct {
		name       string
		maxValue   int64
		machineNum int
	}{
		{name: "zero max", maxValue: 0, machineNum: Machine32},
		{name: "too large max", maxValue: 0x80000000, machineNum: Machine32},
		{name: "unknown machine", maxValue: 1024, machineNum: 16},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			f, err := NewFuncFilterWithMachineNumE(tt.maxValue, tt.machineNum, func(s string) int64 { return 0 })
			if err == nil || f != nil {
				t.Fatalf("NewFuncFilterWithMachineNumE() = %#v, %v; want nil invalid-input error", f, err)
			}
			assertBloomFilterCode(t, err, knifer.ErrCodeInvalidInput)
			if got := NewFuncFilterWithMachineNum(tt.maxValue, tt.machineNum, func(s string) int64 { return 0 }); got != nil {
				t.Fatalf("panic-compatible constructor should return nil on invalid input, got %#v", got)
			}
		})
	}
}

func TestIntMapAndLongMap(t *testing.T) {
	im := NewIntMap(8)
	im.Add(0)
	im.Add(31)
	im.Add(64)
	if !im.Contains(0) || !im.Contains(31) || !im.Contains(64) {
		t.Fatal("intmap contains failed")
	}
	if im.Contains(1) {
		t.Fatal("intmap should not contain 1")
	}
	im.Remove(31)
	if im.Contains(31) {
		t.Fatal("intmap remove failed")
	}

	lm := NewLongMap(4)
	lm.Add(0)
	lm.Add(63)
	lm.Add(128)
	if !lm.Contains(0) || !lm.Contains(63) || !lm.Contains(128) {
		t.Fatal("longmap contains failed")
	}
	lm.Remove(128)
	if lm.Contains(128) {
		t.Fatal("longmap remove failed")
	}
}

func TestHashAlgorithms(t *testing.T) {
	s := "test-bloomFilter"
	checks := map[string]int32{
		"rs":   RsHash(s),
		"js":   JsHash(s),
		"pjw":  PjwHash(s),
		"elf":  ElfHash(s),
		"bkdr": BkdrHash(s),
		"sdbm": SdbmHash(s),
		"djb":  DjbHash(s),
		"ap":   ApHash(s),
		"fnv":  FnvHashString(s),
	}
	for name, v := range checks {
		// Only verify stability: the same string should produce the same result.
		if v != checkAgain(name, s) {
			t.Fatalf("%s is unstable", name)
		}
	}
	if JavaDefaultHash("a") != 97 {
		t.Fatal("javaDefault hash 'a' should be 97")
	}
	if TianlHash("") != 0 {
		t.Fatal("tianl empty should be 0")
	}
}

// checkAgain runs the same algorithm again for stability tests.
func checkAgain(name, s string) int32 {
	switch name {
	case "rs":
		return RsHash(s)
	case "js":
		return JsHash(s)
	case "pjw":
		return PjwHash(s)
	case "elf":
		return ElfHash(s)
	case "bkdr":
		return BkdrHash(s)
	case "sdbm":
		return SdbmHash(s)
	case "djb":
		return DjbHash(s)
	case "ap":
		return ApHash(s)
	case "fnv":
		return FnvHashString(s)
	}
	return 0
}

func TestBloomFilterInterface(t *testing.T) {
	var _ BloomFilter = (*BitSetBloomFilter)(nil)
	var _ BloomFilter = (*BitMapBloomFilter)(nil)
	var _ BloomFilter = (*FuncFilter)(nil)
}

func TestBitSetBloomFilterErrorContract(t *testing.T) {
	bf := NewBitSetBloomFilter(10000, 100, 4)
	err := bf.InitFromFile(filepath.Join(t.TempDir(), "missing.txt"))
	assertBloomFilterCode(t, err, knifer.ErrCodeNotFound)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("InitFromFile should preserve os not-exist cause: %v", err)
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
