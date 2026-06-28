package vresty_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/imajinyun/knifer-go/vresty"
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

func ExamplePutSafe() {
	fmt.Println(exampleSafeMethod(vresty.PutSafe, http.MethodPut))
	// Output: PUT
}

func ExampleDelete() {
	fmt.Println(exampleMethod(vresty.Delete, http.MethodDelete))
	// Output: DELETE
}

func ExampleDeleteSafe() {
	fmt.Println(exampleSafeMethod(vresty.DeleteSafe, http.MethodDelete))
	// Output: DELETE
}

func ExamplePatch() {
	fmt.Println(exampleMethod(vresty.Patch, http.MethodPatch))
	// Output: PATCH
}

func ExamplePatchSafe() {
	fmt.Println(exampleSafeMethod(vresty.PatchSafe, http.MethodPatch))
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

func ExampleHeadSafe() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Method", r.Method)
	}))
	defer server.Close()

	resp := vresty.HeadSafe(server.URL, vresty.WithURLPolicy(localURLPolicy())).Execute()
	fmt.Println(resp.Header("X-Method"), resp.Err())
	// Output: HEAD <nil>
}

func ExampleOptions() {
	fmt.Println(exampleMethod(vresty.Options, http.MethodOptions))
	// Output: OPTIONS
}

func ExampleOptionsSafe() {
	fmt.Println(exampleSafeMethod(vresty.OptionsSafe, http.MethodOptions))
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

func ExampleNewClient_timeout() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Timeout-Client")))
	}))
	defer server.Close()

	client := vresty.NewClient(
		vresty.WithClientRequestOptions(
			vresty.WithTimeout(time.Second),
			vresty.WithHeader("X-Timeout-Client", "configured"),
		),
	)
	resp := client.Get(server.URL).Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: configured <nil>
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

func ExampleWithClientGlobalConfig() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Client-Global")))
	}))
	defer server.Close()

	cfg := vresty.SnapshotGlobalConfig()
	cfg.Headers = vresty.HeaderValues{"X-Client-Global": []string{"yes"}}
	client := vresty.NewIsolatedClient(vresty.WithClientGlobalConfig(cfg))
	resp := client.Get(server.URL).Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: yes <nil>
}

func ExampleWithClientRequestOptions() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Client-Option")))
	}))
	defer server.Close()

	client := vresty.NewIsolatedClient(vresty.WithClientRequestOptions(vresty.WithHeader("X-Client-Option", "yes")))
	resp := client.Get(server.URL).Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: yes <nil>
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

func ExamplePostFormSafeE() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		_, _ = fmt.Fprintf(w, "%s:%s", r.Method, r.Form.Get("name"))
	}))
	defer server.Close()

	body, err := vresty.PostFormSafeE(server.URL, map[string]any{"name": "go"}, vresty.WithURLPolicy(localURLPolicy()))
	fmt.Println(body, err)
	// Output: POST:go <nil>
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

func ExamplePostStringSafeE() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_, _ = fmt.Fprintf(w, "%s:%s", r.Method, body)
	}))
	defer server.Close()

	body, err := vresty.PostStringSafeE(server.URL, "safe", vresty.WithURLPolicy(localURLPolicy()))
	fmt.Println(body, err)
	// Output: POST:safe <nil>
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

func ExamplePostJSONSafeE() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_, _ = fmt.Fprintf(w, "%s:%s", r.Method, body)
	}))
	defer server.Close()

	body, err := vresty.PostJSONSafeE(server.URL, `{"safe":true}`, vresty.WithURLPolicy(localURLPolicy()))
	fmt.Println(body, err)
	// Output: POST:{"safe":true} <nil>
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

