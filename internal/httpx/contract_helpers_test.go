package httpx_test

import (
	"testing"
	"time"

	httpxhttp "github.com/imajinyun/knifer-go/internal/httpx/http"
	httpxresty "github.com/imajinyun/knifer-go/internal/httpx/resty"
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
				resp := httpxhttp.GetSafe(rawURL,
					httpxhttp.WithURLPolicy(httpxhttp.URLPolicy{AllowedSchemes: []string{"http", "https"}, AllowedHosts: []string{host}}),
				).Execute()
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
				resp := httpxresty.GetSafe(rawURL,
					httpxresty.WithURLPolicy(httpxresty.URLPolicy{AllowedSchemes: []string{"http", "https"}, AllowedHosts: []string{host}}),
				).Execute()
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
