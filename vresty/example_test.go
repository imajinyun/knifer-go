package vresty_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/imajinyun/go-knifer/vresty"
	grestry "resty.dev/v3"
)

func ExampleGetStringSafeE() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("safe response"))
	}))
	defer server.Close()

	body, err := vresty.GetStringSafeE(server.URL,
		vresty.WithURLPolicy(vresty.URLPolicy{AllowedSchemes: []string{"http", "https"}, RejectPrivate: false}),
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(body)
	// Output: safe response
}

func ExampleGetStringE() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("plain response"))
	}))
	defer server.Close()

	body, err := vresty.GetStringE(server.URL)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(body)
	// Output: plain response
}

func ExamplePostStringE() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_, _ = fmt.Fprintf(w, "%s:%s", r.Method, body)
	}))
	defer server.Close()

	body, err := vresty.PostStringE(server.URL, "payload")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(body)
	// Output: POST:payload
}

func ExampleDownload() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("download"))
	}))
	defer server.Close()

	var buf bytes.Buffer
	n, err := vresty.Download(server.URL, &buf)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(n, buf.String())
	// Output: 8 download
}

func ExampleBuildBasicAuth() {
	fmt.Println(vresty.BuildBasicAuth("user", "pass"))
	// Output: Basic dXNlcjpwYXNz
}

func ExampleGetWithParamsE() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "page=%s", r.URL.Query().Get("page"))
	}))
	defer server.Close()

	body, err := vresty.GetWithParamsE(server.URL, map[string]any{"page": 2})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(body)
	// Output: page=2
}

func ExamplePostJSONE() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_, _ = fmt.Fprintf(w, "%s:%s", r.Method, body)
	}))
	defer server.Close()

	body, err := vresty.PostJSONE(server.URL, `{"name":"alice"}`)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(body)
	// Output: POST:{"name":"alice"}
}

func ExampleBuildContentType() {
	fmt.Println(vresty.BuildContentType("text/plain", "utf-8"))
	// Output: text/plain;charset=utf-8
}

func ExampleGet() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "%s:%s", r.Method, r.URL.Query().Get("q"))
	}))
	defer server.Close()

	resp := vresty.Get(server.URL).Query("q", "go").Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: GET:go <nil>
}

func ExampleGetSafe() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("safe get"))
	}))
	defer server.Close()

	resp := vresty.GetSafe(server.URL, vresty.WithURLPolicy(vresty.URLPolicy{
		AllowedSchemes: []string{"http"},
		RejectPrivate:  false,
	})).Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: safe get <nil>
}

func ExamplePost() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_, _ = fmt.Fprintf(w, "%s:%s", r.Method, body)
	}))
	defer server.Close()

	resp := vresty.Post(server.URL).BodyString("payload").Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: POST:payload <nil>
}

func ExamplePostSafe() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_, _ = fmt.Fprintf(w, "%s:%s", r.Method, body)
	}))
	defer server.Close()

	resp := vresty.PostSafe(server.URL, vresty.WithURLPolicy(vresty.URLPolicy{
		AllowedSchemes: []string{"http"},
		RejectPrivate:  false,
	})).BodyString("safe").Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: POST:safe <nil>
}

func ExamplePut() {
	fmt.Println(exampleMethod(vresty.Put, http.MethodPut))
	// Output: PUT
}

func ExampleDelete() {
	fmt.Println(exampleMethod(vresty.Delete, http.MethodDelete))
	// Output: DELETE
}

func ExamplePatch() {
	fmt.Println(exampleMethod(vresty.Patch, http.MethodPatch))
	// Output: PATCH
}

func ExampleHead() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Method", r.Method)
	}))
	defer server.Close()

	resp := vresty.Head(server.URL).Execute()
	fmt.Println(resp.Header("X-Method"), resp.Err())
	// Output: HEAD <nil>
}

func ExampleOptions() {
	fmt.Println(exampleMethod(vresty.Options, http.MethodOptions))
	// Output: OPTIONS
}

func ExampleNewRequest() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Method))
	}))
	defer server.Close()

	resp := vresty.NewRequest(vresty.MethodTrace, server.URL).Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: TRACE <nil>
}

