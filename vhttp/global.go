package vhttp

import (
	"net/http"
	"time"

	httpx "github.com/imajinyun/go-knifer/internal/httpx/http"
)

// SetGlobalTimeout sets the global HTTP timeout.
func SetGlobalTimeout(d time.Duration) { httpx.SetGlobalTimeout(d) }

// GetGlobalTimeout returns the global HTTP timeout.
func GetGlobalTimeout() time.Duration { return httpx.GetGlobalTimeout() }

// SetGlobalHeader sets a global HTTP header.
func SetGlobalHeader(name, value string) { httpx.SetGlobalHeader(name, value) }

// AddGlobalHeader adds a global HTTP header value.
func AddGlobalHeader(name, value string) { httpx.AddGlobalHeader(name, value) }

// RemoveGlobalHeader removes a global HTTP header.
func RemoveGlobalHeader(name string) { httpx.RemoveGlobalHeader(name) }

// CloneGlobalHeaders returns cloned global headers.
func CloneGlobalHeaders() http.Header { return httpx.CloneGlobalHeaders() }

// SetGlobalMaxRedirects delegates to the internal httpx implementation.
func SetGlobalMaxRedirects(n int) {
	httpx.SetGlobalMaxRedirects(n)
}

// GetGlobalMaxRedirects delegates to the internal httpx implementation.
func GetGlobalMaxRedirects() int {
	return httpx.GetGlobalMaxRedirects()
}

// SetGlobalFollowRedirects delegates to the internal httpx implementation.
func SetGlobalFollowRedirects(b bool) {
	httpx.SetGlobalFollowRedirects(b)
}

// GetGlobalFollowRedirects delegates to the internal httpx implementation.
func GetGlobalFollowRedirects() bool {
	return httpx.GetGlobalFollowRedirects()
}

// SetGlobalUserAgent delegates to the internal httpx implementation.
func SetGlobalUserAgent(ua string) {
	httpx.SetGlobalUserAgent(ua)
}

// GetGlobalUserAgent delegates to the internal httpx implementation.
func GetGlobalUserAgent() string {
	return httpx.GetGlobalUserAgent()
}

// SetIgnoreEOFError delegates to the internal httpx implementation.
func SetIgnoreEOFError(b bool) {
	httpx.SetIgnoreEOFError(b)
}

// IsIgnoreEOFError delegates to the internal httpx implementation.
func IsIgnoreEOFError() bool {
	return httpx.IsIgnoreEOFError()
}

// SetTrustAnyHost delegates to the internal httpx implementation.
func SetTrustAnyHost(b bool) {
	httpx.SetTrustAnyHost(b)
}

// IsTrustAnyHost delegates to the internal httpx implementation.
func IsTrustAnyHost() bool {
	return httpx.IsTrustAnyHost()
}

// SetGlobalBoundary delegates to the internal httpx implementation.
func SetGlobalBoundary(b string) {
	httpx.SetGlobalBoundary(b)
}

// GetGlobalBoundary delegates to the internal httpx implementation.
func GetGlobalBoundary() string {
	return httpx.GetGlobalBoundary()
}

// SetGlobalDecodeURL delegates to the internal httpx implementation.
func SetGlobalDecodeURL(b bool) {
	httpx.SetGlobalDecodeURL(b)
}

// IsGlobalDecodeURL delegates to the internal httpx implementation.
func IsGlobalDecodeURL() bool {
	return httpx.IsGlobalDecodeURL()
}
