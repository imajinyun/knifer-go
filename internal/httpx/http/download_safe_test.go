package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadFileSafeRejectsPrivateHost(t *testing.T) {
	target := filepath.Join(t.TempDir(), "blocked.txt")
	if _, err := DownloadFileSafe("http://127.0.0.1/config.yaml", target, WithSaveDefaultFilename("blocked.txt")); err == nil {
		t.Fatal("DownloadFileSafe should reject private hosts")
	}
	if _, err := os.Stat(target); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("blocked safe download should not create destination, stat err=%v", err)
	}
}

func TestDownloadFileSafeWithOptions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Mode") != "safe-file" {
			http.Error(w, "missing option header", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte("safe-file-options"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	n, err := DownloadFileSafeWithOptions(srv.URL, dir,
		[]RequestOption{
			WithHeader("X-Mode", "safe-file"),
			WithURLPolicy(URLPolicy{AllowedSchemes: []string{"http", "https"}, RejectPrivate: false}),
		},
		WithSaveDefaultFilename("safe.txt"),
	)
	if err != nil {
		t.Fatalf("DownloadFileSafeWithOptions() error = %v", err)
	}
	if n != int64(len("safe-file-options")) {
		t.Fatalf("size: %d", n)
	}
	data, err := os.ReadFile(filepath.Join(dir, "safe.txt"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(data) != "safe-file-options" {
		t.Fatalf("content: %q", string(data))
	}
}
