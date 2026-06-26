package vhttp_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/imajinyun/knifer-go/vhttp"
)

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

	allowLocal := allowLocalURLPolicy()
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
