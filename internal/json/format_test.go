package json

import (
	"strings"
	"testing"
)

func TestPretty(t *testing.T) {
	obj := NewJSONObject().Set("a", 1).Set("b", NewJSONArray().Add(1).Add(2))
	out := obj.ToStringPretty()
	expect := "{\n    \"a\": 1,\n    \"b\": [\n        1,\n        2\n    ]\n}"
	if out != expect {
		t.Fatalf("pretty mismatch:\n%s\n--\n%s", out, expect)
	}
}

func TestWithFormatIndent(t *testing.T) {
	in := `{"a":1}`
	out := FormatJSONStrWithOptions(in, WithFormatIndent("  "))
	if !strings.Contains(out, "\n  \"a\": 1") {
		t.Fatalf("WithFormatIndent = %q", out)
	}
}

func TestFormatJSONStr(t *testing.T) {
	in := `{"a":1,"b":[1,2],"c":"x"}`
	out := FormatJSONStr(in)
	if !strings.Contains(out, "\n") {
		t.Fatalf("expect formatted: %q", out)
	}
	custom := FormatJSONStrWithOptions(in, WithFormatIndentWidth(2), WithFormatSpaceAfterKey(false))
	if !strings.Contains(custom, "\n  \"a\":1") {
		t.Fatalf("custom format = %q", custom)
	}
}

func TestIsJSONWithOptions(t *testing.T) {
	called := false
	valid := func(data []byte) bool {
		called = true
		return string(data) == "custom"
	}
	if !IsJSONWithOptions("custom", WithJSONValidFunc(valid)) || !called {
		t.Fatalf("IsJSONWithOptions called=%v", called)
	}
	if !IsJSONObjWithOptions("{custom}", WithJSONValidFunc(func([]byte) bool { return true })) {
		t.Fatal("IsJSONObjWithOptions should use custom validator")
	}
	if !IsJSONArrayWithOptions("[custom]", WithJSONValidFunc(func([]byte) bool { return true })) {
		t.Fatal("IsJSONArrayWithOptions should use custom validator")
	}
}

func TestIsJSONHelpers(t *testing.T) {
	if !IsJSON(`{"a":1}`) || !IsJSONObj(`{"a":1}`) {
		t.Fatalf("obj")
	}
	if !IsJSONArray(`[1,2]`) {
		t.Fatalf("array")
	}
	if IsJSON("not json") {
		t.Fatalf("invalid")
	}
}
