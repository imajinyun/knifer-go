package vresty_test

import (
	"testing"
	"time"

	"github.com/imajinyun/go-knifer/vresty"
)

func TestFacadeCloneGlobalHeaders(t *testing.T) {
	previous := vresty.SnapshotGlobalConfig()
	defer vresty.ConfigureGlobalConfig(previous)

	vresty.SetGlobalHeader("X-Facade", "one")
	vresty.AddGlobalHeader("X-Facade", "two")

	headers := vresty.CloneGlobalHeaders()
	if got := headers["X-Facade"]; len(got) != 2 || got[0] != "one" || got[1] != "two" {
		t.Fatalf("CloneGlobalHeaders()[X-Facade] = %v, want [one two]", got)
	}
}

func TestFacadeScopedGlobalConfig(t *testing.T) {
	previous := vresty.SnapshotGlobalConfig()
	defer vresty.ConfigureGlobalConfig(previous)

	vresty.ResetGlobalConfig()
	vresty.WithScopedGlobalConfig(vresty.GlobalConfig{
		Timeout:          3 * time.Second,
		MaxRedirects:     1,
		MaxResponseBytes: 32,
		FollowRedirects:  false,
		DefaultUserAgent: "facade-scope-agent",
		Headers:          vresty.HeaderValues{"X-Facade-Scope": []string{"inner"}},
		CookieDisabled:   true,
	}, func() {
		cfg := vresty.SnapshotGlobalConfig()
		if cfg.Timeout != 3*time.Second || cfg.MaxRedirects != 1 || cfg.MaxResponseBytes != 32 || cfg.FollowRedirects || cfg.DefaultUserAgent != "facade-scope-agent" || cfg.Headers["X-Facade-Scope"][0] != "inner" || !cfg.CookieDisabled {
			t.Fatalf("facade scoped config = %#v", cfg)
		}
	})

	cfg := vresty.SnapshotGlobalConfig()
	if cfg.Timeout != 30*time.Second || cfg.MaxRedirects != 10 || cfg.MaxResponseBytes != 64<<20 || !cfg.FollowRedirects || len(cfg.Headers["X-Facade-Scope"]) != 0 || cfg.CookieDisabled {
		t.Fatalf("facade config not restored after scoped helper: %#v", cfg)
	}
}
