package http

import (
	"net/http"
	"strings"
	"testing"
	"time"
)

// Covers the utility toolkit-http HttpGlobalConfigTest.

func TestGlobalTimeout(t *testing.T) {
	old := GetGlobalTimeout()
	defer SetGlobalTimeout(old)

	SetGlobalTimeout(7 * time.Second)
	if got := GetGlobalTimeout(); got != 7*time.Second {
		t.Fatalf("timeout: %v", got)
	}
}

func TestGlobalUserAgent(t *testing.T) {
	old := GetGlobalUserAgent()
	defer SetGlobalUserAgent(old)

	SetGlobalUserAgent("gokit-test/1.0")
	if got := GetGlobalUserAgent(); got != "gokit-test/1.0" {
		t.Fatalf("ua: %q", got)
	}
}

func TestGlobalFollowRedirects(t *testing.T) {
	old := GetGlobalFollowRedirects()
	defer SetGlobalFollowRedirects(old)

	SetGlobalFollowRedirects(false)
	if GetGlobalFollowRedirects() {
		t.Fatal("expected false")
	}
}

func TestGlobalMaxRedirects(t *testing.T) {
	old := GetGlobalMaxRedirects()
	defer SetGlobalMaxRedirects(old)

	SetGlobalMaxRedirects(3)
	if got := GetGlobalMaxRedirects(); got != 3 {
		t.Fatalf("max: %d", got)
	}
}

func TestGlobalMaxResponseBytes(t *testing.T) {
	old := GetGlobalMaxResponseBytes()
	defer SetGlobalMaxResponseBytes(old)

	SetGlobalMaxResponseBytes(123)
	if got := GetGlobalMaxResponseBytes(); got != 123 {
		t.Fatalf("max response bytes: %d", got)
	}
}

func TestGlobalIgnoreEOFError(t *testing.T) {
	old := IsIgnoreEOFError()
	defer SetIgnoreEOFError(old)

	SetIgnoreEOFError(false)
	if IsIgnoreEOFError() {
		t.Fatal("expected false")
	}
}

func TestGlobalHeadersDefault(t *testing.T) {
	headers := CloneGlobalHeaders()
	if headers.Get("User-Agent") == "" {
		t.Fatal("default UA missing")
	}
	if headers.Get("Accept") == "" {
		t.Fatal("default Accept missing")
	}
	if got := headers.Get("Accept-Encoding"); strings.Contains(got, "br") {
		t.Fatalf("default Accept-Encoding = %q should not advertise br without brotli decoding support", got)
	}
}

func TestGlobalHeadersSetAndRemove(t *testing.T) {
	SetGlobalHeader("X-Test", "v1")
	defer RemoveGlobalHeader("X-Test")

	headers := CloneGlobalHeaders()
	if headers.Get("X-Test") != "v1" {
		t.Fatalf("X-Test: %q", headers.Get("X-Test"))
	}

	RemoveGlobalHeader("X-Test")
	if got := CloneGlobalHeaders().Get("X-Test"); got != "" {
		t.Fatalf("after remove: %q", got)
	}
}

func TestGlobalCookieJar(t *testing.T) {
	jar := GetCookieJar()
	if jar == nil {
		t.Fatal("default jar should not be nil")
	}
	CloseCookie()
	if GetCookieJar() != nil {
		t.Fatal("after close should be nil")
	}
	// Restore the default jar.
	SetCookieJar(jar)
	if GetCookieJar() == nil {
		t.Fatal("restored jar nil")
	}

	// Customize the jar.
	var custom http.CookieJar
	SetCookieJar(custom)
	if GetCookieJar() != nil {
		t.Fatal("custom nil jar")
	}
	SetCookieJar(jar)
}