func ExampleDownloadFileSafe() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("safe-file"))
	}))
	defer server.Close()

	dir, _ := os.MkdirTemp("", "vresty-safe-download-*")
	defer os.RemoveAll(dir)
	dest := filepath.Join(dir, "download.txt")
	n, err := vresty.DownloadFileSafeWithOptions(server.URL, dest,
		[]vresty.RequestOption{vresty.WithURLPolicy(vresty.URLPolicy{AllowedSchemes: []string{"http"}, RejectPrivate: false})},
		vresty.WithSaveOverwrite(true),
	)
	data, _ := os.ReadFile(dest)
	fmt.Println(n, string(data), err)
	// Output: 9 safe-file <nil>
}

func ExampleDownloadFileSafeWithOptions() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Safe-File")))
	}))
	defer server.Close()

	dir, _ := os.MkdirTemp("", "vresty-safe-download-*")
	defer os.RemoveAll(dir)
	dest := filepath.Join(dir, "download.txt")
	n, err := vresty.DownloadFileSafeWithOptions(
		server.URL,
		dest,
		[]vresty.RequestOption{
			vresty.WithURLPolicy(localURLPolicy()),
			vresty.WithHeader("X-Safe-File", "safe-file-options"),
		},
		vresty.WithSaveOverwrite(true),
	)
	data, _ := os.ReadFile(dest)
	fmt.Println(n, string(data), err)
	// Output: 17 safe-file-options <nil>
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

