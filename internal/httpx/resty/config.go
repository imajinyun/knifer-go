package resty

import (
	"sync"
	"time"
)

// HeaderValues stores HTTP header values without depending on net/http types.
type HeaderValues map[string][]string

// GlobalConfig captures resty package-level defaults for explicit request construction.
type GlobalConfig struct {
	Timeout          time.Duration
	MaxRedirects     int
	MaxResponseBytes int64
	FollowRedirects  bool
	DefaultUserAgent string
	Headers          HeaderValues
	CookieDisabled   bool
}

var (
	globalMu               sync.RWMutex
	globalTimeout          = defaultGlobalTimeout
	globalMaxRedirects     = defaultGlobalMaxRedirects
	globalMaxResponseBytes = int64(defaultGlobalMaxResponseBytes)
	globalFollowRedirects  = defaultGlobalFollowRedirects
	globalDefaultUserAgent = ""

	globalHeadersMu sync.RWMutex
	globalHeaders   = defaultGlobalHeaders()

	cookieMu       sync.RWMutex
	cookieDisabled bool
)

const (
	defaultGlobalTimeout          = 0 * time.Second
	defaultGlobalMaxRedirects     = 10
	defaultGlobalMaxResponseBytes = 64 << 20
	defaultGlobalFollowRedirects  = true
)

func defaultGlobalHeaders() HeaderValues {
	headers := HeaderValues{}
	setHeader(headers, string(HeaderAccept), "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	setHeader(headers, string(HeaderAcceptEncoding), "gzip, deflate")
	setHeader(headers, string(HeaderAcceptLanguage), "zh-CN,zh;q=0.8")
	setHeader(headers, string(HeaderUserAgent),
		"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) "+
			"Chrome/72.0.3626.109 Safari/537.36")
	return headers
}

// SnapshotGlobalConfig returns a copy of the package-level resty defaults.
func SnapshotGlobalConfig() GlobalConfig {
	globalMu.RLock()
	cfg := GlobalConfig{
		Timeout:          globalTimeout,
		MaxRedirects:     globalMaxRedirects,
		MaxResponseBytes: globalMaxResponseBytes,
		FollowRedirects:  globalFollowRedirects,
		DefaultUserAgent: globalDefaultUserAgent,
	}
	globalMu.RUnlock()
	cfg.Headers = CloneGlobalHeaders()
	cfg.CookieDisabled = isCookieDisabled()
	return cfg
}

// ResetGlobalConfig restores package-level resty defaults, including headers and cookies.
func ResetGlobalConfig() { applyGlobalConfig(defaultGlobalConfig()) }

// ConfigureGlobalConfig replaces package-level resty defaults with cfg.
func ConfigureGlobalConfig(cfg GlobalConfig) { applyGlobalConfig(cfg) }

// WithScopedGlobalConfig runs fn with cfg installed as package-level resty defaults,
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
		FollowRedirects:  defaultGlobalFollowRedirects,
		DefaultUserAgent: "",
		Headers:          defaultGlobalHeaders(),
		CookieDisabled:   false,
	}
}

func applyGlobalConfig(cfg GlobalConfig) {
	globalMu.Lock()
	globalTimeout = cfg.Timeout
	globalMaxRedirects = cfg.MaxRedirects
	globalMaxResponseBytes = cfg.MaxResponseBytes
	globalFollowRedirects = cfg.FollowRedirects
	globalDefaultUserAgent = cfg.DefaultUserAgent
	globalMu.Unlock()

	globalHeadersMu.Lock()
	globalHeaders = cloneHeaders(cfg.Headers)
	globalHeadersMu.Unlock()

	cookieMu.Lock()
	cookieDisabled = cfg.CookieDisabled
	cookieMu.Unlock()
}

func isolatedGlobalConfig() GlobalConfig {
	return GlobalConfig{FollowRedirects: defaultGlobalFollowRedirects, MaxRedirects: defaultGlobalMaxRedirects, MaxResponseBytes: defaultGlobalMaxResponseBytes}
}

func cloneHeaders(headers HeaderValues) HeaderValues {
	out := HeaderValues{}
	for k, v := range headers {
		out[k] = append([]string(nil), v...)
	}
	return out
}

// SetGlobalTimeout sets the global default timeout.
func SetGlobalTimeout(d time.Duration) { globalMu.Lock(); defer globalMu.Unlock(); globalTimeout = d }

// GetGlobalTimeout returns the global default timeout.
func GetGlobalTimeout() time.Duration {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalTimeout
}

// SetGlobalMaxRedirects sets the global maximum redirect count.
func SetGlobalMaxRedirects(n int) { globalMu.Lock(); defer globalMu.Unlock(); globalMaxRedirects = n }

// GetGlobalMaxRedirects returns the global maximum redirect count.
func GetGlobalMaxRedirects() int {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalMaxRedirects
}

// SetGlobalMaxResponseBytes sets the global maximum response bytes read into memory.
// Non-positive values disable the limit.
func SetGlobalMaxResponseBytes(n int64) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalMaxResponseBytes = n
}

// GetGlobalMaxResponseBytes returns the global maximum response bytes read into memory.
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

// SetGlobalHeader sets a global default request header.
func SetGlobalHeader(name, value string) {
	globalHeadersMu.Lock()
	defer globalHeadersMu.Unlock()
	setHeader(globalHeaders, name, value)
}

// AddGlobalHeader appends a global default request header value.
func AddGlobalHeader(name, value string) {
	globalHeadersMu.Lock()
	defer globalHeadersMu.Unlock()
	globalHeaders[name] = append(globalHeaders[name], value)
}

// RemoveGlobalHeader removes a global default request header.
func RemoveGlobalHeader(name string) {
	globalHeadersMu.Lock()
	defer globalHeadersMu.Unlock()
	delete(globalHeaders, name)
}

// CloneGlobalHeaders returns a copy of global default request headers.
func CloneGlobalHeaders() HeaderValues {
	globalHeadersMu.RLock()
	defer globalHeadersMu.RUnlock()
	return cloneHeaders(globalHeaders)
}

// CloseCookie disables global cookie management.
func CloseCookie() {
	cookieMu.Lock()
	defer cookieMu.Unlock()
	cookieDisabled = true
}

func isCookieDisabled() bool {
	cookieMu.RLock()
	defer cookieMu.RUnlock()
	return cookieDisabled
}

func setHeader(headers HeaderValues, name, value string) {
	headers[name] = []string{value}
}

func getHeader(headers HeaderValues, name string) string {
	if values := headers[name]; len(values) > 0 {
		return values[0]
	}
	return ""
}
