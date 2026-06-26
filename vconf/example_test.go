package vconf_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/imajinyun/knifer-go/vconf"
)

type exampleServerConfig struct {
	Host    string `conf:"host"`
	Port    int    `conf:"port"`
	Enabled bool   `conf:"enabled"`
}

func ExampleParse() {
	cfg, err := vconf.Parse("app.name=knifer-go\napp.port=8080\n")
	if err != nil {
		fmt.Println(err)
		return
	}

	port := cfg.GetInt("app.port", 0)
	fmt.Println(cfg.Get("app.name"))
	fmt.Println(port)
	// Output:
	// knifer-go
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
	cfg, err := vconf.ParseBytes([]byte("name=knifer-go\n"))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(cfg.Get("name"))
	// Output: knifer-go
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

func ExampleNew() {
	cfg := vconf.New()
	cfg.Set("app.name", "knifer-go")
	cfg.SetByGroup("server", "port", "8080")

	fmt.Println(cfg.Get("app.name"))
	fmt.Println(cfg.GetIntByGroup("server", "port", 0))
	// Output:
	// knifer-go
	// 8080
}

func ExampleParseByExtWithOptions() {
	cfg, err := vconf.ParseByExtWithOptions(
		"app.custom",
		[]byte("ignored"),
		vconf.WithParserForExt("custom", func([]byte) (*vconf.Conf, error) {
			parsed := vconf.New()
			parsed.Set("source", "custom")
			return parsed, nil
		}),
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(cfg.Get("source"))
	// Output: custom
}

func ExampleParseTOML() {
	cfg, err := vconf.ParseTOML("title = 'demo'\n[server]\nport = 8080\nenabled = true\n")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(cfg.Get("title"))
	fmt.Println(cfg.GetByGroup("server", "port"))
	fmt.Println(cfg.GetBoolByGroup("server", "enabled", false))
	// Output:
	// demo
	// 8080
	// true
}

func ExampleConf_Bind() {
	cfg, err := vconf.Parse("host=localhost\nport=8080\nenabled=true\n")
	if err != nil {
		fmt.Println(err)
		return
	}

	var server exampleServerConfig
	if err := cfg.Bind(&server); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(server.Host)
	fmt.Println(server.Port)
	fmt.Println(server.Enabled)
	// Output:
	// localhost
	// 8080
	// true
}

func ExampleConf_Clone() {
	cfg := vconf.New()
	cfg.Set("name", "original")

	clone := cfg.Clone()
	clone.Set("name", "clone")

	fmt.Println(cfg.Get("name"))
	fmt.Println(clone.Get("name"))
	// Output:
	// original
	// clone
}

func ExampleConf_ExpandWithOptions() {
	cfg := vconf.New()
	cfg.Set("host", "localhost")
	cfg.Set("dsn", "http://${host}:${ENV:PORT}")

	expanded := cfg.ExpandWithOptions(vconf.WithEnvLookup(func(name string) string {
		if name == "PORT" {
			return "8080"
		}
		return ""
	}))

	fmt.Println(expanded.Get("dsn"))
	// Output: http://localhost:8080
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }
