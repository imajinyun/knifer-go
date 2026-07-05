package vcodec

import "testing"

var benchmarkCodecSink string

func BenchmarkBase64Encode(b *testing.B) {
	data := []byte("knifer-go benchmark payload")
	for i := 0; i < b.N; i++ {
		benchmarkCodecSink = Base64Encode(data)
	}
}

func BenchmarkBase64Decode(b *testing.B) {
	encoded := Base64Encode([]byte("knifer-go benchmark payload"))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out, err := Base64Decode(encoded)
		if err != nil {
			b.Fatal(err)
		}
		benchmarkCodecSink = string(out)
	}
}

func BenchmarkHexEncode(b *testing.B) {
	data := []byte("knifer-go benchmark payload")
	for i := 0; i < b.N; i++ {
		benchmarkCodecSink = HexEncode(data)
	}
}
