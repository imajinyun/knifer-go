package vurl_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/imajinyun/knifer-go/vurl"
)

func TestFacadeResourceMaxBytesAndStatusErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/large":
			w.Header().Set("Content-Length", "10")
			_, _ = w.Write([]byte("0123456789"))
		default:
			http.Error(w, "nope", http.StatusTeapot)
		}
	}))
	defer server.Close()

	if _, err := vurl.OpenWithOptions(server.URL+"/large", vurl.WithMaxBytes(2)); err == nil {
		t.Fatal("OpenWithOptions content-length max bytes error = nil")
	}
	if _, err := vurl.ContentLengthWithOptions(server.URL+"/large", vurl.WithMaxBytes(2)); err == nil {
		t.Fatal("ContentLengthWithOptions max bytes error = nil")
	}
	rc, err := vurl.OpenWithOptions(server.URL+"/large", vurl.WithMaxBytes(4), vurl.WithHTTPClient(&http.Client{Transport: stripLengthTransport{base: http.DefaultTransport}}))
	if err != nil {
		t.Fatalf("OpenWithOptions limited reader: %v", err)
	}
	_, err = io.ReadAll(rc)
	_ = rc.Close()
	if err == nil {
		t.Fatal("limited reader overflow error = nil")
	}
	if _, err := vurl.OpenWithOptions(server.URL+"/status", vurl.WithCheckStatus(true)); err == nil {
		t.Fatal("OpenWithOptions check status error = nil")
	}
}

type stripLengthTransport struct{ base http.RoundTripper }

func (t stripLengthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	resp.ContentLength = -1
	resp.Header.Del("Content-Length")
	return resp, nil
}
