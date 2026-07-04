package vzip_test

import (
	archivezip "archive/zip"
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
	"github.com/imajinyun/knifer-go/vzip"
)

func TestFacadeZipAndExtraction(t *testing.T) {
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
}

func TestFacadeZipAppendAndUnzipOptions(t *testing.T) {
	tmp := t.TempDir()
	archive := filepath.Join(tmp, "append-default.zip")
	if err := vzip.ZipEntries(archive, vzip.EntryData{Name: "base.txt", Data: []byte("base")}); err != nil {
		t.Fatalf("ZipEntries: %v", err)
	}
	extra := filepath.Join(tmp, "extra.txt")
	if err := os.WriteFile(extra, []byte("extra"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := vzip.Append(archive, extra); err != nil {
		t.Fatalf("Append: %v", err)
	}
	if got, err := vzip.GetBytes(archive, "extra.txt"); err != nil || string(got) != "extra" {
		t.Fatalf("Append content = %q, %v", got, err)
	}

	defaultDest, err := vzip.Unzip(archive)
	if err != nil {
		t.Fatalf("Unzip: %v", err)
	}
	if got, err := os.ReadFile(filepath.Join(defaultDest, "base.txt")); err != nil || string(got) != "base" {
		t.Fatalf("Unzip default output = %q, %v", got, err)
	}
	if err := vzip.UnzipToLimit(archive, filepath.Join(tmp, "limit"), 1); err == nil {
		t.Fatal("UnzipToLimit should reject content larger than limit")
	}
}

func TestFacadeUnzipRejectsPathTraversal(t *testing.T) {
	for _, name := range []string{"../evil.txt", `..\evil.txt`} {
		t.Run(name, func(t *testing.T) {
			tmp := t.TempDir()
			archive := filepath.Join(tmp, "bad.zip")
			var buf bytes.Buffer
			zw := archivezip.NewWriter(&buf)
			w, err := zw.Create(name)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := w.Write([]byte("bad")); err != nil {
				t.Fatal(err)
			}
			if err := zw.Close(); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(archive, buf.Bytes(), 0o644); err != nil {
				t.Fatal(err)
			}

			if err := vzip.UnzipTo(archive, filepath.Join(tmp, "dest")); !errors.Is(err, knifer.ErrCodeInvalidInput) {
				t.Fatalf("UnzipTo traversal error = %v, want invalid input", err)
			}
			if _, err := os.Stat(filepath.Join(tmp, "evil.txt")); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("path traversal wrote outside destination, stat err=%v", err)
			}
		})
	}
}

func TestFacadeUnzipRejectsSymlinkEscape(t *testing.T) {
	tmp := t.TempDir()
	archive := filepath.Join(tmp, "symlink.zip")
	if err := vzip.ZipEntries(archive, vzip.EntryData{Name: "link/payload.txt", Data: []byte("bad")}); err != nil {
		t.Fatalf("ZipEntries: %v", err)
	}
	dest := filepath.Join(tmp, "dest")
	outside := filepath.Join(tmp, "outside")
	if err := os.MkdirAll(filepath.Join(dest, "link"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := vzip.UnzipToWithOptions(archive, dest, vzip.WithEvalSymlinks(func(path string) (string, error) {
		if filepath.Clean(path) == filepath.Join(dest, "link") {
			return outside, nil
		}
		return path, nil
	})); !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("UnzipToWithOptions symlink escape error = %v, want invalid input", err)
	}
}
