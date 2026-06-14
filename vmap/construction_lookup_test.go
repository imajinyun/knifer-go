package vmap

import (
	"reflect"
	"sort"
	"testing"
)

func TestMapFacade(t *testing.T) {
	fromPairs := FromPairs(
		Pair[string, int]{Key: "a", Value: 1},
		Pair[string, int]{Key: "b", Value: 2},
	)
	if fromPairs["a"] != 1 || fromPairs["b"] != 2 {
		t.Fatalf("FromPairs failed: %v", fromPairs)
	}
	fromAny, err := OfE[string, int]("a", 1, "b", 2)
	if err != nil || fromAny["a"] != 1 || fromAny["b"] != 2 {
		t.Fatalf("OfE failed: %v, %v", fromAny, err)
	}

	m := map[string]int{"a": 1, "b": 2}
	if IsEmpty(m) || !IsNotEmpty(m) {
		t.Fatal("empty checks failed")
	}
	keys := Keys(m)
	sort.Strings(keys)
	if len(keys) != 2 || keys[0] != "a" || keys[1] != "b" {
		t.Fatalf("Keys failed: %v", keys)
	}
	values := Values(m)
	sort.Ints(values)
	if len(values) != 2 || values[0] != 1 || values[1] != 2 {
		t.Fatalf("Values failed: %v", values)
	}
	inv := Inverse(m)
	if inv[1] != "a" || inv[2] != "b" {
		t.Fatalf("Inverse failed: %v", inv)
	}
	merged := Merge(map[string]int{"a": 1}, map[string]int{"a": 9, "b": 2})
	if merged["a"] != 9 || merged["b"] != 2 {
		t.Fatalf("Merge failed: %v", merged)
	}
}

func TestMapConstructionLookupAndPredicateFacades(t *testing.T) {
	if got := New[string, int](); got == nil || len(got) != 0 {
		t.Fatalf("New = %#v", got)
	}
	if got := NewWithCap[string, int](-1); got == nil || len(got) != 0 {
		t.Fatalf("NewWithCap negative = %#v", got)
	}
	if got := Of[string, int]("a", 1, "a", 2); got["a"] != 2 {
		t.Fatalf("Of duplicate = %#v", got)
	}
	if _, err := OfE[string, int]("a"); err == nil {
		t.Fatal("OfE odd args error = nil")
	}
	if _, err := OfE[string, int](1, 2); err == nil {
		t.Fatal("OfE wrong key type error = nil")
	}
	if _, err := OfE[string, int]("a", "bad"); err == nil {
		t.Fatal("OfE wrong value type error = nil")
	}

	original := map[string]int{"a": 1}
	if got := OrEmpty(original); !reflect.DeepEqual(got, original) {
		t.Fatalf("OrEmpty existing = %#v", got)
	}
	if got := OrEmpty[string, int](nil); got == nil || len(got) != 0 {
		t.Fatalf("OrEmpty nil = %#v", got)
	}

	m := map[string]int{"a": 1, "b": 2, "c": 3}
	if !ContainsKey(m, "a") || ContainsKey(m, "missing") {
		t.Fatal("ContainsKey facade returned unexpected result")
	}
	if !ContainsValue(m, 2) || ContainsValue(m, 9) {
		t.Fatal("ContainsValue facade returned unexpected result")
	}
	if !Some(m, func(_ string, v int) bool { return v > 2 }) {
		t.Fatal("Some = false, want true")
	}
	if Every(m, func(_ string, v int) bool { return v < 3 }) {
		t.Fatal("Every = true, want false")
	}
	if got := Get(m, "missing"); got != 0 {
		t.Fatalf("Get missing = %d", got)
	}
	if got := GetOr(m, "missing", 9); got != 9 {
		t.Fatalf("GetOr missing = %d", got)
	}
	if got, ok := GetAny(m, "missing", "b"); !ok || got != 2 {
		t.Fatalf("GetAny = %d, %v", got, ok)
	}
	if _, ok := GetAny(m, "missing"); ok {
		t.Fatal("GetAny missing ok = true")
	}
	if k, v, ok := Find(m, func(_ string, v int) bool { return v == 3 }); !ok || k != "c" || v != 3 {
		t.Fatalf("Find = %q, %d, %v", k, v, ok)
	}
	if k, ok := FindKey(m, func(v int) bool { return v == 2 }); !ok || k != "b" {
		t.Fatalf("FindKey = %q, %v", k, ok)
	}
}
