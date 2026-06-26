package json

import (
	"errors"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestPathGetPut(t *testing.T) {
	src := `{"a":{"b":[10,20,{"c":"hit"}]}}`
	v, _ := Parse(src)
	if got := GetByPath(v, "a.b[2].c"); got != "hit" {
		t.Fatalf("path get: %v", got)
	}
	if got := GetByPath(v, "$.a.b[0]"); got != int64(10) {
		t.Fatalf("path get with $: %v", got)
	}
	obj := v.(*JSONObject)
	if err := obj.PutByPath("a.b[1]", "X"); err != nil {
		t.Fatalf("put: %v", err)
	}
	if got := obj.GetByPath("a.b[1]"); got != "X" {
		t.Fatalf("after put: %v", got)
	}
}

func TestPathCreatesIntermediate(t *testing.T) {
	obj := NewJSONObject()
	if err := obj.PutByPath("a.b.c", 42); err != nil {
		t.Fatalf("put: %v", err)
	}
	if obj.GetByPath("a.b.c") != int64(42) {
		t.Fatalf("nested put")
	}
}

func TestPutByPathRejectsNegativeIndexWithoutPanic(t *testing.T) {
	cases := []struct {
		name string
		root any
		path string
	}{
		{name: "root array", root: NewJSONArray().Add("safe"), path: "[-1]"},
		{name: "nested array", root: NewJSONObject().Set("items", NewJSONArray().Add("safe")), path: "items[-1]"},
		{name: "intermediate array", root: NewJSONArray(), path: "[-1].name"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			panicked := false
			var err error
			func() {
				defer func() {
					if recover() != nil {
						panicked = true
					}
				}()
				err = PutByPath(tt.root, tt.path, "bad")
			}()
			if panicked {
				t.Fatalf("PutByPath(%q) panicked", tt.path)
			}
			if err == nil || !errors.Is(err, knifer.ErrCodeInvalidInput) {
				t.Fatalf("PutByPath(%q) err = %v, want invalid input", tt.path, err)
			}
		})
	}
}

func TestJSONArraySetNegativeIndexLeavesArrayUnchanged(t *testing.T) {
	arr := NewJSONArray().Add("safe")
	arr.Set(-1, "bad")
	if arr.Len() != 1 || arr.GetString(0) != "safe" {
		t.Fatalf("Set(-1) should leave array unchanged: %s", arr.String())
	}
}
