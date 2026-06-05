package zip

import (
	archivezip "archive/zip"
	"bytes"
	"compress/flate"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	knifer "github.com/imajinyun/go-knifer"
)

func TestZipFilesUnzipGetAndList(t *testing.T) {
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
	archive := filepath.Join(tmp, "out.zip")
	if err := ZipFiles(archive, false, src); err != nil {
		t.Fatalf("ZipFiles: %v", err)
	}
	data, err := GetBytes(archive, "a.txt")
	if err != nil || string(data) != "a" {
		t.Fatalf("GetBytes: %q %v", data, err)
	}
	names, err := ListFileNames(archive, "")
	if err != nil {
		t.Fatalf("ListFileNames: %v", err)
	}
	sort.Strings(names)
	if !reflect.DeepEqual(names, []string{"a.txt"}) {
		t.Fatalf("names: %#v", names)
	}
	dest := filepath.Join(tmp, "dest")
	if err := UnzipTo(archive, dest); err != nil {
		t.Fatalf("UnzipTo: %v", err)
	}
	if got, err := os.ReadFile(filepath.Join(dest, "nested", "b.txt")); err != nil || string(got) != "b" {
		t.Fatalf("unzipped: %q %v", got, err)
	}
	if info, err := os.Stat(filepath.Join(dest, "empty")); err != nil || !info.IsDir() {
		t.Fatalf("empty directory was not restored, info=%v err=%v", info, err)
	}
}

func TestZipEntriesAppendReadAndLimit(t *testing.T) {
	tmp := t.TempDir()
	archive := filepath.Join(tmp, "entries.zip")
	if err := ZipEntries(archive, EntryData{Name: "a.txt", Data: []byte("a")}); err != nil {
		t.Fatalf("ZipEntries: %v", err)
	}
	appendFile := filepath.Join(tmp, "b.txt")
	if err := os.WriteFile(appendFile, []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Append(archive, appendFile); err != nil {
		t.Fatalf("Append: %v", err)
	}
	seen := map[string]bool{}
	if err := Read(archive, func(f *archivezip.File) error {
		seen[f.Name] = true
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !seen["a.txt"] || !seen["b.txt"] {
		t.Fatalf("seen: %#v", seen)
	}
	assertZipCode(t, UnzipToLimit(archive, filepath.Join(tmp, "limited"), 1), knifer.ErrCodeInvalidInput)
	_, err := Get(archive, "missing.txt")
	assertZipCode(t, err, knifer.ErrCodeNotFound)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing entry error should preserve os.ErrNotExist: %v", err)
	}
}

func TestArchiveOptions(t *testing.T) {
	tmp := t.TempDir()
	archive := filepath.Join(tmp, "entries.zip")
	entries := []EntryData{{Name: "a.txt", Data: []byte("abcd")}}
	if err := ZipEntriesWithOptions(archive, entries, WithFilePerm(0o600), WithCompressionLevel(flate.BestSpeed)); err != nil {
		t.Fatalf("ZipEntriesWithOptions: %v", err)
	}
	info, err := os.Stat(archive)
	if err != nil {
		t.Fatalf("stat archive: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("archive perm = %o, want 600", got)
	}
	if err := ZipEntriesWithOptions(archive, entries, WithOverwrite(false)); !errors.Is(err, os.ErrExist) {
		t.Fatalf("overwrite false err = %v, want exists", err)
	}
	if _, err := GetBytesWithOptions(archive, "a.txt", WithMaxBytes(3)); err == nil {
		t.Fatal("GetBytesWithOptions over limit error = nil")
	}
	dest := filepath.Join(tmp, "dest")
	if err := UnzipToWithOptions(archive, dest, WithDirPerm(0o700), WithFilePerm(0o600), WithPreserveMode(false)); err != nil {
		t.Fatalf("UnzipToWithOptions: %v", err)
	}
	fileInfo, err := os.Stat(filepath.Join(dest, "a.txt"))
	if err != nil {
		t.Fatalf("stat extracted: %v", err)
	}
	if got := fileInfo.Mode().Perm(); got != 0o600 {
		t.Fatalf("extracted perm = %o, want 600", got)
	}
	if err := UnzipToWithOptions(archive, dest, WithOverwrite(false)); !errors.Is(err, os.ErrExist) {
		t.Fatalf("unzip overwrite false err = %v, want exists", err)
	}
}

func TestGzipAndZlib(t *testing.T) {
	data := []byte("hello compression")
	gz, err := Gzip(data)
	if err != nil {
		t.Fatalf("Gzip: %v", err)
	}
	out, err := UnGzip(gz)
	if err != nil || !bytes.Equal(out, data) {
		t.Fatalf("UnGzip: %q %v", out, err)
	}
	z, err := ZlibLevel(data, 6)
	if err != nil {
		t.Fatalf("ZlibLevel: %v", err)
	}
	out, err = UnZlib(z)
	if err != nil || !bytes.Equal(out, data) {
		t.Fatalf("UnZlib: %q %v", out, err)
	}
	if _, err := UnGzipWithOptions(gz, WithMaxBytes(3)); err == nil {
		t.Fatal("UnGzipWithOptions over limit error = nil")
	}
	if _, err := UnZlibWithOptions(z, WithMaxBytes(3)); err == nil {
		t.Fatal("UnZlibWithOptions over limit error = nil")
	}
}

func TestUnzipRejectsPathTraversal(t *testing.T) {
	tmp := t.TempDir()
	archive := filepath.Join(tmp, "bad.zip")
	var buf bytes.Buffer
	zw := archivezip.NewWriter(&buf)
	w, err := zw.Create("../evil.txt")
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
	assertZipCode(t, UnzipTo(archive, filepath.Join(tmp, "dest")), knifer.ErrCodeInvalidInput)
}

func TestZipErrorContract(t *testing.T) {
	_, err := GetStream(nil)
	assertZipCode(t, err, knifer.ErrCodeInvalidInput)
	assertZipCode(t, UnzipReaderToLimit(nil, t.TempDir(), -1), knifer.ErrCodeInvalidInput)

	var buf bytes.Buffer
	err = ZipEntriesToWriter(&buf, EntryData{Name: "../evil.txt", Data: []byte("bad")})
	assertZipCode(t, err, knifer.ErrCodeInvalidInput)
}

func assertZipCode(t *testing.T, err error, code knifer.ErrCode) {
	t.Helper()
	if err == nil {
		t.Fatalf("err = nil, want %s", code)
	}
	if !errors.Is(err, code) {
		t.Fatalf("errors.Is(%v, %s) = false", err, code)
	}
	got, ok := knifer.CodeOf(err)
	if !ok || got != code {
		t.Fatalf("CodeOf(%v) = %q, %v; want %q, true", err, got, ok, code)
	}
}
