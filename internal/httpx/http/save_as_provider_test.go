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
	"testing"
)

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

func TestNilSaveProviderOptionsDoNotOverwriteConfiguredProviders(t *testing.T) {
	stat := func(string) (os.FileInfo, error) { return nil, os.ErrNotExist }
	mkdirAll := func(string, fs.FileMode) error { return nil }
	openFile := func(string, int, fs.FileMode) (io.WriteCloser, error) {
		return nopWriteCloser{Writer: io.Discard}, nil
	}
	cfg := applySaveOptions([]SaveOption{
		WithSaveStat(stat), WithSaveStat(nil),
		WithSaveMkdirAll(mkdirAll), WithSaveMkdirAll(nil),
		WithSaveOpenFile(openFile), WithSaveOpenFile(nil),
	})
	if cfg.stat == nil || cfg.mkdirAll == nil || cfg.openFile == nil {
		t.Fatalf("nil save provider option overwrote configured provider: %#v", cfg)
	}
}

type nopWriteCloser struct{ io.Writer }

func (w nopWriteCloser) Close() error { return nil }

type closeErrorWriteCloser struct {
	io.Writer
	err error
}

func (w closeErrorWriteCloser) Close() error { return w.err }

func TestSaveAsReturnsCloseError(t *testing.T) {
	closeErr := errors.New("close failed")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("provider-save"))
	}))
	defer srv.Close()

	n, err := Get(srv.URL).Execute().SaveAs("/virtual/out.txt",
		WithSaveMkdirAll(func(string, fs.FileMode) error { return nil }),
		WithSaveOpenFile(func(string, int, fs.FileMode) (io.WriteCloser, error) {
			return closeErrorWriteCloser{Writer: io.Discard, err: closeErr}, nil
		}),
	)
	if n != int64(len("provider-save")) {
		t.Fatalf("SaveAs close error bytes = %d, want %d", n, len("provider-save"))
	}
	if !errors.Is(err, closeErr) {
		t.Fatalf("SaveAs close error = %v, want close cause", err)
	}
}

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
