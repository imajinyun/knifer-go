package http

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	knifer "github.com/imajinyun/go-knifer"
)

func TestSaveAsStreamsWithoutCachingBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("x", 64*1024)))
	}))
	defer srv.Close()

	resp := Get(srv.URL).Execute()
	target := filepath.Join(t.TempDir(), "stream.bin")
	n, err := resp.SaveAs(target)
	if err != nil {
		t.Fatalf("SaveAs() error = %v", err)
	}
	if n != 64*1024 {
		t.Fatalf("SaveAs() wrote %d bytes, want %d", n, 64*1024)
	}
	if resp.body != nil {
		t.Fatalf("SaveAs() should stream to file without caching response body, cached %d bytes", len(resp.body))
	}
	if got := resp.Bytes(); got != nil {
		t.Fatalf("Bytes() after streamed SaveAs = %q, want nil", string(got))
	}
	if !errors.Is(resp.Err(), knifer.ErrCodeUnsupported) {
		t.Fatalf("Err() after Bytes() on consumed body = %v, want unsupported", resp.Err())
	}
}

func TestSaveAsHonorsMaxResponseBytes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("too-large"))
	}))
	defer srv.Close()

	target := filepath.Join(t.TempDir(), "limited.txt")
	n, err := Get(srv.URL, WithMaxResponseBytes(3)).Execute().SaveAs(target)
	if !errors.Is(err, knifer.ErrCodeUnsupported) {
		t.Fatalf("SaveAs() error = %v, want unsupported", err)
	}
	if n != 3 {
		t.Fatalf("SaveAs() wrote %d bytes, want exactly the configured limit", n)
	}
	data, readErr := os.ReadFile(target)
	if readErr != nil {
		t.Fatalf("read partial file: %v", readErr)
	}
	if string(data) != "too" {
		t.Fatalf("partial file = %q, want only bytes within the limit", string(data))
	}
}

func TestSaveAsUsesCachedBodyAfterBodyRead(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("cached"))
	}))
	defer srv.Close()

	resp := Get(srv.URL).Execute()
	if got := resp.Body(); got != "cached" {
		t.Fatalf("Body() = %q, want cached", got)
	}
	target := filepath.Join(t.TempDir(), "cached.txt")
	if _, err := resp.SaveAs(target); err != nil {
		t.Fatalf("SaveAs() after Body() error = %v", err)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "cached" {
		t.Fatalf("saved content = %q, want cached", data)
	}
}

func TestSaveAsOptions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("saved"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	target := filepath.Join(dir, "out.txt")
	if err := os.WriteFile(target, []byte("old"), 0o644); err != nil {
		t.Fatalf("write old: %v", err)
	}
	if _, err := Get(srv.URL).Execute().SaveAs(target, WithSaveOverwrite(false)); err == nil {
		t.Fatal("SaveAs overwrite false should fail")
	}
	missing := filepath.Join(dir, "missing", "out.txt")
	if _, err := Get(srv.URL).Execute().SaveAs(missing, WithSaveCreateParents(false)); err == nil {
		t.Fatal("SaveAs without parent creation should fail")
	}
	if _, err := DownloadFile(srv.URL, target); err != nil {
		t.Fatalf("DownloadFile overwrite default: %v", err)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(data) != "saved" {
		t.Fatalf("content = %q", data)
	}
}

func TestSaveAsProviderOptions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("provider-save"))
	}))
	defer srv.Close()

	var mkdirPath string
	var mkdirPerm fs.FileMode
	var openPath string
	var openFlag int
	var openPerm fs.FileMode
	var written bytes.Buffer
	n, err := Get(srv.URL).Execute().SaveAs("/virtual/out.txt",
		WithSaveMkdirAll(func(path string, perm fs.FileMode) error {
			mkdirPath, mkdirPerm = path, perm
			return nil
		}),
		WithSaveOpenFile(func(path string, flag int, perm fs.FileMode) (io.WriteCloser, error) {
			openPath, openFlag, openPerm = path, flag, perm
			return nopWriteCloser{Writer: &written}, nil
		}),
		WithSaveDirPerm(0o700), WithSaveFilePerm(0o600),
	)
	if err != nil || n != int64(len("provider-save")) {
		t.Fatalf("SaveAs provider n=%d err=%v", n, err)
	}
	if mkdirPath != "/virtual" || mkdirPerm != 0o700 || openPath != "/virtual/out.txt" || openPerm != 0o600 || openFlag&os.O_CREATE == 0 || written.String() != "provider-save" {
		t.Fatalf("providers mkdir=%q/%v open=%q flag=%#x perm=%v content=%q", mkdirPath, mkdirPerm, openPath, openFlag, openPerm, written.String())
	}
}

type nopWriteCloser struct{ io.Writer }

func (w nopWriteCloser) Close() error { return nil }

func TestSaveAsDefaultFilenameOption(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("fallback"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	if _, err := Get(srv.URL).Execute().SaveAs(dir, WithSaveDefaultFilename("fallback.bin")); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "fallback.bin")); err != nil {
		t.Fatalf("fallback file missing: %v", err)
	}
}
