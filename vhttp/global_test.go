package vhttp_test

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/imajinyun/knifer-go/vhttp"
	"github.com/imajinyun/knifer-go/vurl"
)

func TestFacadeHelperNamesWithoutHTTPPrefix(t *testing.T) {
	previous := vhttp.SnapshotGlobalConfig()
	defer vhttp.ConfigureGlobalConfig(previous)

	vhttp.SetGlobalTimeout(2 * time.Second)
	if got := vhttp.GetGlobalTimeout(); got != 2*time.Second {
		t.Fatalf("GetGlobalTimeout() = %v, want 2s", got)
	}

	vhttp.SetGlobalHeader("X-Test", "a")
	vhttp.AddGlobalHeader("X-Test", "b")
	if got := vhttp.CloneGlobalHeaders().Values("X-Test"); len(got) != 2 {
		t.Fatalf("CloneGlobalHeaders().Values(X-Test) = %v, want 2 values", got)
	}
	vhttp.RemoveGlobalHeader("X-Test")
	if got := vhttp.CloneGlobalHeaders().Values("X-Test"); len(got) != 0 {
		t.Fatalf("after RemoveGlobalHeader values = %v, want empty", got)
	}

	if got := vhttp.BuildBasicAuth("aladdin", "opensesame"); got != "Basic YWxhZGRpbjpvcGVuc2VzYW1l" {
		t.Fatalf("BuildBasicAuth() = %q", got)
	}
	if got := vurl.EncodeQueryMap(map[string]any{"q": "go", "page": 1}); !strings.Contains(got, "q=go") || !strings.Contains(got, "page=1") {
		t.Fatalf("EncodeQueryMap() = %q", got)
	}
}

func TestFacadeScopedGlobalConfig(t *testing.T) {
	previous := vhttp.SnapshotGlobalConfig()
	defer vhttp.ConfigureGlobalConfig(previous)

	vhttp.ResetGlobalConfig()
	vhttp.WithScopedGlobalConfig(vhttp.GlobalConfig{
		Timeout:          3 * time.Second,
		MaxRedirects:     1,
		MaxResponseBytes: 32,
		IgnoreEOFError:   true,
		FollowRedirects:  false,
		DefaultUserAgent: "facade-scope-agent",
		Boundary:         "facade-boundary",
		Headers:          http.Header{"X-Facade-Scope": []string{"inner"}},
		CookieJar:        nil,
	}, func() {
		cfg := vhttp.SnapshotGlobalConfig()
		if cfg.Timeout != 3*time.Second || cfg.MaxRedirects != 1 || cfg.MaxResponseBytes != 32 || cfg.FollowRedirects || cfg.DefaultUserAgent != "facade-scope-agent" || cfg.Headers.Get("X-Facade-Scope") != "inner" || cfg.CookieJar != nil {
			t.Fatalf("facade scoped config = %#v", cfg)
		}
	})

	cfg := vhttp.SnapshotGlobalConfig()
	if cfg.Timeout != 30*time.Second || cfg.MaxRedirects != 10 || cfg.MaxResponseBytes != 64<<20 || !cfg.FollowRedirects || cfg.Headers.Get("X-Facade-Scope") != "" || cfg.CookieJar == nil {
		t.Fatalf("facade config not restored after scoped helper: %#v", cfg)
	}
}
