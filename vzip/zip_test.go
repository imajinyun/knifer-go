package vzip_test

import (
	archivezip "archive/zip"
	"bytes"
	"compress/flate"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	knifer "github.com/imajinyun/go-knifer"
	"github.com/imajinyun/go-knifer/vzip"
)

func TestFacadeZipAndCompression(t *testing.T) {
	tmp := t.TempDir()
	archive := filepath.Join(tmp, "data.zip")
	if err := vzip.ZipEntries(archive, vzip.EntryData{Name: "hello.txt", Data: []byte("hello")}); err != nil {
		t.Fatalf("ZipEntries: %v", err)
	}
	data, err := vzip.GetBytes(archive, "hello.txt")
	if err != nil || string(data) != "hello" {
		t.Fatalf("GetBytes: %q %v", data, err)
	}
	dest := filepath.Join(tmp, "dest")
	if err := vzip.UnzipTo(archive, dest); err != nil {
		t.Fatalf("UnzipTo: %v", err)
	}
	if got, err := os.ReadFile(filepath.Join(dest, "hello.txt")); err != nil || string(got) != "hello" {
		t.Fatalf("unzipped: %q %v", got, err)
	}
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

func TestFacadeZipCreationUsingOptions(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "keep.txt"), []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "skip.log"), []byte("skip"), 0o644); err != nil {
		t.Fatal(err)
	}

	archive := filepath.Join(tmp, "filtered.zip")
	filter := func(path string, info os.FileInfo) bool {
		return info.IsDir() || filepath.Ext(path) == ".txt"
	}
	if err := vzip.ZipFilesUsingOptions(archive, []string{src}, vzip.WithSourceDir(true), vzip.WithFileFilter(filter)); err != nil {
		t.Fatalf("ZipFilesUsingOptions: %v", err)
	}
	data, err := vzip.GetBytes(archive, "src/keep.txt")
	if err != nil || string(data) != "keep" {
		t.Fatalf("GetBytes keep = %q, %v", data, err)
	}
	if _, err := vzip.GetBytes(archive, "src/skip.log"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("skip.log err = %v, want not exist", err)
	}

	var buf bytes.Buffer
	if err := vzip.ZipToWriterUsingOptions(&buf, []string{src}, vzip.WithFileFilter(filter)); err != nil {
		t.Fatalf("ZipToWriterUsingOptions: %v", err)
	}
	bufReader, err := archivezip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	if len(bufReader.File) != 1 || bufReader.File[0].Name != "keep.txt" {
		t.Fatalf("writer archive entries = %#v", bufReader.File)
	}
	entry, err := vzip.GetStream(bufReader.File[0])
	if err != nil {
		t.Fatalf("GetStream: %v", err)
	}
	defer func() { _ = entry.Close() }()
	if _, err := io.ReadAll(entry); err != nil {
		t.Fatalf("read entry: %v", err)
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
