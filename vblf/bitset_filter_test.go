package vblf_test

import (
	"io"
	"strings"
	"testing"

	"github.com/imajinyun/knifer-go/vblf"
)

func TestFacadeBitSetBloomFilter(t *testing.T) {
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

func TestFacadeBitSetBloomFilterWithOptions(t *testing.T) {
	bitset := vblf.NewBitSetBloomFilterWithOptions(
		vblf.WithBitSetCapacity(1000),
		vblf.WithExpectedElements(5),
		vblf.WithHashFunctionNumber(3),
	)
	if !bitset.Add("hello") || !bitset.Contains("hello") {
		t.Fatal("expected NewBitSetBloomFilterWithOptions filter to contain value")
	}
}

func TestFacadeBloomFilterFileOptions(t *testing.T) {
	bf := vblf.NewBitSetBloomFilter(1000, 5, 3)
	openedPath := ""
	err := bf.InitFromFileWithOptions("virtual.txt", vblf.WithOpenFile(func(path string) (io.ReadCloser, error) {
		openedPath = path
		return io.NopCloser(strings.NewReader("facade\n")), nil
	}))
	if err != nil {
		t.Fatalf("InitFromFileWithOptions: %v", err)
	}
	if openedPath != "virtual.txt" || !bf.Contains("facade") {
		t.Fatalf("custom open not applied path=%q contains=%v", openedPath, bf.Contains("facade"))
	}

	bf = vblf.NewBitSetBloomFilter(1000, 5, 3)
	if err := bf.InitFromReader(strings.NewReader("reader\n")); err != nil || !bf.Contains("reader") {
		t.Fatalf("InitFromReader contains=%v err=%v", bf.Contains("reader"), err)
	}
}

func TestFacadeInitFromFileWithOptionsStandalone(t *testing.T) {
	bf := vblf.NewBitSetBloomFilter(1000, 5, 3)
	var opened bool
	err := vblf.InitFromFileWithOptions(bf, "virtual.data", vblf.WithOpenFile(func(path string) (io.ReadCloser, error) {
		opened = true
		return io.NopCloser(strings.NewReader("standalone\n")), nil
	}))
	if err != nil || !opened || !bf.Contains("standalone") {
		t.Fatalf("InitFromFileWithOptions standalone err=%v opened=%v", err, opened)
	}
}

func TestFacadeInitFromReaderStandalone(t *testing.T) {
	bf := vblf.NewBitSetBloomFilter(1000, 5, 3)
	if err := vblf.InitFromReader(bf, strings.NewReader("standalone\n")); err != nil || !bf.Contains("standalone") {
		t.Fatalf("InitFromReader standalone err=%v", err)
	}
}
