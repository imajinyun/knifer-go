package vhttp_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/imajinyun/go-knifer/vhttp"
)

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
