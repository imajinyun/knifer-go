package vzip_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
	"github.com/imajinyun/knifer-go/vzip"
)

func TestFacadeZipOptions(t *testing.T) {
	tmp := t.TempDir()
	archive := filepath.Join(tmp, "options.zip")
	if err := vzip.ZipEntriesWithOptions(
		archive,
		[]vzip.EntryData{{Name: "hello.txt", Data: []byte("hello")}},
		vzip.WithFilePerm(0o600),
		vzip.WithCompressionLevel(1),
	); err != nil {
		t.Fatalf("ZipEntriesWithOptions: %v", err)
	}
	if err := vzip.ZipEntriesWithOptions(
		archive,
		[]vzip.EntryData{{Name: "hello.txt", Data: []byte("hello")}},
		vzip.WithOverwrite(false),
	); err == nil {
		t.Fatal("ZipEntriesWithOptions should reject overwrite=false for existing archive")
	}

	data, err := vzip.GetBytesWithOptions(archive, "hello.txt", vzip.WithMaxBytes(5))
	if err != nil || string(data) != "hello" {
		t.Fatalf("GetBytesWithOptions = %q, %v", data, err)
	}
	if _, err := vzip.GetBytesWithOptions(archive, "hello.txt", vzip.WithMaxBytes(4)); err == nil {
		t.Fatal("GetBytesWithOptions should reject content larger than max bytes")
	}

	dest := filepath.Join(tmp, "dest")
	if err := vzip.UnzipToWithOptions(archive, dest, vzip.WithDirPerm(0o700), vzip.WithFilePerm(0o600)); err != nil {
		t.Fatalf("UnzipToWithOptions: %v", err)
	}
	if got, err := os.ReadFile(filepath.Join(dest, "hello.txt")); err != nil || string(got) != "hello" {
		t.Fatalf("unzipped via options: %q %v", got, err)
	}

	gz, err := vzip.GzipString("hello")
	if err != nil {
		t.Fatalf("GzipString: %v", err)
	}
	if _, err := vzip.UnGzipWithOptions(gz, vzip.WithMaxBytes(4)); err == nil {
		t.Fatal("UnGzipWithOptions should reject content larger than max bytes")
	}
	zlibBytes, err := vzip.Zlib([]byte("hello"))
	if err != nil {
		t.Fatalf("Zlib: %v", err)
	}
	if _, err := vzip.UnZlibWithOptions(zlibBytes, vzip.WithMaxBytes(4)); err == nil {
		t.Fatal("UnZlibWithOptions should reject content larger than max bytes")
	}
}

func TestFacadeZipErrorContract(t *testing.T) {
	_, err := vzip.GetStream(nil)
	if err == nil {
		t.Fatal("GetStream(nil) error = nil, want invalid input")
	}
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("errors.Is(err, ErrCodeInvalidInput) = false: %v", err)
	}
	code, ok := knifer.CodeOf(err)
	if !ok || code != knifer.ErrCodeInvalidInput {
		t.Fatalf("CodeOf(err) = %q, %v; want invalid input", code, ok)
	}
	var zipErr *vzip.Error
	if !errors.As(err, &zipErr) {
		t.Fatalf("errors.As(err, *vzip.Error) = false: %v", err)
	}
}
