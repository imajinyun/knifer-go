package http

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAsOptions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("saved"))
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
	missing := filepath.Join(dir, "missing", "out.txt")
	if _, err := Get(srv.URL).Execute().SaveAs(missing, WithSaveCreateParents(false)); err == nil {
		t.Fatal("SaveAs without parent creation should fail")
	}
	if _, err := DownloadFile(srv.URL, target); err != nil {
		t.Fatalf("DownloadFile overwrite default: %v", err)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(data) != "saved" {
		t.Fatalf("content = %q", data)
	}
}
