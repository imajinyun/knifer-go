package vresty_test

import (
	"bytes"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/imajinyun/go-knifer/vresty"
	grestry "resty.dev/v3"
)

func TestFacadeGetString(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("facade"))
	}))
	defer srv.Close()

	if got := vresty.GetString(srv.URL); got != "facade" {
		t.Fatalf("GetString() = %q, want facade", got)
	}
}

func TestFacadeBuildBasicAuth(t *testing.T) {
	if got := vresty.BuildBasicAuth("u", "p"); got != "Basic dTpw" {
		t.Fatalf("BuildBasicAuth() = %q, want Basic dTpw", got)
	}
}

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

func TestFacadeRequestOptions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Opt") + ":" + r.Header.Get("User-Agent")))
	}))
	defer srv.Close()

	resp := vresty.Get(srv.URL,
		vresty.WithHeader("X-Opt", "yes"),
		vresty.WithUserAgent("vresty-test/1.0"),
	).Execute()
	if resp.Err() != nil {
		t.Fatal(resp.Err())
	}
	if got := resp.Body(); got != "yes:vresty-test/1.0" {
		t.Fatalf("Body() = %q, want option headers", got)
	}
}

func TestFacadeRestyClientFactoryProvider(t *testing.T) {
	vresty.ResetDefaultRestyClientProvider()
	t.Cleanup(vresty.ResetDefaultRestyClientProvider)

	called := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Factory")))
	}))
	defer server.Close()

	resp := vresty.Get(server.URL,
		vresty.WithRestyClientFactory(func() *grestry.Client {
			called++
			return grestry.New().SetHeader("X-Factory", "per-call")
		}),
	).Execute()
	if resp.Err() != nil {
		t.Fatal(resp.Err())
	}
	if called != 1 || resp.Body() != "per-call" {
		t.Fatalf("factory called=%d body=%q", called, resp.Body())
	}

	vresty.ConfigureDefaultRestyClientProvider(func() *grestry.Client {
		called++
		return grestry.New().SetHeader("X-Factory", "default")
	})
	resp = vresty.Get(server.URL).Execute()
	if resp.Err() != nil {
		t.Fatal(resp.Err())
	}
	if called != 2 || resp.Body() != "default" {
		t.Fatalf("default provider called=%d body=%q", called, resp.Body())
	}
}

func TestFacadeCreateWithOptions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/final", http.StatusFound)
			return
		}
		_, _ = w.Write([]byte(r.Method + ":" + r.Header.Get("X-Create")))
	}))
	defer srv.Close()

	getResp := vresty.CreateGetWithOptions(srv.URL+"/redirect", false, vresty.WithHeader("X-Create", "get")).Execute()
	if getResp.Err() != nil {
		t.Fatal(getResp.Err())
	}
	if got := getResp.Status(); got != http.StatusFound {
		t.Fatalf("CreateGetWithOptions status = %d, want 302", got)
	}

	postResp := vresty.CreatePostWithOptions(srv.URL, vresty.WithHeader("X-Create", "post")).Execute()
	if postResp.Err() != nil {
		t.Fatal(postResp.Err())
	}
	if got := postResp.Body(); got != "POST:post" {
		t.Fatalf("CreatePostWithOptions body = %q, want POST:post", got)
	}
}

func TestFacadeUtilityWrappers(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			body, _ := io.ReadAll(r.Body)
			_, _ = w.Write([]byte("post:" + string(body) + ":" + r.Header.Get("X-Util")))
			return
		}
		_, _ = w.Write([]byte(r.URL.Query().Get("q") + ":" + r.Header.Get("X-Util")))
	}))
	defer srv.Close()

	if got := vresty.GetWithParamsWithOptions(srv.URL, map[string]any{"q": "go"}, vresty.WithHeader("X-Util", "get")); got != "go:get" {
		t.Fatalf("GetWithParamsWithOptions() = %q, want go:get", got)
	}
	if got := vresty.PostStringWithOptions(srv.URL, "body", vresty.WithHeader("X-Util", "post")); got != "post:body:post" {
		t.Fatalf("PostStringWithOptions() = %q, want post:body:post", got)
	}
	if !vresty.IsHTTP("http://example.com") || !vresty.IsHTTPS("https://example.com") {
		t.Fatal("IsHTTP/IsHTTPS wrappers returned false")
	}
	if got := vresty.ToParams(map[string]any{"q": "go"}); got != "q=go" {
		t.Fatalf("ToParams() = %q, want q=go", got)
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
	if cfg.Timeout != 0 || cfg.MaxRedirects != 10 || cfg.MaxResponseBytes != 64<<20 || !cfg.FollowRedirects || len(cfg.Headers["X-Facade-Scope"]) != 0 || cfg.CookieDisabled {
		t.Fatalf("facade config not restored after scoped helper: %#v", cfg)
	}
}

func TestFacadeSaveProviderOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("vresty-save"))
	}))
	defer server.Close()

	var mkdirPath string
	var mkdirPerm fs.FileMode
	var openPath string
	var openFlag int
	var openPerm fs.FileMode
	var written bytes.Buffer
	n, err := vresty.Get(server.URL).Execute().SaveAs("/virtual/out.txt",
		vresty.WithSaveMkdirAll(func(path string, perm fs.FileMode) error {
			mkdirPath, mkdirPerm = path, perm
			return nil
		}),
		vresty.WithSaveOpenFile(func(path string, flag int, perm fs.FileMode) (io.WriteCloser, error) {
			openPath, openFlag, openPerm = path, flag, perm
			return nopWriteCloser{Writer: &written}, nil
		}),
		vresty.WithSaveDirPerm(0o700), vresty.WithSaveFilePerm(0o600),
	)
	if err != nil || n != int64(len("vresty-save")) {
		t.Fatalf("SaveAs provider n=%d err=%v", n, err)
	}
	if mkdirPath != "/virtual" || mkdirPerm != 0o700 || openPath != "/virtual/out.txt" || openPerm != 0o600 || openFlag&os.O_CREATE == 0 || written.String() != "vresty-save" {
		t.Fatalf("providers mkdir=%q/%v open=%q flag=%#x perm=%v content=%q", mkdirPath, mkdirPerm, openPath, openFlag, openPerm, written.String())
	}
}

type nopWriteCloser struct{ io.Writer }

func (w nopWriteCloser) Close() error { return nil }
