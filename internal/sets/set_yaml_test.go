package sets

import (
	"testing"
)

func TestSetString(t *testing.T) {
	s := New("a", "b")
	got := s.String()
	if len(got) < 4 || got[:3] != "set" {
		t.Fatalf("String() = %q, want set[...]", got)
	}
}

func TestSetMarshalYAML(t *testing.T) {
	s := New(1, 2, 3)
	got, err := s.MarshalYAML()
	if err != nil {
		t.Fatal(err)
	}
	items, ok := got.([]int)
	if !ok {
		t.Fatalf("MarshalYAML returned %T, want []int", got)
	}
	if len(items) != 3 {
		t.Fatalf("MarshalYAML returned %v items, want 3", len(items))
	}
}

func TestSetUnmarshalYAML(t *testing.T) {
	var s Set[string]
	err := s.UnmarshalYAML(func(v any) error {
		out := v.(*[]string)
		*out = []string{"x", "y"}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !s.Contains("x") || !s.Contains("y") || len(s) != 2 {
		t.Fatalf("UnmarshalYAML resulted in %v, want {x, y}", s.Members())
	}
}

func TestSetUnmarshalYAMLError(t *testing.T) {
	var s Set[string]
	err := s.UnmarshalYAML(func(v any) error {
		return nil // return nil without writing to v — s should remain empty
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(s) != 0 {
		t.Fatalf("UnmarshalYAML empty result = %v", s.Members())
	}
}
