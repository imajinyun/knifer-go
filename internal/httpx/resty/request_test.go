package resty

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	knifer "github.com/imajinyun/go-knifer"
	grestry "resty.dev/v3"
)

func TestGetWithQueryAndHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("q") != "go" {
			t.Fatalf("query q = %q, want go", r.URL.Query().Get("q"))
		}
		if r.Header.Get("X-Test") != "yes" {
			t.Fatalf("X-Test = %q, want yes", r.Header.Get("X-Test"))
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	resp := Get(srv.URL).Query("q", "go").Header("X-Test", "yes").Execute()
	if resp.Err() != nil {
		t.Fatalf("Execute() error = %v", resp.Err())
	}
	if !resp.IsOK() || resp.Body() != "ok" {
		t.Fatalf("status/body = %d/%q, want 2xx/ok", resp.Status(), resp.Body())
	}
}

func TestPostForm(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm() error = %v", err)
		}
		if got := r.Form.Get("name"); got != "resty" {
			t.Fatalf("form name = %q, want resty", got)
		}
		_, _ = w.Write([]byte("posted"))
	}))
	defer srv.Close()

	got, err := PostFormE(srv.URL, map[string]any{"name": "resty"})
	if err != nil {
		t.Fatalf("PostFormE() error = %v", err)
	}
	if got != "posted" {
		t.Fatalf("PostFormE() = %q, want posted", got)
	}
}

func TestPostJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Header.Get("Content-Type"), string(ContentTypeJSON)) {
			t.Fatalf("Content-Type = %q, want application/json", r.Header.Get("Content-Type"))
		}
		_, _ = w.Write([]byte("json"))
	}))
	defer srv.Close()

	got, err := PostJSONE(srv.URL, `{"ok":true}`)
	if err != nil {
		t.Fatalf("PostJSONE() error = %v", err)
	}
	if got != "json" {
		t.Fatalf("PostJSONE() = %q, want json", got)
	}
}

func TestRequestJSONMarshalProvider(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	called := false
	resp := Post(srv.URL, WithJSONMarshalFunc(func(any) ([]byte, error) {
		called = true
		return []byte(`{"provided":true}`), nil
	})).BodyJSONValue(map[string]any{"ignored": true}).Execute()
	if resp.Err() != nil {
		t.Fatalf("Execute() error = %v", resp.Err())
	}
	if !called || resp.Body() != `{"provided":true}` {
		t.Fatalf("marshal provider called=%v body=%q", called, resp.Body())
	}
}

func TestRequestJSONUnmarshalProvider(t *testing.T) {
	type result struct {
		Name string `json:"name"`
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"ignored"}`))
	}))
	defer srv.Close()

	called := false
	out := &result{}
	resp := Get(srv.URL, WithJSONUnmarshalFunc(func(_ []byte, dst any) error {
		called = true
		return json.Unmarshal([]byte(`{"name":"provided"}`), dst)
	})).Result(out).Execute()
	if resp.Err() != nil {
		t.Fatalf("Execute() error = %v", resp.Err())
	}
	if !called || out.Name != "provided" || resp.Result() == nil {
		t.Fatalf("unmarshal provider called=%v result=%+v raw=%v", called, out, resp.Result())
	}
}

func TestRequestJSONDecodeReadOptions(t *testing.T) {
	type result struct {
		Name string `json:"name"`
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"abcdef"}`))
	}))
	defer srv.Close()

	tooLarge := &result{}
	resp := Get(srv.URL,
		WithMaxDecodeBytes(3),
		WithJSONUnmarshalFunc(json.Unmarshal),
	).Result(tooLarge).Execute()
	if resp.Err() == nil {
		t.Fatal("Execute() with max decode bytes error = nil")
	}

	readAllCalled := false
	out := &result{}
	resp = Get(srv.URL,
		WithJSONDecodeReadAllFunc(func(io.Reader) ([]byte, error) {
			readAllCalled = true
			return []byte(`{"name":"provided"}`), nil
		}),
		WithJSONUnmarshalFunc(json.Unmarshal),
	).Result(out).Execute()
	if resp.Err() != nil || !readAllCalled || out.Name != "provided" {
		t.Fatalf("custom decode readAll called=%v out=%+v err=%v", readAllCalled, out, resp.Err())
	}
}

func TestResponseReadLimitOptions(t *testing.T) {
	previous := SnapshotGlobalConfig()
	defer ConfigureGlobalConfig(previous)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("abcdef"))
	}))
	defer srv.Close()

	limited := Get(srv.URL, WithMaxResponseBytes(3)).Execute()
	if got := limited.Bytes(); len(got) != 0 || limited.Err() == nil {
		t.Fatalf("limited Bytes() = %q err=%v, want max bytes error", string(got), limited.Err())
	}

	SetGlobalMaxResponseBytes(3)
	globalLimited := Get(srv.URL).Execute()
	SetGlobalMaxResponseBytes(0)
	if got := globalLimited.Bytes(); len(got) != 0 || globalLimited.Err() == nil {
		t.Fatalf("global limited Bytes() = %q err=%v, want max bytes error", string(got), globalLimited.Err())
	}

	unlimited := Get(srv.URL, WithMaxResponseBytes(0)).Execute()
	if got := unlimited.Body(); got != "abcdef" || unlimited.Err() != nil {
		t.Fatalf("unlimited override Body() = %q err=%v", got, unlimited.Err())
	}
}

func TestResponseHeadersCookiesAndLength(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Cookie"); !strings.Contains(got, "k=v") {
			t.Fatalf("Cookie = %q, want k=v", got)
		}
		w.Header().Set("X-Test", "yes")
		w.Header().Add("Set-Cookie", "sid=abc; Path=/")
		_, _ = w.Write([]byte("hello"))
	}))
	defer srv.Close()

	resp := Get(srv.URL).Cookie("k", "v").Execute()
	if resp.Err() != nil {
		t.Fatalf("Execute() error = %v", resp.Err())
	}
	if got := resp.Headers()["X-Test"]; len(got) != 1 || got[0] != "yes" {
		t.Fatalf("Headers()[X-Test] = %v, want [yes]", got)
	}
	cookies := resp.Cookies()
	if len(cookies) != 1 || cookies[0].Name != "sid" || cookies[0].Value != "abc" {
		t.Fatalf("Cookies() = %+v, want sid=abc", cookies)
	}
	if got := resp.ContentLength(); got != int64(len("hello")) {
		t.Fatalf("ContentLength() = %d, want %d", got, len("hello"))
	}
}

func TestGlobalHeadersArePlainValues(t *testing.T) {
	SetGlobalHeader("X-Resty-Plain", "one")
	AddGlobalHeader("X-Resty-Plain", "two")
	defer RemoveGlobalHeader("X-Resty-Plain")

	headers := CloneGlobalHeaders()
	if got := headers["X-Resty-Plain"]; len(got) != 2 || got[0] != "one" || got[1] != "two" {
		t.Fatalf("CloneGlobalHeaders()[X-Resty-Plain] = %v, want [one two]", got)
	}
}

func TestTimeoutReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		_, _ = w.Write([]byte("late"))
	}))
	defer srv.Close()

	resp := Get(srv.URL).Timeout(time.Millisecond).Execute()
	if resp.Err() == nil {
		t.Fatal("Execute() error is nil, want timeout error")
	}
}

func TestStringHelpersReturnErrorsExplicitly(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Method + ":" + r.URL.Query().Get("k")))
	}))
	defer srv.Close()

	body, err := GetWithParamsE(srv.URL, map[string]any{"k": "v"})
	if err != nil || body != "GET:v" {
		t.Fatalf("GetWithParamsE = %q, %v", body, err)
	}

	if body, err = PostStringE(srv.URL, "payload"); err != nil || body != "POST:" {
		t.Fatalf("PostStringE = %q, %v", body, err)
	}

	if _, err = GetStringE("http://[::1"); err == nil {
		t.Fatal("GetStringE invalid URL error = nil")
	}
	if _, err = DownloadBytesE("http://[::1"); err == nil {
		t.Fatal("DownloadBytesE invalid URL error = nil")
	}
	if _, err = GetStringSafeE(srv.URL); err == nil {
		t.Fatal("GetStringSafeE local URL error = nil, want private address rejection")
	}
}

func TestRequestOptionsOverrideGlobalDefaults(t *testing.T) {
	oldUA := GetGlobalUserAgent()
	oldFollow := GetGlobalFollowRedirects()
	defer SetGlobalUserAgent(oldUA)
	defer SetGlobalFollowRedirects(oldFollow)

	SetGlobalUserAgent("global-resty-agent")
	SetGlobalFollowRedirects(false)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/start" {
			http.Redirect(w, r, "/end", http.StatusFound)
			return
		}
		_, _ = w.Write([]byte(r.Header.Get("X-Req") + ":" + r.Header.Get("User-Agent")))
	}))
	defer srv.Close()

	resp := Get(srv.URL+"/start",
		WithHeader("X-Req", "per-call"),
		WithUserAgent("request-resty-agent"),
		WithFollowRedirects(true),
	).Execute()
	if resp.Err() != nil {
		t.Fatalf("Execute() error = %v", resp.Err())
	}
	if got := resp.Body(); got != "per-call:request-resty-agent" {
		t.Fatalf("Body() = %q, want per-call options to override globals", got)
	}
}