func TestSnapshotGlobalConfigClonesMutableDefaults(t *testing.T) {
	oldUA := GetGlobalUserAgent()
	oldTimeout := GetGlobalTimeout()
	oldFollow := GetGlobalFollowRedirects()
	oldMax := GetGlobalMaxRedirects()
	oldMaxResponse := GetGlobalMaxResponseBytes()
	jar := GetCookieJar()
	defer SetGlobalUserAgent(oldUA)
	defer SetGlobalTimeout(oldTimeout)
	defer SetGlobalFollowRedirects(oldFollow)
	defer SetGlobalMaxRedirects(oldMax)
	defer SetGlobalMaxResponseBytes(oldMaxResponse)
	defer SetCookieJar(jar)
	defer RemoveGlobalHeader("X-Snapshot")

	SetGlobalUserAgent("snapshot-agent")
	SetGlobalTimeout(9 * time.Second)
	SetGlobalFollowRedirects(false)
	SetGlobalMaxRedirects(2)
	SetGlobalMaxResponseBytes(321)
	SetGlobalHeader("X-Snapshot", "old")
	CloseCookie()

	cfg := SnapshotGlobalConfig()
	if cfg.DefaultUserAgent != "snapshot-agent" || cfg.Timeout != 9*time.Second || cfg.FollowRedirects || cfg.MaxRedirects != 2 || cfg.MaxResponseBytes != 321 {
		t.Fatalf("snapshot scalar config = %#v", cfg)
	}
	if cfg.CookieJar != nil || cfg.Headers.Get("X-Snapshot") != "old" {
		t.Fatalf("snapshot mutable config cookie=%v header=%q", cfg.CookieJar, cfg.Headers.Get("X-Snapshot"))
	}
	cfg.Headers.Set("X-Snapshot", "changed")
	if got := CloneGlobalHeaders().Get("X-Snapshot"); got != "old" {
		t.Fatalf("snapshot headers should be cloned; global header = %q", got)
	}
}

func TestResetGlobalConfigRestoresDefaults(t *testing.T) {
	previous := SnapshotGlobalConfig()
	defer ConfigureGlobalConfig(previous)

	SetGlobalTimeout(9 * time.Second)
	SetGlobalMaxRedirects(2)
	SetGlobalMaxResponseBytes(3)
	SetGlobalFollowRedirects(false)
	SetIgnoreEOFError(false)
	SetGlobalDecodeURL(true)
	SetGlobalUserAgent("mutated-agent")
	SetGlobalBoundary("mutated-boundary")
	SetGlobalHeader("X-Reset", "mutated")
	CloseCookie()

	ResetGlobalConfig()
	cfg := SnapshotGlobalConfig()
	if cfg.Timeout != 0 || cfg.MaxRedirects != 10 || cfg.MaxResponseBytes != defaultGlobalMaxResponseBytes || !cfg.FollowRedirects || !cfg.IgnoreEOFError || cfg.DecodeURL || cfg.DefaultUserAgent != "" || cfg.Boundary != "--------------------gokitFormBoundary" {
		t.Fatalf("reset scalar config = %#v", cfg)
	}
	if cfg.Headers.Get("X-Reset") != "" || cfg.Headers.Get("User-Agent") == "" || cfg.CookieJar == nil {
		t.Fatalf("reset mutable config header=%q ua=%q cookie=%v", cfg.Headers.Get("X-Reset"), cfg.Headers.Get("User-Agent"), cfg.CookieJar)
	}
}

func TestWithScopedGlobalConfigRestoresPreviousDefaults(t *testing.T) {
	previous := SnapshotGlobalConfig()
	defer ConfigureGlobalConfig(previous)

	ConfigureGlobalConfig(GlobalConfig{
		Timeout:          time.Second,
		MaxRedirects:     4,
		MaxResponseBytes: 64,
		IgnoreEOFError:   true,
		FollowRedirects:  true,
		DefaultUserAgent: "outer-agent",
		Boundary:         "outer-boundary",
		Headers:          http.Header{"X-Scope": []string{"outer"}},
		CookieJar:        previous.CookieJar,
	})

	WithScopedGlobalConfig(GlobalConfig{
		Timeout:          2 * time.Second,
		MaxRedirects:     1,
		MaxResponseBytes: 32,
		IgnoreEOFError:   false,
		FollowRedirects:  false,
		DefaultUserAgent: "inner-agent",
		Boundary:         "inner-boundary",
		Headers:          http.Header{"X-Scope": []string{"inner"}},
		CookieJar:        nil,
	}, func() {
		cfg := SnapshotGlobalConfig()
		if cfg.Timeout != 2*time.Second || cfg.MaxRedirects != 1 || cfg.MaxResponseBytes != 32 || cfg.FollowRedirects || cfg.IgnoreEOFError || cfg.DefaultUserAgent != "inner-agent" || cfg.Headers.Get("X-Scope") != "inner" || cfg.CookieJar != nil {
			t.Fatalf("scoped inner config = %#v", cfg)
		}
	})

	cfg := SnapshotGlobalConfig()
	if cfg.Timeout != time.Second || cfg.MaxRedirects != 4 || cfg.MaxResponseBytes != 64 || !cfg.FollowRedirects || !cfg.IgnoreEOFError || cfg.DefaultUserAgent != "outer-agent" || cfg.Boundary != "outer-boundary" || cfg.Headers.Get("X-Scope") != "outer" || cfg.CookieJar != previous.CookieJar {
		t.Fatalf("scoped restored config = %#v", cfg)
	}
}
