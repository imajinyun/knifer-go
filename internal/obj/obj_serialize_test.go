package obj

import (
	"reflect"
	"testing"
)

func TestCloneSerializeAndDeserialize(t *testing.T) {
	src := sample{Name: "n", Tags: []string{"a"}}
	clone, err := Clone(src)
	if err != nil {
		t.Fatalf("Clone: %v", err)
	}
	clone.Tags[0] = "b"
	if src.Tags[0] != "a" {
		t.Fatal("clone is not independent")
	}
	data, err := Serialize(src)
	if err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	var out sample
	if err := Deserialize(data, &out); err != nil || !reflect.DeepEqual(out, src) {
		t.Fatalf("Deserialize: %#v %v", out, err)
	}
}
