package vresty_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/imajinyun/knifer-go/vresty"
)

func TestFacadeClientAndSafeRequestWrappers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Method", r.Method)
		if r.Method != http.MethodHead {
			_, _ = w.Write([]byte(r.Method + ":" + r.Header.Get("X-Client-Default")))
		}
	}))
	defer server.Close()

	client := vresty.NewClient(vresty.WithClientRequestOptions(vresty.WithHeader("X-Client-Default", "shared")))
	if got := client.Get(server.URL).Execute().Body(); got != "GET:shared" {
		t.Fatalf("client.Get body = %q", got)
	}
	if got := client.Post(server.URL).Execute().Body(); got != "POST:shared" {
		t.Fatalf("client.Post body = %q", got)
	}
	if got := client.NewRequest(vresty.MethodPut, server.URL).Execute().Body(); got != "PUT:shared" {
		t.Fatalf("client.NewRequest body = %q", got)
	}

	cfg := vresty.SnapshotGlobalConfig()
	cfg.Headers["X-Client-Default"] = []string{"configured"}
	if got := vresty.NewClientWithConfig(cfg).Get(server.URL).Execute().Body(); got != "GET:configured" {
		t.Fatalf("NewClientWithConfig body = %q", got)
	}
	if got := vresty.NewIsolatedClient(vresty.WithClientGlobalConfig(cfg)).Get(server.URL).Execute().Body(); got != "GET:configured" {
		t.Fatalf("NewIsolatedClient body = %q", got)
	}

	if resp := vresty.GetSafe(server.URL).Execute(); resp.Err() == nil {
		t.Fatal("GetSafe(localhost default policy) error = nil")
	}
	allowLocal := vresty.WithURLPolicy(vresty.URLPolicy{AllowedSchemes: []string{"http", "https"}, RejectPrivate: false})
	tests := []struct {
		name   string
		method string
		req    *vresty.Request
	}{
		{name: "post safe", method: http.MethodPost, req: vresty.PostSafe(server.URL, allowLocal)},
		{name: "put safe", method: http.MethodPut, req: vresty.PutSafe(server.URL, allowLocal)},
		{name: "delete safe", method: http.MethodDelete, req: vresty.DeleteSafe(server.URL, allowLocal)},
		{name: "patch safe", method: http.MethodPatch, req: vresty.PatchSafe(server.URL, allowLocal)},
		{name: "head safe", method: http.MethodHead, req: vresty.HeadSafe(server.URL, allowLocal)},
		{name: "options safe", method: http.MethodOptions, req: vresty.OptionsSafe(server.URL, allowLocal)},
		{name: "new safe", method: http.MethodTrace, req: vresty.NewSafeRequest(vresty.MethodTrace, server.URL, allowLocal)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := tt.req.Execute()
			if resp.Err() != nil {
				t.Fatalf("Execute: %v", resp.Err())
			}
			if got := resp.Header("X-Method"); got != tt.method {
				t.Fatalf("method header = %q, want %q", got, tt.method)
			}
		})
	}
}
