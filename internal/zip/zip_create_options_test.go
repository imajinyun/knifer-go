package zip

import (
	archivezip "archive/zip"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func newZipCreateFilterSource(t *testing.T) (string, string, FileFilter) {
	t.Helper()
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
	filter := func(path string, info os.FileInfo) bool {
		return info.IsDir() || filepath.Ext(path) == ".txt"
	}
	return tmp, src, filter
}

func TestZipFilesWithOptionsFiltersSources(t *testing.T) {
	tmp, src, filter := newZipCreateFilterSource(t)
	archive := filepath.Join(tmp, "filtered.zip")
	if err := ZipFilesWithOptions(archive, []string{src}, WithSourceDir(true), WithFileFilter(filter)); err != nil {
		t.Fatalf("ZipFilesWithOptions: %v", err)
	}
	data, err := GetBytes(archive, "src/keep.txt")
	if err != nil || string(data) != "keep" {
		t.Fatalf("GetBytes keep = %q, %v", data, err)
	}
	if _, err := GetBytes(archive, "src/skip.log"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("skip.log err = %v, want not exist", err)
	}
}

func TestNilFileFilterOptionDoesNotClearPreviousFilter(t *testing.T) {
	tmp, src, filter := newZipCreateFilterSource(t)
	archive := filepath.Join(tmp, "filtered.zip")
	if err := ZipFilesWithOptions(archive, []string{src}, WithSourceDir(true), WithFileFilter(filter), WithFileFilter(nil)); err != nil {
		t.Fatalf("ZipFilesWithOptions: %v", err)
	}
	if _, err := GetBytes(archive, "src/skip.log"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("skip.log err = %v, want not exist", err)
	}
}

func TestZipToWriterWithOptionsFiltersSources(t *testing.T) {
	_, src, filter := newZipCreateFilterSource(t)
	var buf bytes.Buffer
	if err := ZipToWriterWithOptions(&buf, []string{src}, WithFileFilter(filter)); err != nil {
		t.Fatalf("ZipToWriterWithOptions: %v", err)
	}
	r, err := archivezip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	for _, f := range r.File {
		if f.Name == "skip.log" {
			t.Fatal("ZipToWriterWithOptions should filter skip.log")
		}
		if f.Name == "keep.txt" {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("open keep.txt: %v", err)
			}
			content, err := io.ReadAll(rc)
			_ = rc.Close()
			if err != nil || string(content) != "keep" {
				t.Fatalf("keep.txt = %q, %v", content, err)
			}
			return
		}
	}
	t.Fatal("keep.txt not found in writer archive")
}
