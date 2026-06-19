package vhttp_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

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
