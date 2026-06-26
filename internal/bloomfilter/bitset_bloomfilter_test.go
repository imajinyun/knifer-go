package bloomfilter

import (
	"io"
	"strings"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
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
