package http

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAsUsesCachedBodyAfterBodyRead(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("cached"))
	}))
	defer srv.Close()

	resp := Get(srv.URL).Execute()
	if got := resp.Body(); got != "cached" {
		t.Fatalf("Body() = %q, want cached", got)
	}
	target := filepath.Join(t.TempDir(), "cached.txt")
	if _, err := resp.SaveAs(target); err != nil {
		t.Fatalf("SaveAs() after Body() error = %v", err)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "cached" {
		t.Fatalf("saved content = %q, want cached", data)
	}
}
