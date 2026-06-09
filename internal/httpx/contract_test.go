package httpx_test

import (
	stdhttp "net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	httpxhttp "github.com/imajinyun/go-knifer/internal/httpx/http"
	httpxresty "github.com/imajinyun/go-knifer/internal/httpx/resty"
)

type contractResponse struct {
	status int
	body   string
	err    error
}

type httpContractBackend struct {
	name            string
	reset           func(t *testing.T)
	snapshotTimeout func() time.Duration
	getBasic        func(rawURL string) contractResponse
	postJSON        func(rawURL string) contractResponse
	getNoFollow     func(rawURL string) contractResponse
	getFollow       func(rawURL string) contractResponse
	getWithMaxBytes func(rawURL string, maxBytes int64) contractResponse
	clientSnapshot  func(t *testing.T, rawURL string) contractResponse
	isolatedClient  func(t *testing.T, rawURL string) contractResponse
	invalidURL      func() error
	safeURL         func(rawURL string) error
	safeAllowed     func(rawURL, host string) contractResponse
}

func TestHTTPImplementationsContract(t *testing.T) {
	for _, backend := range httpContractBackends() {
		backend := backend
		t.Run(backend.name, func(t *testing.T) {
			backend.reset(t)

			t.Run("bounded default timeout", func(t *testing.T) {
				if got := backend.snapshotTimeout(); got <= 0 {
					t.Fatalf("SnapshotGlobalConfig().Timeout = %v, want positive timeout", got)
				}
			})

			t.Run("request basics", func(t *testing.T) {
				srv := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
					_, _ = w.Write([]byte(r.Method + ":" + r.URL.Query().Get("q") + ":" + r.Header.Get("X-Contract") + ":" + r.Header.Get("User-Agent")))
				}))
				defer srv.Close()

				resp := backend.getBasic(srv.URL)
				if resp.err != nil || resp.status != stdhttp.StatusOK || resp.body != "GET:go:yes:contract-agent" {
					t.Fatalf("GET contract status=%d body=%q err=%v", resp.status, resp.body, resp.err)
				}
			})

			t.Run("post json content type", func(t *testing.T) {
				srv := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
					_, _ = w.Write([]byte(r.Method + ":" + r.Header.Get("Content-Type")))
				}))
				defer srv.Close()

				resp := backend.postJSON(srv.URL)
				if resp.err != nil || resp.status != stdhttp.StatusOK || resp.body != "POST:application/json;charset=UTF-8" {
					t.Fatalf("POST JSON contract status=%d body=%q err=%v", resp.status, resp.body, resp.err)
				}
			})

			t.Run("redirect controls", func(t *testing.T) {
				srv := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
					if r.URL.Path == "/start" {
						stdhttp.Redirect(w, r, "/end", stdhttp.StatusFound)
						return
					}
					_, _ = w.Write([]byte("end"))
				}))
				defer srv.Close()

				noFollow := backend.getNoFollow(srv.URL + "/start")
				if noFollow.err != nil || noFollow.status != stdhttp.StatusFound {
					t.Fatalf("no-follow status=%d body=%q err=%v", noFollow.status, noFollow.body, noFollow.err)
				}

				follow := backend.getFollow(srv.URL + "/start")
				if follow.err != nil || follow.status != stdhttp.StatusOK || follow.body != "end" {
					t.Fatalf("follow status=%d body=%q err=%v", follow.status, follow.body, follow.err)
				}
			})

			t.Run("max response bytes", func(t *testing.T) {
				srv := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
					_, _ = w.Write([]byte("abcdef"))
				}))
				defer srv.Close()

				resp := backend.getWithMaxBytes(srv.URL, 3)
				if resp.err == nil || resp.body != "" {
					t.Fatalf("limited body=%q err=%v, want max bytes error", resp.body, resp.err)
				}
			})

			t.Run("global snapshot and isolated client", func(t *testing.T) {
				srv := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
					_, _ = w.Write([]byte(r.Header.Get("X-Contract-Global")))
				}))
				defer srv.Close()

				snapshot := backend.clientSnapshot(t, srv.URL)
				if snapshot.err != nil || snapshot.body != "snapshot" {
					t.Fatalf("client snapshot body=%q err=%v", snapshot.body, snapshot.err)
				}

				isolated := backend.isolatedClient(t, srv.URL)
				if isolated.err != nil || isolated.body != "" {
					t.Fatalf("isolated client body=%q err=%v, want no global headers", isolated.body, isolated.err)
				}
			})

			t.Run("invalid and safe urls", func(t *testing.T) {
				if err := backend.invalidURL(); err == nil {
					t.Fatal("invalid URL error = nil")
				}
				for _, rawURL := range []string{"file:///tmp/secret.txt", "http://127.0.0.1/config.yaml"} {
					if err := backend.safeURL(rawURL); err == nil {
						t.Fatalf("safe request to %q error = nil", rawURL)
					}
				}

				srv := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
					_, _ = w.Write([]byte("safe"))
				}))
				defer srv.Close()

				parsed, err := url.Parse(srv.URL)
				if err != nil {
					t.Fatalf("parse server URL: %v", err)
				}
				resp := backend.safeAllowed(srv.URL, parsed.Hostname())
				if resp.err != nil || resp.status != stdhttp.StatusOK || resp.body != "safe" {
					t.Fatalf("safe allowed status=%d body=%q err=%v", resp.status, resp.body, resp.err)
				}
			})
		})
	}
}