func ExampleNewSafeRequest() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Method))
	}))
	defer server.Close()

	resp := vresty.NewSafeRequest(vresty.MethodGet, server.URL, vresty.WithURLPolicy(vresty.URLPolicy{
		AllowedSchemes: []string{"http"},
		RejectPrivate:  false,
	})).Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: GET <nil>
}

func ExampleNewIsolatedRequest() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Isolated")))
	}))
	defer server.Close()

	resp := vresty.NewIsolatedRequest(
		vresty.MethodGet,
		server.URL,
		vresty.WithHeader("X-Isolated", "yes"),
	).Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: yes <nil>
}

func ExampleNewRequestWithConfig() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Config")))
	}))
	defer server.Close()

	cfg := vresty.SnapshotGlobalConfig()
	cfg.Headers = vresty.HeaderValues{"X-Config": []string{"from-config"}}
	resp := vresty.NewRequestWithConfig(vresty.MethodGet, server.URL, cfg).Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: from-config <nil>
}

func ExampleNewClient() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Client")))
	}))
	defer server.Close()

	client := vresty.NewClient(vresty.WithClientRequestOptions(vresty.WithHeader("X-Client", "default")))
	resp := client.Get(server.URL).Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: default <nil>
}

func ExampleNewIsolatedClient() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Client")))
	}))
	defer server.Close()

	client := vresty.NewIsolatedClient(vresty.WithClientRequestOptions(vresty.WithHeader("X-Client", "isolated")))
	resp := client.Get(server.URL).Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: isolated <nil>
}

func ExampleNewClientWithConfig() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Client")))
	}))
	defer server.Close()

	cfg := vresty.SnapshotGlobalConfig()
	cfg.Headers = vresty.HeaderValues{"X-Client": []string{"configured"}}
	client := vresty.NewClientWithConfig(cfg)
	resp := client.Get(server.URL).Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: configured <nil>
}

func ExampleGetStringEWithOptions() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Token")))
	}))
	defer server.Close()

	body, err := vresty.GetStringEWithOptions(server.URL, vresty.WithHeader("X-Token", "abc"))
	fmt.Println(body, err)
	// Output: abc <nil>
}

func ExampleGetWithParamsEWithOptions() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "%s:%s", r.URL.Query().Get("name"), r.URL.Query().Get("page"))
	}))
	defer server.Close()

	body, err := vresty.GetWithParamsEWithOptions(
		server.URL,
		map[string]any{"name": "go", "page": 2},
		vresty.WithHeader("X-Trace", "trace-1"),
	)
	fmt.Println(body, err)
	// Output: go:2 <nil>
}

func ExampleGetWithTimeoutE() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	body, err := vresty.GetWithTimeoutE(server.URL, time.Second)
	fmt.Println(body, err)
	// Output: ok <nil>
}

func ExampleGetWithTimeoutEWithOptions() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Timeout")))
	}))
	defer server.Close()

	body, err := vresty.GetWithTimeoutEWithOptions(
		server.URL,
		time.Second,
		vresty.WithHeader("X-Timeout", "set"),
	)
	fmt.Println(body, err)
	// Output: set <nil>
}

func ExamplePostFormE() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		_, _ = fmt.Fprintf(w, "%s:%s", r.Method, r.Form.Get("name"))
	}))
	defer server.Close()

	body, err := vresty.PostFormE(server.URL, map[string]any{"name": "go"})
	fmt.Println(body, err)
	// Output: POST:go <nil>
}

func ExamplePostFormEWithOptions() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		_, _ = fmt.Fprintf(w, "%s:%s", r.Form.Get("name"), r.Header.Get("X-Form"))
	}))
	defer server.Close()

	body, err := vresty.PostFormEWithOptions(
		server.URL,
		map[string]any{"name": "go"},
		vresty.WithHeader("X-Form", "yes"),
	)
	fmt.Println(body, err)
	// Output: go:yes <nil>
}

func ExamplePostStringEWithOptions() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("Content-Type")))
	}))
	defer server.Close()

	body, err := vresty.PostStringEWithOptions(
		server.URL,
		"payload",
		vresty.WithContentType("text/plain;charset=utf-8"),
	)
	fmt.Println(body, err)
	// Output: text/plain;charset=utf-8 <nil>
}

