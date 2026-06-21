package vhttp_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/imajinyun/go-knifer"
	"github.com/imajinyun/go-knifer/vhttp"
)

func ExampleNewError() {
	err := vhttp.NewError("no response", nil)
	fmt.Println(errors.Is(err, knifer.ErrCodeInternal))
	// Output: true
}

func ExampleGetStringE() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	body, err := vhttp.GetStringE(server.URL)
	if err != nil {
		panic(err)
	}
	fmt.Println(body)
	// Output: ok
}

func ExampleGetStringSafeE() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("safe"))
	}))
	defer server.Close()

	body, err := vhttp.GetStringSafeE(server.URL,
		vhttp.WithURLPolicy(vhttp.URLPolicy{AllowedSchemes: []string{"http", "https"}, RejectPrivate: false}),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(body)
	// Output: safe
}

func ExamplePostStringE() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_, _ = fmt.Fprintf(w, "%s:%s", r.Method, body)
	}))
	defer server.Close()

	body, err := vhttp.PostStringE(server.URL, "payload")
	if err != nil {
		panic(err)
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
	n, err := vhttp.Download(server.URL, &buf)
	if err != nil {
		panic(err)
	}
	fmt.Println(n, buf.String())
	// Output: 8 download
}

func ExampleBuildContentType() {
	fmt.Println(vhttp.BuildContentType("application/json", "utf-8"))
	fmt.Println(vhttp.BuildContentType("text/plain", ""))
	// Output:
	// application/json;charset=utf-8
	// text/plain
}

func ExampleGetCharsetFromHTML() {
	html := `<html><head><meta charset="big5"></head></html>`
	fmt.Println(vhttp.GetCharsetFromHTML(html))
	// Output: big5
}

func ExamplePostJSONE() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_, _ = fmt.Fprintf(w, "%s:%s:%s", r.Method, r.Header.Get("Content-Type"), body)
	}))
	defer server.Close()

	body, err := vhttp.PostJSONE(server.URL, `{"ok":true}`)
	if err != nil {
		panic(err)
	}
	fmt.Println(body)
	// Output: POST:application/json;charset=UTF-8:{"ok":true}
}

func ExampleBuildBasicAuth() {
	fmt.Println(vhttp.BuildBasicAuth("user", "pass"))
	// Output: Basic dXNlcjpwYXNz
}

func ExampleGet() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "%s:%s", r.Method, r.URL.Query().Get("q"))
	}))
	defer server.Close()

	resp := vhttp.Get(server.URL).Query("q", "go").Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: GET:go <nil>
}

func ExampleGetSafe() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("safe get"))
	}))
	defer server.Close()

	resp := vhttp.GetSafe(server.URL, vhttp.WithURLPolicy(vhttp.URLPolicy{
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

	resp := vhttp.Post(server.URL).BodyString("payload").Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: POST:payload <nil>
}

func ExamplePostSafe() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_, _ = fmt.Fprintf(w, "%s:%s", r.Method, body)
	}))
	defer server.Close()

	resp := vhttp.PostSafe(server.URL, vhttp.WithURLPolicy(vhttp.URLPolicy{
		AllowedSchemes: []string{"http"},
		RejectPrivate:  false,
	})).BodyString("safe").Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: POST:safe <nil>
}

func ExamplePut() {
	fmt.Println(exampleMethod(vhttp.Put, http.MethodPut))
	// Output: PUT
}

func ExampleDelete() {
	fmt.Println(exampleMethod(vhttp.Delete, http.MethodDelete))
	// Output: DELETE
}

func ExamplePatch() {
	fmt.Println(exampleMethod(vhttp.Patch, http.MethodPatch))
	// Output: PATCH
}

func ExampleHead() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Method", r.Method)
	}))
	defer server.Close()

	resp := vhttp.Head(server.URL).Execute()
	fmt.Println(resp.Header("X-Method"), resp.Err())
	// Output: HEAD <nil>
}

func ExampleOptions() {
	fmt.Println(exampleMethod(vhttp.Options, http.MethodOptions))
	// Output: OPTIONS
}

func ExampleNewRequest() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Method))
	}))
	defer server.Close()

	resp := vhttp.NewRequest(vhttp.MethodTrace, server.URL).Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: TRACE <nil>
}

func ExampleNewRequest_customHeader() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Trace")))
	}))
	defer server.Close()

	resp := vhttp.NewRequest(vhttp.MethodGet, server.URL).
		Header("X-Trace", "reader-facing").
		Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: reader-facing <nil>
}

func ExampleNewSafeRequest() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Method))
	}))
	defer server.Close()

	resp := vhttp.NewSafeRequest(vhttp.MethodGet, server.URL, vhttp.WithURLPolicy(vhttp.URLPolicy{
		AllowedSchemes: []string{"http"},
		RejectPrivate:  false,
	})).Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: GET <nil>
}