func httpContractBackends() []httpContractBackend {
	return []httpContractBackend{
		{
			name: "stdlib-http",
			reset: func(t *testing.T) {
				previous := httpxhttp.SnapshotGlobalConfig()
				httpxhttp.ResetGlobalConfig()
				t.Cleanup(func() { httpxhttp.ConfigureGlobalConfig(previous) })
			},
			snapshotTimeout: func() time.Duration { return httpxhttp.SnapshotGlobalConfig().Timeout },
			getBasic: func(rawURL string) contractResponse {
				resp := httpxhttp.Get(rawURL, httpxhttp.WithHeader("X-Contract", "yes"), httpxhttp.WithUserAgent("contract-agent")).Query("q", "go").Execute()
				return stdlibHTTPContractResponse(resp)
			},
			postJSON: func(rawURL string) contractResponse {
				resp := httpxhttp.Post(rawURL).BodyJSON(`{"ok":true}`).Execute()
				return stdlibHTTPContractResponse(resp)
			},
			getNoFollow: func(rawURL string) contractResponse {
				resp := httpxhttp.Get(rawURL, httpxhttp.WithFollowRedirects(false)).Execute()
				return stdlibHTTPContractResponse(resp)
			},
			getFollow: func(rawURL string) contractResponse {
				resp := httpxhttp.Get(rawURL, httpxhttp.WithFollowRedirects(true)).Execute()
				return stdlibHTTPContractResponse(resp)
			},
			getWithMaxBytes: func(rawURL string, maxBytes int64) contractResponse {
				resp := httpxhttp.Get(rawURL, httpxhttp.WithMaxResponseBytes(maxBytes)).Execute()
				body := resp.Body()
				return contractResponse{status: resp.Status(), body: body, err: resp.Err()}
			},
			clientSnapshot: func(t *testing.T, rawURL string) contractResponse {
				httpxhttp.SetGlobalHeader("X-Contract-Global", "snapshot")
				httpxhttp.SetGlobalUserAgent("contract-snapshot-agent")
				client := httpxhttp.NewClient()
				httpxhttp.SetGlobalHeader("X-Contract-Global", "mutated")
				httpxhttp.SetGlobalUserAgent("mutated-agent")
				resp := client.Get(rawURL).Execute()
				return stdlibHTTPContractResponse(resp)
			},
			isolatedClient: func(t *testing.T, rawURL string) contractResponse {
				httpxhttp.SetGlobalHeader("X-Contract-Global", "global")
				httpxhttp.SetGlobalUserAgent("global-agent")
				resp := httpxhttp.NewIsolatedClient().Get(rawURL).Execute()
				return stdlibHTTPContractResponse(resp)
			},
			invalidURL: func() error { return httpxhttp.Get("http://[::1").Execute().Err() },
			safeURL:    func(rawURL string) error { return httpxhttp.GetSafe(rawURL).Execute().Err() },
			safeAllowed: func(rawURL, host string) contractResponse {
				resp := httpxhttp.GetSafe(rawURL, httpxhttp.WithAllowedHosts(host)).Execute()
				return stdlibHTTPContractResponse(resp)
			},
		},
		{
			name: "resty",
			reset: func(t *testing.T) {
				previous := httpxresty.SnapshotGlobalConfig()
				httpxresty.ResetGlobalConfig()
				t.Cleanup(func() { httpxresty.ConfigureGlobalConfig(previous) })
			},
			snapshotTimeout: func() time.Duration { return httpxresty.SnapshotGlobalConfig().Timeout },
			getBasic: func(rawURL string) contractResponse {
				resp := httpxresty.Get(rawURL, httpxresty.WithHeader("X-Contract", "yes"), httpxresty.WithUserAgent("contract-agent")).Query("q", "go").Execute()
				return restyContractResponse(resp)
			},
			postJSON: func(rawURL string) contractResponse {
				resp := httpxresty.Post(rawURL).BodyJSON(`{"ok":true}`).Execute()
				return restyContractResponse(resp)
			},
			getNoFollow: func(rawURL string) contractResponse {
				resp := httpxresty.Get(rawURL, httpxresty.WithFollowRedirects(false)).Execute()
				return restyContractResponse(resp)
			},
			getFollow: func(rawURL string) contractResponse {
				resp := httpxresty.Get(rawURL, httpxresty.WithFollowRedirects(true)).Execute()
				return restyContractResponse(resp)
			},
			getWithMaxBytes: func(rawURL string, maxBytes int64) contractResponse {
				resp := httpxresty.Get(rawURL, httpxresty.WithMaxResponseBytes(maxBytes)).Execute()
				body := resp.Body()
				return contractResponse{status: resp.Status(), body: body, err: resp.Err()}
			},
			clientSnapshot: func(t *testing.T, rawURL string) contractResponse {
				httpxresty.SetGlobalHeader("X-Contract-Global", "snapshot")
				httpxresty.SetGlobalUserAgent("contract-snapshot-agent")
				client := httpxresty.NewClient()
				httpxresty.SetGlobalHeader("X-Contract-Global", "mutated")
				httpxresty.SetGlobalUserAgent("mutated-agent")
				resp := client.Get(rawURL).Execute()
				return restyContractResponse(resp)
			},
			isolatedClient: func(t *testing.T, rawURL string) contractResponse {
				httpxresty.SetGlobalHeader("X-Contract-Global", "global")
				httpxresty.SetGlobalUserAgent("global-agent")
				resp := httpxresty.NewIsolatedClient().Get(rawURL).Execute()
				return restyContractResponse(resp)
			},
			invalidURL: func() error { return httpxresty.Get("http://[::1").Execute().Err() },
			safeURL:    func(rawURL string) error { return httpxresty.GetSafe(rawURL).Execute().Err() },
			safeAllowed: func(rawURL, host string) contractResponse {
				resp := httpxresty.GetSafe(rawURL, httpxresty.WithAllowedHosts(host)).Execute()
				return restyContractResponse(resp)
			},
		},
	}
}

func stdlibHTTPContractResponse(resp *httpxhttp.HTTPResponse) contractResponse {
	body := resp.Body()
	return contractResponse{status: resp.Status(), body: body, err: resp.Err()}
}

func restyContractResponse(resp *httpxresty.HTTPResponse) contractResponse {
	body := resp.Body()
	return contractResponse{status: resp.Status(), body: body, err: resp.Err()}
}
