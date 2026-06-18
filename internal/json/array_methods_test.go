package json

import (
	"testing"
)

func TestArrayConfig(t *testing.T) {
	cfg := NewConfig()
	arr := NewJSONArrayWithConfig(cfg)
	if arr.Config() != cfg {
		t.Fatal("Config() mismatch")
	}
}

func TestArrayGetOrDefault(t *testing.T) {
	arr := NewJSONArray().Add(42)
	if got := arr.GetOrDefault(0, -1); got != int64(42) {
		t.Fatalf("GetOrDefault(0) = %v", got)
	}
	if got := arr.GetOrDefault(5, -1); got != -1 {
		t.Fatalf("GetOrDefault(missing) = %v", got)
	}
}

func TestArrayRange(t *testing.T) {
	arr := NewJSONArray().Add(1).Add(2).Add(3)
	var sum int64
	arr.Range(func(i int, v any) bool {
		sum += v.(int64)
		return true
	})
	if sum != 6 {
		t.Fatalf("Range sum = %d", sum)
	}
	sum = 0
	arr.Range(func(i int, v any) bool {
		sum += v.(int64)
		return false // stop immediately
	})
	if sum != 1 {
		t.Fatalf("Range stop = %d", sum)
	}
}

func TestArrayGetInt64AndFloat64(t *testing.T) {
	arr := NewJSONArray().Add(42).Add(3.14)
	if got := arr.GetInt64(0); got != 42 {
		t.Fatalf("GetInt64(0) = %d", got)
	}
	if got := arr.GetInt64(9); got != 0 {
		t.Fatalf("GetInt64(out-of-range) = %d", got)
	}
	if got := arr.GetFloat64(1); got != 3.14 {
		t.Fatalf("GetFloat64(1) = %v", got)
	}
	if got := arr.GetFloat64Or(1, 0); got != 3.14 {
		t.Fatalf("GetFloat64Or(1) = %v", got)
	}
	if got := arr.GetFloat64Or(9, 1.5); got != 1.5 {
		t.Fatalf("GetFloat64Or(default) = %v", got)
	}
}

func TestArrayGetJSONObjectAndArray(t *testing.T) {
	inner := NewJSONObject().Set("k", "v")
	nested := NewJSONArray().Add(inner).Add(NewJSONArray().Add(1))
	if got := nested.GetJSONObject(0); got == nil || got.GetString("k") != "v" {
		t.Fatalf("GetJSONObject(0) = %v", got)
	}
	if got := nested.GetJSONArray(1); got == nil || got.Len() != 1 {
		t.Fatalf("GetJSONArray(1) = %v", got)
	}
	if got := nested.GetJSONArray(9); got != nil {
		t.Fatalf("GetJSONArray(out-of-range) = %v", got)
	}
}

func TestArrayStringAndToString(t *testing.T) {
	arr := NewJSONArray().Add(1).Add("x")
	if s := arr.String(); s != `[1,"x"]` {
		t.Fatalf("String = %s", s)
	}
	if s := arr.ToString(); s != `[1,"x"]` {
		t.Fatalf("ToString = %s", s)
	}
}

func TestArrayMarshalJSON(t *testing.T) {
	arr := NewJSONArray().Add(1).Add("x")
	b, err := arr.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}
	if string(b) != `[1,"x"]` {
		t.Fatalf("MarshalJSON = %s", b)
	}
}

func TestArrayGetByPathAndPutByPath(t *testing.T) {
	arr := NewJSONArray().Add(NewJSONObject().Set("k", "v"))
	if got := arr.GetByPath("[0].k"); got != "v" {
		t.Fatalf("GetByPath = %v", got)
	}
	if err := arr.PutByPath("[0].k", "new"); err != nil {
		t.Fatalf("PutByPath error = %v", err)
	}
	if got := arr.GetByPath("[0].k"); got != "new" {
		t.Fatalf("after PutByPath = %v", got)
	}
}
