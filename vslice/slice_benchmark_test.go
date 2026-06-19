package vslice

import "testing"

var benchSliceResult any

type benchSliceSize struct {
	name string
	size int
}

func sliceBenchSizes() []benchSliceSize {
	return []benchSliceSize{
		{name: "empty", size: 0},
		{name: "small", size: 16},
		{name: "medium", size: 1024},
		{name: "large", size: 4096},
	}
}

func makeBenchSlice(n int) []int {
	values := make([]int, n)
	for i := range values {
		values[i] = i
	}
	return values
}

func BenchmarkFilter(b *testing.B) {
	for _, tt := range sliceBenchSizes() {
		values := makeBenchSlice(tt.size)
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchSliceResult = Filter(values, func(v int) bool { return v%2 == 0 })
			}
		})
	}
}

func BenchmarkMap(b *testing.B) {
	for _, tt := range sliceBenchSizes() {
		values := makeBenchSlice(tt.size)
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchSliceResult = Map(values, func(v int) int { return v * 2 })
			}
		})
	}
}

func BenchmarkMapErr(b *testing.B) {
	for _, tt := range sliceBenchSizes() {
		values := makeBenchSlice(tt.size)
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				result, err := MapErr(values, func(v int) (int, error) { return v * 2, nil })
				if err != nil {
					b.Fatal(err)
				}
				benchSliceResult = result
			}
		})
	}
}

func BenchmarkWindow(b *testing.B) {
	for _, tt := range sliceBenchSizes() {
		values := makeBenchSlice(tt.size)
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchSliceResult = Window(values, 8)
			}
		})
	}
}

func BenchmarkZip2(b *testing.B) {
	for _, tt := range sliceBenchSizes() {
		left := makeBenchSlice(tt.size)
		right := makeBenchSlice(tt.size)
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchSliceResult = Zip2(left, right)
			}
		})
	}
}

func BenchmarkUniq(b *testing.B) {
	for _, tt := range sliceBenchSizes() {
		values := makeBenchSlice(tt.size)
		if len(values) > 1 {
			for i := 1; i < len(values); i += 2 {
				values[i] = values[i-1]
			}
		}
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchSliceResult = Uniq(values)
			}
		})
	}
}
