package vhttp_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/imajinyun/knifer-go/vhttp"
)

func TestFacadeTransportProviderOption(t *testing.T) {
	calls := 0
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(req.Header.Get("X-Transport"))),
			Header:     http.Header{},
			Request:    req,
		}, nil
	})
	resp := vhttp.Get("https://example.com",
		vhttp.WithHeader("X-Transport", "facade"),
		vhttp.WithTransportProvider(func() http.RoundTripper {
			calls++
			return transport
		}),
	).Execute()
	if resp.Err() != nil {
		t.Fatal(resp.Err())
	}
	if calls != 1 || resp.Body() != "facade" {
		t.Fatalf("transport provider calls=%d body=%q", calls, resp.Body())
	}
}

func TestFacadeDefaultTransportProviderLifecycle(t *testing.T) {
	custom := &http.Transport{MaxIdleConnsPerHost: 5}
	vhttp.ConfigureDefaultTransportProvider(func() *http.Transport { return custom })
	t.Cleanup(vhttp.ResetDefaultTransport)

	providerCalls := 0
	resp := vhttp.Get("https://example.com",
		vhttp.WithTransportProvider(func() http.RoundTripper {
			providerCalls++
			return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("ok")), Header: http.Header{}, Request: req}, nil
			})
		}),
	).Execute()
	if resp.Err() != nil || resp.Body() != "ok" || providerCalls != 1 {
		t.Fatalf("per-request transport provider resp=%q err=%v calls=%d", resp.Body(), resp.Err(), providerCalls)
	}

	vhttp.ResetDefaultTransport()
}
