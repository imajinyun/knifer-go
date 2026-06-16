package vstr

import "testing"

var benchStringResult any

type benchStringSize struct {
	name string
	size int
}

func stringBenchSizes() []benchStringSize {
	return []benchStringSize{
		{name: "empty", size: 0},
		{name: "small", size: 16},
		{name: "medium", size: 1024},
		{name: "large", size: 4096},
	}
}

func makeBenchString(n int) string {
	buf := make([]byte, n)
	for i := range buf {
		buf[i] = byte('a' + i%26)
	}
	return string(buf)
}

func BenchmarkReverse(b *testing.B) {
	for _, tt := range stringBenchSizes() {
		value := makeBenchString(tt.size)
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchStringResult = Reverse(value)
			}
		})
	}
}

func BenchmarkToCamelCase(b *testing.B) {
	for _, tt := range stringBenchSizes() {
		value := makeBenchString(tt.size)
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchStringResult = ToCamelCase(value)
			}
		})
	}
}

func BenchmarkContains(b *testing.B) {
	for _, tt := range stringBenchSizes() {
		value := makeBenchString(tt.size)
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchStringResult = Contains(value, "z")
			}
		})
	}
}
