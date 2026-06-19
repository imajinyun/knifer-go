package vconf_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/imajinyun/go-knifer/vconf"
)

func ExampleParse() {
	cfg, err := vconf.Parse("app.name=go-knifer\napp.port=8080\n")
	if err != nil {
		fmt.Println(err)
		return
	}

	port := cfg.GetInt("app.port", 0)
	fmt.Println(cfg.Get("app.name"))
	fmt.Println(port)
	// Output:
	// go-knifer
	// 8080
}

func ExampleLoadRemoteSafeWithOptions() {
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("feature.enabled=true\n")),
			Request:    req,
		}, nil
	})}
	lookup := func(context.Context, string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("93.184.216.34")}, nil
	}

	cfg, err := vconf.LoadRemoteSafeWithOptions("https://config.example/app.setting", vconf.LoadOptions{
		RemoteClient:       client,
		RemoteAllowedHosts: []string{"config.example"},
		LookupIP:           lookup,
		MaxBytes:           64,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	enabled := cfg.GetBool("feature.enabled", false)
	fmt.Println(enabled)
	// Output: true
}

func ExampleParseBytes() {
	cfg, err := vconf.ParseBytes([]byte("name=go-knifer\n"))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(cfg.Get("name"))
	// Output: go-knifer
}

func ExampleParseYAML() {
	cfg, err := vconf.ParseYAML("server:\n  port: 8080\n")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(cfg.GetByGroup("server", "port"))
	// Output: 8080
}

func ExampleMerge() {
	base, _ := vconf.Parse("name=base\nport=8080\n")
	override, _ := vconf.Parse("name=override\n")
	cfg := vconf.Merge(base, override)
	fmt.Println(cfg.Get("name"))
	fmt.Println(cfg.Get("port"))
	// Output:
	// override
	// 8080
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }
