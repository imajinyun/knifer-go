package vhttp_test

import (
	"context"
	"crypto/tls"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"testing"
	"time"

	"github.com/imajinyun/knifer-go/vhttp"
)

func TestFacadeRequestOptionWrappers(t *testing.T) {
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

func TestFacadeWithMaxRedirects(t *testing.T) {
	if vhttp.WithMaxRedirects(5) == nil {
		t.Fatal("WithMaxRedirects returned nil")
	}
	if vhttp.WithFollowRedirects(false) == nil {
		t.Fatal("WithFollowRedirects returned nil")
	}
}
