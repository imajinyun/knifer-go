package vhttp_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/imajinyun/go-knifer/vhttp"
)

func TestFacadeUsesNamesWithoutHTTPPrefix(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != string(vhttp.MethodGet) {
			t.Fatalf("method = %q, want GET", r.Method)
		}
		if got := r.URL.Query().Get("lang"); got != "go" {
			t.Fatalf("query lang = %q, want go", got)
		}
		if got := r.Header.Get("X-Client"); got != "go-knifer" {
			t.Fatalf("header X-Client = %q, want go-knifer", got)
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	req := vhttp.Get(server.URL).
		Query("lang", "go").
		Header("X-Client", "go-knifer")

	resp := executeRequest(req)
	if resp.Err() != nil {
		t.Fatal(resp.Err())
	}
	if got := resp.Body(); got != "ok" {
		t.Fatalf("Body() = %q, want ok", got)
	}
}

func TestFacadeRequestFollowRedirectOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Opt") + ":" + r.Header.Get("User-Agent")))
	}))
	defer server.Close()

	resp := vhttp.Get(
		server.URL,
		vhttp.WithHeader("X-Opt", "yes"),
		vhttp.WithUserAgent("vhttp-test/1.0"),
	).Execute()
	if resp.Err() != nil {
		t.Fatal(resp.Err())
	}
	if got := resp.Body(); got != "yes:vhttp-test/1.0" {
		t.Fatalf("Body() = %q, want option headers", got)
	}
}

func TestFacadeRequestCloneAndSingleUse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			b, _ := io.ReadAll(r.Body)
			_, _ = w.Write(b)
			return
		}
		_, _ = w.Write([]byte(r.URL.Query().Get("q") + ":" + r.Header.Get("X-Token")))
	}))
	defer server.Close()

	base := vhttp.Get(server.URL).Query("q", "base").Header("X-Token", "base")
	clone := base.Clone().Header("X-Token", "clone")
	if got := base.Execute().Body(); got != "base:base" {
		t.Fatalf("base Body() = %q", got)
	}
	if got := clone.Execute().Body(); got != "base:clone" {
		t.Fatalf("clone Body() = %q", got)
	}

	req := vhttp.Post(server.URL).BodyReader(strings.NewReader("payload"))
	if got := req.Execute().Body(); got != "payload" {
		t.Fatalf("first reader body = %q", got)
	}
	if resp := req.Execute(); resp.Err() == nil {
		t.Fatal("second Execute() should reject reader-backed body reuse")
	}
}

func TestFacadeRequestOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/final", http.StatusFound)
			return
		}
		_, _ = w.Write([]byte(r.Method + ":" + r.Header.Get("X-Create")))
	}))
	defer server.Close()

	getResp := vhttp.Get(server.URL+"/redirect", vhttp.WithFollowRedirects(false), vhttp.WithHeader("X-Create", "get")).Execute()
	if getResp.Err() != nil {
		t.Fatal(getResp.Err())
	}
	if got := getResp.Status(); got != http.StatusFound {
		t.Fatalf("Get status = %d, want 302", got)
	}

	postResp := vhttp.Post(server.URL, vhttp.WithHeader("X-Create", "post")).Execute()
	if postResp.Err() != nil {
		t.Fatal(postResp.Err())
	}
	if got := postResp.Body(); got != "POST:post" {
		t.Fatalf("Post body = %q, want POST:post", got)
	}
}

func TestFacadeAdditionalHTTPMethodsAndGlobalAccessors(t *testing.T) {
	var lastMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastMethod = r.Method
		w.Header().Set("X-Method", r.Method)
		if r.Method != http.MethodHead {
			_, _ = w.Write([]byte(r.Method))
		}
	}))
	defer server.Close()

	tests := []struct {
		name   string
		method string
		req    *vhttp.Request
	}{
		{name: "put", method: http.MethodPut, req: vhttp.Put(server.URL)},
		{name: "delete", method: http.MethodDelete, req: vhttp.Delete(server.URL)},
		{name: "patch", method: http.MethodPatch, req: vhttp.Patch(server.URL)},
		{name: "head", method: http.MethodHead, req: vhttp.Head(server.URL)},
		{name: "options", method: http.MethodOptions, req: vhttp.Options(server.URL)},
		{name: "new request", method: http.MethodTrace, req: vhttp.NewRequest(vhttp.MethodTrace, server.URL)},
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

	previous := vhttp.SnapshotGlobalConfig()
	defer vhttp.ConfigureGlobalConfig(previous)
	vhttp.SetGlobalMaxRedirects(3)
	vhttp.SetGlobalMaxResponseBytes(99)
	vhttp.SetGlobalFollowRedirects(false)
	vhttp.SetGlobalUserAgent("vhttp-extra/1.0")
	vhttp.SetIgnoreEOFError(true)
	vhttp.SetGlobalBoundary("boundary-extra")
	vhttp.SetGlobalDecodeURL(true)
	if vhttp.GetGlobalMaxRedirects() != 3 || vhttp.GetGlobalMaxResponseBytes() != 99 || vhttp.GetGlobalFollowRedirects() || vhttp.GetGlobalUserAgent() != "vhttp-extra/1.0" || !vhttp.IsIgnoreEOFError() || vhttp.GetGlobalBoundary() != "boundary-extra" || !vhttp.IsGlobalDecodeURL() {
		t.Fatalf("global accessors snapshot = %#v", vhttp.SnapshotGlobalConfig())
	}
}
