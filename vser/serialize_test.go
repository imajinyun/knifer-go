package vser_test

import (
	"testing"

	"github.com/imajinyun/go-knifer/vser"
)

type record struct {
	Name string
	Tags []string
}

func TestFacadeSerializeCloneAndDeserialize(t *testing.T) {
	src := record{Name: "go", Tags: []string{"tool"}}
	clone, err := vser.Clone(src)
	if err != nil {
		t.Fatalf("Clone: %v", err)
	}
	clone.Tags[0] = "sdk"
	if src.Tags[0] != "tool" {
		t.Fatal("clone changed source")
	}
	data, err := vser.Serialize(src)
	if err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	out, err := vser.DeserializeTo[record](data, record{})
	if err != nil {
		t.Fatalf("DeserializeTo: %v", err)
	}
	if out.Name != src.Name || len(out.Tags) != 1 || out.Tags[0] != "tool" {
		t.Fatalf("DeserializeTo mismatch: %#v", out)
	}
}
