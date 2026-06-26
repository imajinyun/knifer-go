package vresty_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/imajinyun/knifer-go/vresty"
)

func TestFacadeRequestGlobalHelpers(t *testing.T) {
	previous := vresty.SnapshotGlobalConfig()
	defer vresty.ConfigureGlobalConfig(previous)
	vresty.SetGlobalMaxRedirects(4)
	vresty.SetGlobalMaxResponseBytes(123)
	vresty.SetGlobalFollowRedirects(false)
	vresty.SetGlobalUserAgent("vresty-extra/1.0")
	vresty.CloseCookie()
	vresty.SetGlobalHeader("X-Extra", "one")
	vresty.RemoveGlobalHeader("X-Extra")
	cfg := vresty.SnapshotGlobalConfig()
	if vresty.GetGlobalMaxRedirects() != 4 || vresty.GetGlobalMaxResponseBytes() != 123 || vresty.GetGlobalFollowRedirects() || vresty.GetGlobalUserAgent() != "vresty-extra/1.0" || !cfg.CookieDisabled || len(cfg.Headers["X-Extra"]) != 0 {
		t.Fatalf("global config = %#v", cfg)
	}
}

func TestFacadeRequestGlobalConfigAPIs(t *testing.T) {
	previous := vresty.SnapshotGlobalConfig()
	defer vresty.ConfigureGlobalConfig(previous)

	vresty.SetGlobalTimeout(321 * time.Millisecond)
	vresty.SetGlobalHeader("X-Facade-Config", "global")

	cfg := vresty.SnapshotGlobalConfig()
	cfg.Headers["X-Facade-Config"][0] = "snapshot"
	cfg.DefaultUserAgent = "facade-config-agent"
	cfg.Headers["User-Agent"] = []string{"facade-config-agent"}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Facade-Config") + ":" + r.Header.Get("User-Agent")))
	}))
	defer srv.Close()

	resp := vresty.NewRequestWithConfig(vresty.MethodGet, srv.URL, cfg).Execute()
	if resp.Err() != nil {
		t.Fatal(resp.Err())
	}
	if got := resp.Body(); got != "snapshot:facade-config-agent" {
		t.Fatalf("NewRequestWithConfig body = %q", got)
	}

	resp = vresty.NewIsolatedRequest(vresty.MethodGet, srv.URL, vresty.WithGlobalConfig(cfg)).Execute()
	if resp.Err() != nil {
		t.Fatal(resp.Err())
	}
	if got := resp.Body(); got != "snapshot:facade-config-agent" {
		t.Fatalf("NewIsolatedRequest WithGlobalConfig body = %q", got)
	}
}