func ExampleGetStringEWithOptions() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Token")))
	}))
	defer server.Close()

	body, err := vhttp.GetStringEWithOptions(server.URL, vhttp.WithHeader("X-Token", "abc"))
	fmt.Println(body, err)
	// Output: abc <nil>
}

func ExampleGetWithParamsE() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "%s:%s", r.URL.Query().Get("name"), r.URL.Query().Get("page"))
	}))
	defer server.Close()

	body, err := vhttp.GetWithParamsE(server.URL, map[string]any{"name": "go", "page": 2})
	fmt.Println(body, err)
	// Output: go:2 <nil>
}

func ExampleGetWithTimeoutE() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	body, err := vhttp.GetWithTimeoutE(server.URL, time.Second)
	fmt.Println(body, err)
	// Output: ok <nil>
}

func ExamplePostFormE() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		_, _ = fmt.Fprintf(w, "%s:%s", r.Method, r.Form.Get("name"))
	}))
	defer server.Close()

	body, err := vhttp.PostFormE(server.URL, map[string]any{"name": "go"})
	fmt.Println(body, err)
	// Output: POST:go <nil>
}

func ExamplePostJSONEWithOptions() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Trace")))
	}))
	defer server.Close()

	body, err := vhttp.PostJSONEWithOptions(server.URL, `{"ok":true}`, vhttp.WithHeader("X-Trace", "trace-1"))
	fmt.Println(body, err)
	// Output: trace-1 <nil>
}

func ExamplePostStringEWithOptions() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("Content-Type")))
	}))
	defer server.Close()

	body, err := vhttp.PostStringEWithOptions(
		server.URL,
		"payload",
		vhttp.WithContentType("text/plain;charset=utf-8"),
	)
	fmt.Println(body, err)
	// Output: text/plain;charset=utf-8 <nil>
}

func ExampleDownloadBytesE() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("bytes"))
	}))
	defer server.Close()

	body, err := vhttp.DownloadBytesE(server.URL)
	fmt.Println(string(body), err)
	// Output: bytes <nil>
}

func ExampleDownloadBytesEWithOptions() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("limited"))
	}))
	defer server.Close()

	body, err := vhttp.DownloadBytesEWithOptions(server.URL, vhttp.WithMaxResponseBytes(7))
	fmt.Println(string(body), err)
	// Output: limited <nil>
}

func ExampleDownloadStringE() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("text"))
	}))
	defer server.Close()

	body, err := vhttp.DownloadStringE(server.URL, "")
	fmt.Println(body, err)
	// Output: text <nil>
}

func ExampleDownloadSafe() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("safe"))
	}))
	defer server.Close()

	var buf bytes.Buffer
	n, err := vhttp.DownloadSafe(server.URL, &buf, vhttp.WithURLPolicy(vhttp.URLPolicy{
		AllowedSchemes: []string{"http"},
		RejectPrivate:  false,
	}))
	fmt.Println(n, buf.String(), err)
	// Output: 4 safe <nil>
}

func ExampleDownloadFileSafe() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("safe-file"))
	}))
	defer server.Close()

	dir, _ := os.MkdirTemp("", "vhttp-safe-download-*")
	defer os.RemoveAll(dir)
	dest := filepath.Join(dir, "download.txt")
	n, err := vhttp.DownloadFileSafeWithOptions(server.URL, dest,
		[]vhttp.RequestOption{vhttp.WithURLPolicy(vhttp.URLPolicy{AllowedSchemes: []string{"http"}, RejectPrivate: false})},
		vhttp.WithSaveOverwrite(true),
	)
	data, _ := os.ReadFile(dest)
	fmt.Println(n, string(data), err)
	// Output: 9 safe-file <nil>
}

func ExampleWithHeader() {
	fmt.Println(exampleHeader(vhttp.WithHeader("X-Mode", "one")))
	// Output: one
}

func ExampleWithHeaders() {
	fmt.Println(exampleHeader(vhttp.WithHeaders(map[string]string{"X-Mode": "batch"})))
	// Output: batch
}

func ExampleWithUserAgent() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.UserAgent()))
	}))
	defer server.Close()

	resp := vhttp.Get(server.URL, vhttp.WithUserAgent("go-knifer-example")).Execute()
	fmt.Println(resp.Body())
	// Output: go-knifer-example
}

func ExampleWithTimeout() {
	fmt.Println(vhttp.WithTimeout(time.Second) != nil)
	// Output: true
}

func ExampleWithTransport() {
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(req.Method + ":mock")),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})

	resp := vhttp.Get("https://example.invalid", vhttp.WithTransport(transport)).Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: GET:mock <nil>
}

func ExampleWithResponseReadAllFunc() {
	readAllCalled := false
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("body")),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})

	resp := vhttp.Get("https://example.invalid",
		vhttp.WithTransport(transport),
		vhttp.WithResponseReadAllFunc(func(r io.Reader) ([]byte, error) {
			readAllCalled = true
			return io.ReadAll(r)
		}),
	).Execute()
	fmt.Println(resp.Body(), readAllCalled)
	// Output: body true
}

