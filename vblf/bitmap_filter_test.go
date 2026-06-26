package vblf_test

import (
	"testing"

	"github.com/imajinyun/knifer-go/vblf"
)

func TestFacadeBitMapBloomFilterWithOptions(t *testing.T) {
	bitmap := vblf.NewBitMapBloomFilterWithOptions(
		vblf.WithBitMapSize(5),
		vblf.WithBloomFilters(vblf.NewFNVFilter(1<<20), vblf.NewRSFilter(1<<20)),
	)
	if !bitmap.Add("world") || !bitmap.Contains("world") {
		t.Fatal("expected options-created bitmap filter to contain value")
	}
}

func TestFacadeBitMap(t *testing.T) {
	bm := vblf.NewIntMap(100)
	bm.Add(42)
	if !bm.Contains(42) {
		t.Fatal("expected bitmap to contain 42")
	}
}

func TestFacadeNewBitMapBloomFilter(t *testing.T) {
	bf := vblf.NewBitMapBloomFilter(5)
	if bf == nil {
		t.Fatal("NewBitMapBloomFilter returned nil")
	}
	bf.Add("test")
	if !bf.Contains("test") {
		t.Fatal("expected NewBitMapBloomFilter to contain 'test'")
	}
}

func TestFacadeNewBitMapBloomFilterWithFilters(t *testing.T) {
	f1 := vblf.NewDefaultBloomFilter(1 << 20)
	f2 := vblf.NewDefaultBloomFilter(1 << 20)
	bf := vblf.NewBitMapBloomFilterWithFilters(5, f1, f2)
	if bf == nil {
		t.Fatal("NewBitMapBloomFilterWithFilters returned nil")
	}
	bf.Add("test")
	if !bf.Contains("test") {
		t.Fatal("expected filter to contain 'test'")
	}
}
