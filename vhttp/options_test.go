package vhttp_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/imajinyun/go-knifer/vhttp"
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

func TestFacadeResponseDecodeOptions(t *testing.T) {
	gzipServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		_, _ = gz.Write([]byte("gzipped"))
		_ = gz.Close()
	}))
	defer gzipServer.Close()

	compressed := vhttp.Get(gzipServer.URL, vhttp.WithAutoDecodeResponse(false)).Execute().Bytes()
	if bytes.Contains(compressed, []byte("gzipped")) || len(compressed) == 0 {
		t.Fatalf("body should remain compressed, got %q", compressed)
	}

	customServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "upper")
		_, _ = w.Write([]byte("hello"))
	}))
	defer customServer.Close()

	decoder := func(r io.Reader) (io.ReadCloser, error) {
		data, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}
		return io.NopCloser(strings.NewReader(strings.ToUpper(string(data)))), nil
	}
	if got := vhttp.Get(customServer.URL, vhttp.WithContentDecoder("upper", decoder)).Execute().Body(); got != "HELLO" {
		t.Fatalf("custom decoded body = %q", got)
	}
}

func TestFacadeAdditionalOptionWrappers(t *testing.T) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	previous := vhttp.SnapshotGlobalConfig()
	defer vhttp.ConfigureGlobalConfig(previous)
	vhttp.SetCookieJar(jar)
	if vhttp.GetCookieJar() == nil {
		t.Fatal("GetCookieJar returned nil after SetCookieJar")
	}
	vhttp.CloseCookie()
	if vhttp.GetCookieJar() != nil {
		t.Fatal("CloseCookie did not clear global jar")
	}

	requestFactoryCalled := false
	readAllCalled := false
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/plain"}},
			Body:       io.NopCloser(strings.NewReader(req.Method + ":" + req.Header.Get("X-A"))),
			Request:    req,
		}, nil
	})
	client := &http.Client{Transport: transport}
	cfg := vhttp.SnapshotGlobalConfig()
	cfg.Headers.Set("X-A", "from-config")
	resp := vhttp.NewIsolatedRequest(vhttp.MethodPost, "https://example.com",
		vhttp.WithGlobalConfig(cfg),
		vhttp.WithTimeout(time.Second),
		vhttp.WithHeaders(map[string]string{"X-A": "one"}),
		vhttp.WithTLSConfig(&tls.Config{MinVersion: tls.VersionTLS12}),
		vhttp.WithTransport(transport),
		vhttp.WithClient(client),
		vhttp.WithCookieJar(jar),
		vhttp.WithContentType(string(vhttp.ContentTypeTextPlain)),
		vhttp.WithCharset("utf-8"),
		vhttp.WithMaxResponseBytes(64),
		vhttp.WithResponseReadAllFunc(func(r io.Reader) ([]byte, error) {
			readAllCalled = true
			return io.ReadAll(r)
		}),
		vhttp.WithRequestFactory(func(method, rawURL string, body io.Reader) (*http.Request, error) {
			requestFactoryCalled = true
			return http.NewRequest(method, rawURL, body)
		}),
		vhttp.WithMultipartWriterFactory(func(w io.Writer) vhttp.MultipartWriter {
			return multipart.NewWriter(w)
		}),
	).BodyString("payload").Execute()
	if resp.Err() != nil {
		t.Fatalf("NewIsolatedRequest Execute: %v", resp.Err())
	}
	if got := resp.Body(); got != "POST:one" {
		t.Fatalf("option response = %q", got)
	}
	if !requestFactoryCalled || !readAllCalled {
		t.Fatalf("providers called request=%v readAll=%v", requestFactoryCalled, readAllCalled)
	}

	cfg.Headers.Set("X-A", "configured")
	if got := vhttp.NewRequestWithConfig(vhttp.MethodGet, "https://example.com", cfg, vhttp.WithTransport(transport)).Execute().Body(); got != "GET:configured" {
		t.Fatalf("NewRequestWithConfig body = %q", got)
	}
	if resp := vhttp.Get("http://public.example",
		vhttp.WithAllowedHosts("public.example"),
		vhttp.WithLookupIP(func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("8.8.8.8")}, nil
		}),
		vhttp.WithTransport(transport),
	).Execute(); resp.Err() != nil {
		t.Fatalf("allowed host request error = %v", resp.Err())
	}
}
