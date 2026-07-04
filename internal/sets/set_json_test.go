package sets

import (
	"encoding/json"
	"testing"
	"unicode/utf8"
)

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

func TestNilJSONProviderOptionsDoNotOverwriteConfiguredProviders(t *testing.T) {
	marshal := func(any) ([]byte, error) { return []byte("[]"), nil }
	unmarshal := func([]byte, any) error { return nil }
	cfg := applyJSONOptions([]JSONOption{
		WithSetMarshalFunc(marshal),
		WithSetMarshalFunc(nil),
		WithSetUnmarshalFunc(unmarshal),
		WithSetUnmarshalFunc(nil),
	})
	if cfg.marshal == nil || cfg.unmarshal == nil {
		t.Fatalf("nil json provider option overwrote configured provider: %#v", cfg)
	}
}

func FuzzSetJSONRoundTrip(f *testing.F) {
	for _, seed := range []string{"", "go", "go,knifer", "重复,重复,value"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed string) {
		if !utf8.ValidString(seed) {
			t.Skip()
		}
		members := []string{}
		if seed != "" {
			members = append(members, seed)
		}
		members = append(members, seed+"-suffix")
		original := New(members...)
		b, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}
		var decoded Set[string]
		if err := json.Unmarshal(b, &decoded); err != nil {
			t.Fatalf("Unmarshal(%q) error = %v", b, err)
		}
		if !decoded.Equal(original) {
			t.Fatalf("decoded = %v, want %v", decoded.Members(), original.Members())
		}
	})
}
