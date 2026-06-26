package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
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
