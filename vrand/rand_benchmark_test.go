package vrand_test

import (
	"testing"

	"github.com/imajinyun/knifer-go/vrand"
)

var (
	benchmarkRandBytes []byte
	benchmarkRandErr   error
)

func BenchmarkSecureBytes(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		benchmarkRandBytes, benchmarkRandErr = vrand.SecureBytes(32)
		if benchmarkRandErr != nil {
			b.Fatalf("SecureBytes: %v", benchmarkRandErr)
		}
	}
}
