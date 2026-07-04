package http

import (
	"errors"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadFileToFile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("file-content"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	target := filepath.Join(dir, "out.txt")
	n, err := DownloadFile(srv.URL, target)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n != int64(len("file-content")) {
		t.Fatalf("size: %d", n)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(data) != "file-content" {
		t.Fatalf("content: %q", string(data))
	}
}

func TestDownloadFileWithOptions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Mode") != "file" {
			http.Error(w, "missing option header", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte("file-options"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	n, err := DownloadFileWithOptions(srv.URL, dir, []RequestOption{WithHeader("X-Mode", "file")}, WithSaveDefaultFilename("out.txt"))
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n != int64(len("file-options")) {
		t.Fatalf("size: %d", n)
	}
	data, err := os.ReadFile(filepath.Join(dir, "out.txt"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(data) != "file-options" {
		t.Fatalf("content: %q", string(data))
	}
}

func TestDownloadFileToDirectory(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("D"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	// dest is a directory, so the file name should come from the URL path.
	url := srv.URL + "/foo.bin"
	if _, err := DownloadFile(url, dir); err != nil {
		t.Fatalf("err: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "foo.bin")); err != nil {
		t.Fatalf("file should exist: %v", err)
	}
}

func TestDownloadFileReturnsCloseError(t *testing.T) {
	closeErr := errors.New("close failed")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("file-close"))
	}))
	defer srv.Close()

	n, err := DownloadFile(srv.URL, "/virtual/out.txt",
		WithSaveMkdirAll(func(string, fs.FileMode) error { return nil }),
		WithSaveOpenFile(func(string, int, fs.FileMode) (io.WriteCloser, error) {
			return closeErrorWriteCloser{Writer: io.Discard, err: closeErr}, nil
		}),
	)
	if n != int64(len("file-close")) {
		t.Fatalf("DownloadFile close error bytes = %d, want %d", n, len("file-close"))
	}
	if !errors.Is(err, closeErr) {
		t.Fatalf("DownloadFile close error = %v, want close cause", err)
	}
}
