package vmap

import "testing"

var benchMapResult any

type benchMapSize struct {
	name string
	size int
}

func mapBenchSizes() []benchMapSize {
	return []benchMapSize{
		{name: "empty", size: 0},
		{name: "small", size: 16},
		{name: "medium", size: 1024},
		{name: "large", size: 4096},
	}
}

func makeBenchMap(n int) map[int]int {
	m := make(map[int]int, n)
	for i := 0; i < n; i++ {
		m[i] = i
	}
	return m
}

func BenchmarkFilter(b *testing.B) {
	for _, tt := range mapBenchSizes() {
		m := makeBenchMap(tt.size)
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchMapResult = Filter(m, func(_ int, v int) bool { return v%2 == 0 })
			}
		})
	}
}

func BenchmarkSortedKeys(b *testing.B) {
	for _, tt := range mapBenchSizes() {
		m := makeBenchMap(tt.size)
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchMapResult = SortedKeys(m)
			}
		})
	}
}

func BenchmarkMerge(b *testing.B) {
	for _, tt := range mapBenchSizes() {
		left := makeBenchMap(tt.size)
		right := makeBenchMap(tt.size)
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchMapResult = Merge(left, right)
			}
		})
	}
}
