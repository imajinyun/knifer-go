package resty

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestAdditionalDownloadWrappers(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("download"))
	}))
	defer srv.Close()
	allowLocal := WithURLPolicy(URLPolicy{AllowedSchemes: []string{"http", "https"}, RejectPrivate: false})

	if got, err := DownloadStringE(srv.URL, ""); err != nil || got != "download" {
		t.Fatalf("DownloadStringE = %q, %v", got, err)
	}
	if got, err := DownloadStringEWithOptions(srv.URL, "", WithMaxResponseBytes(64)); err != nil || got != "download" {
		t.Fatalf("DownloadStringEWithOptions = %q, %v", got, err)
	}
	if got, err := DownloadStringSafeE(srv.URL, "", allowLocal); err != nil || got != "download" {
		t.Fatalf("DownloadStringSafeE = %q, %v", got, err)
	}
	var buf bytes.Buffer
	if n, err := Download(srv.URL, &buf); err != nil || n != int64(len("download")) || buf.String() != "download" {
		t.Fatalf("Download n=%d body=%q err=%v", n, buf.String(), err)
	}
	buf.Reset()
	if n, err := DownloadWithOptions(srv.URL, &buf, WithMaxResponseBytes(64)); err != nil || n != int64(len("download")) || buf.String() != "download" {
		t.Fatalf("DownloadWithOptions n=%d body=%q err=%v", n, buf.String(), err)
	}
	buf.Reset()
	if n, err := DownloadSafe(srv.URL, &buf, allowLocal); err != nil || n != int64(len("download")) || buf.String() != "download" {
		t.Fatalf("DownloadSafe n=%d body=%q err=%v", n, buf.String(), err)
	}
	if b, err := DownloadBytesSafeE(srv.URL, allowLocal); err != nil || string(b) != "download" {
		t.Fatalf("DownloadBytesSafeE = %q, %v", b, err)
	}
	dir := t.TempDir()
	if n, err := DownloadFileSafeWithOptions(srv.URL, filepath.Join(dir, "safe.txt"), []RequestOption{allowLocal}, WithSaveOverwrite(true)); err != nil || n != int64(len("download")) {
		t.Fatalf("DownloadFileSafeWithOptions n=%d err=%v", n, err)
	}
	if _, err := DownloadFileSafe(srv.URL, filepath.Join(dir, "blocked.txt")); err == nil {
		t.Fatal("DownloadFileSafe default policy error = nil")
	}
}

func TestDownloadFileReturnsCloseError(t *testing.T) {
	closeErr := errors.New("close failed")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("resty-close"))
	}))
	defer srv.Close()

	n, err := DownloadFile(srv.URL, "/virtual/resty.txt",
		WithSaveMkdirAll(func(string, fs.FileMode) error { return nil }),
		WithSaveOpenFile(func(string, int, fs.FileMode) (io.WriteCloser, error) {
			return closeErrorWriteCloser{Writer: io.Discard, err: closeErr}, nil
		}),
	)
	if n != int64(len("resty-close")) {
		t.Fatalf("DownloadFile close error bytes = %d, want %d", n, len("resty-close"))
	}
	if !errors.Is(err, closeErr) {
		t.Fatalf("DownloadFile close error = %v, want close cause", err)
	}
}
