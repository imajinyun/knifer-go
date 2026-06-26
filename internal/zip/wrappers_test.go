package zip

import (
	"archive/zip"
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

// TestConvenienceWrappersRoundTrip exercises the thin convenience wrappers that
// delegate to their *WithOptions counterparts, using a real temp directory.
func TestConvenienceWrappersRoundTrip(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "hello.txt")
	if err := os.WriteFile(src, []byte("hello world"), 0o600); err != nil {
		t.Fatalf("write src: %v", err)
	}

	// Zip -> default sibling .zip path.
	zipPath, err := Zip(src)
	if err != nil || zipPath != filepath.Join(dir, "hello.zip") {
		t.Fatalf("Zip path=%q err=%v", zipPath, err)
	}

	// ReadFile / Open the produced archive.
	if _, err := ReadFile(zipPath); err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	rc, err := Open(zipPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	_ = rc.Close()

	// Get and Read entries.
	entry, err := Get(zipPath, "hello.txt")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	_ = entry.Close()
	seen := 0
	if err := Read(zipPath, func(*zip.File) error { seen++; return nil }); err != nil || seen != 1 {
		t.Fatalf("Read seen=%d err=%v", seen, err)
	}

	// Unzip back to a sibling directory.
	dest, err := Unzip(zipPath)
	if err != nil {
		t.Fatalf("Unzip: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dest, "hello.txt"))
	if err != nil || string(got) != "hello world" {
		t.Fatalf("extracted = %q err=%v", got, err)
	}
}

func TestZipToAndDataAndStreamsWrappers(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(src, []byte("data"), 0o600); err != nil {
		t.Fatalf("write src: %v", err)
	}

	if err := ZipTo(src, filepath.Join(dir, "a.zip"), false); err != nil {
		t.Fatalf("ZipTo: %v", err)
	}
	if err := ZipData(filepath.Join(dir, "data.zip"), "note.txt", "inline"); err != nil {
		t.Fatalf("ZipData: %v", err)
	}
	if err := ZipBytes(filepath.Join(dir, "bytes.zip"), "blob.bin", []byte{1, 2, 3}); err != nil {
		t.Fatalf("ZipBytes: %v", err)
	}
	if err := ZipFilesFilter(filepath.Join(dir, "filtered.zip"), false, nil, src); err != nil {
		t.Fatalf("ZipFilesFilter: %v", err)
	}
	streams := []StreamEntry{{Name: "s.txt", Reader: bytes.NewReader([]byte("stream"))}}
	if err := ZipStreams(filepath.Join(dir, "streams.zip"), streams...); err != nil {
		t.Fatalf("ZipStreams: %v", err)
	}

	var buf bytes.Buffer
	if err := ZipStreamsToWriter(&buf, streams...); err != nil {
		t.Fatalf("ZipStreamsToWriter: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("ZipStreamsToWriter produced empty output")
	}

	// Verify one of the inline archives is readable.
	got, err := GetBytes(filepath.Join(dir, "data.zip"), "note.txt")
	if err != nil || string(got) != "inline" {
		t.Fatalf("GetBytes = %q err=%v", got, err)
	}
}

func TestGzipZlibStringFileWrappers(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "c.txt")
	content := "compress me please, compress me please"
	if err := os.WriteFile(src, []byte(content), 0o600); err != nil {
		t.Fatalf("write src: %v", err)
	}

	// gzip string/file/reader convenience wrappers and their inverse.
	gz, err := GzipString(content)
	if err != nil {
		t.Fatalf("GzipString: %v", err)
	}
	if s, err := UnGzipString(gz); err != nil || s != content {
		t.Fatalf("UnGzipString = %q err=%v", s, err)
	}
	if _, err := GzipFile(src); err != nil {
		t.Fatalf("GzipFile: %v", err)
	}
	if _, err := GzipReader(bytes.NewReader([]byte(content)), len(content)); err != nil {
		t.Fatalf("GzipReader: %v", err)
	}
	if _, err := Gunzip(gz); err != nil {
		t.Fatalf("Gunzip: %v", err)
	}

	// zlib string/file/reader convenience wrappers and their inverse.
	zl, err := ZlibString(content, -1)
	if err != nil {
		t.Fatalf("ZlibString: %v", err)
	}
	if s, err := UnZlibString(zl); err != nil || s != content {
		t.Fatalf("UnZlibString = %q err=%v", s, err)
	}
	if _, err := ZlibFile(src, -1); err != nil {
		t.Fatalf("ZlibFile: %v", err)
	}
	if _, err := UnZlibReader(bytes.NewReader(zl), len(zl)); err != nil {
		t.Fatalf("UnZlibReader: %v", err)
	}
	if _, err := Unzlib(zl); err != nil {
		t.Fatalf("Unzlib: %v", err)
	}
}

// TestArchiveProviderOptionsAreApplied confirms each provider-injection option
// installs its function into the resolved config.
func TestArchiveProviderOptionsAreApplied(t *testing.T) {
	marker := errors.New("marker")
	cfg := applyArchiveOptions([]ArchiveOption{
		WithCompressionMethod(zip.Store),
		WithReadFile(func(string) ([]byte, error) { return nil, marker }),
		WithLstat(func(string) (os.FileInfo, error) { return nil, marker }),
		WithReadDir(func(string) ([]os.DirEntry, error) { return nil, marker }),
		WithReadlink(func(string) (string, error) { return "", marker }),
		WithRemove(func(string) error { return marker }),
		WithRename(func(string, string) error { return marker }),
		WithCreateTemp(func(string, string) (TempFile, error) { return nil, marker }),
	})

	if cfg.compressionMethod != zip.Store {
		t.Fatalf("compressionMethod = %d", cfg.compressionMethod)
	}
	if _, err := cfg.readFile("x"); !errors.Is(err, marker) {
		t.Fatal("readFile not applied")
	}
	if _, err := cfg.lstat("x"); !errors.Is(err, marker) {
		t.Fatal("lstat not applied")
	}
	if _, err := cfg.readDir("x"); !errors.Is(err, marker) {
		t.Fatal("readDir not applied")
	}
	if _, err := cfg.readlink("x"); !errors.Is(err, marker) {
		t.Fatal("readlink not applied")
	}
	if err := cfg.remove("x"); !errors.Is(err, marker) {
		t.Fatal("remove not applied")
	}
	if err := cfg.rename("x", "y"); !errors.Is(err, marker) {
		t.Fatal("rename not applied")
	}
	if _, err := cfg.createTemp("x", "y"); !errors.Is(err, marker) {
		t.Fatal("createTemp not applied")
	}

	// nil providers are ignored, leaving defaults in place.
	def := applyArchiveOptions([]ArchiveOption{WithReadFile(nil), WithRemove(nil)})
	if def.readFile == nil || def.remove == nil {
		t.Fatal("nil providers should leave defaults intact")
	}
}

func TestZipErrorBranches(t *testing.T) {
	// Error() with cause includes both message and cause.
	werr := wrapZipError(knifer.ErrCodeInternal, "boom", errors.New("root"))
	if got := werr.Error(); got != "boom: root" {
		t.Fatalf("Error() = %q", got)
	}
	// wrapZipError with nil cause returns nil.
	if wrapZipError(knifer.ErrCodeInternal, "ignored", nil) != nil {
		t.Fatal("wrapZipError(nil) should be nil")
	}

	e := invalidInputf("bad %s", "input")
	if !e.Is(knifer.ErrCodeInvalidInput) {
		t.Fatal("Is(code) should match")
	}
	if !e.Is(&ZipError{Code: knifer.ErrCodeInvalidInput}) {
		t.Fatal("Is(*ZipError same code) should match")
	}
	if e.Is(knifer.ErrCodeInternal) || e.Is(errors.New("x")) || e.Is(nil) {
		t.Fatal("Is should not match unrelated targets")
	}
	if e.ErrorCode() != knifer.ErrCodeInvalidInput {
		t.Fatalf("ErrorCode() = %q", e.ErrorCode())
	}
	if got := e.Unwrap(); got != nil {
		t.Fatalf("Unwrap() = %v", got)
	}

	// nil receiver behaviour.
	var nilErr *ZipError
	if nilErr.Error() != "" || nilErr.ErrorCode() != "" || nilErr.Unwrap() != nil || nilErr.Is(knifer.ErrCodeInternal) {
		t.Fatal("nil *ZipError methods should be zero-valued")
	}
}
