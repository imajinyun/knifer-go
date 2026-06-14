package conf

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLoadRemoteWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Config-Token") != "secret" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte("app:\n  name: remote"))
	}))
	defer server.Close()
	calledFactory := false
	c, err := LoadRemoteWithOptions(server.URL+"/app.yaml", LoadOptions{
		Timeout: time.Second,
		Headers: http.Header{"X-Config-Token": []string{"secret"}},
		RequestFactory: func(ctx context.Context, rawURL string) (*http.Request, error) {
			calledFactory = true
			return http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !calledFactory {
		t.Fatal("request factory was not called")
	}
	if got := c.GetByGroup("app", "name"); got != "remote" {
		t.Fatalf("remote app.name = %q", got)
	}
	if _, err := LoadRemoteWithOptions(server.URL+"/app.yaml", LoadOptions{MaxBytes: 3, Headers: http.Header{"X-Config-Token": []string{"secret"}}}); err == nil {
		t.Fatal("LoadRemoteWithOptions max bytes error = nil")
	}
}
