package vblf_test

import (
	"testing"

	"github.com/imajinyun/knifer-go/vblf"
)

func TestFacadeFuncFilter(t *testing.T) {
	f := vblf.NewDefaultBloomFilter(1000)
	f.Add("test")
	if !f.Contains("test") {
		t.Fatal("expected func filter to contain 'test'")
	}
}

func TestFacadeFuncFilterWithOptions(t *testing.T) {
	fn := vblf.NewFuncFilterWithOptions(
		vblf.WithMaxValue(1000),
		vblf.WithMachineNum(vblf.BloomMachine64),
		vblf.WithHashFunc(func(s string) int64 { return int64(vblf.JavaDefaultHash(s)) }),
	)
	if !fn.Add("test") || !fn.Contains("test") {
		t.Fatal("expected options-created func filter to contain value")
	}

	alias := vblf.NewFuncFilterWithOptions(vblf.WithMaxValue(1000), vblf.WithHashFunc(func(s string) int64 {
		return int64(vblf.JavaDefaultHash(s))
	}))
	if !alias.Add("alias") || !alias.Contains("alias") {
		t.Fatal("expected NewFuncFilterWithOptions filter to contain value")
	}
}

func TestFacadeNewFuncFilter(t *testing.T) {
	fn := vblf.NewFuncFilter(1000, func(s string) int64 { return int64(len(s)) })
	if fn == nil {
		t.Fatal("NewFuncFilter returned nil")
	}
	fn.Add("hello")
	if !fn.Contains("hello") {
		t.Fatal("expected NewFuncFilter to contain 'hello'")
	}
}