func TestClientUsesCapturedConfig(t *testing.T) {
	oldUA := GetGlobalUserAgent()
	oldFollow := GetGlobalFollowRedirects()
	defer SetGlobalUserAgent(oldUA)
	defer SetGlobalFollowRedirects(oldFollow)

	SetGlobalUserAgent("client-agent")
	SetGlobalFollowRedirects(false)
	client := NewClient()
	SetGlobalUserAgent("mutated-agent")
	SetGlobalFollowRedirects(true)

	req := client.Get("https://example.com")
	if req.userAgent != "client-agent" {
		t.Fatalf("client request userAgent = %q, want captured client-agent", req.userAgent)
	}
	if req.followRedir == nil || *req.followRedir {
		t.Fatalf("client request followRedirects = %v, want captured false", req.followRedir)
	}

	isolated := NewIsolatedClient().Get("https://example.com")
	if isolated.userAgent != "" || isolated.followRedir == nil || !*isolated.followRedir {
		t.Fatalf("isolated client defaults ua=%q follow=%v", isolated.userAgent, isolated.followRedir)
	}
}

func TestRestyClientFactoryProviderLifecycle(t *testing.T) {
	ResetDefaultRestyClientProvider()
	t.Cleanup(ResetDefaultRestyClientProvider)

	defaultCalled := 0
	ConfigureDefaultRestyClientProvider(func() *grestry.Client {
		defaultCalled++
		return grestry.New()
	})
	client := NewIsolatedRequest(MethodGet, "http://example.com").buildClient()
	if client == nil || defaultCalled != 1 {
		t.Fatalf("default provider client=%v called=%d", client, defaultCalled)
	}

	perCallCalled := 0
	client = NewIsolatedRequest(MethodGet, "http://example.com", WithRestyClientFactory(func() *grestry.Client {
		perCallCalled++
		return grestry.New()
	})).buildClient()
	if client == nil || perCallCalled != 1 || defaultCalled != 1 {
		t.Fatalf("per-call factory client=%v perCall=%d default=%d", client, perCallCalled, defaultCalled)
	}

	client = NewIsolatedRequest(MethodGet, "http://example.com", WithRestyClientFactory(func() *grestry.Client { return nil })).buildClient()
	if client == nil || defaultCalled != 2 {
		t.Fatalf("nil per-call factory client=%v default=%d", client, defaultCalled)
	}

	ResetDefaultRestyClientProvider()
	client = NewIsolatedRequest(MethodGet, "http://example.com").buildClient()
	if client == nil {
		t.Fatal("reset default provider should create a client")
	}
}

func TestNewRequestWithOptionsAppliesRequestOptions(t *testing.T) {
	getReq := Get("http://example.com", WithFollowRedirects(false), WithHeader("X-Create", "get"), WithUserAgent("create-get-agent"))
	if getReq.followRedir == nil || *getReq.followRedir {
		t.Fatalf("followRedir: %v", getReq.followRedir)
	}
	if got := getReq.headers["X-Create"]; len(got) != 1 || got[0] != "get" {
		t.Fatalf("Get header = %q, want get", got)
	}
	if got := getReq.userAgent; got != "create-get-agent" {
		t.Fatalf("Get userAgent = %q", got)
	}

	postReq := Post("http://example.com", WithHeader("X-Create", "post"))
	if postReq.method != MethodPost {
		t.Fatalf("Post method = %v, want POST", postReq.method)
	}
	if got := postReq.headers["X-Create"]; len(got) != 1 || got[0] != "post" {
		t.Fatalf("Post header = %q, want post", got)
	}
}

