package vresty_test

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/imajinyun/go-knifer/vresty"
)

func TestFacadeGetString(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("facade"))
	}))
	defer srv.Close()

	got, err := vresty.GetStringE(srv.URL)
	if err != nil {
		t.Fatalf("GetStringE() error = %v", err)
	}
	if got != "facade" {
		t.Fatalf("GetStringE() = %q, want facade", got)
	}
}

func TestFacadeRequestFollowRedirectOptions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Opt") + ":" + r.Header.Get("User-Agent")))
	}))
	defer srv.Close()

	resp := vresty.Get(srv.URL,
		vresty.WithHeader("X-Opt", "yes"),
		vresty.WithUserAgent("vresty-test/1.0"),
	).Execute()
	if resp.Err() != nil {
		t.Fatal(resp.Err())
	}
	if got := resp.Body(); got != "yes:vresty-test/1.0" {
		t.Fatalf("Body() = %q, want option headers", got)
	}
}

func TestFacadeRequestOptions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/final", http.StatusFound)
			return
		}
		_, _ = w.Write([]byte(r.Method + ":" + r.Header.Get("X-Create")))
	}))
	defer srv.Close()

	getResp := vresty.Get(srv.URL+"/redirect", vresty.WithFollowRedirects(false), vresty.WithHeader("X-Create", "get")).Execute()
	if getResp.Err() != nil {
		t.Fatal(getResp.Err())
	}
	if got := getResp.Status(); got != http.StatusFound {
		t.Fatalf("Get status = %d, want 302", got)
	}

	postResp := vresty.Post(srv.URL, vresty.WithHeader("X-Create", "post")).Execute()
	if postResp.Err() != nil {
		t.Fatal(postResp.Err())
	}
	if got := postResp.Body(); got != "POST:post" {
		t.Fatalf("Post body = %q, want POST:post", got)
	}
}

func TestFacadeAdditionalMethodsGlobalAndContentHelpers(t *testing.T) {
	var lastMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastMethod = r.Method
		w.Header().Set("X-Method", r.Method)
		if r.Method != http.MethodHead {
			_, _ = w.Write([]byte(r.Method))
		}
	}))
	defer srv.Close()

	tests := []struct {
		name   string
		method string
		req    *vresty.Request
	}{
		{name: "put", method: http.MethodPut, req: vresty.Put(srv.URL)},
		{name: "delete", method: http.MethodDelete, req: vresty.Delete(srv.URL)},
		{name: "patch", method: http.MethodPatch, req: vresty.Patch(srv.URL)},
		{name: "head", method: http.MethodHead, req: vresty.Head(srv.URL)},
		{name: "options", method: http.MethodOptions, req: vresty.Options(srv.URL)},
		{name: "new request", method: http.MethodTrace, req: vresty.NewRequest(vresty.MethodTrace, srv.URL)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := tt.req.Execute()
			if resp.Err() != nil {
				t.Fatalf("Execute: %v", resp.Err())
			}
			if lastMethod != tt.method {
				t.Fatalf("server method = %q, want %q", lastMethod, tt.method)
			}
			if got := resp.Header("X-Method"); got != tt.method {
				t.Fatalf("response method header = %q, want %q", got, tt.method)
			}
		})
	}

	previous := vresty.SnapshotGlobalConfig()
	defer vresty.ConfigureGlobalConfig(previous)
	vresty.SetGlobalMaxRedirects(4)
	vresty.SetGlobalMaxResponseBytes(123)
	vresty.SetGlobalFollowRedirects(false)
	vresty.SetGlobalUserAgent("vresty-extra/1.0")
	vresty.CloseCookie()
	vresty.SetGlobalHeader("X-Extra", "one")
	vresty.RemoveGlobalHeader("X-Extra")
	cfg := vresty.SnapshotGlobalConfig()
	if vresty.GetGlobalMaxRedirects() != 4 || vresty.GetGlobalMaxResponseBytes() != 123 || vresty.GetGlobalFollowRedirects() || vresty.GetGlobalUserAgent() != "vresty-extra/1.0" || !cfg.CookieDisabled || len(cfg.Headers["X-Extra"]) != 0 {
		t.Fatalf("global config = %#v", cfg)
	}

	if got := vresty.BuildContentType("application/json", "utf-8"); got != "application/json;charset=utf-8" {
		t.Fatalf("BuildContentType = %q", got)
	}
	if got := vresty.GuessContentType("<root/>"); got != vresty.ContentTypeXML {
		t.Fatalf("GuessContentType = %q", got)
	}
	if !vresty.IsDefaultContentType("") || !vresty.IsFormURLEncoded("application/x-www-form-urlencoded; charset=utf-8") {
		t.Fatal("content type predicates returned unexpected result")
	}
	if got := vresty.URLWithForm("https://example.com/path?x=1", map[string]any{"q": "go"}); got != "https://example.com/path?x=1&q=go" {
		t.Fatalf("URLWithForm = %q", got)
	}
	if got := vresty.GetCharsetFromContentTypeWithOptions("text/plain; enc=gbk", vresty.WithCharsetRegexp(regexp.MustCompile(`enc=([a-z0-9-]+)`))); got != "gbk" {
		t.Fatalf("GetCharsetFromContentTypeWithOptions = %q", got)
	}
	if got := vresty.GetCharsetFromHTMLWithOptions(`<meta data-charset="big5">`, vresty.WithMetaCharsetRegexp(regexp.MustCompile(`data-charset="([^"]+)"`))); got != "big5" {
		t.Fatalf("GetCharsetFromHTMLWithOptions = %q", got)
	}
	if got := vresty.GetMimeType("payload.zip"); got != "application/zip" {
		t.Fatalf("GetMimeType = %q", got)
	}
}

func TestFacadeRequestGlobalConfigAPIs(t *testing.T) {
	previous := vresty.SnapshotGlobalConfig()
	defer vresty.ConfigureGlobalConfig(previous)

	vresty.SetGlobalTimeout(321 * time.Millisecond)
	vresty.SetGlobalHeader("X-Facade-Config", "global")

	cfg := vresty.SnapshotGlobalConfig()
	cfg.Headers["X-Facade-Config"][0] = "snapshot"
	cfg.DefaultUserAgent = "facade-config-agent"
	cfg.Headers["User-Agent"] = []string{"facade-config-agent"}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Facade-Config") + ":" + r.Header.Get("User-Agent")))
	}))
	defer srv.Close()

	resp := vresty.NewRequestWithConfig(vresty.MethodGet, srv.URL, cfg).Execute()
	if resp.Err() != nil {
		t.Fatal(resp.Err())
	}
	if got := resp.Body(); got != "snapshot:facade-config-agent" {
		t.Fatalf("NewRequestWithConfig body = %q", got)
	}

	resp = vresty.NewIsolatedRequest(vresty.MethodGet, srv.URL, vresty.WithGlobalConfig(cfg)).Execute()
	if resp.Err() != nil {
		t.Fatal(resp.Err())
	}
	if got := resp.Body(); got != "snapshot:facade-config-agent" {
		t.Fatalf("NewIsolatedRequest WithGlobalConfig body = %q", got)
	}
}

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
