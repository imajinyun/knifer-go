// Package http is aligned with the utility toolkit-http and provides HTTP client, download,
// Cookie, UserAgent, SimpleServer, and related utilities.
//
// This package is the standard-library based HTTP implementation for vhttp. Use
// internal/httpx/resty through vresty when a Resty-based chainable client is desired.
//
// Engine-agnostic protocol types (Method, Header, ContentType, HTTPError) are
// re-exported from internal/httpx/internal/shared.
//
// Unlike the utility toolkit-http, this package wraps Go's standard net/http library and
// provides a chainable API:
//
//	body := http.Get("https://example.com").Execute().Body()
//	resp := http.NewRequest(http.MethodPost, url,
//	            http.WithTimeout(5*time.Second),
//	            http.WithHeader("X-Client", "knifer-go"),
//	        ).
//	            Form(map[string]any{"a": 1}).
//	            Execute()
package http