func TestSnapshotGlobalConfigAndExplicitRequestConfig(t *testing.T) {
	oldTimeout := GetGlobalTimeout()
	oldMaxRedirects := GetGlobalMaxRedirects()
	oldMaxResponse := GetGlobalMaxResponseBytes()
	oldFollow := GetGlobalFollowRedirects()
	oldUA := GetGlobalUserAgent()
	defer SetGlobalTimeout(oldTimeout)
	defer SetGlobalMaxRedirects(oldMaxRedirects)
	defer SetGlobalMaxResponseBytes(oldMaxResponse)
	defer SetGlobalFollowRedirects(oldFollow)
	defer SetGlobalUserAgent(oldUA)
	defer RemoveGlobalHeader("X-Snapshot")

	SetGlobalTimeout(123 * time.Millisecond)
	SetGlobalMaxRedirects(3)
	SetGlobalMaxResponseBytes(321)
	SetGlobalFollowRedirects(false)
	SetGlobalUserAgent("snapshot-agent")
	SetGlobalHeader("X-Snapshot", "one")

	cfg := SnapshotGlobalConfig()
	SetGlobalHeader("X-Snapshot", "mutated")
	cfg.Headers["X-Snapshot"][0] = "cfg"

	req := NewRequestWithConfig(MethodGet, "http://example.com", cfg)
	if req.timeout != 123*time.Millisecond || req.maxRedirects != 3 || req.maxResponse != 321 || req.followRedir == nil || *req.followRedir || req.userAgent != "snapshot-agent" {
		t.Fatalf("request config not applied: timeout=%v max=%d maxResponse=%d follow=%v ua=%q", req.timeout, req.maxRedirects, req.maxResponse, req.followRedir, req.userAgent)
	}
	if got := req.headers["X-Snapshot"]; len(got) != 1 || got[0] != "cfg" {
		t.Fatalf("explicit config headers = %v, want [cfg]", got)
	}
	if got := CloneGlobalHeaders()["X-Snapshot"]; len(got) != 1 || got[0] != "mutated" {
		t.Fatalf("snapshot should be detached from globals, global header = %v", got)
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

func TestNewIsolatedRequestDoesNotReadGlobals(t *testing.T) {
	oldTimeout := GetGlobalTimeout()
	oldMaxRedirects := GetGlobalMaxRedirects()
	oldFollow := GetGlobalFollowRedirects()
	oldUA := GetGlobalUserAgent()
	defer SetGlobalTimeout(oldTimeout)
	defer SetGlobalMaxRedirects(oldMaxRedirects)
	defer SetGlobalFollowRedirects(oldFollow)
	defer SetGlobalUserAgent(oldUA)
	defer RemoveGlobalHeader("X-Isolated")

	SetGlobalTimeout(time.Second)
	SetGlobalMaxRedirects(1)
	SetGlobalFollowRedirects(false)
	SetGlobalUserAgent("global-agent")
	SetGlobalHeader("X-Isolated", "global")

	req := NewIsolatedRequest(MethodGet, "http://example.com")
	if req.timeout != defaultGlobalTimeout || req.maxRedirects != 10 || req.maxResponse != defaultGlobalMaxResponseBytes || req.followRedir == nil || !*req.followRedir || req.userAgent != "" {
		t.Fatalf("isolated request leaked globals: timeout=%v max=%d maxResponse=%d follow=%v ua=%q", req.timeout, req.maxRedirects, req.maxResponse, req.followRedir, req.userAgent)
	}
	if got := req.headers["X-Isolated"]; len(got) != 0 {
		t.Fatalf("isolated request should not include global header: %v", got)
	}
}

func TestWithGlobalConfigOptionOverridesConstructionDefaults(t *testing.T) {
	cfg := GlobalConfig{
		Timeout:          250 * time.Millisecond,
		MaxRedirects:     2,
		MaxResponseBytes: 456,
		FollowRedirects:  false,
		DefaultUserAgent: "option-agent",
		Headers:          HeaderValues{"X-Config": []string{"yes"}},
	}
	req := NewIsolatedRequest(MethodGet, "http://example.com", WithGlobalConfig(cfg), WithHeader("X-Req", "ok"))
	if req.timeout != 250*time.Millisecond || req.maxRedirects != 2 || req.maxResponse != 456 || req.followRedir == nil || *req.followRedir || req.userAgent != "option-agent" {
		t.Fatalf("WithGlobalConfig not applied: timeout=%v max=%d maxResponse=%d follow=%v ua=%q", req.timeout, req.maxRedirects, req.maxResponse, req.followRedir, req.userAgent)
	}
	if got := req.headers["X-Config"]; len(got) != 1 || got[0] != "yes" {
		t.Fatalf("config header = %v, want [yes]", got)
	}
	if got := req.headers["X-Req"]; len(got) != 1 || got[0] != "ok" {
		t.Fatalf("request header after config = %v, want [ok]", got)
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

func TestRequestOptionContentTypeAndCharset(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("Content-Type")))
	}))
	defer srv.Close()

	got := Post(srv.URL, WithCharset("GBK"), WithContentType("text/custom")).BodyString("hello").Execute().Body()
	if got != "text/custom" {
		t.Fatalf("Content-Type = %q, want text/custom", got)
	}

	got = Post(srv.URL, WithCharset("GBK")).BodyJSON(`{"ok":true}`).Execute().Body()
	if got != "application/json;charset=GBK" {
		t.Fatalf("JSON Content-Type = %q", got)
	}
}

