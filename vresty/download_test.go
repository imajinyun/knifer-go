package vresty_test

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

	"github.com/imajinyun/knifer-go/vresty"
)

func TestFacadeSafeShortcutAndDownloadHelpers(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			body, _ := io.ReadAll(r.Body)
			_, _ = w.Write([]byte(r.Method + ":" + string(body)))
			return
		}
		_, _ = w.Write([]byte("download"))
	}))
	defer srv.Close()

	allowLocal := vresty.WithURLPolicy(vresty.URLPolicy{AllowedSchemes: []string{"http", "https"}, RejectPrivate: false})
	if _, err := vresty.GetStringSafeE(srv.URL); err == nil {
		t.Fatal("GetStringSafeE(localhost default policy) error = nil")
	}
	if got, err := vresty.GetStringSafeE(srv.URL, allowLocal); err != nil || got != "download" {
		t.Fatalf("GetStringSafeE allowed = %q, %v", got, err)
	}
	if got, err := vresty.PostStringSafeE(srv.URL, "body", allowLocal); err != nil || got != "POST:body" {
		t.Fatalf("PostStringSafeE allowed = %q, %v", got, err)
	}
	if got, err := vresty.DownloadBytesSafeE(srv.URL, allowLocal); err != nil || string(got) != "download" {
		t.Fatalf("DownloadBytesSafeE allowed = %q, %v", got, err)
	}
	var buf bytes.Buffer
	if n, err := vresty.Download(srv.URL, &buf); err != nil || n != int64(len("download")) || buf.String() != "download" {
		t.Fatalf("Download n=%d body=%q err=%v", n, buf.String(), err)
	}
}

func TestFacadeDownloadAndFileWrappers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("download-text"))
	}))
	defer server.Close()

	allowLocal := vresty.WithURLPolicy(vresty.URLPolicy{AllowedSchemes: []string{"http", "https"}, RejectPrivate: false})
	var buf bytes.Buffer
	if n, err := vresty.DownloadSafe(server.URL, &buf, allowLocal); err != nil || n != int64(len("download-text")) || buf.String() != "download-text" {
		t.Fatalf("DownloadSafe n=%d body=%q err=%v", n, buf.String(), err)
	}
	if b, err := vresty.DownloadBytesE(server.URL); err != nil || string(b) != "download-text" {
		t.Fatalf("DownloadBytesE = %q, %v", b, err)
	}
	if b, err := vresty.DownloadBytesEWithOptions(server.URL, vresty.WithMaxResponseBytes(64)); err != nil || string(b) != "download-text" {
		t.Fatalf("DownloadBytesEWithOptions = %q, %v", b, err)
	}
	if got, err := vresty.DownloadStringE(server.URL, ""); err != nil || got != "download-text" {
		t.Fatalf("DownloadStringE = %q, %v", got, err)
	}
	if got, err := vresty.DownloadStringEWithOptions(server.URL, "", vresty.WithMaxResponseBytes(64)); err != nil || got != "download-text" {
		t.Fatalf("DownloadStringEWithOptions = %q, %v", got, err)
	}
	if got, err := vresty.DownloadStringSafeE(server.URL, "", allowLocal); err != nil || got != "download-text" {
		t.Fatalf("DownloadStringSafeE = %q, %v", got, err)
	}

	dir := t.TempDir()
	file := filepath.Join(dir, "plain.txt")
	if n, err := vresty.DownloadFile(server.URL, file); err != nil || n != int64(len("download-text")) {
		t.Fatalf("DownloadFile n=%d err=%v", n, err)
	}
	if data, err := os.ReadFile(file); err != nil || string(data) != "download-text" {
		t.Fatalf("DownloadFile content = %q, %v", data, err)
	}
	fileWithOpts := filepath.Join(dir, "with-options.txt")
	if n, err := vresty.DownloadFileWithOptions(server.URL, fileWithOpts, []vresty.RequestOption{vresty.WithMaxResponseBytes(64)}, vresty.WithSaveOverwrite(true)); err != nil || n != int64(len("download-text")) {
		t.Fatalf("DownloadFileWithOptions n=%d err=%v", n, err)
	}
	safeFile := filepath.Join(dir, "safe.txt")
	if n, err := vresty.DownloadFileSafe(server.URL, safeFile, vresty.WithSaveOverwrite(true)); err == nil || n != 0 {
		t.Fatalf("DownloadFileSafe default policy n=%d err=%v, want private host rejection", n, err)
	}
	if n, err := vresty.DownloadFileSafeWithOptions(server.URL, safeFile, []vresty.RequestOption{allowLocal}, vresty.WithSaveOverwrite(true)); err != nil || n != int64(len("download-text")) {
		t.Fatalf("DownloadFileSafeWithOptions n=%d err=%v", n, err)
	}
}

func TestFacadeDownloadFileReturnsCloseError(t *testing.T) {
	closeErr := errors.New("close failed")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("download-text"))
	}))
	defer server.Close()

	n, err := vresty.DownloadFile(server.URL, "/virtual/resty.txt",
		vresty.WithSaveMkdirAll(func(string, fs.FileMode) error { return nil }),
		vresty.WithSaveOpenFile(func(string, int, fs.FileMode) (io.WriteCloser, error) {
			return closeErrorWriteCloser{Writer: io.Discard, err: closeErr}, nil
		}),
	)
	if n != int64(len("download-text")) {
		t.Fatalf("DownloadFile close error bytes = %d, want %d", n, len("download-text"))
	}
	if !errors.Is(err, closeErr) {
		t.Fatalf("DownloadFile close error = %v, want close cause", err)
	}
}
