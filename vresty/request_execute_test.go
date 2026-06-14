package vresty_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

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
