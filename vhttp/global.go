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

// SnapshotGlobalConfig returns a consistent copy of the current package-level HTTP defaults.
func SnapshotGlobalConfig() GlobalConfig { return httpx.SnapshotGlobalConfig() }

// ResetGlobalConfig restores package-level HTTP defaults, including headers and cookie jar.
func ResetGlobalConfig() { httpx.ResetGlobalConfig() }

// ConfigureGlobalConfig replaces package-level HTTP defaults with cfg.
func ConfigureGlobalConfig(cfg GlobalConfig) { httpx.ConfigureGlobalConfig(cfg) }

// WithScopedGlobalConfig runs fn with cfg installed as package-level HTTP defaults,
// then restores the previous defaults.
func WithScopedGlobalConfig(cfg GlobalConfig, fn func()) { httpx.WithScopedGlobalConfig(cfg, fn) }

// SetGlobalMaxRedirects delegates to the internal httpx implementation.
func SetGlobalMaxRedirects(n int) {
	httpx.SetGlobalMaxRedirects(n)
}

// GetGlobalMaxRedirects delegates to the internal httpx implementation.
func GetGlobalMaxRedirects() int {
	return httpx.GetGlobalMaxRedirects()
}

// SetGlobalMaxResponseBytes sets the global maximum response bytes read by response Bytes/Body helpers.
// Non-positive values disable the limit.
func SetGlobalMaxResponseBytes(n int64) { httpx.SetGlobalMaxResponseBytes(n) }

// GetGlobalMaxResponseBytes returns the global maximum response bytes read by response Bytes/Body helpers.
func GetGlobalMaxResponseBytes() int64 { return httpx.GetGlobalMaxResponseBytes() }

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
