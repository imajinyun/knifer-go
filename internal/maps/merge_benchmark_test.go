package maps

import "testing"

func BenchmarkMerge_TwoMaps(b *testing.B) {
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

func BenchmarkMerge_FiveMaps(b *testing.B) {
	for _, tt := range mapBenchSizes() {
		ms := []map[int]int{
			makeBenchMap(tt.size),
			makeBenchMap(tt.size),
			makeBenchMap(tt.size),
			makeBenchMap(tt.size),
			makeBenchMap(tt.size),
		}
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchMapResult = Merge(ms...)
			}
		})
	}
}

func BenchmarkMergeFunc_Sum(b *testing.B) {
	add := func(o, n int) int { return o + n }
	for _, tt := range mapBenchSizes() {
		left := makeBenchMap(tt.size)
		right := makeBenchMap(tt.size)
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchMapResult = MergeFunc(add, left, right)
			}
		})
	}
}
