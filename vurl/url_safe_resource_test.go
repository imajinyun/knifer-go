package vurl_test

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/imajinyun/go-knifer/vurl"
)

func TestFacadeSafeResourceHelpersRejectLocalhost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "2")
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	if rc, err := vurl.OpenSafe(server.URL); err == nil {
		_ = rc.Close()
		t.Fatal("OpenSafe(localhost) error = nil, want private host rejection")
	}
	if _, err := vurl.OpenSafeWithOptions(server.URL); err == nil {
		t.Fatal("OpenSafeWithOptions(localhost) error = nil, want private host rejection")
	}
	if _, err := vurl.ContentLengthSafe(server.URL); err == nil {
		t.Fatal("ContentLengthSafe(localhost) error = nil, want private host rejection")
	}
	if _, err := vurl.ContentLengthSafeWithOptions(server.URL); err == nil {
		t.Fatal("ContentLengthSafeWithOptions(localhost) error = nil, want private host rejection")
	}
}

func TestFacadeSafeResourcePolicyAllowsPublicHost(t *testing.T) {
	body := "public-body"
	client := &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if got := req.Header.Get("X-Safe"); got != "one" {
			t.Fatalf("safe header = %q", got)
		}
		return &http.Response{
			StatusCode:    http.StatusOK,
			ContentLength: int64(len(body)),
			Header:        http.Header{"Content-Length": []string{"11"}},
			Body:          io.NopCloser(strings.NewReader(body)),
			Request:       req,
		}, nil
	})}
	lookup := func(context.Context, string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("8.8.8.8")}, nil
	}
	opts := []vurl.ResourceOption{
		vurl.WithHTTPClient(client),
		vurl.WithHeaders(http.Header{"X-Safe": []string{"one"}}),
		vurl.WithTimeout(time.Second),
		vurl.WithLookupIP(lookup),
		vurl.WithAllowedHosts("public.example"),
		vurl.WithRejectPrivateHosts(true),
		vurl.WithMaxBytes(64),
	}
	rc, err := vurl.OpenSafeWithOptions("http://public.example/resource", opts...)
	if err != nil {
		t.Fatalf("OpenSafeWithOptions public: %v", err)
	}
	data, err := io.ReadAll(rc)
	_ = rc.Close()
	if err != nil || string(data) != body {
		t.Fatalf("OpenSafeWithOptions data = %q, %v", data, err)
	}
	if length, err := vurl.ContentLengthSafeWithOptions("http://public.example/resource", opts...); err != nil || length != int64(len(body)) {
		t.Fatalf("ContentLengthSafeWithOptions = %d, %v", length, err)
	}
	if _, err := vurl.OpenSafeWithOptions("ftp://public.example/resource", opts...); err == nil {
		t.Fatal("OpenSafeWithOptions disallowed safe scheme error = nil")
	}
	if _, err := vurl.OpenSafeWithOptions("http://other.example/resource", opts...); err == nil {
		t.Fatal("OpenSafeWithOptions disallowed host error = nil")
	}
	if _, err := vurl.OpenSafeWithOptions("http://private.example/resource", vurl.WithHTTPClient(client), vurl.WithLookupIP(func(context.Context, string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("10.0.0.1")}, nil
	})); err == nil {
		t.Fatal("OpenSafeWithOptions private resolver error = nil")
	}
}

func TestFacadeSafeResourceRejectsUnsafeRedirect(t *testing.T) {
	client := &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusFound,
			Header:     http.Header{"Location": []string{"http://127.0.0.1/private"}},
			Body:       io.NopCloser(strings.NewReader("")),
			Request:    req,
		}, nil
	})}
	lookup := func(context.Context, string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("93.184.216.34")}, nil
	}
	if rc, err := vurl.OpenSafeWithOptions("http://example.com/redirect",
		vurl.WithHTTPClient(client),
		vurl.WithAllowedHosts("example.com"),
		vurl.WithLookupIP(lookup),
	); err == nil {
		_ = rc.Close()
		t.Fatal("OpenSafeWithOptions unsafe redirect error = nil")
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }
