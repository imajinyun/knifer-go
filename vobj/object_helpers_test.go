package vobj_test

import (
	"testing"

	"github.com/imajinyun/knifer-go/vobj"
)

func TestFacadeObjectHelpers(t *testing.T) {
	if !vobj.Equal(1, int64(1)) || !vobj.Contains([]string{"go", "tool"}, "go") {
		t.Fatal("equality or contains failed")
	}
	if !vobj.IsEmpty([]int{}) || vobj.Length(map[string]int{"a": 1}) != 1 {
		t.Fatal("empty or length failed")
	}
	name := "go"
	if vobj.DefaultIfNil(&name, "x") != "go" || vobj.DefaultIfNil[string](nil, "x") != "x" {
		t.Fatal("defaults failed")
	}
	if got := vobj.Apply(&name, func(s string) int { return len(s) }); got != 2 {
		t.Fatal("apply failed")
	}
}
