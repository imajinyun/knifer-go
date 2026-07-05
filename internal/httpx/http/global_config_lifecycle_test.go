package http

import (
	"net/http"
	"sync"
	"testing"
	"time"
)

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
	if cfg.Timeout != defaultGlobalTimeout || cfg.MaxRedirects != 10 || cfg.MaxResponseBytes != defaultGlobalMaxResponseBytes || !cfg.FollowRedirects || !cfg.IgnoreEOFError || cfg.DecodeURL || cfg.DefaultUserAgent != "" || cfg.Boundary != "--------------------gokitFormBoundary" {
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

func TestGlobalConfigConcurrentMutationAndSnapshot(t *testing.T) {
	previous := SnapshotGlobalConfig()
	t.Cleanup(func() { ConfigureGlobalConfig(previous) })

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		idx := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				ConfigureGlobalConfig(GlobalConfig{
					Timeout:          time.Duration(idx+1) * time.Second,
					MaxRedirects:     idx + 1,
					MaxResponseBytes: int64(idx + 1),
					IgnoreEOFError:   idx%2 == 0,
					DecodeURL:        idx%2 == 1,
					FollowRedirects:  idx%2 == 0,
					DefaultUserAgent: "agent",
					Boundary:         "boundary",
					Headers:          http.Header{"X-Concurrent": []string{"configured"}},
					CookieJar:        previous.CookieJar,
				})
				SetGlobalHeader("X-Concurrent", "set")
				AddGlobalHeader("X-Concurrent", "add")
				RemoveGlobalHeader("X-Removed")
				SetCookieJar(previous.CookieJar)
				if idx%3 == 0 {
					CloseCookie()
				}
			}
		}()
	}
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				cfg := SnapshotGlobalConfig()
				if cfg.Headers == nil {
					t.Error("snapshot headers should not be nil")
				}
				_ = GetGlobalTimeout()
				_ = GetGlobalMaxRedirects()
				_ = GetGlobalMaxResponseBytes()
				_ = GetGlobalFollowRedirects()
				_ = GetGlobalUserAgent()
				_ = IsIgnoreEOFError()
				_ = IsGlobalDecodeURL()
				_ = GetGlobalBoundary()
				_ = CloneGlobalHeaders()
				_ = GetCookieJar()
			}
		}()
	}
	wg.Wait()
}
