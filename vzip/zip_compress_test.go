package vzip_test

import (
	"bytes"
	"compress/flate"
	"os"
	"path/filepath"
	"testing"

	"github.com/imajinyun/knifer-go/vzip"
)

func TestFacadeZipAppendAndGzipOptions(t *testing.T) {
	tmp := t.TempDir()
	archive := filepath.Join(tmp, "append.zip")
	if err := vzip.ZipEntries(archive, vzip.EntryData{Name: "a.txt", Data: []byte("a")}); err != nil {
		t.Fatalf("ZipEntries: %v", err)
	}
	keepFile := filepath.Join(tmp, "keep.txt")
	if err := os.WriteFile(keepFile, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	filter := func(path string, info os.FileInfo) bool {
		return info.IsDir() || filepath.Ext(path) == ".txt"
	}
	if err := vzip.AppendWithOptions(archive, keepFile, vzip.WithFileFilter(filter), vzip.WithCompressionLevel(flate.BestSpeed)); err != nil {
		t.Fatalf("AppendWithOptions: %v", err)
	}
	data, err := vzip.GetBytes(archive, "keep.txt")
	if err != nil || string(data) != "keep" {
		t.Fatalf("appended keep.txt = %q, %v", data, err)
	}

	payload := []byte("hello gzip options")
	gz, err := vzip.GzipWithOptions(payload, vzip.WithCompressionLevel(flate.BestSpeed))
	if err != nil {
		t.Fatalf("GzipWithOptions: %v", err)
	}
	out, err := vzip.UnGzip(gz)
	if err != nil || !bytes.Equal(out, payload) {
		t.Fatalf("UnGzip = %q, %v", out, err)
	}
	gz, err = vzip.GzipReaderWithOptions(bytes.NewReader(payload), 0, vzip.WithCompressionLevel(flate.NoCompression))
	if err != nil {
		t.Fatalf("GzipReaderWithOptions: %v", err)
	}
	out, err = vzip.UnGzip(gz)
	if err != nil || !bytes.Equal(out, payload) {
		t.Fatalf("UnGzip reader = %q, %v", out, err)
	}
}

func TestFacadeGzipZlibAndReadFileHelpers(t *testing.T) {
	tmp := t.TempDir()
	source := filepath.Join(tmp, "payload.txt")
	payload := []byte("payload for compression helpers")
	if err := os.WriteFile(source, payload, 0o600); err != nil {
		t.Fatal(err)
	}

	gz, err := vzip.GzipFile(source)
	if err != nil {
		t.Fatalf("GzipFile: %v", err)
	}
	gunzip, err := vzip.UnGzipReader(bytes.NewReader(gz), len(payload))
	if err != nil || !bytes.Equal(gunzip, payload) {
		t.Fatalf("UnGzipReader = %q, %v", gunzip, err)
	}

	zlibBytes, err := vzip.ZlibString("hello zlib", flate.BestSpeed)
	if err != nil {
		t.Fatalf("ZlibString: %v", err)
	}
	zlibText, err := vzip.UnZlibString(zlibBytes)
	if err != nil || zlibText != "hello zlib" {
		t.Fatalf("UnZlibString = %q, %v", zlibText, err)
	}
	levelBytes, err := vzip.ZlibLevel(payload, flate.BestCompression)
	if err != nil {
		t.Fatalf("ZlibLevel: %v", err)
	}
	levelOut, err := vzip.UnZlibReader(bytes.NewReader(levelBytes), len(payload))
	if err != nil || !bytes.Equal(levelOut, payload) {
		t.Fatalf("UnZlibReader = %q, %v", levelOut, err)
	}
	readerBytes, err := vzip.ZlibReader(bytes.NewReader(payload), flate.NoCompression, len(payload))
	if err != nil {
		t.Fatalf("ZlibReader: %v", err)
	}
	if out, err := vzip.UnZlib(readerBytes); err != nil || !bytes.Equal(out, payload) {
		t.Fatalf("UnZlib reader bytes = %q, %v", out, err)
	}

	read, err := vzip.ReadFile(source)
	if err != nil || !bytes.Equal(read, payload) {
		t.Fatalf("ReadFile = %q, %v", read, err)
	}
	read, err = vzip.ReadFileWithOptions("/virtual/payload.txt", vzip.WithReadFile(func(path string) ([]byte, error) {
		if path != "/virtual/payload.txt" {
			return nil, os.ErrNotExist
		}
		return []byte("virtual"), nil
	}))
	if err != nil || string(read) != "virtual" {
		t.Fatalf("ReadFileWithOptions = %q, %v", read, err)
	}
}
