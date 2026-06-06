package vhttp

import (
	"crypto/tls"
	"io"
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

// WithTLSConfig sets a per-request TLS config. It is ignored when WithClient or WithTransport is set.
func WithTLSConfig(cfg *tls.Config) RequestOption { return httpx.WithTLSConfig(cfg) }

// WithTransport sets a per-request RoundTripper and takes precedence over WithTLSConfig.
func WithTransport(t http.RoundTripper) RequestOption { return httpx.WithTransport(t) }

// WithTransportProvider sets a per-request RoundTripper provider evaluated when the request is built.
func WithTransportProvider(provider func() http.RoundTripper) RequestOption {
	return httpx.WithTransportProvider(provider)
}

// ConfigureDefaultTransportProvider sets the provider used to initialize the shared default transport.
func ConfigureDefaultTransportProvider(provider func() *http.Transport) {
	httpx.ConfigureDefaultTransportProvider(provider)
}

// ResetDefaultTransport clears the cached shared default transport and restores the standard provider.
func ResetDefaultTransport() { httpx.ResetDefaultTransport() }

// WithClient sets a per-request HTTP client and takes precedence over WithTransport and WithTLSConfig.
func WithClient(c *http.Client) RequestOption { return httpx.WithClient(c) }

// WithCookieJar sets a per-request CookieJar. nil disables cookie management for this request.
func WithCookieJar(jar http.CookieJar) RequestOption { return httpx.WithCookieJar(jar) }

// WithGlobalConfig initializes request defaults from a captured global configuration snapshot.
func WithGlobalConfig(cfg GlobalConfig) RequestOption { return httpx.WithGlobalConfig(cfg) }

// WithUserAgent sets a per-request User-Agent.
func WithUserAgent(ua string) RequestOption { return httpx.WithUserAgent(ua) }

// WithContentType sets a per-request Content-Type at construction time.
func WithContentType(ct string) RequestOption { return httpx.WithContentType(ct) }

// WithCharset sets a per-request charset at construction time.
func WithCharset(charset string) RequestOption { return httpx.WithCharset(charset) }

// WithAutoDecodeResponse controls whether response bodies are decoded by Content-Encoding.
func WithAutoDecodeResponse(autoDecode bool) RequestOption {
	return httpx.WithAutoDecodeResponse(autoDecode)
}

// WithContentDecoder registers a per-request response body decoder for encoding.
func WithContentDecoder(encoding string, decoder func(io.Reader) (io.ReadCloser, error)) RequestOption {
	return httpx.WithContentDecoder(encoding, decoder)
}

// WithRequestFactory sets the HTTP request factory used at execution time.
func WithRequestFactory(newRequest NewRequestFunc) RequestOption {
	return httpx.WithRequestFactory(newRequest)
}

// WithMultipartWriterFactory sets the multipart writer factory used when building multipart request bodies.
func WithMultipartWriterFactory(factory MultipartWriterFactory) RequestOption {
	return httpx.WithMultipartWriterFactory(factory)
}
