package vconf_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/imajinyun/go-knifer/vconf"
)

func TestLoadFilesAndRemoteWithOptionsFacade(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "base.setting")
	main := filepath.Join(dir, "main.setting")
	if err := os.WriteFile(base, []byte("name=base\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(main, []byte("name=main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	merged, err := vconf.LoadFilesWithOptions(vconf.LoadOptions{MaxBytes: 64}, base, main)
	if err != nil {
		t.Fatalf("LoadFilesWithOptions() error = %v", err)
	}
	if got := merged.Get("name"); got != "main" {
		t.Fatalf("LoadFilesWithOptions() name = %q, want main", got)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Config-Token"); got != "secret" {
			t.Fatalf("remote header X-Config-Token = %q, want secret", got)
		}
		_, _ = w.Write([]byte("remote: true\n"))
	}))
	defer server.Close()
	remote, err := vconf.LoadRemoteWithOptions(server.URL+"/app.yaml", vconf.LoadOptions{
		Headers:  http.Header{"X-Config-Token": []string{"secret"}},
		Timeout:  time.Second,
		MaxBytes: 64,
	})
	if err != nil {
		t.Fatalf("LoadRemoteWithOptions() error = %v", err)
	}
	if got := remote.Get("remote"); got != "true" {
		t.Fatalf("LoadRemoteWithOptions() remote = %q, want true", got)
	}
}
