package http

import (
	"crypto/tls"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestRequestOptionTLSConfig(t *testing.T) {
	client := Get("https://example.com", WithTLSConfig(&tls.Config{ServerName: "example.com"})).buildClient()
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("transport type = %T", client.Transport)
	}
	if transport.TLSClientConfig == nil || transport.TLSClientConfig.ServerName != "example.com" {
		t.Fatalf("TLS config = %#v", transport.TLSClientConfig)
	}
}

func TestNilClientAndTransportOptionsDoNotOverwriteConfiguredProviders(t *testing.T) {
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("transport")),
			Request:    req,
		}, nil
	})
	client := &http.Client{Transport: transport}
	resp := Get("https://example.com", WithClient(client), WithClient(nil), WithTransport(transport), WithTransport(nil)).Execute()
	if resp.Err() != nil {
		t.Fatalf("Execute error = %v", resp.Err())
	}
	if got := resp.Body(); got != "transport" {
		t.Fatalf("Body = %q, want transport", got)
	}
}
