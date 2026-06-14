package vblf_test

import (
	"testing"

	"github.com/imajinyun/go-knifer/vblf"
)

func TestFacadeHashFunctions(t *testing.T) {
	// smoke test: hash functions should return consistent values
	h1 := vblf.BloomRSHash("abc")
	h2 := vblf.BloomRSHash("abc")
	if h1 != h2 {
		t.Fatal("hash function should be deterministic")
	}
}

func TestFacadeHashAliasesAndConstructors(t *testing.T) {
	inputs := []struct {
		name string
		hash func(string) int32
	}{
		{"rs", vblf.RsHash},
		{"js", vblf.JsHash},
		{"pjw", vblf.PjwHash},
		{"elf", vblf.ElfHash},
		{"bkdr", vblf.BkdrHash},
		{"sdbm", vblf.SdbmHash},
		{"djb", vblf.DjbHash},
		{"ap", vblf.ApHash},
		{"fnv", vblf.FnvHashString},
		{"bloom-js", vblf.BloomJSHash},
		{"bloom-elf", vblf.BloomELFHash},
		{"bloom-bkdr", vblf.BloomBKDRHash},
		{"bloom-sdbm", vblf.BloomSDBMHash},
		{"bloom-djb", vblf.BloomDJBHash},
		{"bloom-fnv", vblf.BloomFNVHash},
	}
	for _, in := range inputs {
		first := in.hash("abc")
		second := in.hash("abc")
		if first != second {
			t.Fatalf("%s hash should be deterministic", in.name)
		}
	}
	hfFirst := vblf.HfHash("abc")
	hfSecond := vblf.HfHash("abc")
	hfIPFirst := vblf.HfIpHash("127.0.0.1")
	hfIPSecond := vblf.HfIpHash("127.0.0.1")
	tianlFirst := vblf.TianlHash("abc")
	tianlSecond := vblf.TianlHash("abc")
	if hfFirst != hfSecond || hfIPFirst != hfIPSecond || tianlFirst != tianlSecond {
		t.Fatal("int64 hash aliases should be deterministic")
	}
	javaFirst := vblf.JavaDefaultHash("abc")
	javaSecond := vblf.JavaDefaultHash("abc")
	if javaFirst != javaSecond {
		t.Fatal("JavaDefaultHash should be deterministic")
	}

	constructors := []func(int64) *vblf.FuncFilter{
		vblf.NewDefaultFilter,
		vblf.NewDefaultBloomFilter,
		vblf.NewELFFilter,
		vblf.NewFNVFilter,
		vblf.NewHfFilter,
		vblf.NewHfIpFilter,
		vblf.NewJSFilter,
		vblf.NewPJWFilter,
		vblf.NewRSFilter,
		vblf.NewSDBMFilter,
		vblf.NewTianlFilter,
	}
	for _, newFilter := range constructors {
		filter := newFilter(1 << 20)
		if !filter.Add("abc") || !filter.Contains("abc") {
			t.Fatal("constructor-created filter should contain added value")
		}
	}
}

func TestFacadeFilterConstructorsWithErrors(t *testing.T) {
	if _, err := vblf.NewFuncFilterE(1<<20, vblf.HfHash); err != nil {
		t.Fatalf("NewFuncFilterE: %v", err)
	}
	if _, err := vblf.NewFuncFilterWithMachineNumE(1<<20, vblf.BloomMachine64, vblf.HfHash); err != nil {
		t.Fatalf("NewFuncFilterWithMachineNumE: %v", err)
	}
	if f := vblf.NewFuncFilterWithMachineNum(1<<20, vblf.BloomMachine32, vblf.HfHash); !f.Add("x") || !f.Contains("x") {
		t.Fatal("NewFuncFilterWithMachineNum filter failed")
	}
	if f := vblf.NewFuncFilterWithOptions(vblf.WithMaxValue(1<<20), vblf.WithMachineNum(vblf.BloomMachine64), vblf.WithHashFunc(vblf.HfHash)); !f.Add("x") || !f.Contains("x") {
		t.Fatal("NewFuncFilterWithOptions filter failed")
	}
	if _, err := vblf.NewFuncFilterWithOptionsE(vblf.WithMaxValue(1<<20), vblf.WithHashFunc(vblf.HfHash)); err != nil {
		t.Fatalf("NewFuncFilterWithOptionsE: %v", err)
	}
	if _, err := vblf.NewBitMapBloomFilterE(5); err != nil {
		t.Fatalf("NewBitMapBloomFilterE: %v", err)
	}
	if _, err := vblf.NewBitMapBloomFilterWithOptionsE(vblf.WithBitMapSize(5)); err != nil {
		t.Fatalf("NewBitMapBloomFilterWithOptionsE: %v", err)
	}
	if _, err := vblf.NewBitMapBloomFilterWithFiltersE(5, vblf.NewRSFilter(1<<20)); err != nil {
		t.Fatalf("NewBitMapBloomFilterWithFiltersE: %v", err)
	}
	if _, err := vblf.NewBitSetBloomFilterE(1000, 5, 3); err != nil {
		t.Fatalf("NewBitSetBloomFilterE: %v", err)
	}
	if _, err := vblf.NewBitSetBloomFilterWithOptionsE(vblf.WithBitSetCapacity(1000), vblf.WithExpectedElements(5), vblf.WithHashFunctionNumber(3)); err != nil {
		t.Fatalf("NewBitSetBloomFilterWithOptionsE: %v", err)
	}
	longMap := vblf.NewLongMap(100)
	longMap.Add(7)
	if !longMap.Contains(7) {
		t.Fatal("NewLongMap should contain added value")
	}
}
