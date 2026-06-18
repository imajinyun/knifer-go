package json

import "testing"

func TestObjectConfig(t *testing.T) {
	cfg := NewConfig()
	obj := NewJSONObjectWithConfig(cfg)
	if obj.Config() != cfg {
		t.Fatal("Config() mismatch")
	}
}

func TestObjectHas(t *testing.T) {
	obj := NewJSONObject().Set("a", 1)
	if !obj.Has("a") {
		t.Fatal("Has('a') = false")
	}
	if obj.Has("missing") {
		t.Fatal("Has('missing') = true")
	}
}

func TestObjectGetOrDefault(t *testing.T) {
	obj := NewJSONObject().Set("a", 42)
	if got := obj.GetOrDefault("a", -1); got != int64(42) {
		t.Fatalf("GetOrDefault('a') = %v", got)
	}
	if got := obj.GetOrDefault("missing", "def"); got != "def" {
		t.Fatalf("GetOrDefault(missing) = %v", got)
	}
}

func TestObjectPutRemoveToMap(t *testing.T) {
	obj := NewJSONObject()
	obj.Put("a", 1).Put("b", 2)
	if obj.Len() != 2 {
		t.Fatalf("after Put, Len = %d", obj.Len())
	}
	if !obj.Remove("a") {
		t.Fatal("Remove('a') = false")
	}
	if obj.Remove("missing") {
		t.Fatal("Remove('missing') = true")
	}
	m := obj.ToMap()
	if len(m) != 1 || m["b"] != int64(2) {
		t.Fatalf("ToMap = %v", m)
	}
}

func TestObjectGetIntOr(t *testing.T) {
	obj := NewJSONObject().Set("a", 42)
	if got := obj.GetIntOr("a", -1); got != 42 {
		t.Fatalf("GetIntOr('a') = %d", got)
	}
	if got := obj.GetIntOr("missing", -1); got != -1 {
		t.Fatalf("GetIntOr(missing) = %d", got)
	}
}

func TestObjectToString(t *testing.T) {
	obj := NewJSONObject().Set("a", 1)
	if s := obj.ToString(); s != `{"a":1}` {
		t.Fatalf("ToString = %s", s)
	}
}

func TestObjectMarshalJSON(t *testing.T) {
	obj := NewJSONObject().Set("a", 1)
	b, err := obj.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}
	if string(b) != `{"a":1}` {
		t.Fatalf("MarshalJSON = %s", b)
	}
}

func TestObjectGetByPathAndPutByPath(t *testing.T) {
	obj := NewJSONObject().Set("nested", NewJSONObject().Set("k", "v"))
	if got := obj.GetByPath("nested.k"); got != "v" {
		t.Fatalf("GetByPath = %v", got)
	}
	if err := obj.PutByPath("nested.k", "new"); err != nil {
		t.Fatalf("PutByPath error = %v", err)
	}
	if got := obj.GetByPath("nested.k"); got != "new" {
		t.Fatalf("after PutByPath = %v", got)
	}
}
