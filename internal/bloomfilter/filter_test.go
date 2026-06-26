package bloomfilter

import (
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

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

func TestBloomFilterInterface(t *testing.T) {
	var _ BloomFilter = (*BitSetBloomFilter)(nil)
	var _ BloomFilter = (*BitMapBloomFilter)(nil)
	var _ BloomFilter = (*FuncFilter)(nil)
}

func TestNamedHashFilterConstructors(t *testing.T) {
	if f := NewHfFilter(1024); f == nil || !f.Add("x") || !f.Contains("x") {
		t.Fatal("NewHfFilter failed")
	}
	if f := NewHfIpFilter(1024); f == nil || !f.Add("x") || !f.Contains("x") {
		t.Fatal("NewHfIpFilter failed")
	}
	if f := NewTianlFilter(1024); f == nil || !f.Add("x") || !f.Contains("x") {
		t.Fatal("NewTianlFilter failed")
	}
}
