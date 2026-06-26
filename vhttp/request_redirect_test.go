package vhttp_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/imajinyun/knifer-go/vhttp"
)

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

func TestFacadeRequestCreationOptions(t *testing.T) {
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
