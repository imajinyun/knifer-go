package vzip_test

import (
	"bytes"
	"testing"

	"github.com/imajinyun/knifer-go/vzip"
)

func TestFacadeGzipAndZlibRoundTrip(t *testing.T) {
	gz, err := vzip.GzipString("hello")
	if err != nil {
		t.Fatalf("GzipString: %v", err)
	}
	text, err := vzip.UnGzipString(gz)
	if err != nil || text != "hello" {
		t.Fatalf("UnGzipString: %q %v", text, err)
	}
	dataBytes := []byte("hello the utility toolkit zip facade")
	gzipBytes, err := vzip.Gzip(dataBytes)
	if err != nil {
		t.Fatalf("Gzip: %v", err)
	}
	out, err := vzip.Gunzip(gzipBytes)
	if err != nil || !bytes.Equal(out, dataBytes) {
		t.Fatalf("Gunzip: %q %v", out, err)
	}
	zlibBytes, err := vzip.Zlib(dataBytes)
	if err != nil {
		t.Fatalf("Zlib: %v", err)
	}
	out, err = vzip.Unzlib(zlibBytes)
	if err != nil || !bytes.Equal(out, dataBytes) {
		t.Fatalf("Unzlib: %q %v", out, err)
	}
}
