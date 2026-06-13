package vhttp_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/imajinyun/go-knifer/vhttp"
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

func TestFacadeSafeShortcutHelpers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if r.Method == http.MethodPost {
			_, _ = w.Write([]byte(r.Method + ":" + string(body)))
			return
		}
		_, _ = w.Write([]byte(r.Method))
	}))
	defer server.Close()

	allowLocal := vhttp.WithURLPolicy(vhttp.URLPolicy{AllowedSchemes: []string{"http", "https"}, RejectPrivate: false})
	if got, err := vhttp.GetStringSafeE(server.URL, allowLocal); err != nil || got != "GET" {
		t.Fatalf("GetStringSafeE allowed = %q, %v", got, err)
	}
	if _, err := vhttp.GetStringSafeE(server.URL); err == nil {
		t.Fatal("GetStringSafeE(localhost default policy) error = nil")
	}
	if got, err := vhttp.PostFormSafeE(server.URL, map[string]any{"name": "safe"}, allowLocal); err != nil || got != "POST:name=safe" {
		t.Fatalf("PostFormSafeE = %q, %v", got, err)
	}
	if got, err := vhttp.PostJSONSafeE(server.URL, `{"safe":true}`, allowLocal); err != nil || got != `POST:{"safe":true}` {
		t.Fatalf("PostJSONSafeE = %q, %v", got, err)
	}
	if got, err := vhttp.PostStringSafeE(server.URL, "safe-string", allowLocal); err != nil || got != "POST:safe-string" {
		t.Fatalf("PostStringSafeE = %q, %v", got, err)
	}
}

func TestFacadeSafeRequestsAndDownloadWrappers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Method", r.Method)
		if r.Method != http.MethodHead {
			_, _ = w.Write([]byte("download-text"))
		}
	}))
	defer server.Close()

	allowLocal := vhttp.WithURLPolicy(vhttp.URLPolicy{AllowedSchemes: []string{"http", "https"}, RejectPrivate: false})
	tests := []struct {
		name   string
		method string
		req    *vhttp.Request
	}{
		{name: "get safe", method: http.MethodGet, req: vhttp.GetSafe(server.URL, allowLocal)},
		{name: "post safe", method: http.MethodPost, req: vhttp.PostSafe(server.URL, allowLocal)},
		{name: "put safe", method: http.MethodPut, req: vhttp.PutSafe(server.URL, allowLocal)},
		{name: "delete safe", method: http.MethodDelete, req: vhttp.DeleteSafe(server.URL, allowLocal)},
		{name: "patch safe", method: http.MethodPatch, req: vhttp.PatchSafe(server.URL, allowLocal)},
		{name: "head safe", method: http.MethodHead, req: vhttp.HeadSafe(server.URL, allowLocal)},
		{name: "options safe", method: http.MethodOptions, req: vhttp.OptionsSafe(server.URL, allowLocal)},
		{name: "new safe", method: http.MethodTrace, req: vhttp.NewSafeRequest(vhttp.MethodTrace, server.URL, allowLocal)},
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

	var buf bytes.Buffer
	if n, err := vhttp.Download(server.URL, &buf); err != nil || n != int64(len("download-text")) || buf.String() != "download-text" {
		t.Fatalf("Download n=%d body=%q err=%v", n, buf.String(), err)
	}
	buf.Reset()
	if n, err := vhttp.DownloadWithOptions(server.URL, &buf, vhttp.WithMaxResponseBytes(64)); err != nil || n != int64(len("download-text")) || buf.String() != "download-text" {
		t.Fatalf("DownloadWithOptions n=%d body=%q err=%v", n, buf.String(), err)
	}
	buf.Reset()
	if n, err := vhttp.DownloadSafe(server.URL, &buf, allowLocal); err != nil || n != int64(len("download-text")) || buf.String() != "download-text" {
		t.Fatalf("DownloadSafe n=%d body=%q err=%v", n, buf.String(), err)
	}
	if b, err := vhttp.DownloadBytesE(server.URL); err != nil || string(b) != "download-text" {
		t.Fatalf("DownloadBytesE = %q, %v", b, err)
	}
	if b, err := vhttp.DownloadBytesEWithOptions(server.URL, vhttp.WithMaxResponseBytes(64)); err != nil || string(b) != "download-text" {
		t.Fatalf("DownloadBytesEWithOptions = %q, %v", b, err)
	}
	if b, err := vhttp.DownloadBytesSafeE(server.URL, allowLocal); err != nil || string(b) != "download-text" {
		t.Fatalf("DownloadBytesSafeE = %q, %v", b, err)
	}
	if got, err := vhttp.DownloadStringE(server.URL, ""); err != nil || got != "download-text" {
		t.Fatalf("DownloadStringE = %q, %v", got, err)
	}
	if got, err := vhttp.DownloadStringEWithOptions(server.URL, "", vhttp.WithMaxResponseBytes(64)); err != nil || got != "download-text" {
		t.Fatalf("DownloadStringEWithOptions = %q, %v", got, err)
	}
	if got, err := vhttp.DownloadStringSafeE(server.URL, "", allowLocal); err != nil || got != "download-text" {
		t.Fatalf("DownloadStringSafeE = %q, %v", got, err)
	}

	dir := t.TempDir()
	file := filepath.Join(dir, "plain.txt")
	if n, err := vhttp.DownloadFile(server.URL, file); err != nil || n != int64(len("download-text")) {
		t.Fatalf("DownloadFile n=%d err=%v", n, err)
	}
	fileWithOpts := filepath.Join(dir, "with-options.txt")
	if n, err := vhttp.DownloadFileWithOptions(server.URL, fileWithOpts, []vhttp.RequestOption{vhttp.WithMaxResponseBytes(64)}, vhttp.WithSaveOverwrite(true)); err != nil || n != int64(len("download-text")) {
		t.Fatalf("DownloadFileWithOptions n=%d err=%v", n, err)
	}
	if n, err := vhttp.DownloadFileSafe(server.URL, filepath.Join(dir, "blocked.txt")); err == nil || n != 0 {
		t.Fatalf("DownloadFileSafe default policy n=%d err=%v, want private host rejection", n, err)
	}
}
