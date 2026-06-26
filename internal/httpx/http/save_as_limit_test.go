package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

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