func TestRequestOptionTLSConfig(t *testing.T) {
	c := Get("https://example.com", WithTLSConfig(&tls.Config{ServerName: "example.com"})).buildClient()
	if c == nil {
		t.Fatal("client is nil")
	}
}

func TestSafeRequestRejectsPrivateAndUnsafeRedirects(t *testing.T) {
	if err := GetSafe("file:///tmp/secret.txt").Execute().Err(); err == nil {
		t.Fatal("GetSafe should reject non-HTTP schemes")
	}
	if err := GetSafe("http://127.0.0.1/config.yaml").Execute().Err(); err == nil {
		t.Fatal("GetSafe should reject loopback hosts by default")
	}
	if err := GetSafe("http://224.0.0.1/config.yaml").Execute().Err(); err == nil {
		t.Fatal("GetSafe should reject multicast hosts by default")
	}
	if err := GetSafe("http://0.0.0.0/config.yaml").Execute().Err(); err == nil {
		t.Fatal("GetSafe should reject unspecified hosts by default")
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "http://127.0.0.1/private", http.StatusFound)
			return
		}
		_, _ = w.Write([]byte("safe"))
	}))
	defer srv.Close()
	serverURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	resp := GetSafe(srv.URL,
		WithURLPolicy(URLPolicy{AllowedSchemes: []string{"http", "https"}, AllowedHosts: []string{serverURL.Hostname()}}),
	).Execute()
	if resp.Err() != nil || resp.Body() != "safe" {
		t.Fatalf("GetSafe allowed public policy host body=%q err=%v", resp.Body(), resp.Err())
	}
	if err := GetSafe(srv.URL+"/redirect",
		WithURLPolicy(URLPolicy{AllowedSchemes: []string{"http", "https"}, AllowedHosts: []string{serverURL.Hostname()}}),
	).Execute().Err(); err == nil {
		t.Fatal("GetSafe should reject unsafe redirect targets")
	}
}

func TestSafeRequestAllowedHostsDoesNotBypassPrivateRejection(t *testing.T) {
	if err := GetSafe("http://127.0.0.1/config.yaml", WithAllowedHosts("127.0.0.1")).Execute().Err(); err == nil {
		t.Fatal("GetSafe should reject allowlisted private hosts when RejectPrivate is enabled")
	}

	client := grestry.New().SetTransport(restyRoundTripperFunc(func(*http.Request) (*http.Response, error) {
		t.Fatal("unsafe request reached base transport")
		return nil, nil
	}))
	resp := GetSafe("http://example.com/config.yaml",
		WithAllowedHosts("example.com"),
		WithRestyClient(client),
		WithLookupIP(func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		}),
	).Execute()
	if resp.Err() == nil {
		t.Fatal("GetSafe should reject allowlisted hosts that resolve private during RoundTrip")
	}
}

func TestSafeRequestRevalidatesHostAtRoundTrip(t *testing.T) {
	lookups := [][]net.IP{{net.ParseIP("93.184.216.34")}, {net.ParseIP("127.0.0.1")}}
	lookupCount := 0
	client := grestry.New().SetTransport(restyRoundTripperFunc(func(*http.Request) (*http.Response, error) {
		t.Fatal("unsafe request reached base transport")
		return nil, nil
	}))
	resp := GetSafe("http://example.com/config.yaml",
		WithRestyClient(client),
		WithLookupIP(func(context.Context, string) ([]net.IP, error) {
			if lookupCount >= len(lookups) {
				return lookups[len(lookups)-1], nil
			}
			ips := lookups[lookupCount]
			lookupCount++
			return ips, nil
		}),
	).Execute()
	if resp.Err() == nil {
		t.Fatal("GetSafe should reject a host that resolves private during RoundTrip")
	}
	if lookupCount != 2 {
		t.Fatalf("lookup count = %d, want 2", lookupCount)
	}
}