func ExamplePostJSONEWithOptions() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Trace")))
	}))
	defer server.Close()

	body, err := vresty.PostJSONEWithOptions(server.URL, `{"ok":true}`, vresty.WithHeader("X-Trace", "trace-1"))
	fmt.Println(body, err)
	// Output: trace-1 <nil>
}

func ExampleDownloadWithOptions() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Download")))
	}))
	defer server.Close()

	var buf bytes.Buffer
	n, err := vresty.DownloadWithOptions(server.URL, &buf, vresty.WithHeader("X-Download", "body"))
	fmt.Println(n, buf.String(), err)
	// Output: 4 body <nil>
}

func ExampleDownloadSafe() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("safe"))
	}))
	defer server.Close()

	var buf bytes.Buffer
	n, err := vresty.DownloadSafe(server.URL, &buf, vresty.WithURLPolicy(vresty.URLPolicy{
		AllowedSchemes: []string{"http"},
		RejectPrivate:  false,
	}))
	fmt.Println(n, buf.String(), err)
	// Output: 4 safe <nil>
}

func ExampleDownloadBytesE() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("bytes"))
	}))
	defer server.Close()

	body, err := vresty.DownloadBytesE(server.URL)
	fmt.Println(string(body), err)
	// Output: bytes <nil>
}

func ExampleDownloadBytesEWithOptions() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("limited"))
	}))
	defer server.Close()

	body, err := vresty.DownloadBytesEWithOptions(server.URL, vresty.WithMaxResponseBytes(7))
	fmt.Println(string(body), err)
	// Output: limited <nil>
}

func ExampleDownloadBytesSafeE() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("safe-bytes"))
	}))
	defer server.Close()

	body, err := vresty.DownloadBytesSafeE(server.URL, vresty.WithURLPolicy(vresty.URLPolicy{
		AllowedSchemes: []string{"http"},
		RejectPrivate:  false,
	}))
	fmt.Println(string(body), err)
	// Output: safe-bytes <nil>
}

func ExampleDownloadStringE() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("text"))
	}))
	defer server.Close()

	body, err := vresty.DownloadStringE(server.URL, "")
	fmt.Println(body, err)
	// Output: text <nil>
}

func ExampleDownloadStringEWithOptions() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("text-options"))
	}))
	defer server.Close()

	body, err := vresty.DownloadStringEWithOptions(server.URL, "", vresty.WithMaxResponseBytes(12))
	fmt.Println(body, err)
	// Output: text-options <nil>
}

func ExampleDownloadStringSafeE() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("safe-text"))
	}))
	defer server.Close()

	body, err := vresty.DownloadStringSafeE(server.URL, "", vresty.WithURLPolicy(vresty.URLPolicy{
		AllowedSchemes: []string{"http"},
		RejectPrivate:  false,
	}))
	fmt.Println(body, err)
	// Output: safe-text <nil>
}

func ExampleDownloadFile() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("file"))
	}))
	defer server.Close()

	dir, _ := os.MkdirTemp("", "vresty-example-*")
	defer os.RemoveAll(dir)
	dest := filepath.Join(dir, "download.txt")
	n, err := vresty.DownloadFile(server.URL, dest)
	data, _ := os.ReadFile(dest)
	fmt.Println(n, string(data), err)
	// Output: 4 file <nil>
}

func ExampleDownloadFileWithOptions() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("file-options"))
	}))
	defer server.Close()

	dir, _ := os.MkdirTemp("", "vresty-example-*")
	defer os.RemoveAll(dir)
	dest := filepath.Join(dir, "download.txt")
	n, err := vresty.DownloadFileWithOptions(
		server.URL,
		dest,
		[]vresty.RequestOption{vresty.WithMaxResponseBytes(12)},
		vresty.WithSaveOverwrite(true),
	)
	data, _ := os.ReadFile(dest)
	fmt.Println(n, string(data), err)
	// Output: 12 file-options <nil>
}

func ExampleWithHeader() {
	fmt.Println(exampleHeader(vresty.WithHeader("X-Mode", "one")))
	// Output: one
}

