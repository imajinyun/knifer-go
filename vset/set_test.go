package vset_test

import (
	"encoding/json"
	"testing"

	"github.com/imajinyun/go-knifer/vset"
)

func TestVSetFacadeConstructorsAndMethods(t *testing.T) {
	stringSet := vset.NewString("a", "b")
	stringSet.Add("c")
	if !stringSet.Contains("c") {
		t.Fatal("String.Add() item should exist")
	}
	if got := stringSet.Sub(vset.NewString("a")); !got.Equal(vset.NewString("b", "c")) {
		t.Fatalf("String.Sub() = %v", got.Members())
	}

	intSet := vset.NewInt(1, 2).Union(vset.NewInt(2, 3))
	if !intSet.Equal(vset.NewInt(1, 2, 3)) {
		t.Fatalf("Int.Union() = %v", intSet.Members())
	}
	if got := vset.NewInt32(1, 2).Intersect(vset.NewInt32(2)); !got.Equal(vset.NewInt32(2)) {
		t.Fatalf("Int32.Intersect() = %v", got.Members())
	}
	if got := vset.NewInt64(1, 2).Sub(vset.NewInt64(1)); !got.Equal(vset.NewInt64(2)) {
		t.Fatalf("Int64.Sub() = %v", got.Members())
	}
	if got := vset.NewUint(1, 2).Union(vset.NewUint(3)); !got.Equal(vset.NewUint(1, 2, 3)) {
		t.Fatalf("Uint.Union() = %v", got.Members())
	}
	if got := vset.NewUint32(1, 2).Intersect(vset.NewUint32(2)); !got.Equal(vset.NewUint32(2)) {
		t.Fatalf("Uint32.Intersect() = %v", got.Members())
	}
	if got := vset.NewUint64(1, 2).Sub(vset.NewUint64(1)); !got.Equal(vset.NewUint64(2)) {
		t.Fatalf("Uint64.Sub() = %v", got.Members())
	}
}

func TestVSetFacadeJSONRoundTrip(t *testing.T) {
	original := vset.NewString("go", "knifer")
	b, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}

	var decoded vset.String
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Equal(original) {
		t.Fatalf("decoded = %v, want %v", decoded.Members(), original.Members())
	}
}
