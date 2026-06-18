package json

import (
	"testing"
)

func TestWithConfigOption(t *testing.T) {
	cfg := NewConfig()
	opt := WithConfig(cfg)
	if opt == nil {
		t.Fatal("WithConfig() = nil")
	}
}

func TestWithUnmarshalFuncOption(t *testing.T) {
	opt := WithUnmarshalFunc(func([]byte, any) error { return nil })
	if opt == nil {
		t.Fatal("WithUnmarshalFunc() = nil")
	}
}

func TestWithParseConfigOption(t *testing.T) {
	cfg := NewConfig()
	opt := WithParseConfig(cfg)
	if opt == nil {
		t.Fatal("WithParseConfig() = nil")
	}
}

func TestWithParseOptions(t *testing.T) {
	if WithParseIntFunc(func(string, int, int) (int64, error) { return 0, nil }) == nil {
		t.Fatal("WithParseIntFunc() = nil")
	}
	if WithParseFloatFunc(func(string, int) (float64, error) { return 0, nil }) == nil {
		t.Fatal("WithParseFloatFunc() = nil")
	}
	if WithParseBoolFunc(func(string) (bool, error) { return false, nil }) == nil {
		t.Fatal("WithParseBoolFunc() = nil")
	}
}

func TestToJSONStrVariants(t *testing.T) {
	v := map[string]any{"a": 1}
	out, err := ToJSONStrIndent(v, 2)
	if err != nil {
		t.Fatalf("ToJSONStrIndent error = %v", err)
	}
	if len(out) == 0 {
		t.Fatal("ToJSONStrIndent empty")
	}

	cfg := NewConfig()
	out2, err := ToJSONStrWithConfig(v, cfg)
	if err != nil {
		t.Fatalf("ToJSONStrWithConfig error = %v", err)
	}
	if len(out2) == 0 {
		t.Fatal("ToJSONStrWithConfig empty")
	}

	out3, err := ToJSONPrettyStrWithConfig(v, cfg)
	if err != nil {
		t.Fatalf("ToJSONPrettyStrWithConfig error = %v", err)
	}
	if len(out3) == 0 {
		t.Fatal("ToJSONPrettyStrWithConfig empty")
	}
}

func TestPutByPath(t *testing.T) {
	obj := NewJSONObject()
	if err := PutByPath(obj, "a.b", 42); err != nil {
		t.Fatalf("PutByPath error = %v", err)
	}
	if got := GetByPath(obj, "a.b"); got != int64(42) {
		t.Fatalf("after PutByPath, GetByPath = %v", got)
	}
}
