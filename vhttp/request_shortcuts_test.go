package vhttp_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/imajinyun/knifer-go/vhttp"
)

func TestFacadeShortcutRequestHelpers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		switch r.URL.Path {
		case "/get":
			_, _ = w.Write([]byte(r.Method + ":" + r.URL.Query().Get("q")))
		case "/form":
			_, _ = w.Write([]byte(r.Method + ":" + string(body) + ":" + r.Header.Get("X-Shortcut")))
		case "/json":
			_, _ = w.Write([]byte(r.Method + ":" + r.Header.Get("Content-Type") + ":" + string(body)))
		case "/string":
			_, _ = w.Write([]byte(r.Method + ":" + string(body)))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	got, err := vhttp.GetStringE(server.URL + "/get")
	if err != nil || got != "GET:" {
		t.Fatalf("GetStringE = %q, %v", got, err)
	}
	got, err = vhttp.GetStringEWithOptions(server.URL+"/get?q=go", vhttp.WithHeader("X-Shortcut", "yes"))
	if err != nil || got != "GET:go" {
		t.Fatalf("GetStringEWithOptions = %q, %v", got, err)
	}
	got, err = vhttp.GetWithTimeoutE(server.URL+"/get", time.Second)
	if err != nil || got != "GET:" {
		t.Fatalf("GetWithTimeoutE = %q, %v", got, err)
	}
	got, err = vhttp.GetWithTimeoutEWithOptions(server.URL+"/get?q=timeout", time.Second, vhttp.WithHeader("X-Shortcut", "yes"))
	if err != nil || got != "GET:timeout" {
		t.Fatalf("GetWithTimeoutEWithOptions = %q, %v", got, err)
	}
	got, err = vhttp.GetWithParamsE(server.URL+"/get", map[string]any{"q": "params"})
	if err != nil || got != "GET:params" {
		t.Fatalf("GetWithParamsE = %q, %v", got, err)
	}
	got, err = vhttp.GetWithParamsEWithOptions(server.URL+"/get", map[string]any{"q": "params2"}, vhttp.WithHeader("X-Shortcut", "yes"))
	if err != nil || got != "GET:params2" {
		t.Fatalf("GetWithParamsEWithOptions = %q, %v", got, err)
	}
	got, err = vhttp.PostFormE(server.URL+"/form", map[string]any{"name": "alice"})
	if err != nil || !strings.Contains(got, "POST:") || !strings.Contains(got, "name=alice") {
		t.Fatalf("PostFormE = %q, %v", got, err)
	}
	got, err = vhttp.PostFormEWithOptions(server.URL+"/form", map[string]any{"name": "bob"}, vhttp.WithHeader("X-Shortcut", "hdr"))
	if err != nil || !strings.Contains(got, "name=bob") || !strings.HasSuffix(got, ":hdr") {
		t.Fatalf("PostFormEWithOptions = %q, %v", got, err)
	}
	got, err = vhttp.PostJSONE(server.URL+"/json", `{"name":"json"}`)
	if err != nil || !strings.Contains(got, `{"name":"json"}`) || !strings.Contains(got, "application/json") {
		t.Fatalf("PostJSONE = %q, %v", got, err)
	}
	got, err = vhttp.PostJSONEWithOptions(server.URL+"/json", `{"name":"json2"}`, vhttp.WithHeader("X-Shortcut", "hdr"))
	if err != nil || !strings.Contains(got, `{"name":"json2"}`) {
		t.Fatalf("PostJSONEWithOptions = %q, %v", got, err)
	}
	got, err = vhttp.PostStringE(server.URL+"/string", "plain")
	if err != nil || got != "POST:plain" {
		t.Fatalf("PostStringE = %q, %v", got, err)
	}
	got, err = vhttp.PostStringEWithOptions(server.URL+"/string", "plain2", vhttp.WithHeader("X-Shortcut", "hdr"))
	if err != nil || got != "POST:plain2" {
		t.Fatalf("PostStringEWithOptions = %q, %v", got, err)
	}
}
