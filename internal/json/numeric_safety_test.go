package json

import (
	"math"
	"testing"
)

func TestUnsignedIntegerWrapPreservesLargeValues(t *testing.T) {
	large := uint64(math.MaxInt64) + 1
	obj := NewJSONObject().Set("large", large).Set("small", uint64(9))

	raw, ok := obj.Get("large")
	if !ok {
		t.Fatal("large key missing")
	}
	if got, ok := raw.(uint64); !ok || got != large {
		t.Fatalf("large raw = %#v (%T), want uint64(%d)", raw, raw, large)
	}
	if got := obj.GetString("large"); got != "9223372036854775808" {
		t.Fatalf("large string = %q", got)
	}
	if got := obj.GetInt64Or("large", -7); got != -7 {
		t.Fatalf("large int64 fallback = %d, want -7", got)
	}
	if got := obj.GetInt64("small"); got != 9 {
		t.Fatalf("small uint64 should remain int64-compatible, got %d", got)
	}
	if got := obj.String(); got != `{"large":9223372036854775808,"small":9}` {
		t.Fatalf("serialized large uint64 = %s", got)
	}
}

func TestParsedLargeUnsignedIntegerStaysExact(t *testing.T) {
	obj, err := ParseObj(`{"n":9223372036854775808}`)
	if err != nil {
		t.Fatalf("ParseObj: %v", err)
	}
	raw, ok := obj.Get("n")
	if !ok {
		t.Fatal("n key missing")
	}
	if got, ok := raw.(uint64); !ok || got != uint64(math.MaxInt64)+1 {
		t.Fatalf("n raw = %#v (%T), want exact uint64", raw, raw)
	}
	if got := obj.GetInt64Or("n", -1); got != -1 {
		t.Fatalf("n int64 fallback = %d, want -1", got)
	}
	if got := obj.String(); got != `{"n":9223372036854775808}` {
		t.Fatalf("serialized parsed large uint64 = %s", got)
	}
}

func TestInt64GetterRejectsUnsafeFloatValues(t *testing.T) {
	obj := NewJSONObject()
	obj.Set("nan", math.NaN())
	obj.Set("inf", math.Inf(1))
	obj.Set("tooLarge", float64(uint64(math.MaxInt64)+1))
	obj.Set("valid", 12.75)

	tests := []struct {
		name string
		key  string
		want int64
	}{
		{name: "nan", key: "nan", want: -5},
		{name: "inf", key: "inf", want: -5},
		{name: "tooLarge", key: "tooLarge", want: -5},
		{name: "valid truncates", key: "valid", want: 12},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := obj.GetInt64Or(tt.key, -5); got != tt.want {
				t.Fatalf("GetInt64Or(%q) = %d, want %d", tt.key, got, tt.want)
			}
		})
	}
}
