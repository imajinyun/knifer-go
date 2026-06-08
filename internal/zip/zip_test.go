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
	"time"

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

func TestDecompressHelpersEnforceConfiguredLimit(t *testing.T) {
	data := bytes.Repeat([]byte("x"), 16)
	gz, err := Gzip(data)
	if err != nil {
		t.Fatalf("Gzip: %v", err)
	}
	zl, err := Zlib(data)
	if err != nil {
		t.Fatalf("Zlib: %v", err)
	}

	if got := applyDecompressOptions(nil).maxBytes; got != DefaultUnzipMaxBytes {
		t.Fatalf("default decompression max bytes = %d, want %d", got, DefaultUnzipMaxBytes)
	}
	if _, err := UnGzipWithOptions(gz, WithMaxBytes(8)); err == nil {
		t.Fatal("UnGzip should enforce configured decompression limit")
	}
	if _, err := UnZlibWithOptions(zl, WithMaxBytes(8)); err == nil {
		t.Fatal("UnZlib should enforce configured decompression limit")
	}
	if out, err := UnGzipWithOptions(gz, WithMaxBytes(0)); err != nil || !bytes.Equal(out, data) {
		t.Fatalf("UnGzip explicit unlimited = %q, %v", out, err)
	}
	if out, err := UnZlibWithOptions(zl, WithMaxBytes(0)); err != nil || !bytes.Equal(out, data) {
		t.Fatalf("UnZlib explicit unlimited = %q, %v", out, err)
	}
}

