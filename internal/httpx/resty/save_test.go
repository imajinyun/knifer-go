package resty

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

	knifer "github.com/imajinyun/knifer-go"
)

func TestSaveAsOptions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("resty-save"))
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
	if _, err := DownloadFile(srv.URL, target); err != nil {
		t.Fatalf("DownloadFile overwrite default: %v", err)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(data) != "resty-save" {
		t.Fatalf("content = %q", data)
	}
}

func TestSaveAsRejectsUnsafeContentDispositionFilename(t *testing.T) {
	tests := []string{
		`attachment; filename="../outside"`,
		`attachment; filename="..\outside"`,
		`attachment; filename="/tmp/outside"`,
	}
	for _, cd := range tests {
		t.Run(cd, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Disposition", cd)
				_, _ = w.Write([]byte("unsafe"))
			}))
			defer srv.Close()

			dir := t.TempDir()
			_, err := Get(srv.URL).Execute().SaveAs(dir)
			if !errors.Is(err, knifer.ErrCodeInvalidInput) {
				t.Fatalf("SaveAs error = %v, want invalid input", err)
			}
			if _, statErr := os.Stat(filepath.Join(dir, "outside")); !errors.Is(statErr, os.ErrNotExist) {
				t.Fatalf("unsafe file should not be created, stat err = %v", statErr)
			}
		})
	}
}

func TestSaveAsProviderOptions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("resty-provider-save"))
	}))
	defer srv.Close()

	var mkdirPath string
	var mkdirPerm fs.FileMode
	var openPath string
	var openFlag int
	var openPerm fs.FileMode
	var written bytes.Buffer
	n, err := Get(srv.URL).Execute().SaveAs("/virtual/resty.txt",
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
	if err != nil || n != int64(len("resty-provider-save")) {
		t.Fatalf("SaveAs provider n=%d err=%v", n, err)
	}
	if mkdirPath != "/virtual" || mkdirPerm != 0o700 || openPath != "/virtual/resty.txt" || openPerm != 0o600 || openFlag&os.O_CREATE == 0 || written.String() != "resty-provider-save" {
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

func TestSaveAsReturnsCloseError(t *testing.T) {
	closeErr := errors.New("close failed")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("resty-provider-save"))
	}))
	defer srv.Close()

	n, err := Get(srv.URL).Execute().SaveAs("/virtual/resty.txt",
		WithSaveMkdirAll(func(string, fs.FileMode) error { return nil }),
		WithSaveOpenFile(func(string, int, fs.FileMode) (io.WriteCloser, error) {
			return closeErrorWriteCloser{Writer: io.Discard, err: closeErr}, nil
		}),
	)
	if n != int64(len("resty-provider-save")) {
		t.Fatalf("SaveAs close error bytes = %d, want %d", n, len("resty-provider-save"))
	}
	if !errors.Is(err, closeErr) {
		t.Fatalf("SaveAs close error = %v, want close cause", err)
	}
}