func ExampleResponse_Result() {
	type payload struct {
		OK bool `json:"ok"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	resp := vresty.Get(server.URL).Result(&payload{}).Execute()
	result := resp.Result().(*payload)
	fmt.Println(result.OK, resp.Err())
	// Output: true <nil>
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

func ExampleWithGlobalConfig() {
	cfg := vresty.SnapshotGlobalConfig()
	cfg.Headers = vresty.HeaderValues{"X-Global-Config": []string{"yes"}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Global-Config")))
	}))
	defer server.Close()

	resp := vresty.Get(server.URL, vresty.WithGlobalConfig(cfg)).Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: yes <nil>
}

func ExampleWithUserAgent() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.UserAgent()))
	}))
	defer server.Close()

	resp := vresty.Get(server.URL, vresty.WithUserAgent("knifer-go-resty-example")).Execute()
	fmt.Println(resp.Body())
	// Output: knifer-go-resty-example
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

func ExampleWithFollowRedirects() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/start" {
			http.Redirect(w, r, "/next", http.StatusFound)
			return
		}
		_, _ = w.Write([]byte("followed"))
	}))
	defer server.Close()

	resp := vresty.Get(server.URL+"/start", vresty.WithFollowRedirects(false)).Execute()
	fmt.Println(resp.Status(), resp.Header("Location"))
	// Output: 302 /next
}

func ExampleWithMaxRedirects() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/loop", http.StatusFound)
	}))
	defer server.Close()

	resp := vresty.Get(server.URL+"/loop", vresty.WithMaxRedirects(1)).Execute()
	fmt.Println(resp.Err() != nil)
	// Output: true
}

func ExampleWithTLSConfig() {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("tls"))
	}))
	defer server.Close()

	resp := vresty.Get(server.URL, vresty.WithTLSConfig(&tls.Config{InsecureSkipVerify: true})).Execute() //nolint:gosec // Example uses httptest's self-signed certificate.
	fmt.Println(resp.Body(), resp.Err())
	// Output: tls <nil>
}

func ExampleWithCookieDisabled() {
	fmt.Println(vresty.WithCookieDisabled(true) != nil)
	// Output: true
}

func ExampleWithMaxResponseBytes() {
	fmt.Println(vresty.WithMaxResponseBytes(64) != nil)
	// Output: true
}

func ExampleWithMaxDecodeBytes() {
	type payload struct {
		Name string `json:"name"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"toolong"}`))
	}))
	defer server.Close()

	var out payload
	resp := vresty.Get(server.URL,
		vresty.WithMaxDecodeBytes(4),
		vresty.WithJSONUnmarshalFunc(json.Unmarshal),
	).Result(&out).Execute()
	fmt.Println(resp.Err() != nil)
	// Output: true
}

func ExampleWithJSONMarshalFunc() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_, _ = w.Write(body)
	}))
	defer server.Close()

	called := false
	resp := vresty.Post(server.URL, vresty.WithJSONMarshalFunc(func(any) ([]byte, error) {
		called = true
		return []byte(`{"provided":true}`), nil
	})).BodyJSONValue(map[string]any{"ignored": true}).Execute()
	fmt.Println(resp.Body(), called)
	// Output: {"provided":true} true
}

func ExampleWithJSONUnmarshalFunc() {
	type payload struct {
		Name string `json:"name"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"ignored"}`))
	}))
	defer server.Close()

	var out payload
	called := false
	resp := vresty.Get(server.URL, vresty.WithJSONUnmarshalFunc(func(_ []byte, dst any) error {
		called = true
		return json.Unmarshal([]byte(`{"name":"provided"}`), dst)
	})).Result(&out).Execute()
	fmt.Println(out.Name, called, resp.Err())
	// Output: provided true <nil>
}

func ExampleWithJSONDecodeReadAllFunc() {
	type payload struct {
		Name string `json:"name"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"ignored"}`))
	}))
	defer server.Close()

	var out payload
	readCalled := false
	resp := vresty.Get(server.URL,
		vresty.WithJSONDecodeReadAllFunc(func(io.Reader) ([]byte, error) {
			readCalled = true
			return []byte(`{"name":"provided"}`), nil
		}),
		vresty.WithJSONUnmarshalFunc(json.Unmarshal),
	).Result(&out).Execute()
	fmt.Println(out.Name, readCalled, resp.Err())
	// Output: provided true <nil>
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

func ExampleConfigureDefaultRestyClientProvider() {
	vresty.ConfigureDefaultRestyClientProvider(func() *grestry.Client {
		client := grestry.New()
		client.SetTransport(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("provider")),
				Header:     make(http.Header),
				Request:    req,
			}, nil
		}))
		return client
	})
	defer vresty.ResetDefaultRestyClientProvider()

	resp := vresty.Get("https://example.invalid").Execute()
	fmt.Println(resp.Body(), resp.Err())
	// Output: provider <nil>
}

func ExampleResetDefaultRestyClientProvider() {
	vresty.ConfigureDefaultRestyClientProvider(nil)
	resp := vresty.Get("https://example.invalid", vresty.WithRestyClientFactory(func() *grestry.Client {
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
	fmt.Println(resp.Body(), resp.Err())
	// Output: factory <nil>
}

func ExampleConfigureGlobalConfig() {
	previous := vresty.SnapshotGlobalConfig()
	defer vresty.ConfigureGlobalConfig(previous)

	cfg := previous
	cfg.DefaultUserAgent = "configured-agent"
	vresty.ConfigureGlobalConfig(cfg)
	fmt.Println(vresty.GetGlobalUserAgent())
	// Output: configured-agent
}

func ExampleResetGlobalConfig() {
	previous := vresty.SnapshotGlobalConfig()
	defer vresty.ConfigureGlobalConfig(previous)

	vresty.SetGlobalUserAgent("temporary-agent")
	vresty.ResetGlobalConfig()
	fmt.Println(vresty.GetGlobalUserAgent() == "")
	// Output: true
}

func ExampleSnapshotGlobalConfig() {
	cfg := vresty.SnapshotGlobalConfig()
	fmt.Println(cfg.Timeout > 0, cfg.Headers != nil)
	// Output: true true
}

func ExampleSetGlobalTimeout() {
	previous := vresty.SnapshotGlobalConfig()
	defer vresty.ConfigureGlobalConfig(previous)

	vresty.SetGlobalTimeout(2 * time.Second)
	fmt.Println(vresty.GetGlobalTimeout())
	// Output: 2s
}

func ExampleGetGlobalTimeout() {
	fmt.Println(vresty.GetGlobalTimeout() > 0)
	// Output: true
}

func ExampleSetGlobalMaxRedirects() {
	previous := vresty.SnapshotGlobalConfig()
	defer vresty.ConfigureGlobalConfig(previous)

	vresty.SetGlobalMaxRedirects(3)
	fmt.Println(vresty.GetGlobalMaxRedirects())
	// Output: 3
}

func ExampleGetGlobalMaxRedirects() {
	fmt.Println(vresty.GetGlobalMaxRedirects() > 0)
	// Output: true
}

func ExampleSetGlobalMaxResponseBytes() {
	previous := vresty.SnapshotGlobalConfig()
	defer vresty.ConfigureGlobalConfig(previous)

	vresty.SetGlobalMaxResponseBytes(1024)
	fmt.Println(vresty.GetGlobalMaxResponseBytes())
	// Output: 1024
}

func ExampleGetGlobalMaxResponseBytes() {
	fmt.Println(vresty.GetGlobalMaxResponseBytes() > 0)
	// Output: true
}

func ExampleSetGlobalFollowRedirects() {
	previous := vresty.SnapshotGlobalConfig()
	defer vresty.ConfigureGlobalConfig(previous)

	vresty.SetGlobalFollowRedirects(false)
	fmt.Println(vresty.GetGlobalFollowRedirects())
	// Output: false
}

func ExampleGetGlobalFollowRedirects() {
	fmt.Println(vresty.GetGlobalFollowRedirects())
	// Output: true
}

func ExampleSetGlobalUserAgent() {
	previous := vresty.SnapshotGlobalConfig()
	defer vresty.ConfigureGlobalConfig(previous)

	vresty.SetGlobalUserAgent("global-agent")
	fmt.Println(vresty.GetGlobalUserAgent())
	// Output: global-agent
}

func ExampleGetGlobalUserAgent() {
	previous := vresty.SnapshotGlobalConfig()
	defer vresty.ConfigureGlobalConfig(previous)

	vresty.SetGlobalUserAgent("read-agent")
	fmt.Println(vresty.GetGlobalUserAgent())
	// Output: read-agent
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

func ExampleCloneGlobalHeaders() {
	previous := vresty.SnapshotGlobalConfig()
	defer vresty.ConfigureGlobalConfig(previous)

	vresty.SetGlobalHeader("X-Example", "one")
	headers := vresty.CloneGlobalHeaders()
	headers["X-Example"][0] = "mutated"
	fmt.Println(vresty.CloneGlobalHeaders()["X-Example"][0])
	// Output: one
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

func ExampleWithCharsetRegexp() {
	custom := regexp.MustCompile(`(?i)encoding=([^;]+)`)
	fmt.Println(vresty.GetCharsetFromContentTypeWithOptions("text/plain; encoding=big5", vresty.WithCharsetRegexp(custom)))
	// Output: big5
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

func ExampleWithMetaCharsetRegexp() {
	custom := regexp.MustCompile(`(?i)data-charset=["']?([^"'>\s]+)`)
	fmt.Println(vresty.GetCharsetFromHTMLWithOptions(`<html data-charset="shift_jis">`, vresty.WithMetaCharsetRegexp(custom)))
	// Output: shift_jis
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

func ExampleWithSaveFilePerm() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("perm"))
	}))
	defer server.Close()

	dir, _ := os.MkdirTemp("", "vresty-save-*")
	defer os.RemoveAll(dir)
	dest := filepath.Join(dir, "download.txt")
	_, err := vresty.DownloadFile(server.URL, dest, vresty.WithSaveFilePerm(0o600))
	info, _ := os.Stat(dest)
	fmt.Printf("%#o %v\n", info.Mode().Perm(), err)
	// Output: 0600 <nil>
}