func ExampleWithHeaders() {
	fmt.Println(exampleHeader(vresty.WithHeaders(map[string]string{"X-Mode": "batch"})))
	// Output: batch
}

func ExampleWithUserAgent() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.UserAgent()))
	}))
	defer server.Close()

	resp := vresty.Get(server.URL, vresty.WithUserAgent("go-knifer-resty-example")).Execute()
	fmt.Println(resp.Body())
	// Output: go-knifer-resty-example
}

func ExampleWithContentType() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("Content-Type")))
	}))
	defer server.Close()

	resp := vresty.Post(server.URL, vresty.WithContentType("text/plain;charset=utf-8")).BodyString("payload").Execute()
	fmt.Println(resp.Body())
	// Output: text/plain;charset=utf-8
}

func ExampleWithCharset() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("Content-Type")))
	}))
	defer server.Close()

	resp := vresty.Post(server.URL, vresty.WithCharset("utf-8")).BodyJSON(`{"ok":true}`).Execute()
	fmt.Println(resp.Body())
	// Output: application/json;charset=utf-8
}

func ExampleWithTimeout() {
	fmt.Println(vresty.WithTimeout(time.Second) != nil)
	// Output: true
}

func ExampleWithMaxResponseBytes() {
	fmt.Println(vresty.WithMaxResponseBytes(64) != nil)
	// Output: true
}

func ExampleWithURLPolicy() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("policy"))
	}))
	defer server.Close()

	resp := vresty.GetSafe(server.URL, vresty.WithURLPolicy(vresty.URLPolicy{
		AllowedSchemes: []string{"http"},
		RejectPrivate:  false,
	})).Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: policy <nil>
}

func ExampleWithAllowedHosts() {
	restyClient := grestry.New()
	restyClient.SetTransport(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(req.Host)),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	}))

	resp := vresty.GetSafe("http://public.example/resource",
		vresty.WithAllowedHosts("public.example"),
		vresty.WithLookupIP(func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("203.0.113.10")}, nil
		}),
		vresty.WithRestyClient(restyClient),
	).Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: public.example <nil>
}

func ExampleWithLookupIP() {
	lookupCalled := false
	resp := vresty.GetSafe("http://private.example/resource",
		vresty.WithLookupIP(func(context.Context, string) ([]net.IP, error) {
			lookupCalled = true
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		}),
	).Execute()
	fmt.Println(lookupCalled, resp.Err() != nil)
	// Output: true true
}

func ExampleWithRestyClient() {
	restyClient := grestry.New()
	restyClient.SetTransport(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(req.Method + ":mock")),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	}))

	resp := vresty.Get("https://example.invalid", vresty.WithRestyClient(restyClient)).Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: GET:mock <nil>
}

func ExampleWithRestyClientFactory() {
	factoryCalled := false
	resp := vresty.Get("https://example.invalid", vresty.WithRestyClientFactory(func() *grestry.Client {
		factoryCalled = true
		client := grestry.New()
		client.SetTransport(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("factory")),
				Header:     make(http.Header),
				Request:    req,
			}, nil
		}))
		return client
	})).Execute()
	fmt.Println(resp.Body(), factoryCalled)
	// Output: factory true
}

func ExampleSetGlobalHeader() {
	previous := vresty.SnapshotGlobalConfig()
	defer vresty.ConfigureGlobalConfig(previous)

	vresty.SetGlobalHeader("X-Example", "one")
	fmt.Println(vresty.CloneGlobalHeaders()["X-Example"][0])
	// Output: one
}

func ExampleAddGlobalHeader() {
	previous := vresty.SnapshotGlobalConfig()
	defer vresty.ConfigureGlobalConfig(previous)

	vresty.AddGlobalHeader("X-Example", "one")
	vresty.AddGlobalHeader("X-Example", "two")
	fmt.Println(strings.Join(vresty.CloneGlobalHeaders()["X-Example"], ","))
	// Output: one,two
}

func ExampleRemoveGlobalHeader() {
	previous := vresty.SnapshotGlobalConfig()
	defer vresty.ConfigureGlobalConfig(previous)

	vresty.SetGlobalHeader("X-Example", "one")
	vresty.RemoveGlobalHeader("X-Example")
	fmt.Println(len(vresty.CloneGlobalHeaders()["X-Example"]) == 0)
	// Output: true
}

