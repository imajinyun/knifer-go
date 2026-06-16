package vnum

import "testing"

var benchNumResult any

type benchNumSize struct {
	name string
	size int
}

func numBenchSizes() []benchNumSize {
	return []benchNumSize{
		{name: "empty", size: 0},
		{name: "small", size: 16},
		{name: "medium", size: 1024},
		{name: "large", size: 4096},
	}
}

func makeBenchNums(n int) []float64 {
	values := make([]float64, n)
	for i := range values {
		values[i] = float64(i) + 0.5
	}
	return values
}

func BenchmarkAdd(b *testing.B) {
	for _, tt := range numBenchSizes() {
		values := makeBenchNums(tt.size)
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchNumResult = Add(values...)
			}
		})
	}
}

func BenchmarkMax(b *testing.B) {
	for _, tt := range numBenchSizes() {
		values := make([]int, tt.size)
		for i := range values {
			values[i] = i
		}
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchNumResult = Max(values...)
			}
		})
	}
}

func BenchmarkCalculate(b *testing.B) {
	for b.Loop() {
		benchNumResult, _ = Calculate("1+2*3-4/2")
	}
}