func ExampleWithURLPolicy() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("policy"))
	}))
	defer server.Close()

	resp := vhttp.GetSafe(server.URL, vhttp.WithURLPolicy(vhttp.URLPolicy{
		AllowedSchemes: []string{"http"},
		RejectPrivate:  false,
	})).Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: policy <nil>
}

func ExampleWithAllowedHosts() {
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(req.Host)),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})

	resp := vhttp.GetSafe("http://public.example/resource",
		vhttp.WithAllowedHosts("public.example"),
		vhttp.WithLookupIP(func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("203.0.113.10")}, nil
		}),
		vhttp.WithTransport(transport),
	).Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: public.example <nil>
}

func ExampleWithLookupIP() {
	lookupCalled := false
	resp := vhttp.GetSafe("http://private.example/resource",
		vhttp.WithLookupIP(func(context.Context, string) ([]net.IP, error) {
			lookupCalled = true
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		}),
		vhttp.WithTransport(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("unused")), Header: make(http.Header)}, nil
		})),
	).Execute()
	fmt.Println(lookupCalled, resp.Err() != nil)
	// Output: true true
}

func ExampleSetGlobalHeader() {
	previous := vhttp.SnapshotGlobalConfig()
	defer vhttp.ConfigureGlobalConfig(previous)

	vhttp.SetGlobalHeader("X-Example", "one")
	fmt.Println(vhttp.CloneGlobalHeaders().Get("X-Example"))
	// Output: one
}

func ExampleAddGlobalHeader() {
	previous := vhttp.SnapshotGlobalConfig()
	defer vhttp.ConfigureGlobalConfig(previous)

	vhttp.AddGlobalHeader("X-Example", "one")
	vhttp.AddGlobalHeader("X-Example", "two")
	fmt.Println(strings.Join(vhttp.CloneGlobalHeaders().Values("X-Example"), ","))
	// Output: one,two
}

func ExampleRemoveGlobalHeader() {
	previous := vhttp.SnapshotGlobalConfig()
	defer vhttp.ConfigureGlobalConfig(previous)

	vhttp.SetGlobalHeader("X-Example", "one")
	vhttp.RemoveGlobalHeader("X-Example")
	fmt.Println(vhttp.CloneGlobalHeaders().Get("X-Example") == "")
	// Output: true
}

func ExampleWithScopedGlobalConfig() {
	previous := vhttp.SnapshotGlobalConfig()
	cfg := previous
	cfg.DefaultUserAgent = "scoped-agent"

	var inside string
	vhttp.WithScopedGlobalConfig(cfg, func() {
		inside = vhttp.GetGlobalUserAgent()
	})
	fmt.Println(inside, vhttp.GetGlobalUserAgent() == previous.DefaultUserAgent)
	// Output: scoped-agent true
}

func ExampleSetCookieJar() {
	previous := vhttp.SnapshotGlobalConfig()
	defer vhttp.ConfigureGlobalConfig(previous)

	jar, _ := cookiejar.New(nil)
	vhttp.SetCookieJar(jar)
	fmt.Println(vhttp.GetCookieJar() != nil)
	// Output: true
}

func ExampleCloseCookie() {
	previous := vhttp.SnapshotGlobalConfig()
	defer vhttp.ConfigureGlobalConfig(previous)

	jar, _ := cookiejar.New(nil)
	vhttp.SetCookieJar(jar)
	vhttp.CloseCookie()
	fmt.Println(vhttp.GetCookieJar() == nil)
	// Output: true
}

func ExampleHTMLEscape() {
	fmt.Println(vhttp.HTMLEscape(`<b>go</b>`))
	// Output: &lt;b&gt;go&lt;/b&gt;
}

func ExampleHTMLUnescape() {
	fmt.Println(vhttp.HTMLUnescape("&lt;b&gt;go&lt;/b&gt;"))
	// Output: <b>go</b>
}

func ExampleCleanHTML() {
	fmt.Println(vhttp.CleanHTML(`<p>Hello <b>Go</b></p>`))
	// Output: Hello Go
}

func ExampleFilterHTMLTag() {
	fmt.Println(vhttp.FilterHTMLTag(`<p>Hello <b>Go</b></p>`, "b"))
	// Output: <p>Hello </p>
}

func ExampleParseUserAgent() {
	ua := vhttp.ParseUserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X) AppleWebKit/537.36 Chrome/120.0 Safari/537.36")
	fmt.Println(ua.Browser, ua.OS, ua.Engine, ua.Platform)
	// Output: Chrome macOS WebKit Macintosh
}

func ExampleIsRedirected() {
	fmt.Println(vhttp.IsRedirected(http.StatusFound), vhttp.IsRedirected(http.StatusOK))
	// Output: true false
}

func exampleMethod(newRequest func(string, ...vhttp.RequestOption) *vhttp.Request, want string) string {
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

func exampleHeader(opt vhttp.RequestOption) string {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Mode")))
	}))
	defer server.Close()

	return vhttp.Get(server.URL, opt).Execute().Body()
}
