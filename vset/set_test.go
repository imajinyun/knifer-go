package vset_test

import (
	"encoding/json"
	"sort"
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

func TestVSetGenericFacade(t *testing.T) {
	s := vset.New("a", "b")
	s.Add("c")

	if !s.Equal(vset.New("a", "b", "c")) {
		t.Fatalf("generic set = %v, want a/b/c", s.Members())
	}
	if got := s.Sub(vset.New("a")); !got.Equal(vset.New("b", "c")) {
		t.Fatalf("generic Sub() = %v, want b/c", got.Members())
	}
}

func TestVSetGenericContainsUnionIntersect(t *testing.T) {
	s := vset.New(1, 2, 3)
	if !s.Contains(2) || s.Contains(4) {
		t.Fatal("Set.Contains failed")
	}
	u := s.Union(vset.New(3, 4, 5))
	if !u.Contains(4) || !u.Contains(1) || u.Contains(6) {
		t.Fatal("Set.Union failed")
	}
	i := s.Intersect(vset.New(2, 3, 4))
	if !i.Contains(2) || i.Contains(1) || !i.Contains(3) {
		t.Fatal("Set.Intersect failed")
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

func TestVSetGenericJSONRoundTrip(t *testing.T) {
	original := vset.New(1, 2, 3)
	b, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}

	var decoded vset.Set[int]
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Equal(original) {
		t.Fatalf("decoded = %v, want %v", decoded.Members(), original.Members())
	}
}

func TestVSetFacadeExplicitJSONAndYAMLHelpers(t *testing.T) {
	s := vset.New("a", "b")
	s.Remove("b")
	if !s.Equal(vset.New("a")) {
		t.Fatalf("Remove() set = %v", s.Members())
	}
	if text := s.String(); text == "" {
		t.Fatal("String() returned empty text")
	}

	marshalCalled := false
	b, err := s.MarshalJSONWithOptions(vset.WithSetMarshalFunc(func(v any) ([]byte, error) {
		marshalCalled = true
		return json.Marshal(v)
	}))
	if err != nil || !marshalCalled {
		t.Fatalf("MarshalJSONWithOptions called=%v err=%v", marshalCalled, err)
	}

	unmarshalCalled := false
	var decoded vset.Set[string]
	if err := decoded.UnmarshalJSONWithOptions(b, vset.WithSetUnmarshalFunc(func(data []byte, v any) error {
		unmarshalCalled = true
		return json.Unmarshal(data, v)
	})); err != nil {
		t.Fatal(err)
	}
	if !unmarshalCalled || !decoded.Equal(s) {
		t.Fatalf("UnmarshalJSONWithOptions called=%v decoded=%v", unmarshalCalled, decoded.Members())
	}

	yamlValue, err := s.MarshalYAML()
	if err != nil {
		t.Fatal(err)
	}
	members, ok := yamlValue.([]string)
	if !ok || len(members) != 1 || members[0] != "a" {
		t.Fatalf("MarshalYAML = %#v", yamlValue)
	}

	var yamlDecoded vset.Set[string]
	err = yamlDecoded.UnmarshalYAML(func(v any) error {
		out := v.(*[]string)
		*out = []string{"x", "y"}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	got := yamlDecoded.Members()
	sort.Strings(got)
	if len(got) != 2 || got[0] != "x" || got[1] != "y" {
		t.Fatalf("UnmarshalYAML members = %v", got)
	}
}
