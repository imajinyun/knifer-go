package sets

import (
	"encoding/json"
	"testing"
)

func TestStringSetOperations(t *testing.T) {
	s := NewString("a", "b", "b")
	s.Add("c")
	s.Remove("a", "missing")

	if s.Contains("a") {
		t.Fatal("removed item should not exist")
	}
	if !s.Equal(NewString("b", "c")) {
		t.Fatalf("set = %v, want b/c", s.Members())
	}

	if got := s.Sub(NewString("c")); !got.Equal(NewString("b")) {
		t.Fatalf("Sub() = %v, want b", got.Members())
	}
	if got := s.Union(NewString("d")); !got.Equal(NewString("b", "c", "d")) {
		t.Fatalf("Union() = %v, want b/c/d", got.Members())
	}
	if got := s.Intersect(NewString("c", "d")); !got.Equal(NewString("c")) {
		t.Fatalf("Intersect() = %v, want c", got.Members())
	}
}

func TestGenericSetOperations(t *testing.T) {
	s := New("a", "b", "b")
	s.Add("c")
	s.Remove("a", "missing")

	if !s.Equal(New("b", "c")) {
		t.Fatalf("generic string set = %v, want b/c", s.Members())
	}
	if got := s.Sub(New("c")); !got.Equal(New("b")) {
		t.Fatalf("generic Sub() = %v, want b", got.Members())
	}
	if got := s.Union(New("d")); !got.Equal(New("b", "c", "d")) {
		t.Fatalf("generic Union() = %v, want b/c/d", got.Members())
	}
	if got := s.Intersect(New("c", "d")); !got.Equal(New("c")) {
		t.Fatalf("generic Intersect() = %v, want c", got.Members())
	}
}

func TestGenericSetWithStructValues(t *testing.T) {
	type key struct {
		ID   int
		Name string
	}

	a := key{ID: 1, Name: "a"}
	b := key{ID: 2, Name: "b"}
	c := key{ID: 3, Name: "c"}

	s := New(a, b)
	if !s.Contains(a) {
		t.Fatal("generic struct set should contain inserted key")
	}
	if got := s.Union(New(c)); !got.Equal(New(a, b, c)) {
		t.Fatalf("struct Union() = %v, want all keys", got.Members())
	}
}

func TestNumericSetOperations(t *testing.T) {
	if got := NewInt(1, 2, 3).Sub(NewInt(2)); !got.Equal(NewInt(1, 3)) {
		t.Fatalf("Int.Sub() = %v", got.Members())
	}
	if got := NewInt32(1, 2).Union(NewInt32(2, 3)); !got.Equal(NewInt32(1, 2, 3)) {
		t.Fatalf("Int32.Union() = %v", got.Members())
	}
	if got := NewInt64(1, 2, 3).Intersect(NewInt64(2, 3, 4)); !got.Equal(NewInt64(2, 3)) {
		t.Fatalf("Int64.Intersect() = %v", got.Members())
	}
	if got := NewUint(1, 2, 3).Sub(NewUint(1)); !got.Equal(NewUint(2, 3)) {
		t.Fatalf("Uint.Sub() = %v", got.Members())
	}
	if got := NewUint32(1, 2).Union(NewUint32(3)); !got.Equal(NewUint32(1, 2, 3)) {
		t.Fatalf("Uint32.Union() = %v", got.Members())
	}
	if got := NewUint64(1, 2, 3).Intersect(NewUint64(3, 4)); !got.Equal(NewUint64(3)) {
		t.Fatalf("Uint64.Intersect() = %v", got.Members())
	}
}

func TestSetJSONRoundTrip(t *testing.T) {
	original := NewInt(3, 1, 2)
	b, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}

	var decoded Int
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Equal(original) {
		t.Fatalf("decoded = %v, want %v", decoded.Members(), original.Members())
	}
}

func TestGenericSetJSONRoundTrip(t *testing.T) {
	original := New("go", "knifer")
	b, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}

	var decoded Set[string]
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Equal(original) {
		t.Fatalf("decoded = %v, want %v", decoded.Members(), original.Members())
	}
}

func TestSetJSONWithOptions(t *testing.T) {
	original := New("go")
	marshalCalled := false
	b, err := original.MarshalJSONWithOptions(WithSetMarshalFunc(func(v any) ([]byte, error) {
		marshalCalled = true
		return json.Marshal(v)
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !marshalCalled {
		t.Fatal("custom marshal provider was not used")
	}

	unmarshalCalled := false
	var decoded Set[string]
	if err := decoded.UnmarshalJSONWithOptions(b, WithSetUnmarshalFunc(func(data []byte, v any) error {
		unmarshalCalled = true
		return json.Unmarshal(data, v)
	})); err != nil {
		t.Fatal(err)
	}
	if !unmarshalCalled || !decoded.Equal(original) {
		t.Fatalf("unmarshalCalled=%v decoded=%v", unmarshalCalled, decoded.Members())
	}
}
