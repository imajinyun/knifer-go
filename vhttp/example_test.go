package vhttp_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/imajinyun/knifer-go"
	"github.com/imajinyun/knifer-go/vhttp"
)

func ExampleNewError() {
	err := vhttp.NewError("no response", nil)
	fmt.Println(errors.Is(err, knifer.ErrCodeInternal))
	// Output: true
}

func ExampleNewErrorWithCode() {
	err := vhttp.NewErrorWithCode(knifer.ErrCodeInvalidInput, "bad url", nil)
	fmt.Println(errors.Is(err, knifer.ErrCodeInvalidInput))
	// Output: true
}

func ExampleErrorf() {
	err := vhttp.Errorf("request %s", "failed")
	fmt.Println(err.Error())
	// Output: request failed
}

func ExampleErrorfWithCode() {
	err := vhttp.ErrorfWithCode(knifer.ErrCodeUnsupported, "status %d", http.StatusTeapot)
	fmt.Println(errors.Is(err, knifer.ErrCodeUnsupported))
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

func ExamplePutSafe() {
	fmt.Println(exampleSafeMethod(vhttp.PutSafe, http.MethodPut))
	// Output: PUT
}

func ExampleDelete() {
	fmt.Println(exampleMethod(vhttp.Delete, http.MethodDelete))
	// Output: DELETE
}

func ExampleDeleteSafe() {
	fmt.Println(exampleSafeMethod(vhttp.DeleteSafe, http.MethodDelete))
	// Output: DELETE
}

func ExamplePatch() {
	fmt.Println(exampleMethod(vhttp.Patch, http.MethodPatch))
	// Output: PATCH
}

func ExamplePatchSafe() {
	fmt.Println(exampleSafeMethod(vhttp.PatchSafe, http.MethodPatch))
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

func ExampleHeadSafe() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Method", r.Method)
	}))
	defer server.Close()

	resp := vhttp.HeadSafe(server.URL, localURLPolicy()).Execute()
	fmt.Println(resp.Header("X-Method"), resp.Err())
	// Output: HEAD <nil>
}

func ExampleOptions() {
	fmt.Println(exampleMethod(vhttp.Options, http.MethodOptions))
	// Output: OPTIONS
}

func ExampleOptionsSafe() {
	fmt.Println(exampleSafeMethod(vhttp.OptionsSafe, http.MethodOptions))
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

func ExampleNewIsolatedRequest() {
	resp := vhttp.NewIsolatedRequest(vhttp.MethodGet, "https://example.invalid",
		vhttp.WithTransport(exampleTextTransport("isolated")),
	).Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: isolated <nil>
}

func ExampleNewRequestWithConfig() {
	cfg := vhttp.SnapshotGlobalConfig()
	cfg.Headers.Set("X-Config", "config")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Config")))
	}))
	defer server.Close()

	resp := vhttp.NewRequestWithConfig(vhttp.MethodGet, server.URL, cfg).Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: config <nil>
}

func ExampleNewClient() {
	client := vhttp.NewClient(vhttp.WithClientRequestOptions(vhttp.WithTransport(exampleTextTransport("client"))))
	resp := client.Get("https://example.invalid").Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: client <nil>
}

func ExampleNewIsolatedClient() {
	client := vhttp.NewIsolatedClient(vhttp.WithClientRequestOptions(vhttp.WithTransport(exampleTextTransport("isolated-client"))))
	resp := client.Get("https://example.invalid").Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: isolated-client <nil>
}

func ExampleNewClientWithConfig() {
	cfg := vhttp.SnapshotGlobalConfig()
	cfg.Headers.Set("X-Client-Config", "yes")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Client-Config")))
	}))
	defer server.Close()

	client := vhttp.NewClientWithConfig(cfg)
	resp := client.Get(server.URL).Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: yes <nil>
}

func ExampleWithClientGlobalConfig() {
	cfg := vhttp.SnapshotGlobalConfig()
	cfg.Headers.Set("X-Client-Global", "yes")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Client-Global")))
	}))
	defer server.Close()

	client := vhttp.NewIsolatedClient(vhttp.WithClientGlobalConfig(cfg))
	resp := client.Get(server.URL).Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: yes <nil>
}

func ExampleWithClientRequestOptions() {
	client := vhttp.NewIsolatedClient(vhttp.WithClientRequestOptions(vhttp.WithHeader("X-Client", "yes")))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Client")))
	}))
	defer server.Close()

	resp := client.Get(server.URL).Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: yes <nil>
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

func ExampleGetWithParamsEWithOptions() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "%s:%s", r.URL.Query().Get("name"), r.Header.Get("X-Trace"))
	}))
	defer server.Close()

	body, err := vhttp.GetWithParamsEWithOptions(server.URL, map[string]any{"name": "go"}, vhttp.WithHeader("X-Trace", "trace"))
	fmt.Println(body, err)
	// Output: go:trace <nil>
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

func ExampleGetWithTimeoutEWithOptions() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Timeout")))
	}))
	defer server.Close()

	body, err := vhttp.GetWithTimeoutEWithOptions(server.URL, time.Second, vhttp.WithHeader("X-Timeout", "set"))
	fmt.Println(body, err)
	// Output: set <nil>
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

func ExamplePostFormEWithOptions() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		_, _ = fmt.Fprintf(w, "%s:%s", r.Form.Get("name"), r.Header.Get("X-Form"))
	}))
	defer server.Close()

	body, err := vhttp.PostFormEWithOptions(server.URL, map[string]any{"name": "go"}, vhttp.WithHeader("X-Form", "ok"))
	fmt.Println(body, err)
	// Output: go:ok <nil>
}

func ExamplePostFormSafeE() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		_, _ = fmt.Fprintf(w, "%s:%s", r.Method, r.Form.Get("name"))
	}))
	defer server.Close()

	body, err := vhttp.PostFormSafeE(server.URL, map[string]any{"name": "safe"}, localURLPolicy())
	fmt.Println(body, err)
	// Output: POST:safe <nil>
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

func ExamplePostJSONSafeE() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_, _ = fmt.Fprintf(w, "%s:%s", r.Header.Get("Content-Type"), body)
	}))
	defer server.Close()

	body, err := vhttp.PostJSONSafeE(server.URL, `{"safe":true}`, localURLPolicy())
	fmt.Println(body, err)
	// Output: application/json;charset=UTF-8:{"safe":true} <nil>
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

func ExamplePostStringSafeE() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_, _ = fmt.Fprintf(w, "%s:%s", r.Method, body)
	}))
	defer server.Close()

	body, err := vhttp.PostStringSafeE(server.URL, "safe-payload", localURLPolicy())
	fmt.Println(body, err)
	// Output: POST:safe-payload <nil>
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

func ExampleDownloadBytesSafeE() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("safe-bytes"))
	}))
	defer server.Close()

	body, err := vhttp.DownloadBytesSafeE(server.URL, localURLPolicy())
	fmt.Println(string(body), err)
	// Output: safe-bytes <nil>
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

func ExampleDownloadStringEWithOptions() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Text")))
	}))
	defer server.Close()

	body, err := vhttp.DownloadStringEWithOptions(server.URL, "", vhttp.WithHeader("X-Text", "option"))
	fmt.Println(body, err)
	// Output: option <nil>
}

func ExampleDownloadStringSafeE() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("safe-text"))
	}))
	defer server.Close()

	body, err := vhttp.DownloadStringSafeE(server.URL, "", localURLPolicy())
	fmt.Println(body, err)
	// Output: safe-text <nil>
}

func ExampleDownloadWithOptions() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Download")))
	}))
	defer server.Close()

	var buf bytes.Buffer
	n, err := vhttp.DownloadWithOptions(server.URL, &buf, vhttp.WithHeader("X-Download", "option"))
	fmt.Println(n, buf.String(), err)
	// Output: 6 option <nil>
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

func ExampleDownloadFile() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("file"))
	}))
	defer server.Close()

	dir, _ := os.MkdirTemp("", "vhttp-download-*")
	defer os.RemoveAll(dir)
	dest := filepath.Join(dir, "download.txt")
	n, err := vhttp.DownloadFile(server.URL, dest)
	data, _ := os.ReadFile(dest)
	fmt.Println(n, string(data), err)
	// Output: 4 file <nil>
}

func ExampleDownloadFileWithOptions() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-File")))
	}))
	defer server.Close()

	dir, _ := os.MkdirTemp("", "vhttp-download-options-*")
	defer os.RemoveAll(dir)
	dest := filepath.Join(dir, "download.txt")
	n, err := vhttp.DownloadFileWithOptions(server.URL, dest,
		[]vhttp.RequestOption{vhttp.WithHeader("X-File", "saved")},
		vhttp.WithSaveOverwrite(true),
	)
	data, _ := os.ReadFile(dest)
	fmt.Println(n, string(data), err)
	// Output: 5 saved <nil>
}

