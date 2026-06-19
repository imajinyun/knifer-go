package vresty_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/imajinyun/go-knifer/vresty"
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
