package vzip_test

import (
	archivezip "archive/zip"
	"bytes"
	"io"
	"path/filepath"
	"testing"

	"github.com/imajinyun/knifer-go/vzip"
)

func TestFacadeZipArchiveReaderAndWriterHelpers(t *testing.T) {
	tmp := t.TempDir()
	archive := filepath.Join(tmp, "helpers.zip")

	if err := vzip.ZipData(archive, "text/data.txt", "hello"); err != nil {
		t.Fatalf("ZipData: %v", err)
	}
	rc, err := vzip.Open(archive)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if len(rc.File) != 1 || rc.File[0].Name != "text/data.txt" {
		_ = rc.Close()
		t.Fatalf("Open files = %#v", rc.File)
	}
	_ = rc.Close()

	names, err := vzip.ListFileNames(archive, "text")
	if err != nil || len(names) != 1 || names[0] != "data.txt" {
		t.Fatalf("ListFileNames = %#v, %v", names, err)
	}
	seen := false
	if err := vzip.Read(archive, func(f *archivezip.File) error {
		seen = f.Name == "text/data.txt"
		return nil
	}); err != nil || !seen {
		t.Fatalf("Read seen=%v err=%v", seen, err)
	}
	reader, err := vzip.Get(archive, "text/data.txt")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	data, err := io.ReadAll(reader)
	_ = reader.Close()
	if err != nil || string(data) != "hello" {
		t.Fatalf("Get data = %q, %v", data, err)
	}

	bytesArchive := filepath.Join(tmp, "bytes.zip")
	if err := vzip.ZipBytes(bytesArchive, "bytes.bin", []byte{1, 2, 3}); err != nil {
		t.Fatalf("ZipBytes: %v", err)
	}
	if got, err := vzip.GetBytes(bytesArchive, "bytes.bin"); err != nil || !bytes.Equal(got, []byte{1, 2, 3}) {
		t.Fatalf("ZipBytes content = %v, %v", got, err)
	}

	var buf bytes.Buffer
	zw := vzip.NewWriter(&buf)
	w, err := zw.Create("manual.txt")
	if err != nil {
		t.Fatalf("Create manual entry: %v", err)
	}
	if _, err := w.Write([]byte("manual")); err != nil {
		t.Fatalf("write manual entry: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close manual writer: %v", err)
	}
	zr, err := archivezip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil || len(zr.File) != 1 || zr.File[0].Name != "manual.txt" {
		t.Fatalf("NewWriter archive = %#v, %v", zr, err)
	}
}