func ExampleDownloadFileSafeWithOptions() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("safe-options"))
	}))
	defer server.Close()

	dir, _ := os.MkdirTemp("", "vhttp-safe-download-options-*")
	defer os.RemoveAll(dir)
	dest := filepath.Join(dir, "download.txt")
	n, err := vhttp.DownloadFileSafeWithOptions(server.URL, dest,
		[]vhttp.RequestOption{localURLPolicy()},
		vhttp.WithSaveOverwrite(true),
	)
	data, _ := os.ReadFile(dest)
	fmt.Println(n, string(data), err)
	// Output: 12 safe-options <nil>
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

	resp := vhttp.Get(server.URL, vhttp.WithUserAgent("knifer-go-example")).Execute()
	fmt.Println(resp.Body())
	// Output: knifer-go-example
}

func ExampleWithTimeout() {
	fmt.Println(vhttp.WithTimeout(time.Second) != nil)
	// Output: true
}

func ExampleWithFollowRedirects() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/next", http.StatusFound)
	}))
	defer server.Close()

	resp := vhttp.Get(server.URL, vhttp.WithFollowRedirects(false)).Execute()
	fmt.Println(resp.Status(), resp.Header("Location"), resp.Err())
	// Output: 302 /next <nil>
}

func ExampleWithMaxRedirects() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/again", http.StatusFound)
	}))
	defer server.Close()

	resp := vhttp.Get(server.URL, vhttp.WithMaxRedirects(1)).Execute()
	fmt.Println(resp.Err() != nil)
	// Output: true
}

func ExampleWithTLSConfig() {
	resp := vhttp.Get("https://example.invalid",
		vhttp.WithTLSConfig(&tls.Config{ServerName: "example.invalid"}),
		vhttp.WithTransport(exampleTextTransport("transport-wins")),
	).Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: transport-wins <nil>
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

func ExampleWithTransportProvider() {
	called := 0
	resp := vhttp.Get("https://example.invalid", vhttp.WithTransportProvider(func() http.RoundTripper {
		called++
		return exampleTextTransport("provided")
	})).Execute()
	fmt.Println(resp.Body(), called, resp.Err())
	// Output: provided 1 <nil>
}

func ExampleConfigureDefaultTransportProvider() {
	vhttp.ConfigureDefaultTransportProvider(func() *http.Transport {
		return &http.Transport{}
	})
	defer vhttp.ResetDefaultTransport()

	fmt.Println("provider configured")
	// Output: provider configured
}

func ExampleResetDefaultTransport() {
	vhttp.ConfigureDefaultTransportProvider(func() *http.Transport {
		return &http.Transport{}
	})
	vhttp.ResetDefaultTransport()

	fmt.Println("provider reset")
	// Output: provider reset
}

func ExampleWithClient() {
	client := &http.Client{Transport: exampleTextTransport("custom-client")}
	resp := vhttp.Get("https://example.invalid", vhttp.WithClient(client)).Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: custom-client <nil>
}

func ExampleWithCookieJar() {
	jar, _ := cookiejar.New(nil)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "stored"})
		_, _ = w.Write([]byte(r.Header.Get("Cookie")))
	}))
	defer server.Close()

	req := vhttp.Get(server.URL, vhttp.WithCookieJar(jar))
	fmt.Println(req.Execute().Body())
	fmt.Println(req.Execute().Body())
	// Output:
	//
	// session=stored
}

func ExampleWithGlobalConfig() {
	cfg := vhttp.SnapshotGlobalConfig()
	cfg.Headers.Set("X-Request-Config", "yes")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Request-Config")))
	}))
	defer server.Close()

	resp := vhttp.Get(server.URL, vhttp.WithGlobalConfig(cfg)).Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: yes <nil>
}

func ExampleWithContentType() {
	fmt.Println(examplePostContentType(vhttp.WithContentType("text/plain;charset=utf-8")))
	// Output: text/plain;charset=utf-8
}

func ExampleWithCharset() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("Content-Type")))
	}))
	defer server.Close()

	resp := vhttp.Post(server.URL, vhttp.WithCharset("GBK")).BodyJSON(`{"ok":true}`).Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: application/json;charset=GBK <nil>
}

