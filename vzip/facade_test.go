package vzip_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/imajinyun/go-knifer/vzip"
)

func TestFacadeArchiveOptionConstructors(t *testing.T) {
	_ = vzip.WithOpen(nil)
	_ = vzip.WithStat(nil)
	_ = vzip.WithLstat(nil)
	_ = vzip.WithReadDir(nil)
	_ = vzip.WithReadlink(nil)
	_ = vzip.WithMkdirAll(nil)
	_ = vzip.WithRemove(nil)
	_ = vzip.WithRename(nil)
	_ = vzip.WithOpenZipReader(nil)
	_ = vzip.WithCreateTemp(nil)
}

func TestFacadeGzipReader(t *testing.T) {
	var buf bytes.Buffer
	data, err := vzip.GzipReader(io.NopCloser(&buf), 64)
	if err != nil {
		t.Fatalf("GzipReader on empty reader: %v", err)
	}
	// GZip output should be non-empty.
	if len(data) == 0 {
		t.Fatal("GzipReader returned empty data")
	}
}

func TestFacadeZlibFile(t *testing.T) {
	// ZlibFile on non-existent path should return an error.
	_, err := vzip.ZlibFile("/nonexistent/path", 6)
	if err == nil {
		t.Fatal("ZlibFile should fail on non-existent file")
	}
}

func TestFacadeGzipDecompress(t *testing.T) {
	// Compress then decompress a known string.
	original := "hello gzip test"
	compressed, err := vzip.GzipString(original)
	if err != nil {
		t.Fatalf("GzipString: %v", err)
	}
	decompressed, err := vzip.UnGzip(compressed)
	if err != nil {
		t.Fatalf("UnGzip: %v", err)
	}
	if string(decompressed) != original {
		t.Fatalf("roundtrip = %q, want %q", string(decompressed), original)
	}

	// UnGzipString
	text, err := vzip.UnGzipString(compressed)
	if err != nil || text != original {
		t.Fatalf("UnGzipString = %q, %v", text, err)
	}

	// Gunzip
	gunzipped, err := vzip.Gunzip(compressed)
	if err != nil || string(gunzipped) != original {
		t.Fatalf("Gunzip = %q, %v", string(gunzipped), err)
	}

	// Zlib roundtrip
	zlibbed, err := vzip.Zlib([]byte(original))
	if err != nil {
		t.Fatalf("Zlib: %v", err)
	}
	unzlibbed, err := vzip.UnZlib(zlibbed)
	if err != nil || string(unzlibbed) != original {
		t.Fatalf("UnZlib = %q, %v", string(unzlibbed), err)
	}

	// ZlibString
	zlibStr, err := vzip.ZlibString(original, 6)
	if err != nil || len(zlibStr) == 0 {
		t.Fatalf("ZlibString: %v", err)
	}
}

func TestFacadeUnGzipReader(t *testing.T) {
	compressed, err := vzip.GzipString("test data")
	if err != nil {
		t.Fatal(err)
	}
	data, err := vzip.UnGzipReader(bytes.NewReader(compressed), len(compressed))
	if err != nil || len(data) == 0 {
		t.Fatalf("UnGzipReader = %v, %v", data, err)
	}
}
