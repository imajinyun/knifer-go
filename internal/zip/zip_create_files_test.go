package zip

import (
	"archive/zip"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	knifer "github.com/imajinyun/go-knifer"
)

func newZipCreateSource(t *testing.T) (string, string) {
	t.Helper()
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	if err := os.MkdirAll(filepath.Join(src, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(src, "empty"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "a.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "nested", "b.txt"), []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	return tmp, src
}

func TestZipFilesRejectsUnsafeDestination(t *testing.T) {
	tmp, src := newZipCreateSource(t)
	assertZipCode(t, ZipFiles(filepath.Join(src, "out.zip"), false, src), knifer.ErrCodeInvalidInput)

	destDir := filepath.Join(tmp, "dest-dir")
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		t.Fatal(err)
	}
	assertZipCode(t, ZipFiles(destDir, false, filepath.Join(src, "a.txt")), knifer.ErrCodeInvalidInput)
	if _, err := os.Stat(filepath.Join(src, "out.zip")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("unsafe archive was created inside source, stat err=%v", err)
	}
}

func TestZipFilesCreatesArchiveFromDirectory(t *testing.T) {
	tmp, src := newZipCreateSource(t)
	archive := filepath.Join(tmp, "out.zip")
	if err := ZipFiles(archive, false, src); err != nil {
		t.Fatalf("ZipFiles: %v", err)
	}

	r, err := zip.OpenReader(archive)
	if err != nil {
		t.Fatalf("OpenReader: %v", err)
	}
	defer func() { _ = r.Close() }()

	names := make([]string, 0, len(r.File))
	for _, f := range r.File {
		names = append(names, f.Name)
	}
	slices.Sort(names)
	want := []string{"a.txt", "empty/", "nested/", "nested/b.txt"}
	if len(names) != len(want) {
		t.Fatalf("archive names = %#v", names)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("archive names = %#v", names)
		}
	}
}
