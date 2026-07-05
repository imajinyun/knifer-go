package resty

import (
	"sync"
	"testing"
	"time"
)

func TestDefaultGlobalTimeoutIsBounded(t *testing.T) {
	previous := SnapshotGlobalConfig()
	defer ConfigureGlobalConfig(previous)

	ResetGlobalConfig()
	if got := GetGlobalTimeout(); got != defaultGlobalTimeout || got <= 0 {
		t.Fatalf("default timeout = %v, want positive %v", got, defaultGlobalTimeout)
	}
}

func TestResetGlobalConfigRestoresDefaults(t *testing.T) {
	previous := SnapshotGlobalConfig()
	defer ConfigureGlobalConfig(previous)

	SetGlobalTimeout(time.Second)
	SetGlobalMaxRedirects(2)
	SetGlobalMaxResponseBytes(3)
	SetGlobalFollowRedirects(false)
	SetGlobalUserAgent("mutated-agent")
	SetGlobalHeader("X-Reset", "mutated")
	CloseCookie()

	ResetGlobalConfig()
	cfg := SnapshotGlobalConfig()
	if cfg.Timeout != defaultGlobalTimeout || cfg.MaxRedirects != 10 || cfg.MaxResponseBytes != defaultGlobalMaxResponseBytes || !cfg.FollowRedirects || cfg.DefaultUserAgent != "" || cfg.CookieDisabled {
		t.Fatalf("reset scalar config = %#v", cfg)
	}
	if got := cfg.Headers["X-Reset"]; len(got) != 0 {
		t.Fatalf("reset retained X-Reset header: %v", got)
	}
	if got := cfg.Headers[string(HeaderUserAgent)]; len(got) == 0 || got[0] == "" {
		t.Fatalf("reset default User-Agent header = %v", got)
	}
}

func TestWithScopedGlobalConfigRestoresPreviousDefaults(t *testing.T) {
	previous := SnapshotGlobalConfig()
	defer ConfigureGlobalConfig(previous)

	ConfigureGlobalConfig(GlobalConfig{
		Timeout:          time.Second,
		MaxRedirects:     4,
		MaxResponseBytes: 64,
		FollowRedirects:  true,
		DefaultUserAgent: "outer-agent",
		Headers:          HeaderValues{"X-Scope": []string{"outer"}},
	})

	WithScopedGlobalConfig(GlobalConfig{
		Timeout:          2 * time.Second,
		MaxRedirects:     1,
		MaxResponseBytes: 32,
		FollowRedirects:  false,
		DefaultUserAgent: "inner-agent",
		Headers:          HeaderValues{"X-Scope": []string{"inner"}},
		CookieDisabled:   true,
	}, func() {
		cfg := SnapshotGlobalConfig()
		if cfg.Timeout != 2*time.Second || cfg.MaxRedirects != 1 || cfg.MaxResponseBytes != 32 || cfg.FollowRedirects || cfg.DefaultUserAgent != "inner-agent" || cfg.Headers["X-Scope"][0] != "inner" || !cfg.CookieDisabled {
			t.Fatalf("scoped inner config = %#v", cfg)
		}
	})

	cfg := SnapshotGlobalConfig()
	if cfg.Timeout != time.Second || cfg.MaxRedirects != 4 || cfg.MaxResponseBytes != 64 || !cfg.FollowRedirects || cfg.DefaultUserAgent != "outer-agent" || cfg.Headers["X-Scope"][0] != "outer" || cfg.CookieDisabled {
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
					FollowRedirects:  idx%2 == 0,
					DefaultUserAgent: "resty-agent",
					Headers:          HeaderValues{"X-Concurrent": []string{"configured"}},
					CookieDisabled:   idx%2 == 1,
				})
				SetGlobalHeader("X-Concurrent", "set")
				AddGlobalHeader("X-Concurrent", "add")
				RemoveGlobalHeader("X-Removed")
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
				_ = CloneGlobalHeaders()
			}
		}()
	}
	wg.Wait()
}
