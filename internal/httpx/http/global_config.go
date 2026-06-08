package http

import (
	"net/http"
	"sync"
	"time"
)

// Global default configuration, aligned with the utility toolkit-http HttpGlobalConfig.
var (
	globalMu               sync.RWMutex
	globalTimeout          = defaultGlobalTimeout // 0 means using the HTTP client's default timeout.
	globalMaxRedirects     = defaultGlobalMaxRedirects
	globalMaxResponseBytes = int64(defaultGlobalMaxResponseBytes)
	globalIgnoreEOFError   = true
	globalDecodeURL        = false
	globalFollowRedirects  = true
	globalDefaultUserAgent = ""
	globalBoundary         = defaultGlobalBoundary
)

const (
	defaultGlobalTimeout          = 0 * time.Second
	defaultGlobalMaxRedirects     = 10
	defaultGlobalMaxResponseBytes = 64 << 20
	defaultGlobalBoundary         = "--------------------gokitFormBoundary"
)

// GlobalConfig is an immutable snapshot of package-level HTTP defaults.
type GlobalConfig struct {
	Timeout          time.Duration
	MaxRedirects     int
	MaxResponseBytes int64
	IgnoreEOFError   bool
	DecodeURL        bool
	FollowRedirects  bool
	DefaultUserAgent string
	Boundary         string
	Headers          http.Header
	CookieJar        http.CookieJar
}

// SnapshotGlobalConfig returns a consistent copy of the current package-level HTTP defaults.
func SnapshotGlobalConfig() GlobalConfig {
	globalMu.RLock()
	cfg := GlobalConfig{
		Timeout:          globalTimeout,
		MaxRedirects:     globalMaxRedirects,
		MaxResponseBytes: globalMaxResponseBytes,
		IgnoreEOFError:   globalIgnoreEOFError,
		DecodeURL:        globalDecodeURL,
		FollowRedirects:  globalFollowRedirects,
		DefaultUserAgent: globalDefaultUserAgent,
		Boundary:         globalBoundary,
	}
	globalMu.RUnlock()
	cfg.Headers = CloneGlobalHeaders()
	cfg.CookieJar = GetCookieJar()
	return cfg
}

// ResetGlobalConfig restores package-level HTTP defaults, including headers and cookie jar.
func ResetGlobalConfig() { applyGlobalConfig(defaultGlobalConfig()) }

// ConfigureGlobalConfig replaces package-level HTTP defaults with cfg.
func ConfigureGlobalConfig(cfg GlobalConfig) { applyGlobalConfig(cfg) }

// WithScopedGlobalConfig runs fn with cfg installed as package-level HTTP defaults,
// then restores the previous defaults. It is intended for tests and serialized setup code.
func WithScopedGlobalConfig(cfg GlobalConfig, fn func()) {
	previous := SnapshotGlobalConfig()
	ConfigureGlobalConfig(cfg)
	defer ConfigureGlobalConfig(previous)
	if fn != nil {
		fn()
	}
}

func defaultGlobalConfig() GlobalConfig {
	return GlobalConfig{
		Timeout:          defaultGlobalTimeout,
		MaxRedirects:     defaultGlobalMaxRedirects,
		MaxResponseBytes: defaultGlobalMaxResponseBytes,
		IgnoreEOFError:   true,
		DecodeURL:        false,
		FollowRedirects:  true,
		DefaultUserAgent: "",
		Boundary:         defaultGlobalBoundary,
		Headers:          defaultGlobalHeaders(),
		CookieJar:        newDefaultCookieJar(),
	}
}

func applyGlobalConfig(cfg GlobalConfig) {
	globalMu.Lock()
	globalTimeout = cfg.Timeout
	globalMaxRedirects = cfg.MaxRedirects
	globalMaxResponseBytes = cfg.MaxResponseBytes
	globalIgnoreEOFError = cfg.IgnoreEOFError
	globalDecodeURL = cfg.DecodeURL
	globalFollowRedirects = cfg.FollowRedirects
	globalDefaultUserAgent = cfg.DefaultUserAgent
	globalBoundary = cfg.Boundary
	globalMu.Unlock()

	globalHeadersMu.Lock()
	globalHeaders = cloneHeader(cfg.Headers)
	globalHeadersMu.Unlock()

	SetCookieJar(cfg.CookieJar)
}

// SetGlobalTimeout sets the global default timeout.
func SetGlobalTimeout(d time.Duration) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalTimeout = d
}

// GetGlobalTimeout returns the global default timeout.
func GetGlobalTimeout() time.Duration {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalTimeout
}

// SetGlobalMaxRedirects sets the global maximum redirect count.
func SetGlobalMaxRedirects(n int) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalMaxRedirects = n
}

// GetGlobalMaxRedirects returns the global maximum redirect count.
func GetGlobalMaxRedirects() int {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalMaxRedirects
}

// SetGlobalMaxResponseBytes sets the global maximum response bytes read by response Bytes/Body helpers.
// Non-positive values disable the limit.
func SetGlobalMaxResponseBytes(n int64) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalMaxResponseBytes = n
}

// GetGlobalMaxResponseBytes returns the global maximum response bytes read by response Bytes/Body helpers.
func GetGlobalMaxResponseBytes() int64 {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalMaxResponseBytes
}

// SetGlobalFollowRedirects sets whether redirects are followed.
func SetGlobalFollowRedirects(b bool) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalFollowRedirects = b
}

// GetGlobalFollowRedirects reports whether redirects are followed.
func GetGlobalFollowRedirects() bool {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalFollowRedirects
}

// SetGlobalUserAgent sets the global default User-Agent.
func SetGlobalUserAgent(ua string) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalDefaultUserAgent = ua
}

// GetGlobalUserAgent returns the global default User-Agent.
func GetGlobalUserAgent() string {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalDefaultUserAgent
}

// SetIgnoreEOFError sets whether EOF errors are ignored.
func SetIgnoreEOFError(b bool) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalIgnoreEOFError = b
}

// IsIgnoreEOFError reports whether EOF errors are ignored.
func IsIgnoreEOFError() bool {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalIgnoreEOFError
}

// SetGlobalBoundary sets the default multipart boundary.
func SetGlobalBoundary(b string) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalBoundary = b
}

// GetGlobalBoundary returns the default multipart boundary.
func GetGlobalBoundary() string {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalBoundary
}

// SetGlobalDecodeURL sets whether URLs are decoded automatically.
func SetGlobalDecodeURL(b bool) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalDecodeURL = b
}

// IsGlobalDecodeURL reports whether URLs are decoded automatically.
func IsGlobalDecodeURL() bool {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalDecodeURL
}
