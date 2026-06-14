package vmap

import (
	"reflect"
	"testing"
)

func TestMapSetAlgebraSelectionMutationComparisonFacades(t *testing.T) {
	a := map[string]int{"a": 1, "b": 2}
	b := map[string]int{"b": 20, "c": 3}
	c := map[string]int{"b": 200, "d": 4}

	dst := map[string]int{"a": 1, "shared": 1}
	MergeWithOverwrite(dst, map[string]int{"shared": 2}, map[string]int{"x": 3})
	if !reflect.DeepEqual(dst, map[string]int{"a": 1, "shared": 2, "x": 3}) {
		t.Fatalf("MergeWithOverwrite dst = %#v", dst)
	}
	MergeWithoutOverwrite(dst, map[string]int{"a": 9, "y": 4})
	if !reflect.DeepEqual(dst, map[string]int{"a": 1, "shared": 2, "x": 3, "y": 4}) {
		t.Fatalf("MergeWithoutOverwrite dst = %#v", dst)
	}

	if got := MergeFunc(func(old, new int) int { return old + new }, a, b, c); !reflect.DeepEqual(got, map[string]int{"a": 1, "b": 222, "c": 3, "d": 4}) {
		t.Fatalf("MergeFunc = %#v", got)
	}
	if got := Intersect(a, b, c); !reflect.DeepEqual(got, map[string]int{"b": 200}) {
		t.Fatalf("Intersect = %#v", got)
	}
	if got := Diff(a, b); !reflect.DeepEqual(got, map[string]int{"a": 1}) {
		t.Fatalf("Diff = %#v", got)
	}
	if got := SymmetricDiff(a, b); !reflect.DeepEqual(got, map[string]int{"a": 1, "c": 3}) {
		t.Fatalf("SymmetricDiff = %#v", got)
	}
	if got := Pick(a, "b", "missing"); !reflect.DeepEqual(got, map[string]int{"b": 2}) {
		t.Fatalf("Pick = %#v", got)
	}
	if got := Omit(a, "a"); !reflect.DeepEqual(got, map[string]int{"b": 2}) {
		t.Fatalf("Omit = %#v", got)
	}

	updated := Update[string, int](nil, a)
	if !reflect.DeepEqual(updated, a) {
		t.Fatalf("Update nil dst = %#v", updated)
	}
	clone := Clone(a)
	clone["a"] = 99
	if a["a"] != 1 || clone["a"] != 99 {
		t.Fatalf("Clone should not alias input: original=%#v clone=%#v", a, clone)
	}
	if got := Clone[string, int](nil); got == nil || len(got) != 0 {
		t.Fatalf("Clone nil = %#v", got)
	}
	Clear(updated)
	if len(updated) != 0 {
		t.Fatalf("Clear = %#v", updated)
	}
	if !Equal(a, map[string]int{"a": 1, "b": 2}) || Equal(a, b) {
		t.Fatal("Equal facade returned unexpected result")
	}
	if !EqualFunc(map[string]int{"a": 1}, map[string]string{"a": "1"}, func(i int, s string) bool { return string(rune('0'+i)) == s }) {
		t.Fatal("EqualFunc = false, want true")
	}
}
