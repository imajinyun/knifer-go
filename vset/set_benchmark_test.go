package vset

import (
	"encoding/json"
	"testing"
)

var (
	benchBoolSink  bool
	benchIntSink   int
	benchSetSink   Set[int]
	benchSliceSink []int
	benchBytesSink []byte
)

func benchmarkSet() Set[int] {
	items := make([]int, 1024)
	for i := range items {
		items[i] = i
	}
	return New(items...)
}

func BenchmarkContains(b *testing.B) {
	set := benchmarkSet()
	for b.Loop() {
		benchBoolSink = set.Contains(777)
	}
}

func BenchmarkUnion(b *testing.B) {
	left := benchmarkSet()
	right := New[int]()
	for i := 512; i < 1536; i++ {
		right.Add(i)
	}
	for b.Loop() {
		benchSetSink = left.Union(right)
	}
}

func BenchmarkIntersect(b *testing.B) {
	left := benchmarkSet()
	right := New[int]()
	for i := 512; i < 1536; i++ {
		right.Add(i)
	}
	for b.Loop() {
		benchSetSink = left.Intersect(right)
	}
}

func BenchmarkSub(b *testing.B) {
	left := benchmarkSet()
	right := New[int]()
	for i := 0; i < 512; i++ {
		right.Add(i)
	}
	for b.Loop() {
		benchSetSink = left.Sub(right)
	}
}

func BenchmarkMembers(b *testing.B) {
	set := benchmarkSet()
	for b.Loop() {
		benchSliceSink = set.Members()
	}
}

func BenchmarkMarshalJSON(b *testing.B) {
	set := benchmarkSet()
	for b.Loop() {
		data, err := json.Marshal(set)
		if err != nil {
			b.Fatal(err)
		}
		benchBytesSink = data
	}
}

func BenchmarkUnmarshalJSON(b *testing.B) {
	data, err := json.Marshal(benchmarkSet())
	if err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		var set Set[int]
		if err := json.Unmarshal(data, &set); err != nil {
			b.Fatal(err)
		}
		benchIntSink = len(set)
	}
}