func TestSaveAsOptions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("resty-save"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	target := filepath.Join(dir, "out.txt")
	if err := os.WriteFile(target, []byte("old"), 0o644); err != nil {
		t.Fatalf("write old: %v", err)
	}
	if _, err := Get(srv.URL).Execute().SaveAs(target, WithSaveOverwrite(false)); err == nil {
		t.Fatal("SaveAs overwrite false should fail")
	}
	if _, err := DownloadFile(srv.URL, target); err != nil {
		t.Fatalf("DownloadFile overwrite default: %v", err)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(data) != "resty-save" {
		t.Fatalf("content = %q", data)
	}
}

func TestSaveAsRejectsUnsafeContentDispositionFilename(t *testing.T) {
	tests := []string{
		`attachment; filename="../outside"`,
		`attachment; filename="..\outside"`,
		`attachment; filename="/tmp/outside"`,
	}
	for _, cd := range tests {
		t.Run(cd, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Disposition", cd)
				_, _ = w.Write([]byte("unsafe"))
			}))
			defer srv.Close()

			dir := t.TempDir()
			_, err := Get(srv.URL).Execute().SaveAs(dir)
			if !errors.Is(err, knifer.ErrCodeInvalidInput) {
				t.Fatalf("SaveAs error = %v, want invalid input", err)
			}
			if _, statErr := os.Stat(filepath.Join(dir, "outside")); !errors.Is(statErr, os.ErrNotExist) {
				t.Fatalf("unsafe file should not be created, stat err = %v", statErr)
			}
		})
	}
}

func TestSaveAsProviderOptions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("resty-provider-save"))
	}))
	defer srv.Close()

	var mkdirPath string
	var mkdirPerm fs.FileMode
	var openPath string
	var openFlag int
	var openPerm fs.FileMode
	var written bytes.Buffer
	n, err := Get(srv.URL).Execute().SaveAs("/virtual/resty.txt",
		WithSaveMkdirAll(func(path string, perm fs.FileMode) error {
			mkdirPath, mkdirPerm = path, perm
			return nil
		}),
		WithSaveOpenFile(func(path string, flag int, perm fs.FileMode) (io.WriteCloser, error) {
			openPath, openFlag, openPerm = path, flag, perm
			return nopWriteCloser{Writer: &written}, nil
		}),
		WithSaveDirPerm(0o700), WithSaveFilePerm(0o600),
	)
	if err != nil || n != int64(len("resty-provider-save")) {
		t.Fatalf("SaveAs provider n=%d err=%v", n, err)
	}
	if mkdirPath != "/virtual" || mkdirPerm != 0o700 || openPath != "/virtual/resty.txt" || openPerm != 0o600 || openFlag&os.O_CREATE == 0 || written.String() != "resty-provider-save" {
		t.Fatalf("providers mkdir=%q/%v open=%q flag=%#x perm=%v content=%q", mkdirPath, mkdirPerm, openPath, openFlag, openPerm, written.String())
	}
}

type nopWriteCloser struct{ io.Writer }

func (w nopWriteCloser) Close() error { return nil }

type restyRoundTripperFunc func(*http.Request) (*http.Response, error)