func ExampleWithAutoDecodeResponse() {
	resp := vhttp.Get("https://example.invalid",
		vhttp.WithTransport(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Encoding": []string{"custom"}},
				Body:       io.NopCloser(strings.NewReader("encoded")),
				Request:    req,
			}, nil
		})),
		vhttp.WithContentDecoder("custom", func(r io.Reader) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("decoded")), nil
		}),
		vhttp.WithAutoDecodeResponse(false),
	).Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: encoded <nil>
}

func ExampleWithMaxResponseBytes() {
	resp := vhttp.Get("https://example.invalid",
		vhttp.WithTransport(exampleTextTransport("too-long")),
		vhttp.WithMaxResponseBytes(3),
	).Execute()
	fmt.Println(resp.Body() == "", resp.Err() != nil)
	// Output: true true
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

func ExampleWithContentDecoder() {
	resp := vhttp.Get("https://example.invalid",
		vhttp.WithTransport(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Encoding": []string{"custom"}},
				Body:       io.NopCloser(strings.NewReader("encoded")),
				Request:    req,
			}, nil
		})),
		vhttp.WithContentDecoder("custom", func(r io.Reader) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("decoded")), nil
		}),
	).Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: decoded <nil>
}

func ExampleWithRequestFactory() {
	called := false
	resp := vhttp.Get("https://example.invalid",
		vhttp.WithRequestFactory(func(method, rawURL string, body io.Reader) (*http.Request, error) {
			called = true
			return http.NewRequest(method, rawURL, body)
		}),
		vhttp.WithTransport(exampleTextTransport("factory")),
	).Execute()
	fmt.Println(resp.Body(), called)
	// Output: factory true
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

func ExampleWithSaveOverwrite() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("new"))
	}))
	defer server.Close()

	dir, _ := os.MkdirTemp("", "vhttp-save-overwrite-*")
	defer os.RemoveAll(dir)
	dest := filepath.Join(dir, "download.txt")
	_ = os.WriteFile(dest, []byte("old"), 0o644)
	_, err := vhttp.DownloadFile(server.URL, dest, vhttp.WithSaveOverwrite(false))
	fmt.Println(err != nil)
	// Output: true
}

func ExampleWithSaveCreateParents() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("body"))
	}))
	defer server.Close()

	dir, _ := os.MkdirTemp("", "vhttp-save-parents-*")
	defer os.RemoveAll(dir)
	dest := filepath.Join(dir, "missing", "download.txt")
	_, err := vhttp.DownloadFile(server.URL, dest, vhttp.WithSaveCreateParents(false))
	fmt.Println(err != nil)
	// Output: true
}

func ExampleWithSaveDefaultFilename() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("named"))
	}))
	defer server.Close()

	dir, _ := os.MkdirTemp("", "vhttp-save-default-*")
	defer os.RemoveAll(dir)
	n, err := vhttp.DownloadFile(server.URL, dir, vhttp.WithSaveDefaultFilename("fallback.txt"))
	data, _ := os.ReadFile(filepath.Join(dir, "fallback.txt"))
	fmt.Println(n, string(data), err)
	// Output: 5 named <nil>
}

func ExampleSetGlobalHeader() {
	previous := vhttp.SnapshotGlobalConfig()
	defer vhttp.ConfigureGlobalConfig(previous)

	vhttp.SetGlobalHeader("X-Example", "one")
	fmt.Println(vhttp.CloneGlobalHeaders().Get("X-Example"))
	// Output: one
}

