package vbean_test

import (
	"testing"

	"github.com/imajinyun/go-knifer/vbean"
)

func TestFacadeCopyProperties(t *testing.T) {
	var dst userModel
	if err := vbean.CopyProperties(userDTO{Name: "alice", Age: "18"}, &dst); err != nil {
		t.Fatalf("CopyProperties() error = %v", err)
	}
	if dst.Name != "alice" || dst.Age != 18 {
		t.Fatalf("dst = %+v", dst)
	}
}

func TestFacadeToMap(t *testing.T) {
	got, err := vbean.ToMap(userDTO{Name: "bob", Age: "20"})
	if err != nil {
		t.Fatalf("ToMap() error = %v", err)
	}
	if got["name"] != "bob" || got["age"] != "20" {
		t.Fatalf("map = %#v", got)
	}
}

func TestFacadeDecodeResult(t *testing.T) {
	type user struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	var dst user
	result, err := vbean.DecodeResult(map[string]any{"name": "alice", "age": "21", "extra": true}, &dst)
	if err != nil {
		t.Fatalf("DecodeResult() error = %v", err)
	}
	if dst != (user{Name: "alice", Age: 21}) {
		t.Fatalf("DecodeResult() dst = %+v", dst)
	}
	assertEqualStrings(t, []string{"age", "name"}, result.Matched)
	assertEqualStrings(t, []string{"extra"}, result.Unused)
}

func TestFacadeDecode(t *testing.T) {
	type user struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	var dst user
	err := vbean.Decode(map[string]any{"name": "alice", "age": "21"}, &dst)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if dst != (user{Name: "alice", Age: 21}) {
		t.Fatalf("Decode() dst = %+v", dst)
	}
}

func TestFacadeMerge(t *testing.T) {
	type user struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	dst := user{Name: "existing", Age: 1}
	err := vbean.Merge(&dst, map[string]any{"name": "alice"}, map[string]any{"age": "22"})
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}
	if dst != (user{Name: "alice", Age: 22}) {
		t.Fatalf("Merge() dst = %+v", dst)
	}
}

func TestFacadeMergeResult(t *testing.T) {
	type user struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	dst := user{Name: "existing", Age: 1}
	result, err := vbean.MergeResult(&dst, map[string]any{"name": "alice"}, map[string]any{"age": "22"})
	if err != nil {
		t.Fatalf("MergeResult() error = %v", err)
	}
	if dst != (user{Name: "alice", Age: 22}) {
		t.Fatalf("MergeResult() dst = %+v", dst)
	}
	assertEqualStrings(t, []string{"age", "name"}, result.Matched)
}

func TestFacadeMergeWithOptions(t *testing.T) {
	type user struct {
		Name string `json:"name"`
		Note string `json:"note"`
	}

	dst := user{Name: "existing", Note: "keep"}
	err := vbean.MergeWithOptions(&dst, []any{
		map[string]any{"name": "alice"},
		map[string]any{"note": ""},
	}, vbean.WithIgnoreEmpty(true))
	if err != nil {
		t.Fatalf("MergeWithOptions() error = %v", err)
	}
	if dst != (user{Name: "alice", Note: "keep"}) {
		t.Fatalf("MergeWithOptions() dst = %+v", dst)
	}
}

func TestFacadeMergeResultWithOptions(t *testing.T) {
	type user struct {
		Name string `json:"name"`
		Note string `json:"note"`
	}

	dst := user{Name: "existing", Note: "keep"}
	result, err := vbean.MergeResultWithOptions(&dst, []any{
		map[string]any{"name": "alice"},
		map[string]any{"note": ""},
	}, vbean.WithIgnoreEmpty(true))
	if err != nil {
		t.Fatalf("MergeResultWithOptions() error = %v", err)
	}
	if dst != (user{Name: "alice", Note: "keep"}) {
		t.Fatalf("MergeResultWithOptions() dst = %+v", dst)
	}
	assertEqualStrings(t, []string{"name"}, result.Matched)
	assertEqualStrings(t, []string{"note"}, result.Skipped)
}