func (f restyRoundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestAdditionalClientFactoriesSafeWrappersAndMethods(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Method", r.Method)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Add("Set-Cookie", "sid=abc; Path=/")
		if r.Method != http.MethodHead {
			_, _ = w.Write([]byte(r.Method + ":" + r.Header.Get("X-Client")))
		}
	}))
	defer srv.Close()

	cfg := SnapshotGlobalConfig()
	cfg.Headers["X-Client"] = []string{"cfg"}
	client := NewClientWithConfig(cfg, WithHeader("X-Client", "opt"))
	if got := client.Get(srv.URL).Execute().Body(); got != "GET:opt" {
		t.Fatalf("client.Get body = %q", got)
	}
	if got := client.Post(srv.URL).Execute().Body(); got != "POST:opt" {
		t.Fatalf("client.Post body = %q", got)
	}
	if got := NewIsolatedClient(WithClientGlobalConfig(cfg), WithClientRequestOptions(WithHeader("X-Client", "isolated"))).NewRequest(MethodPut, srv.URL).Execute().Body(); got != "PUT:isolated" {
		t.Fatalf("NewIsolatedClient body = %q", got)
	}
	if got := (*Client)(nil).NewRequest(MethodDelete, srv.URL).Execute().Header("X-Method"); got != string(MethodDelete) {
		t.Fatalf("nil client NewRequest method = %q", got)
	}

	allowLocal := WithURLPolicy(URLPolicy{AllowedSchemes: []string{"http", "https"}, RejectPrivate: false})
	requests := []*HTTPRequest{
		PostSafe(srv.URL, allowLocal),
		Put(srv.URL),
		PutSafe(srv.URL, allowLocal),
		Delete(srv.URL),
		DeleteSafe(srv.URL, allowLocal),
		Patch(srv.URL),
		PatchSafe(srv.URL, allowLocal),
		Head(srv.URL),
		HeadSafe(srv.URL, allowLocal),
		Options(srv.URL),
		OptionsSafe(srv.URL, allowLocal),
		NewSafeRequest(MethodTrace, srv.URL, allowLocal),
		client.NewSafeRequest(MethodOptions, srv.URL, allowLocal),
	}
	for _, req := range requests {
		resp := req.Execute()
		if resp.Err() != nil {
			t.Fatalf("safe wrapper Execute: %v", resp.Err())
		}
		if resp.Status() == 0 {
			t.Fatal("safe wrapper status = 0")
		}
	}

	resp := Get(srv.URL).Cookie("k", "v").Execute()
	if resp.Err() != nil {
		t.Fatalf("cookie Execute: %v", resp.Err())
	}
	if got := resp.Headers()["X-Method"]; len(got) != 1 || got[0] != http.MethodGet {
		t.Fatalf("Headers()[X-Method] = %v", got)
	}
	if cookies := resp.Cookies(); len(cookies) != 1 || cookies[0].Name != "sid" {
		t.Fatalf("Cookies = %#v", cookies)
	}
	if resp.ContentType() == "" || resp.ContentLength() == 0 || resp.RestyRaw() == nil {
		t.Fatalf("response metadata type=%q length=%d raw=%v", resp.ContentType(), resp.ContentLength(), resp.RestyRaw())
	}
	var out bytes.Buffer
	if n, err := resp.WriteTo(&out); err != nil || n != int64(out.Len()) || !strings.Contains(out.String(), "GET") {
		t.Fatalf("WriteTo n=%d body=%q err=%v", n, out.String(), err)
	}
	if err := resp.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestAdditionalRequestAndUtilityWrappers(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_, _ = w.Write([]byte(r.Method + ":" + r.URL.Query().Get("q") + ":" + r.Header.Get("Authorization") + ":" + string(body)))
	}))
	defer srv.Close()

	req := NewIsolatedRequest(MethodGet, srv.URL).
		Method(MethodPost).
		URL(srv.URL).
		Headers(map[string]string{"X-A": "one"}).
		AddHeader("X-A", "two").
		CookieString("raw=cookie").
		ContentType(string(ContentTypeTextPlain)).
		Charset("utf-8").
		Timeout(time.Second).
		FollowRedirects(false).
		MaxRedirects(1).
		TLSConfig(&tls.Config{MinVersion: tls.VersionTLS12}).
		RestyClient(grestry.New()).
		URLPolicy(URLPolicy{AllowedSchemes: []string{"http", "https"}, RejectPrivate: false}).
		BasicAuth("user", "pass").
		BearerAuth("token").
		Query("q", "one").
		QueryMap(map[string]any{"q": "two"}).
		BodyReader(strings.NewReader("reader-body")).
		ErrorResult(&map[string]any{})
	if req.method != MethodPost || req.rawURL != srv.URL || req.urlPolicy == nil || req.errorResult == nil {
		t.Fatalf("request state method=%s url=%s policy=%#v", req.method, req.rawURL, req.urlPolicy)
	}
	resp := req.Execute()
	if resp.Err() != nil {
		t.Fatalf("Execute: %v", resp.Err())
	}
	if !strings.Contains(resp.Body(), "POST:two:Basic ") || !strings.Contains(resp.Body(), ":reader-body") {
		t.Fatalf("response body = %q", resp.Body())
	}

	if got, err := GetWithTimeoutE(srv.URL, time.Second); err != nil || !strings.HasPrefix(got, "GET:") {
		t.Fatalf("GetWithTimeoutE = %q, %v", got, err)
	}
	if got, err := GetWithTimeoutEWithOptions(srv.URL, time.Second, WithHeader("X-T", "v")); err != nil || !strings.HasPrefix(got, "GET:") {
		t.Fatalf("GetWithTimeoutEWithOptions = %q, %v", got, err)
	}
	if got, err := PostStringSafeE(srv.URL, "safe", WithURLPolicy(URLPolicy{AllowedSchemes: []string{"http", "https"}, RejectPrivate: false})); err != nil || !strings.Contains(got, "POST:::safe") {
		t.Fatalf("PostStringSafeE = %q, %v", got, err)
	}
	if got, err := PostFormSafeE(srv.URL, map[string]any{"a": "b"}, WithURLPolicy(URLPolicy{AllowedSchemes: []string{"http", "https"}, RejectPrivate: false})); err != nil || !strings.HasPrefix(got, "POST:") {
		t.Fatalf("PostFormSafeE = %q, %v", got, err)
	}
	if got, err := PostJSONSafeE(srv.URL, `{"ok":true}`, WithURLPolicy(URLPolicy{AllowedSchemes: []string{"http", "https"}, RejectPrivate: false})); err != nil || !strings.Contains(got, `{"ok":true}`) {
		t.Fatalf("PostJSONSafeE = %q, %v", got, err)
	}

	if !IsHTTP("http://example.com") || !IsHTTPS("https://example.com") {
		t.Fatal("scheme helpers returned false")
	}
	if got := BuildContentType("text/plain", "utf-8"); got != "text/plain;charset=utf-8" {
		t.Fatalf("BuildContentType = %q", got)
	}
	if !IsDefaultContentType("") || !IsFormURLEncoded("application/x-www-form-urlencoded;charset=utf-8") {
		t.Fatal("content type predicates returned unexpected result")
	}
	if got := GetCharsetFromContentTypeWithOptions("text/plain; enc=gbk", WithCharsetRegexp(regexp.MustCompile(`enc=([a-z0-9-]+)`))); got != "gbk" {
		t.Fatalf("GetCharsetFromContentTypeWithOptions = %q", got)
	}
	if got := GetCharsetFromHTMLWithOptions(`<meta data-charset="big5">`, WithMetaCharsetRegexp(regexp.MustCompile(`data-charset="([^"]+)"`))); got != "big5" {
		t.Fatalf("GetCharsetFromHTMLWithOptions = %q", got)
	}
	if got := GetMimeType("payload.JSON"); got != "application/json" {
		t.Fatalf("GetMimeType = %q", got)
	}
	if got := BuildBasicAuth("user", "pass"); !strings.HasPrefix(got, "Basic ") {
		t.Fatalf("BuildBasicAuth = %q", got)
	}
}

