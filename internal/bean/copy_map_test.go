package bean

import (
	"errors"
	"strings"
	"testing"

	knifer "github.com/imajinyun/go-knifer"
)

func TestCopyPropertiesMapToStruct(t *testing.T) {
	src := map[string]any{
		"displayName": "bob",
		"age":         7.9,
		"admin":       1,
		"trace_id":    "t-2",
	}
	var dst targetProfile
	if err := Copy(src, &dst); err != nil {
		t.Fatalf("Copy() error = %v", err)
	}
	if dst.Name != "bob" || dst.Age != 7 || !dst.Admin || dst.Trace != "t-2" {
		t.Fatalf("dst = %+v", dst)
	}
}

func TestCopyPropertiesMapToStructNestedCollections(t *testing.T) {
	type nested struct {
		Value int
	}
	type target struct {
		Bytes  []byte
		Ages   []int
		Flags  []bool
		Labels map[int]uint
		Nested map[string]nested
	}

	var dst target
	err := CopyProperties(map[string]any{
		"bytes":  "abc",
		"ages":   []any{"1", 2.9, true},
		"flags":  [3]int{0, 1, 2},
		"labels": map[string]any{"1": "2", "3": 4.8},
		"nested": map[string]any{"first": map[string]any{"value": "7"}},
	}, &dst)
	if err != nil {
		t.Fatalf("CopyProperties() error = %v", err)
	}
	if string(dst.Bytes) != "abc" {
		t.Fatalf("Bytes = %q", dst.Bytes)
	}
	if len(dst.Ages) != 3 || dst.Ages[0] != 1 || dst.Ages[1] != 2 || dst.Ages[2] != 1 {
		t.Fatalf("Ages = %#v", dst.Ages)
	}
	if len(dst.Flags) != 3 || dst.Flags[0] || !dst.Flags[1] || !dst.Flags[2] {
		t.Fatalf("Flags = %#v", dst.Flags)
	}
	if dst.Labels[1] != 2 || dst.Labels[3] != 4 {
		t.Fatalf("Labels = %#v", dst.Labels)
	}
	if dst.Nested["first"].Value != 7 {
		t.Fatalf("Nested = %#v", dst.Nested)
	}
}

func TestCopyPropertiesMapConversionErrors(t *testing.T) {
	tests := []struct {
		name string
		src  map[string]any
		want string
	}{
		{
			name: "slice element",
			src:  map[string]any{"ages": []any{"bad"}},
			want: "index 0",
		},
		{
			name: "slice source",
			src:  map[string]any{"ages": 12},
			want: "cannot assign int to []int",
		},
		{
			name: "map key",
			src:  map[string]any{"labels": map[string]any{"bad": 1}},
			want: "map key",
		},
		{
			name: "map value",
			src:  map[string]any{"labels": map[string]any{"1": "bad"}},
			want: "map value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type target struct {
				Ages   []int
				Labels map[int]uint
			}
			var dst target
			err := CopyProperties(tt.src, &dst)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("CopyProperties() error = %v, want substring %q", err, tt.want)
			}
			assertBeanInvalidInput(t, err)
		})
	}
}

func TestDecodeResultTracksMatchedSkippedUnused(t *testing.T) {
	type user struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
		Note string `json:"note"`
	}

	var dst user
	result, err := DecodeResult(map[string]any{
		"name":    "alice",
		"age":     "21",
		"note":    "",
		"unknown": true,
	}, &dst, WithIgnoreEmpty(true))
	if err != nil {
		t.Fatalf("DecodeResult() error = %v", err)
	}
	if dst != (user{Name: "alice", Age: 21}) {
		t.Fatalf("DecodeResult() dst = %+v", dst)
	}
	assertEqualStrings(t, []string{"age", "name"}, result.Matched)
	assertEqualStrings(t, []string{"note"}, result.Skipped)
	assertEqualStrings(t, []string{"unknown"}, result.Unused)
}

func TestDecodeStrictUnused(t *testing.T) {
	type user struct {
		Name string `json:"name"`
	}

	var dst user
	err := Decode(map[string]any{"name": "alice", "extra": true}, &dst, WithStrictUnused(true))

	if err == nil {
		t.Fatal("Decode() strict unused error = nil")
	}
	if !strings.Contains(err.Error(), "unused source properties: extra") {
		t.Fatalf("Decode() error = %v", err)
	}
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("Decode() code = %v, want invalid input", err)
	}
}