func ExampleWithScopedGlobalConfig() {
	previous := vresty.SnapshotGlobalConfig()
	cfg := previous
	cfg.DefaultUserAgent = "scoped-agent"

	var inside string
	vresty.WithScopedGlobalConfig(cfg, func() {
		inside = vresty.GetGlobalUserAgent()
	})
	fmt.Println(inside, vresty.GetGlobalUserAgent() == previous.DefaultUserAgent)
	// Output: scoped-agent true
}

func ExampleCloseCookie() {
	previous := vresty.SnapshotGlobalConfig()
	defer vresty.ConfigureGlobalConfig(previous)

	vresty.CloseCookie()
	fmt.Println(vresty.SnapshotGlobalConfig().CookieDisabled)
	// Output: true
}

func ExampleToParams() {
	fmt.Println(vresty.ToParams(map[string]any{"page": 2, "q": "go"}))
	// Output: page=2&q=go
}

func ExampleURLWithForm() {
	fmt.Println(vresty.URLWithForm("https://example.com/search", map[string]any{"q": "go"}))
	// Output: https://example.com/search?q=go
}

func ExampleIsHTTP() {
	fmt.Println(vresty.IsHTTP("http://example.com"), vresty.IsHTTP("https://example.com"))
	// Output: true false
}

func ExampleIsHTTPS() {
	fmt.Println(vresty.IsHTTPS("https://example.com"), vresty.IsHTTPS("http://example.com"))
	// Output: true false
}

func ExampleGetCharsetFromContentType() {
	fmt.Println(vresty.GetCharsetFromContentType("text/html; charset=utf-8"))
	// Output: utf-8
}

func ExampleGetCharsetFromContentTypeWithOptions() {
	custom := regexp.MustCompile(`(?i)encoding=([^;]+)`)
	fmt.Println(vresty.GetCharsetFromContentTypeWithOptions("text/plain; encoding=gbk", vresty.WithCharsetRegexp(custom)))
	// Output: gbk
}

func ExampleGetCharsetFromHTML() {
	fmt.Println(vresty.GetCharsetFromHTML(`<meta charset="utf-8">`))
	// Output: utf-8
}

func ExampleGetCharsetFromHTMLWithOptions() {
	custom := regexp.MustCompile(`(?i)data-charset=["']?([^"'>\s]+)`)
	fmt.Println(vresty.GetCharsetFromHTMLWithOptions(`<html data-charset="gbk">`, vresty.WithMetaCharsetRegexp(custom)))
	// Output: gbk
}

func ExampleGetMimeType() {
	fmt.Println(vresty.GetMimeType("index.html"))
	// Output: text/html
}

func ExampleGuessContentType() {
	fmt.Println(vresty.GuessContentType(`{"ok":true}`))
	// Output: application/json
}

func ExampleIsDefaultContentType() {
	fmt.Println(vresty.IsDefaultContentType("application/x-www-form-urlencoded"), vresty.IsDefaultContentType("application/json"))
	// Output: true false
}

func ExampleIsFormURLEncoded() {
	fmt.Println(vresty.IsFormURLEncoded("application/x-www-form-urlencoded;charset=utf-8"))
	// Output: true
}

func ExampleNewHTTPError() {
	err := vresty.NewHTTPError("request failed", fmt.Errorf("boom"))
	fmt.Println(err != nil)
	// Output: true
}

func ExampleHTTPErrorf() {
	err := vresty.HTTPErrorf("status %d", http.StatusBadGateway)
	fmt.Println(err != nil)
	// Output: true
}

func exampleMethod(newRequest func(string, ...vresty.RequestOption) *vresty.Request, want string) string {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Method))
	}))
	defer server.Close()

	resp := newRequest(server.URL).Execute()
	if resp.Err() != nil {
		return resp.Err().Error()
	}
	if resp.Body() != want {
		return resp.Body()
	}
	return resp.Body()
}

func exampleHeader(opt vresty.RequestOption) string {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Mode")))
	}))
	defer server.Close()

	return vresty.Get(server.URL, opt).Execute().Body()
}
