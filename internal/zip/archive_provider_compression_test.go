package zip

import (
	"bytes"
	"compress/flate"
	"io"
	"os"
	"testing"
)

func TestArchiveProviderOptionsForFileCompression(t *testing.T) {
	data := []byte("provider-data")
	openPath := ""
	statPath := ""
	open := func(path string) (io.ReadCloser, error) {
		openPath = path
		return io.NopCloser(bytes.NewReader(data)), nil
	}
	stat := func(path string) (os.FileInfo, error) {
		statPath = path
		return zipFakeFileInfo{name: path, size: int64(len(data))}, nil
	}
	gz, err := GzipFileWithOptions("virtual.txt", WithOpen(open), WithStat(stat), WithCompressionLevel(flate.BestSpeed))
	if err != nil {
		t.Fatalf("GzipFileWithOptions() error = %v", err)
	}
	if openPath != "virtual.txt" || statPath != "virtual.txt" {
		t.Fatalf("provider paths open=%q stat=%q", openPath, statPath)
	}
	out, err := UnGzip(gz)
	if err != nil || !bytes.Equal(out, data) {
		t.Fatalf("UnGzip(provider gzip) = %q, %v", out, err)
	}
	z, err := ZlibFileWithOptions("virtual.txt", flate.BestSpeed, WithOpen(open), WithStat(stat))
	if err != nil {
		t.Fatalf("ZlibFileWithOptions() error = %v", err)
	}
	out, err = UnZlib(z)
	if err != nil || !bytes.Equal(out, data) {
		t.Fatalf("UnZlib(provider zlib) = %q, %v", out, err)
	}
}