func TestAdditionalDownloadWrappers(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("download"))
	}))
	defer srv.Close()
	allowLocal := WithURLPolicy(URLPolicy{AllowedSchemes: []string{"http", "https"}, RejectPrivate: false})

	if got, err := DownloadStringE(srv.URL, ""); err != nil || got != "download" {
		t.Fatalf("DownloadStringE = %q, %v", got, err)
	}
	if got, err := DownloadStringEWithOptions(srv.URL, "", WithMaxResponseBytes(64)); err != nil || got != "download" {
		t.Fatalf("DownloadStringEWithOptions = %q, %v", got, err)
	}
	if got, err := DownloadStringSafeE(srv.URL, "", allowLocal); err != nil || got != "download" {
		t.Fatalf("DownloadStringSafeE = %q, %v", got, err)
	}
	var buf bytes.Buffer
	if n, err := Download(srv.URL, &buf); err != nil || n != int64(len("download")) || buf.String() != "download" {
		t.Fatalf("Download n=%d body=%q err=%v", n, buf.String(), err)
	}
	buf.Reset()
	if n, err := DownloadWithOptions(srv.URL, &buf, WithMaxResponseBytes(64)); err != nil || n != int64(len("download")) || buf.String() != "download" {
		t.Fatalf("DownloadWithOptions n=%d body=%q err=%v", n, buf.String(), err)
	}
	buf.Reset()
	if n, err := DownloadSafe(srv.URL, &buf, allowLocal); err != nil || n != int64(len("download")) || buf.String() != "download" {
		t.Fatalf("DownloadSafe n=%d body=%q err=%v", n, buf.String(), err)
	}
	if b, err := DownloadBytesSafeE(srv.URL, allowLocal); err != nil || string(b) != "download" {
		t.Fatalf("DownloadBytesSafeE = %q, %v", b, err)
	}
	dir := t.TempDir()
	if n, err := DownloadFileSafeWithOptions(srv.URL, filepath.Join(dir, "safe.txt"), []RequestOption{allowLocal}, WithSaveOverwrite(true)); err != nil || n != int64(len("download")) {
		t.Fatalf("DownloadFileSafeWithOptions n=%d err=%v", n, err)
	}
	if _, err := DownloadFileSafe(srv.URL, filepath.Join(dir, "blocked.txt")); err == nil {
		t.Fatal("DownloadFileSafe default policy error = nil")
	}
}
