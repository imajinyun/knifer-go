package zip

import (
	archivezip "archive/zip"
	"bytes"
	"compress/flate"
	"errors"
	"io"
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

func TestAppendWithOptions(t *testing.T) {
	tmp := t.TempDir()
	archive := filepath.Join(tmp, "entries.zip")
	if err := ZipEntries(archive, EntryData{Name: "a.txt", Data: []byte("a")}); err != nil {
		t.Fatalf("ZipEntries: %v", err)
	}
	keepFile := filepath.Join(tmp, "keep.txt")
	if err := os.WriteFile(keepFile, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	skipFile := filepath.Join(tmp, "skip.log")
	if err := os.WriteFile(skipFile, []byte("skip"), 0o644); err != nil {
		t.Fatal(err)
	}
	filter := func(path string, info os.FileInfo) bool {
		return info.IsDir() || filepath.Ext(path) == ".txt"
	}
	if err := AppendWithOptions(archive, keepFile, WithFileFilter(filter), WithCompressionLevel(flate.BestSpeed)); err != nil {
		t.Fatalf("AppendWithOptions keep: %v", err)
	}
	if err := AppendWithOptions(archive, skipFile, WithFileFilter(filter)); err != nil {
		t.Fatalf("AppendWithOptions skip: %v", err)
	}
	data, err := GetBytes(archive, "keep.txt")
	if err != nil || string(data) != "keep" {
		t.Fatalf("appended keep.txt = %q, %v", data, err)
	}
	if _, err := GetBytes(archive, "skip.log"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("filtered skip.log err = %v, want not exist", err)
	}
}

func TestGzipWithOptions(t *testing.T) {
	data := bytes.Repeat([]byte("abcdef"), 32)
	gz, err := GzipWithOptions(data, WithCompressionLevel(flate.BestSpeed))
	if err != nil {
		t.Fatalf("GzipWithOptions: %v", err)
	}
	out, err := UnGzip(gz)
	if err != nil || !bytes.Equal(out, data) {
		t.Fatalf("UnGzip after GzipWithOptions = %q, %v", out, err)
	}

	gz, err = GzipReaderWithOptions(bytes.NewReader(data), 0, WithCompressionLevel(flate.NoCompression))
	if err != nil {
		t.Fatalf("GzipReaderWithOptions: %v", err)
	}
	out, err = UnGzip(gz)
	if err != nil || !bytes.Equal(out, data) {
		t.Fatalf("UnGzip after GzipReaderWithOptions = %q, %v", out, err)
	}
	if _, err := GzipWithOptions(data, WithCompressionLevel(flate.HuffmanOnly+100)); err == nil {
		t.Fatal("GzipWithOptions should reject invalid compression level")
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

func TestZipCreationUsingOptions(t *testing.T) {
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