func ExampleCloneGlobalHeaders() {
	previous := vhttp.SnapshotGlobalConfig()
	defer vhttp.ConfigureGlobalConfig(previous)

	vhttp.SetGlobalHeader("X-Example", "one")
	cloned := vhttp.CloneGlobalHeaders()
	cloned.Set("X-Example", "changed")
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

func ExampleSnapshotGlobalConfig() {
	cfg := vhttp.SnapshotGlobalConfig()
	fmt.Println(cfg.Timeout > 0, cfg.MaxRedirects > 0)
	// Output: true true
}

func ExampleConfigureGlobalConfig() {
	previous := vhttp.SnapshotGlobalConfig()
	defer vhttp.ConfigureGlobalConfig(previous)

	cfg := previous
	cfg.DefaultUserAgent = "configured-agent"
	vhttp.ConfigureGlobalConfig(cfg)
	fmt.Println(vhttp.GetGlobalUserAgent())
	// Output: configured-agent
}

func ExampleResetGlobalConfig() {
	previous := vhttp.SnapshotGlobalConfig()
	defer vhttp.ConfigureGlobalConfig(previous)

	vhttp.SetGlobalUserAgent("temporary")
	vhttp.ResetGlobalConfig()
	fmt.Println(vhttp.GetGlobalUserAgent() == "")
	// Output: true
}

func ExampleSetGlobalTimeout() {
	previous := vhttp.SnapshotGlobalConfig()
	defer vhttp.ConfigureGlobalConfig(previous)

	vhttp.SetGlobalTimeout(2 * time.Second)
	fmt.Println(vhttp.GetGlobalTimeout())
	// Output: 2s
}

func ExampleGetGlobalTimeout() {
	fmt.Println(vhttp.GetGlobalTimeout() > 0)
	// Output: true
}

func ExampleSetGlobalMaxRedirects() {
	previous := vhttp.SnapshotGlobalConfig()
	defer vhttp.ConfigureGlobalConfig(previous)

	vhttp.SetGlobalMaxRedirects(3)
	fmt.Println(vhttp.GetGlobalMaxRedirects())
	// Output: 3
}

func ExampleGetGlobalMaxRedirects() {
	fmt.Println(vhttp.GetGlobalMaxRedirects() > 0)
	// Output: true
}

func ExampleSetGlobalMaxResponseBytes() {
	previous := vhttp.SnapshotGlobalConfig()
	defer vhttp.ConfigureGlobalConfig(previous)

	vhttp.SetGlobalMaxResponseBytes(1024)
	fmt.Println(vhttp.GetGlobalMaxResponseBytes())
	// Output: 1024
}

func ExampleGetGlobalMaxResponseBytes() {
	fmt.Println(vhttp.GetGlobalMaxResponseBytes() > 0)
	// Output: true
}

func ExampleSetGlobalFollowRedirects() {
	previous := vhttp.SnapshotGlobalConfig()
	defer vhttp.ConfigureGlobalConfig(previous)

	vhttp.SetGlobalFollowRedirects(false)
	fmt.Println(vhttp.GetGlobalFollowRedirects())
	// Output: false
}

func ExampleGetGlobalFollowRedirects() {
	fmt.Println(vhttp.GetGlobalFollowRedirects())
	// Output: true
}

func ExampleSetGlobalUserAgent() {
	previous := vhttp.SnapshotGlobalConfig()
	defer vhttp.ConfigureGlobalConfig(previous)

	vhttp.SetGlobalUserAgent("global-agent")
	fmt.Println(vhttp.GetGlobalUserAgent())
	// Output: global-agent
}

func ExampleGetGlobalUserAgent() {
	previous := vhttp.SnapshotGlobalConfig()
	defer vhttp.ConfigureGlobalConfig(previous)

	vhttp.SetGlobalUserAgent("reader")
	fmt.Println(vhttp.GetGlobalUserAgent())
	// Output: reader
}

func ExampleSetIgnoreEOFError() {
	previous := vhttp.SnapshotGlobalConfig()
	defer vhttp.ConfigureGlobalConfig(previous)

	vhttp.SetIgnoreEOFError(false)
	fmt.Println(vhttp.IsIgnoreEOFError())
	// Output: false
}

func ExampleIsIgnoreEOFError() {
	fmt.Println(vhttp.IsIgnoreEOFError())
	// Output: true
}

func ExampleSetGlobalBoundary() {
	previous := vhttp.SnapshotGlobalConfig()
	defer vhttp.ConfigureGlobalConfig(previous)

	vhttp.SetGlobalBoundary("boundary")
	fmt.Println(vhttp.GetGlobalBoundary())
	// Output: boundary
}

func ExampleGetGlobalBoundary() {
	fmt.Println(vhttp.GetGlobalBoundary() != "")
	// Output: true
}

func ExampleSetGlobalDecodeURL() {
	previous := vhttp.SnapshotGlobalConfig()
	defer vhttp.ConfigureGlobalConfig(previous)

	vhttp.SetGlobalDecodeURL(true)
	fmt.Println(vhttp.IsGlobalDecodeURL())
	// Output: true
}

func ExampleIsGlobalDecodeURL() {
	fmt.Println(vhttp.IsGlobalDecodeURL())
	// Output: false
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

func ExampleCleanHTMLWithOptions() {
	onlyParagraphs := regexp.MustCompile(`(?is)</?p[^>]*>`)
	fmt.Println(vhttp.CleanHTMLWithOptions(`<p>Hello</p><b>Go</b>`, vhttp.WithHTMLTagRegexp(onlyParagraphs)))
	// Output: Hello<b>Go</b>
}

func ExampleFilterHTMLTag() {
	fmt.Println(vhttp.FilterHTMLTag(`<p>Hello <b>Go</b></p>`, "b"))
	// Output: <p>Hello </p>
}

func ExampleFilterHTMLTagWithOptions() {
	compileCalled := false
	out := vhttp.FilterHTMLTagWithOptions(`<p>Hello <em>Go</em></p>`, []string{"em"}, vhttp.WithHTMLFilterCompileFunc(func(pattern string) (*regexp.Regexp, error) {
		compileCalled = true
		return regexp.Compile(pattern)
	}))
	fmt.Println(out, compileCalled)
	// Output: <p>Hello </p> true
}

func ExampleWithHTMLTagRegexp() {
	re := regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	fmt.Println(vhttp.CleanHTMLWithOptions(`<p>Hello</p><script>bad()</script>`, vhttp.WithHTMLTagRegexp(re)))
	// Output: <p>Hello</p>
}

func ExampleWithHTMLCommentRegexp() {
	re := regexp.MustCompile(`(?is)<!--.*?-->`)
	fmt.Println(vhttp.CleanHTMLWithOptions(`<!--drop--><p>Hello</p>`, vhttp.WithHTMLCommentRegexp(re)))
	// Output: Hello
}

func ExampleWithHTMLFilterCompileFunc() {
	fmt.Println(vhttp.FilterHTMLTagWithOptions(`<p>Hello <b>Go</b></p>`, []string{"b"}, vhttp.WithHTMLFilterCompileFunc(regexp.Compile)))
	// Output: <p>Hello </p>
}

func ExampleGetCharsetFromContentType() {
	fmt.Println(vhttp.GetCharsetFromContentType("text/plain;charset=gbk"))
	// Output: gbk
}

func ExampleGetCharsetFromContentTypeWithOptions() {
	re := regexp.MustCompile(`(?i)encoding=([^;]+)`)
	fmt.Println(vhttp.GetCharsetFromContentTypeWithOptions("text/plain;encoding=big5", vhttp.WithCharsetRegexp(re)))
	// Output: big5
}

func ExampleGetCharsetFromHTMLWithOptions() {
	re := regexp.MustCompile(`(?i)data-charset=["']?([^"'>\s]+)`)
	fmt.Println(vhttp.GetCharsetFromHTMLWithOptions(`<meta data-charset="shift_jis">`, vhttp.WithMetaCharsetRegexp(re)))
	// Output: shift_jis
}

func ExampleWithCharsetRegexp() {
	re := regexp.MustCompile(`(?i)encoding=([^;]+)`)
	fmt.Println(vhttp.GetCharsetFromContentTypeWithOptions("text/plain;encoding=utf-16", vhttp.WithCharsetRegexp(re)))
	// Output: utf-16
}

func ExampleWithMetaCharsetRegexp() {
	re := regexp.MustCompile(`(?i)data-charset=["']?([^"'>\s]+)`)
	fmt.Println(vhttp.GetCharsetFromHTMLWithOptions(`<meta data-charset="utf-8">`, vhttp.WithMetaCharsetRegexp(re)))
	// Output: utf-8
}

func ExampleGetMimeType() {
	fmt.Println(vhttp.GetMimeType("report.json"))
	// Output: application/json
}

func ExampleGuessContentType() {
	fmt.Println(vhttp.GuessContentType(`{"ok":true}`))
	fmt.Println(vhttp.GuessContentType(`<xml/>`))
	// Output:
	// application/json
	// application/xml
}

func ExampleIsDefaultContentType() {
	fmt.Println(vhttp.IsDefaultContentType(""))
	fmt.Println(vhttp.IsDefaultContentType("application/json"))
	// Output:
	// true
	// false
}

func ExampleIsFormURLEncoded() {
	fmt.Println(vhttp.IsFormURLEncoded("application/x-www-form-urlencoded;charset=utf-8"))
	// Output: true
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

func ExampleNewSimpleServer() {
	server := vhttp.NewSimpleServer(8080)
	server.AddAction("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("pong"))
	})
	fmt.Println(server != nil)
	// Output: true
}

func ExampleNewSimpleServerWithOptions() {
	server := vhttp.NewSimpleServerWithOptions(8080, vhttp.WithReadHeaderTimeout(time.Second))
	fmt.Println(server != nil)
	// Output: true
}

func ExampleNewSimpleServerAddr() {
	server := vhttp.NewSimpleServerAddr("127.0.0.1:0")
	fmt.Println(server != nil)
	// Output: true
}

func ExampleNewSimpleServerAddrWithOptions() {
	server := vhttp.NewSimpleServerAddrWithOptions("127.0.0.1:0", vhttp.WithIdleTimeout(time.Second))
	fmt.Println(server != nil)
	// Output: true
}

func ExampleCreateServer() {
	server := vhttp.CreateServer(8080)
	fmt.Println(server != nil)
	// Output: true
}

func ExampleCreateServerWithOptions() {
	server := vhttp.CreateServerWithOptions(8080, vhttp.WithWriteTimeout(time.Second))
	fmt.Println(server != nil)
	// Output: true
}

func ExampleWithListenAndServeFunc() {
	called := false
	server := vhttp.NewSimpleServerAddrWithOptions("127.0.0.1:0", vhttp.WithListenAndServeFunc(func(*http.Server) error {
		called = true
		return nil
	}))
	err := server.Start()
	fmt.Println(called, err)
	// Output: true <nil>
}

func ExampleWithAsyncRunner() {
	ran := false
	server := vhttp.NewSimpleServerAddrWithOptions("127.0.0.1:0",
		vhttp.WithListenAndServeFunc(func(*http.Server) error {
			return nil
		}),
		vhttp.WithAsyncRunner(func(fn func()) {
			ran = true
			fn()
		}),
	)
	err, ok := <-server.StartAsync()
	fmt.Println(ran, ok, err)
	// Output: true false <nil>
}

func ExampleResetServerStarters() {
	vhttp.ResetServerStarters()
	fmt.Println("server starters reset")
	// Output: server starters reset
}

func ExampleWithReadHeaderTimeout() {
	server := vhttp.NewSimpleServerAddrWithOptions("127.0.0.1:0", vhttp.WithReadHeaderTimeout(time.Second))
	fmt.Println(server != nil)
	// Output: true
}

func ExampleWithReadTimeout() {
	server := vhttp.NewSimpleServerAddrWithOptions("127.0.0.1:0", vhttp.WithReadTimeout(time.Second))
	fmt.Println(server != nil)
	// Output: true
}

func ExampleWithWriteTimeout() {
	server := vhttp.NewSimpleServerAddrWithOptions("127.0.0.1:0", vhttp.WithWriteTimeout(time.Second))
	fmt.Println(server != nil)
	// Output: true
}

func ExampleWithIdleTimeout() {
	server := vhttp.NewSimpleServerAddrWithOptions("127.0.0.1:0", vhttp.WithIdleTimeout(time.Second))
	fmt.Println(server != nil)
	// Output: true
}

func ExampleWithServerErrorLog() {
	logger := log.New(io.Discard, "", 0)
	server := vhttp.NewSimpleServerAddrWithOptions("127.0.0.1:0", vhttp.WithServerErrorLog(logger))
	fmt.Println(server != nil)
	// Output: true
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

func exampleSafeMethod(newRequest func(string, ...vhttp.RequestOption) *vhttp.Request, want string) string {
	return exampleMethod(func(rawURL string, opts ...vhttp.RequestOption) *vhttp.Request {
		opts = append(opts, localURLPolicy())
		return newRequest(rawURL, opts...)
	}, want)
}

func localURLPolicy() vhttp.RequestOption {
	return vhttp.WithURLPolicy(vhttp.URLPolicy{
		AllowedSchemes: []string{"http"},
		RejectPrivate:  false,
	})
}

func exampleTextTransport(body string) http.RoundTripper {
	return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})
}

func examplePostContentType(opt vhttp.RequestOption) string {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("Content-Type")))
	}))
	defer server.Close()

	return vhttp.Post(server.URL, opt).BodyString("payload").Execute().Body()
}
