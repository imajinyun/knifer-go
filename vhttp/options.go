package vhttp

import (
	"crypto/tls"
	"net/http"
	"time"

	httpx "github.com/imajinyun/go-knifer/internal/httpx/http"
)

// WithTimeout sets a per-request timeout.
func WithTimeout(d time.Duration) RequestOption { return httpx.WithTimeout(d) }

// WithHeader sets one per-request header.
func WithHeader(name, value string) RequestOption { return httpx.WithHeader(name, value) }

// WithHeaders sets per-request headers in batch.
func WithHeaders(headers map[string]string) RequestOption { return httpx.WithHeaders(headers) }

// WithFollowRedirects sets per-request redirect behavior.
func WithFollowRedirects(b bool) RequestOption { return httpx.WithFollowRedirects(b) }

// WithMaxRedirects sets the per-request redirect limit.
func WithMaxRedirects(n int) RequestOption { return httpx.WithMaxRedirects(n) }

// WithSkipTLSVerify sets per-request TLS verification behavior.
func WithSkipTLSVerify(b bool) RequestOption { return httpx.WithSkipTLSVerify(b) }

// WithTLSConfig sets a per-request TLS config.
func WithTLSConfig(cfg *tls.Config) RequestOption { return httpx.WithTLSConfig(cfg) }

// WithTransport sets a per-request RoundTripper.
func WithTransport(t http.RoundTripper) RequestOption { return httpx.WithTransport(t) }

// WithClient sets a per-request HTTP client.
func WithClient(c *http.Client) RequestOption { return httpx.WithClient(c) }

// WithCookieJar sets a per-request CookieJar. nil disables cookie management for this request.
func WithCookieJar(jar http.CookieJar) RequestOption { return httpx.WithCookieJar(jar) }

// WithUserAgent sets a per-request User-Agent.
func WithUserAgent(ua string) RequestOption { return httpx.WithUserAgent(ua) }

// WithContentType sets a per-request Content-Type at construction time.
func WithContentType(ct string) RequestOption { return httpx.WithContentType(ct) }

// WithCharset sets a per-request charset at construction time.
func WithCharset(charset string) RequestOption { return httpx.WithCharset(charset) }
