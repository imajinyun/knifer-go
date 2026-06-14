package vurl_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/imajinyun/go-knifer/vurl"
)

func TestFacadeResourceOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Test"); got != "facade" {
			t.Fatalf("header X-Test = %q, want facade", got)
		}
		w.Header().Set("Content-Length", "5")
		_, _ = w.Write([]byte("hello"))
	}))
	defer server.Close()

	rc, err := vurl.OpenWithOptions(server.URL, vurl.WithHeader("X-Test", "facade"), vurl.WithCheckStatus(true))
	if err != nil {
		t.Fatalf("OpenWithOptions: %v", err)
	}
	defer func() { _ = rc.Close() }()
	data, err := io.ReadAll(rc)
	if err != nil || string(data) != "hello" {
		t.Fatalf("OpenWithOptions data = %q, %v", data, err)
	}

	length, err := vurl.ContentLengthWithOptions(server.URL, vurl.WithHeader("X-Test", "facade"), vurl.WithCheckStatus(true))
	if err != nil || length != 5 {
		t.Fatalf("ContentLengthWithOptions = %d, %v; want 5, nil", length, err)
	}
	if size, err := vurl.SizeWithOptions(server.URL, vurl.WithHeader("X-Test", "facade")); err != nil || size != 5 {
		t.Fatalf("SizeWithOptions = %d, %v; want 5, nil", size, err)
	}
}
