package http

import (
	"testing"
	"time"
)

func TestGlobalTimeout(t *testing.T) {
	old := GetGlobalTimeout()
	defer SetGlobalTimeout(old)

	SetGlobalTimeout(7 * time.Second)
	if got := GetGlobalTimeout(); got != 7*time.Second {
		t.Fatalf("timeout: %v", got)
	}
}

func TestDefaultGlobalTimeoutIsBounded(t *testing.T) {
	previous := SnapshotGlobalConfig()
	defer ConfigureGlobalConfig(previous)

	ResetGlobalConfig()
	if got := GetGlobalTimeout(); got != defaultGlobalTimeout || got <= 0 {
		t.Fatalf("default timeout = %v, want positive %v", got, defaultGlobalTimeout)
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
