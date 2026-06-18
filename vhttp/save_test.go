package vhttp_test

import (
	"bytes"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/imajinyun/go-knifer/vhttp"
)

func TestFacadeSaveProviderOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("vhttp-save"))
	}))
	defer server.Close()

	var mkdirPath string
	var mkdirPerm fs.FileMode
	var openPath string
	var openFlag int
	var openPerm fs.FileMode
	var written bytes.Buffer
	n, err := vhttp.Get(server.URL).Execute().SaveAs("/virtual/out.txt",
		vhttp.WithSaveMkdirAll(func(path string, perm fs.FileMode) error {
			mkdirPath, mkdirPerm = path, perm
			return nil
		}),
		vhttp.WithSaveOpenFile(func(path string, flag int, perm fs.FileMode) (io.WriteCloser, error) {
			openPath, openFlag, openPerm = path, flag, perm
			return nopWriteCloser{Writer: &written}, nil
		}),
		vhttp.WithSaveDirPerm(0o700), vhttp.WithSaveFilePerm(0o600),
	)
	if err != nil || n != int64(len("vhttp-save")) {
		t.Fatalf("SaveAs provider n=%d err=%v", n, err)
	}
	if mkdirPath != "/virtual" || mkdirPerm != 0o700 || openPath != "/virtual/out.txt" || openPerm != 0o600 || openFlag&os.O_CREATE == 0 || written.String() != "vhttp-save" {
		t.Fatalf("providers mkdir=%q/%v open=%q flag=%#x perm=%v content=%q", mkdirPath, mkdirPerm, openPath, openFlag, openPerm, written.String())
	}
}

func TestFacadeSaveOptions(t *testing.T) {
	if vhttp.WithSaveCreateParents(true) == nil {
		t.Fatal("WithSaveCreateParents returned nil")
	}
	if vhttp.WithSaveStat(func(s string) (os.FileInfo, error) { return nil, nil }) == nil {
		t.Fatal("WithSaveStat returned nil")
	}
	if vhttp.WithSaveOverwrite(true) == nil {
		t.Fatal("WithSaveOverwrite returned nil")
	}
}

func TestFacadeDownloadFileSafe(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Mode") != "safe" {
			http.Error(w, "missing option header", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte("vhttp-safe-file"))
	}))
	defer server.Close()

	dir := t.TempDir()
	n, err := vhttp.DownloadFileSafeWithOptions(server.URL, dir,
		[]vhttp.RequestOption{
			vhttp.WithHeader("X-Mode", "safe"),
			vhttp.WithURLPolicy(vhttp.URLPolicy{AllowedSchemes: []string{"http", "https"}, RejectPrivate: false}),
		},
		vhttp.WithSaveDefaultFilename("safe.txt"),
	)
	if err != nil {
		t.Fatalf("DownloadFileSafeWithOptions() error = %v", err)
	}
	if n != int64(len("vhttp-safe-file")) {
		t.Fatalf("DownloadFileSafeWithOptions() n = %d", n)
	}
	data, err := os.ReadFile(filepath.Join(dir, "safe.txt"))
	if err != nil {
		t.Fatalf("read saved file: %v", err)
	}
	if string(data) != "vhttp-safe-file" {
		t.Fatalf("saved file = %q", data)
	}
}
