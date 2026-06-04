package vobj_test

import (
	"testing"

	"github.com/imajinyun/go-knifer/vobj"
)

type record struct {
	Name string
	Tags []string
}

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

func TestFacadeCloneAndSerialize(t *testing.T) {
	src := record{Name: "go", Tags: []string{"tool"}}
	clone, err := vobj.Clone(src)
	if err != nil {
		t.Fatalf("Clone: %v", err)
	}
	clone.Tags[0] = "sdk"
	if src.Tags[0] != "tool" {
		t.Fatal("clone changed source")
	}
	data, err := vobj.Serialize(src)
	if err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	var out record
	if err := vobj.Deserialize(data, &out); err != nil || out.Name != src.Name {
		t.Fatalf("Deserialize: %#v %v", out, err)
	}
}

func TestFacadeSerializeExtended(t *testing.T) {
	src := record{Name: "go", Tags: []string{"tool"}}

	data, err := vobj.Serialize(src)
	if err != nil {
		t.Fatalf("Serialize: %v", err)
	}

	out, err := vobj.DeserializeTo[record](data)
	if err != nil {
		t.Fatalf("DeserializeTo: %v", err)
	}
	if out.Name != src.Name || len(out.Tags) != 1 || out.Tags[0] != "tool" {
		t.Fatalf("DeserializeTo mismatch: %#v", out)
	}

	nilData := vobj.SerializeOrNil(src)
	if nilData == nil {
		t.Fatal("SerializeOrNil should not return nil for valid input")
	}
}
