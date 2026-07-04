package vhttp_test

import (
	"errors"
	"io"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/imajinyun/knifer-go/vhttp"
)

func TestFacadeDownloadFileWrappers(t *testing.T) {
	server := newFacadeDownloadServer(t)
	defer server.Close()

	dir := t.TempDir()
	file := filepath.Join(dir, "plain.txt")
	if n, err := vhttp.DownloadFile(server.URL, file); err != nil || n != int64(len(facadeDownloadText)) {
		t.Fatalf("DownloadFile n=%d err=%v", n, err)
	}
	fileWithOpts := filepath.Join(dir, "with-options.txt")
	if n, err := vhttp.DownloadFileWithOptions(server.URL, fileWithOpts, []vhttp.RequestOption{vhttp.WithMaxResponseBytes(64)}, vhttp.WithSaveOverwrite(true)); err != nil || n != int64(len(facadeDownloadText)) {
		t.Fatalf("DownloadFileWithOptions n=%d err=%v", n, err)
	}
	if n, err := vhttp.DownloadFileSafe(server.URL, filepath.Join(dir, "blocked.txt")); err == nil || n != 0 {
		t.Fatalf("DownloadFileSafe default policy n=%d err=%v, want private host rejection", n, err)
	}
}

func TestFacadeDownloadFileReturnsCloseError(t *testing.T) {
	closeErr := errors.New("close failed")
	server := newFacadeDownloadServer(t)
	defer server.Close()

	n, err := vhttp.DownloadFile(server.URL, "/virtual/out.txt",
		vhttp.WithSaveMkdirAll(func(string, fs.FileMode) error { return nil }),
		vhttp.WithSaveOpenFile(func(string, int, fs.FileMode) (io.WriteCloser, error) {
			return closeErrorWriteCloser{Writer: io.Discard, err: closeErr}, nil
		}),
	)
	if n != int64(len(facadeDownloadText)) {
		t.Fatalf("DownloadFile close error bytes = %d, want %d", n, len(facadeDownloadText))
	}
	if !errors.Is(err, closeErr) {
		t.Fatalf("DownloadFile close error = %v, want close cause", err)
	}
}
