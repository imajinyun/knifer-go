package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestBuildBasicAuth(t *testing.T) {
	if got := BuildBasicAuth("aladdin", "opensesame"); got != "Basic YWxhZGRpbjpvcGVuc2VzYW1l" {
		t.Fatalf("auth: %q", got)
	}
}

func TestStringHelpersWithError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Method + ":" + r.URL.Query().Get("k")))
	}))
	defer srv.Close()

	body, err := GetWithParamsE(srv.URL, map[string]any{"k": "v"})
	if err != nil || body != "GET:v" {
		t.Fatalf("GetWithParamsE = %q, %v", body, err)
	}

	_, err = GetStringE("http://[::1")
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("GetStringE invalid URL error = %v, want invalid input", err)
	}
}

func TestDownloadBytesEWithReadError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("abcdef"))
	}))
	defer srv.Close()

	_, err := DownloadBytesEWithOptions(srv.URL, WithMaxResponseBytes(3))
	if !errors.Is(err, knifer.ErrCodeUnsupported) {
		t.Fatalf("DownloadBytesE error = %v, want unsupported", err)
	}
}
