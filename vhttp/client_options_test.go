package vhttp_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/imajinyun/knifer-go/vhttp"
)

func TestFacadeClientOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Method + ":" + r.Header.Get("X-Client-Default")))
	}))
	defer server.Close()

	client := vhttp.NewClient(vhttp.WithClientRequestOptions(vhttp.WithHeader("X-Client-Default", "shared")))
	if got := client.Get(server.URL).Execute().Body(); got != "GET:shared" {
		t.Fatalf("client.Get body = %q", got)
	}
	if got := client.Post(server.URL).Execute().Body(); got != "POST:shared" {
		t.Fatalf("client.Post body = %q", got)
	}
	if got := client.NewRequest(vhttp.MethodPut, server.URL).Execute().Body(); got != "PUT:shared" {
		t.Fatalf("client.NewRequest body = %q", got)
	}

	cfg := vhttp.SnapshotGlobalConfig()
	cfg.Headers.Set("X-Client-Default", "configured")
	if got := vhttp.NewClientWithConfig(cfg).Get(server.URL).Execute().Body(); got != "GET:configured" {
		t.Fatalf("NewClientWithConfig body = %q", got)
	}
	isolated := vhttp.NewIsolatedClient(vhttp.WithClientGlobalConfig(cfg))
	if got := isolated.Get(server.URL).Execute().Body(); got != "GET:configured" {
		t.Fatalf("NewIsolatedClient body = %q", got)
	}
	if resp := client.GetSafe(server.URL).Execute(); resp.Err() == nil {
		t.Fatal("client.GetSafe(localhost default policy) error = nil")
	}
	if resp := client.PostSafe(server.URL, vhttp.WithURLPolicy(vhttp.URLPolicy{AllowedSchemes: []string{"http", "https"}, RejectPrivate: false})).Execute(); resp.Err() != nil {
		t.Fatalf("client.PostSafe allowed error = %v", resp.Err())
	}
}