func TestGetBytesEnforcesConfiguredLimit(t *testing.T) {
	tmp := t.TempDir()
	archive := filepath.Join(tmp, "limit.zip")
	data := bytes.Repeat([]byte("x"), 16)
	if err := ZipEntries(archive, EntryData{Name: "a.txt", Data: data}); err != nil {
		t.Fatalf("ZipEntries: %v", err)
	}

	if got := applyDecompressOptions(nil).maxBytes; got != DefaultUnzipMaxBytes {
		t.Fatalf("default entry read max bytes = %d, want %d", got, DefaultUnzipMaxBytes)
	}
	if _, err := GetBytesWithOptions(archive, "a.txt", WithMaxBytes(8)); err == nil {
		t.Fatal("GetBytes should enforce configured entry read limit")
	}
	if out, err := GetBytesWithOptions(archive, "a.txt", WithMaxBytes(0)); err != nil || !bytes.Equal(out, data) {
		t.Fatalf("GetBytes explicit unlimited = %q, %v", out, err)
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

func TestUnzipDefaultLimitCanBeOverridden(t *testing.T) {
	tmp := t.TempDir()
	archive := filepath.Join(tmp, "limit.zip")
	entries := []EntryData{{Name: "a.txt", Data: []byte("abcd")}}
	if err := ZipEntries(archive, entries...); err != nil {
		t.Fatalf("ZipEntries: %v", err)
	}
	if err := UnzipToWithOptions(archive, filepath.Join(tmp, "limited"), WithMaxBytes(3)); err == nil {
		t.Fatal("UnzipToWithOptions should reject archives larger than max bytes")
	}
	if err := UnzipToLimit(archive, filepath.Join(tmp, "unlimited"), -1); err != nil {
		t.Fatalf("UnzipToLimit with explicit unlimited limit: %v", err)
	}
	if got := applyUnzipOptions(nil).maxBytes; got != DefaultUnzipMaxBytes {
		t.Fatalf("default unzip max bytes = %d, want %d", got, DefaultUnzipMaxBytes)
	}
	if got := applyUnzipOptions([]ArchiveOption{WithMaxBytes(-1)}).maxBytes; got != -1 {
		t.Fatalf("explicit unlimited unzip max bytes = %d, want -1", got)
	}
}

func TestUnzipEnforcesActualCopiedBytes(t *testing.T) {
	var buf bytes.Buffer
	if err := ZipEntriesToWriter(&buf, EntryData{Name: "a.txt", Data: []byte("abcd")}); err != nil {
		t.Fatalf("ZipEntriesToWriter: %v", err)
	}
	r, err := archivezip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	if len(r.File) != 1 {
		t.Fatalf("archive entries = %d, want 1", len(r.File))
	}
	// Simulate an archive whose declared uncompressed size is smaller than the
	// actual stream. Extraction must enforce the copy-time limit as a second line
	// of defense instead of trusting central-directory metadata only.
	r.File[0].UncompressedSize64 = 1
	dest := filepath.Join(t.TempDir(), "dest")
	if err := UnzipReaderToWithOptions(r, dest, WithMaxBytes(3)); err == nil {
		t.Fatal("UnzipReaderToWithOptions should reject streams exceeding the actual copy limit")
	}
	data, err := os.ReadFile(filepath.Join(dest, "a.txt"))
	if err != nil {
		t.Fatalf("read partial extraction: %v", err)
	}
	if len(data) > 3 {
		t.Fatalf("partial extraction wrote %d bytes, want at most 3", len(data))
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

func TestArchiveProviderOptionsForZipEntries(t *testing.T) {
	var mkdirPath string
	var mkdirPerm os.FileMode
	var openPath string
	var openFlag int
	var openPerm os.FileMode
	var buf bytes.Buffer
	closer := &zipBufferWriteCloser{Buffer: &buf}

	err := ZipEntriesWithOptions("parent/archive.zip", []EntryData{{Name: "a.txt", Data: []byte("a")}},
		WithDirPerm(0o700),
		WithFilePerm(0o600),
		WithMkdirAll(func(path string, perm os.FileMode) error {
			mkdirPath = path
			mkdirPerm = perm
			return nil
		}),
		WithOpenFile(func(path string, flag int, perm os.FileMode) (io.WriteCloser, error) {
			openPath = path
			openFlag = flag
			openPerm = perm
			return closer, nil
		}),
	)
	if err != nil {
		t.Fatalf("ZipEntriesWithOptions() error = %v", err)
	}
	if mkdirPath != "parent" || mkdirPerm != 0o700 {
		t.Fatalf("mkdir = %q/%o, want parent/700", mkdirPath, mkdirPerm)
	}
	if openPath != "parent/archive.zip" || openPerm != 0o600 || openFlag&os.O_TRUNC == 0 {
		t.Fatalf("open = %q/%o/%#x", openPath, openPerm, openFlag)
	}
	if !closer.closed {
		t.Fatal("archive output was not closed")
	}
	r, err := archivezip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("NewReader(provider output): %v", err)
	}
	if len(r.File) != 1 || r.File[0].Name != "a.txt" {
		t.Fatalf("entries = %#v, want a.txt", r.File)
	}
}

func TestArchiveProviderOptionsForFileCompression(t *testing.T) {
	data := []byte("provider-data")
	openPath := ""
	statPath := ""
	open := func(path string) (io.ReadCloser, error) {
		openPath = path
		return io.NopCloser(bytes.NewReader(data)), nil
	}
	stat := func(path string) (os.FileInfo, error) {
		statPath = path
		return zipFakeFileInfo{name: path, size: int64(len(data))}, nil
	}
	gz, err := GzipFileWithOptions("virtual.txt", WithOpen(open), WithStat(stat), WithCompressionLevel(flate.BestSpeed))
	if err != nil {
		t.Fatalf("GzipFileWithOptions() error = %v", err)
	}
	if openPath != "virtual.txt" || statPath != "virtual.txt" {
		t.Fatalf("provider paths open=%q stat=%q", openPath, statPath)
	}
	out, err := UnGzip(gz)
	if err != nil || !bytes.Equal(out, data) {
		t.Fatalf("UnGzip(provider gzip) = %q, %v", out, err)
	}
	z, err := ZlibFileWithOptions("virtual.txt", flate.BestSpeed, WithOpen(open), WithStat(stat))
	if err != nil {
		t.Fatalf("ZlibFileWithOptions() error = %v", err)
	}
	out, err = UnZlib(z)
	if err != nil || !bytes.Equal(out, data) {
		t.Fatalf("UnZlib(provider zlib) = %q, %v", out, err)
	}
}

func TestArchiveProviderOptionsForReadAndExtract(t *testing.T) {
	tmp := t.TempDir()
	archive := filepath.Join(tmp, "entries.zip")
	if err := ZipEntries(archive, EntryData{Name: "a.txt", Data: []byte("a")}); err != nil {
		t.Fatalf("ZipEntries() error = %v", err)
	}
	opened := ""
	openZip := func(path string) (*archivezip.ReadCloser, error) {
		opened = path
		return archivezip.OpenReader(path)
	}
	data, err := GetBytesWithOptions(archive, "a.txt", WithOpenZipReader(openZip))
	if err != nil || string(data) != "a" || opened != archive {
		t.Fatalf("GetBytesWithOptions() = %q, %v, opened=%q", data, err, opened)
	}
	names, err := ListFileNamesWithOptions(archive, "", WithOpenZipReader(openZip))
	if err != nil || !reflect.DeepEqual(names, []string{"a.txt"}) {
		t.Fatalf("ListFileNamesWithOptions() = %v, %v", names, err)
	}
	seen := false
	if err := ReadWithOptions(archive, func(f *archivezip.File) error {
		seen = f.Name == "a.txt"
		return nil
	}, WithOpenZipReader(openZip)); err != nil || !seen {
		t.Fatalf("ReadWithOptions() = %v, seen=%v", err, seen)
	}

	r, err := archivezip.OpenReader(archive)
	if err != nil {
		t.Fatalf("OpenReader() error = %v", err)
	}
	defer func() { _ = r.Close() }()
	var extracted bytes.Buffer
	var mkdirs []string
	if err := UnzipReaderToWithOptions(&r.Reader, "dest",
		WithMkdirAll(func(path string, perm os.FileMode) error {
			mkdirs = append(mkdirs, path)
			return nil
		}),
		WithOpenFile(func(path string, flag int, perm os.FileMode) (io.WriteCloser, error) {
			return &zipBufferWriteCloser{Buffer: &extracted}, nil
		}),
	); err != nil {
		t.Fatalf("UnzipReaderToWithOptions() error = %v", err)
	}
	if extracted.String() != "a" || len(mkdirs) == 0 {
		t.Fatalf("extracted=%q mkdirs=%v", extracted.String(), mkdirs)
	}
}

type zipBufferWriteCloser struct {
	*bytes.Buffer
	closed bool
}

func (w *zipBufferWriteCloser) Close() error {
	w.closed = true
	return nil
}

type zipFakeFileInfo struct {
	name string
	size int64
	dir  bool
}

func (f zipFakeFileInfo) Name() string { return f.name }
func (f zipFakeFileInfo) Size() int64  { return f.size }
func (f zipFakeFileInfo) Mode() os.FileMode {
	if f.dir {
		return os.ModeDir | 0o755
	}
	return 0o644
}
func (f zipFakeFileInfo) ModTime() time.Time { return time.Unix(0, 0) }
func (f zipFakeFileInfo) IsDir() bool        { return f.dir }
func (f zipFakeFileInfo) Sys() any           { return nil }

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