func ExampleWithSaveDirPerm() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("dir"))
	}))
	defer server.Close()

	dir, _ := os.MkdirTemp("", "vresty-save-*")
	defer os.RemoveAll(dir)
	destDir := filepath.Join(dir, "nested")
	dest := filepath.Join(destDir, "download.txt")
	_, err := vresty.DownloadFile(server.URL, dest, vresty.WithSaveDirPerm(0o700))
	info, _ := os.Stat(destDir)
	fmt.Printf("%#o %v\n", info.Mode().Perm(), err)
	// Output: 0700 <nil>
}

func ExampleWithSaveOverwrite() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("new"))
	}))
	defer server.Close()

	dir, _ := os.MkdirTemp("", "vresty-save-*")
	defer os.RemoveAll(dir)
	dest := filepath.Join(dir, "download.txt")
	_ = os.WriteFile(dest, []byte("old"), 0o600)
	_, err := vresty.DownloadFile(server.URL, dest, vresty.WithSaveOverwrite(false))
	data, _ := os.ReadFile(dest)
	fmt.Println(err != nil, string(data))
	// Output: true old
}

func ExampleWithSaveCreateParents() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("body"))
	}))
	defer server.Close()

	dir, _ := os.MkdirTemp("", "vresty-save-*")
	defer os.RemoveAll(dir)
	dest := filepath.Join(dir, "missing", "download.txt")
	_, err := vresty.DownloadFile(server.URL, dest, vresty.WithSaveCreateParents(false))
	fmt.Println(err != nil)
	// Output: true
}

