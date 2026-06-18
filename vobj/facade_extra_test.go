package vobj_test

import (
	"testing"

	"github.com/imajinyun/go-knifer/vobj"
)

func TestFacadeIsNotEmpty(t *testing.T) {
	if vobj.IsNotEmpty("") {
		t.Fatal("IsNotEmpty('') should be false")
	}
	if !vobj.IsNotEmpty("hello") {
		t.Fatal("IsNotEmpty('hello') should be true")
	}
	if vobj.IsNotEmpty(nil) {
		t.Fatal("IsNotEmpty(nil) should be false")
	}
}

func TestFacadeCloneWithOptions(t *testing.T) {
	src := map[string]int{"a": 1}
	got, err := vobj.CloneWithOptions(src)
	if err != nil {
		t.Fatalf("CloneWithOptions error = %v", err)
	}
	if got["a"] != 1 {
		t.Fatalf("CloneWithOptions = %#v", got)
	}
	// Verify it's a deep copy
	src["a"] = 2
	if got["a"] != 1 {
		t.Fatal("CloneWithOptions was not a deep copy")
	}
}

func TestFacadeCloneByStreamWithOptions(t *testing.T) {
	src := []int{1, 2, 3}
	got, err := vobj.CloneByStreamWithOptions(src)
	if err != nil {
		t.Fatalf("CloneByStreamWithOptions error = %v", err)
	}
	if len(got) != 3 || got[0] != 1 {
		t.Fatalf("CloneByStreamWithOptions = %#v", got)
	}
}

func TestFacadeSerializeWithOptions(t *testing.T) {
	data, err := vobj.SerializeWithOptions("hello")
	if err != nil {
		t.Fatalf("SerializeWithOptions error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("SerializeWithOptions returned empty data")
	}
}

func TestFacadeRegisterAndRegisterName(t *testing.T) {
	// Use distinct types to avoid gob duplicate registration panics
	vobj.Register(struct{ X string }{})
	vobj.RegisterName("customType", struct{ Y int }{})
}
