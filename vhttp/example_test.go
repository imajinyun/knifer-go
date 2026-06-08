package vhttp_test

import (
	"errors"
	"fmt"
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
		vhttp.WithAllowedHosts("127.0.0.1"),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(body)
	// Output: safe
}
