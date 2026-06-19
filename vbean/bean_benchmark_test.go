package vbean

import "testing"

var benchBeanFacadeResult any

type benchUser struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func BenchmarkDecodeResult(b *testing.B) {
	src := map[string]any{"name": "alice", "age": "21"}
	b.ReportAllocs()
	for b.Loop() {
		var dst benchUser
		result, err := DecodeResult(src, &dst)
		if err != nil {
			b.Fatal(err)
		}
		benchBeanFacadeResult = result
	}
}

func BenchmarkMerge(b *testing.B) {
	left := map[string]any{"name": "alice"}
	right := map[string]any{"age": "21"}
	b.ReportAllocs()
	for b.Loop() {
		dst := benchUser{Name: "existing", Age: 1}
		if err := Merge(&dst, left, right); err != nil {
			b.Fatal(err)
		}
		benchBeanFacadeResult = dst
	}
}