func ExampleWithSaveDefaultFilename() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("fallback"))
	}))
	defer server.Close()

	dir, _ := os.MkdirTemp("", "vresty-save-*")
	defer os.RemoveAll(dir)
	n, err := vresty.DownloadFile(server.URL, dir, vresty.WithSaveDefaultFilename("fallback.txt"))
	data, _ := os.ReadFile(filepath.Join(dir, "fallback.txt"))
	fmt.Println(n, string(data), err)
	// Output: 8 fallback <nil>
}

func ExampleWithSaveStat() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("stat"))
	}))
	defer server.Close()

	dir, _ := os.MkdirTemp("", "vresty-save-*")
	defer os.RemoveAll(dir)
	statCalled := false
	_, err := vresty.DownloadFile(server.URL, filepath.Join(dir, "download.txt"), vresty.WithSaveStat(func(path string) (os.FileInfo, error) {
		statCalled = true
		return os.Stat(path)
	}))
	fmt.Println(statCalled, err)
	// Output: true <nil>
}

func ExampleWithSaveMkdirAll() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("mkdir"))
	}))
	defer server.Close()

	dir, _ := os.MkdirTemp("", "vresty-save-*")
	defer os.RemoveAll(dir)
	mkdirCalled := false
	dest := filepath.Join(dir, "nested", "download.txt")
	_, err := vresty.DownloadFile(server.URL, dest, vresty.WithSaveMkdirAll(func(path string, perm fs.FileMode) error {
		mkdirCalled = true
		return os.MkdirAll(path, perm)
	}))
	fmt.Println(mkdirCalled, err)
	// Output: true <nil>
}

func ExampleWithSaveOpenFile() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("open"))
	}))
	defer server.Close()

	var buf bytes.Buffer
	openCalled := false
	_, err := vresty.DownloadFile(server.URL, "ignored.txt",
		vresty.WithSaveCreateParents(false),
		vresty.WithSaveOpenFile(func(string, int, fs.FileMode) (io.WriteCloser, error) {
			openCalled = true
			return nopWriteCloser{Writer: &buf}, nil
		}),
	)
	fmt.Println(openCalled, buf.String(), err)
	// Output: true open <nil>
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

func exampleSafeMethod(newRequest func(string, ...vresty.RequestOption) *vresty.Request, want string) string {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Method))
	}))
	defer server.Close()

	resp := newRequest(server.URL, vresty.WithURLPolicy(localURLPolicy())).Execute()
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

func localURLPolicy() vresty.URLPolicy {
	return vresty.URLPolicy{AllowedSchemes: []string{"http", "https"}, RejectPrivate: false}
}
