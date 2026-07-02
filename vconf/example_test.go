package vconf_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/imajinyun/knifer-go/vconf"
)

type exampleServerConfig struct {
	Host    string `conf:"host"`
	Port    int    `conf:"port"`
	Enabled bool   `conf:"enabled"`
}

type exampleSchemaConfig struct {
	Port int    `conf:"port,required,int"`
	Mode string `conf:"mode,choices=dev|prod,default=dev"`
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

func ExampleParseByExt() {
	cfg, err := vconf.ParseByExt("app.yaml", []byte("server:\n  port: 8080\n"))
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(cfg.GetByGroup("server", "port"))
	// Output: 8080
}

func ExampleLoadWithOptions() {
	dir, err := os.MkdirTemp("", "vconf-example-*")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() { _ = os.RemoveAll(dir) }()

	path := filepath.Join(dir, "app.conf")
	if err := os.WriteFile(path, []byte("name=knifer-go\n"), 0o600); err != nil {
		fmt.Println(err)
		return
	}
	cfg, err := vconf.LoadWithOptions(path, vconf.LoadOptions{MaxBytes: 64})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(cfg.Get("name"))
	// Output: knifer-go
}

func ExampleLoadFiles() {
	dir, err := os.MkdirTemp("", "vconf-files-*")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() { _ = os.RemoveAll(dir) }()

	base := filepath.Join(dir, "base.conf")
	override := filepath.Join(dir, "override.conf")
	_ = os.WriteFile(base, []byte("name=base\nport=8080\n"), 0o600)
	_ = os.WriteFile(override, []byte("name=override\n"), 0o600)

	cfg, err := vconf.LoadFiles(base, override)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(cfg.Get("name"))
	fmt.Println(cfg.Get("port"))
	// Output:
	// override
	// 8080
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

func ExampleParseTOMLWithOptions() {
	cfg, err := vconf.ParseTOMLWithOptions("ignored", vconf.WithTOMLUnmarshalFunc(func(_ []byte, out any) error {
		root := out.(*map[string]any)
		*root = map[string]any{"app": map[string]any{"name": "provider"}}
		return nil
	}))
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(cfg.GetByGroup("app", "name"))
	// Output: provider
}

func ExampleParseYAMLFullWithOptions() {
	cfg, err := vconf.ParseYAMLFullWithOptions("ignored", vconf.WithYAMLUnmarshalFunc(func(_ []byte, out any) error {
		root := out.(*any)
		*root = map[string]any{"server": map[string]any{"port": 9090}}
		return nil
	}))
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(cfg.GetByGroup("server", "port"))
	// Output: 9090
}

func ExampleLoadProfile() {
	dir, err := os.MkdirTemp("", "vconf-profile-*")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() { _ = os.RemoveAll(dir) }()

	path := filepath.Join(dir, "app.toml")
	content := "name = \"base\"\n[profile.dev]\nname = \"dev\"\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		fmt.Println(err)
		return
	}

	cfg, err := vconf.LoadProfile(path, "dev")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(cfg.Get("name"))
	// Output: dev
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

func ExampleSchemaFromStruct() {
	schema, err := vconf.SchemaFromStruct(exampleSchemaConfig{})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(len(schema.Fields))
	fmt.Println(schema.Fields[0].Key, schema.Fields[0].Required, schema.Fields[0].Type)
	// Output:
	// 2
	// port true int
}

func ExampleConf_ApplyDefaults() {
	cfg, _ := vconf.Parse("port=8080\n")
	schema, _ := vconf.SchemaFromStruct(exampleSchemaConfig{})
	withDefaults := cfg.ApplyDefaults(schema)

	fmt.Println(withDefaults.Get("mode"))
	// Output: dev
}

func ExampleConf_ValidateSchema() {
	cfg, _ := vconf.Parse("port=8080\nmode=prod\n")
	schema, _ := vconf.SchemaFromStruct(exampleSchemaConfig{})

	fmt.Println(cfg.ValidateSchema(schema))
	// Output: <nil>
}

func ExampleBase64Decrypt() {
	value, err := vconf.Base64Decrypt("a25pZmVyLWdv")

	fmt.Println(value)
	fmt.Println(err)
	// Output:
	// knifer-go
	// <nil>
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
