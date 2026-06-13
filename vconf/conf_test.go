package vconf_test

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	knifer "github.com/imajinyun/go-knifer"
	"github.com/imajinyun/go-knifer/vconf"
)

func TestParseSettingFacade(t *testing.T) {
	s, err := vconf.Parse("name=gokit\ncount=42\nenabled=true\n[server]\nhost=127.0.0.1\nport=8080")
	if err != nil {
		t.Fatal(err)
	}
	if got := s.Get("name"); got != "gokit" {
		t.Fatalf("Get(name) = %q", got)
	}
	if got := s.GetInt("count", 0); got != 42 {
		t.Fatalf("GetInt(count) = %d", got)
	}
	if got := s.GetBool("enabled", false); !got {
		t.Fatal("GetBool(enabled) = false")
	}
	if got := s.GetByGroup("server", "host"); got != "127.0.0.1" {
		t.Fatalf("GetByGroup(server, host) = %q", got)
	}
	s.SetByGroup("server", "scheme", "http")
	if got := s.GetByGroup("server", "scheme"); got != "http" {
		t.Fatalf("SetByGroup() value = %q", got)
	}
	if !reflect.DeepEqual(s.Keys("server"), []string{"host", "port", "scheme"}) {
		t.Fatalf("Keys(server) = %#v", s.Keys("server"))
	}
}

func TestLoadAndParseYAMLFacade(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.setting")
	if err := os.WriteFile(path, []byte("app='demo'\n[db]\nuser=root"), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := vconf.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := s.Get("app"); got != "demo" {
		t.Fatalf("Load Get(app) = %q", got)
	}
	if got := s.GetByGroup("db", "user"); got != "root" {
		t.Fatalf("Load GetByGroup(db, user) = %q", got)
	}

	s, err = vconf.ParseYAML("app: gokit\nserver:\n  port: 8080\n  debug: true")
	if err != nil {
		t.Fatal(err)
	}
	if got := s.GetByGroup("server", "debug"); got != "true" {
		t.Fatalf("ParseYAML server.debug = %q", got)
	}
}

func TestNewAndParseBytesFacade(t *testing.T) {
	s := vconf.New()
	s.Set("k", "v")
	if got := s.GetOrDefault("k", "default"); got != "v" {
		t.Fatalf("GetOrDefault(k) = %q", got)
	}
	parsed, err := vconf.ParseBytes([]byte("x: 1"))
	if err != nil {
		t.Fatal(err)
	}
	if got := parsed.Get("x"); got != "1" {
		t.Fatalf("ParseBytes Get(x) = %q", got)
	}
}

func TestNilConfFacadeReadMethodsAreEmptyAndSafe(t *testing.T) {
	var s *vconf.Conf

	if got := s.Groups(); len(got) != 0 {
		t.Fatalf("Groups() = %v, want empty", got)
	}
	if got := s.Keys("missing"); len(got) != 0 {
		t.Fatalf("Keys(missing) = %v, want empty", got)
	}
	if got := s.ToMap(); len(got) != 0 {
		t.Fatalf("ToMap() = %v, want empty", got)
	}
}

func TestFacadeConfErrorContract(t *testing.T) {
	_, err := vconf.Parse("invalid-line")
	if err == nil {
		t.Fatal("Parse() error = nil, want invalid input")
	}
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("errors.Is(err, ErrCodeInvalidInput) = false: %v", err)
	}
	code, ok := knifer.CodeOf(err)
	if !ok || code != knifer.ErrCodeInvalidInput {
		t.Fatalf("CodeOf(err) = %q, %v; want invalid input", code, ok)
	}
	var confErr *vconf.Error
	if !errors.As(err, &confErr) {
		t.Fatalf("errors.As(err, *vconf.Error) = false: %v", err)
	}
}

func TestAdvancedConfigFacade(t *testing.T) {
	t.Setenv("VCONF_HOST", "env.local")
	s, err := vconf.ParseTOML(`
name = "demo"
base = "http://${ENV:VCONF_HOST}"
[server]
port = 8080
debug = true
tags = ["api", "admin"]
[profile.prod.server]
port = 9090
`)
	if err != nil {
		t.Fatal(err)
	}
	if got := s.GetExpanded("base"); got != "http://env.local" {
		t.Fatalf("GetExpanded(base) = %q", got)
	}

	type serverConf struct {
		Port  int      `conf:"port"`
		Debug bool     `conf:"debug"`
		Tags  []string `conf:"tags"`
	}
	var cfg serverConf
	if err := s.BindGroup("server", &cfg); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cfg, serverConf{Port: 8080, Debug: true, Tags: []string{"api", "admin"}}) {
		t.Fatalf("BindGroup() = %#v", cfg)
	}
	prod := s.ApplyProfile("prod")
	if got := prod.GetByGroup("server", "port"); got != "9090" {
		t.Fatalf("ApplyProfile(prod).server.port = %q", got)
	}
}

func TestFacadeTypedGettersMutationAndSchemaMethods(t *testing.T) {
	s, err := vconf.Parse(`
name=demo
custom_int=custom
custom_bool=yes
[server]
port=8080
enabled=true
`)
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := s.Lookup("", "name"); !ok || got != "demo" {
		t.Fatalf("Lookup default = %q, %v", got, ok)
	}
	if got := s.GetIntWithOptions("custom_int", 0, vconf.WithIntParser(func(value string) (int, error) {
		if value == "custom" {
			return 77, nil
		}
		return 0, errors.New("unexpected")
	})); got != 77 {
		t.Fatalf("GetIntWithOptions = %d", got)
	}
	if got := s.GetBoolByGroupWithOptions("", "custom_bool", false, vconf.WithBoolParser(func(value string) (bool, error) {
		return value == "yes", nil
	})); !got {
		t.Fatal("GetBoolByGroupWithOptions = false")
	}
	if got := s.GetIntByGroupWithOptions("server", "port", 0); got != 8080 {
		t.Fatalf("GetIntByGroupWithOptions = %d", got)
	}
	if got := s.GetBoolByGroupWithOptions("server", "enabled", false); !got {
		t.Fatal("GetBoolByGroupWithOptions server.enabled = false")
	}

	clone := s.Clone()
	clone.Set("name", "clone")
	if s.Get("name") != "demo" || clone.Get("name") != "clone" {
		t.Fatalf("Clone should not alias source: source=%q clone=%q", s.Get("name"), clone.Get("name"))
	}
	clone.Delete("name")
	if _, ok := clone.Lookup("", "name"); ok {
		t.Fatal("Delete did not remove default key")
	}
	clone.DeleteByGroup("server", "enabled")
	if _, ok := clone.Lookup("server", "enabled"); ok {
		t.Fatal("DeleteByGroup did not remove grouped key")
	}
	merged := clone.Merge(vconf.Merge(s))
	if got := merged.Get("name"); got != "demo" {
		t.Fatalf("Merge method name = %q", got)
	}

	schema := vconf.Schema{Fields: []vconf.FieldRule{
		{Group: "server", Key: "host", Default: "127.0.0.1"},
		{Group: "server", Key: "port", Required: true, Type: vconf.TypeInt},
	}}
	withDefaults := s.ApplyDefaults(schema)
	if got := withDefaults.GetByGroup("server", "host"); got != "127.0.0.1" {
		t.Fatalf("ApplyDefaults host = %q", got)
	}
	if err := withDefaults.ValidateSchema(schema); err != nil {
		t.Fatalf("ValidateSchema: %v", err)
	}
	type defaultConfig struct {
		CustomInt int `conf:"custom_int,required,int"`
	}
	if err := withDefaults.ValidateStructWithOptions(defaultConfig{}, vconf.WithSchemaIntParser(func(value string, base int, bitSize int) (int64, error) {
		if value == "custom" {
			return 77, nil
		}
		return 0, errors.New("unexpected")
	})); err != nil {
		t.Fatalf("ValidateStructWithOptions: %v", err)
	}
}

func TestFacadeParserProviderOptions(t *testing.T) {
	parsed, err := vconf.ParseByExtWithOptions("app.custom", []byte("ignored"), vconf.WithParserForExt(".custom", func(data []byte) (*vconf.Conf, error) {
		c := vconf.New()
		c.Set("from", string(data))
		return c, nil
	}))
	if err != nil || parsed.Get("from") != "ignored" {
		t.Fatalf("ParseByExtWithOptions = %#v, %v", parsed, err)
	}

	yamlCalled := false
	_, err = vconf.ParseYAMLFullWithOptions("ignored", vconf.WithYAMLUnmarshalFunc(func(data []byte, out any) error {
		yamlCalled = true
		return errors.New("yaml provider failed")
	}))
	if err == nil || !yamlCalled {
		t.Fatalf("ParseYAMLFullWithOptions err=%v called=%v", err, yamlCalled)
	}

	tomlCalled := false
	_, err = vconf.ParseTOMLWithOptions("ignored", vconf.WithTOMLUnmarshalFunc(func(data []byte, out any) error {
		tomlCalled = true
		return errors.New("toml provider failed")
	}))
	if err == nil || !tomlCalled {
		t.Fatalf("ParseTOMLWithOptions err=%v called=%v", err, tomlCalled)
	}
}

func TestExpandOptionsFacade(t *testing.T) {
	s, err := vconf.Parse(`
host=${ENV:VCFG_HOST}
base=http://${host}:${port:8080}
`)
	if err != nil {
		t.Fatal(err)
	}
	lookup := func(key string) string {
		if key == "VCFG_HOST" {
			return "facade.local"
		}
		return ""
	}
	if got := s.GetExpandedWithOptions("host", vconf.WithEnvLookup(lookup)); got != "facade.local" {
		t.Fatalf("GetExpandedWithOptions(host) = %q", got)
	}
	if got := s.GetExpandedWithOptions("base", vconf.WithEnvLookup(lookup)); got != "http://facade.local:8080" {
		t.Fatalf("GetExpandedWithOptions(base) = %q", got)
	}
	if got := s.ExpandWithOptions(vconf.WithEnvLookup(lookup)).Get("host"); got != "facade.local" {
		t.Fatalf("ExpandWithOptions host = %q", got)
	}
}

func TestLoadProfileAndWatchFacade(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.yaml")
	if err := os.WriteFile(path, []byte("app:\n  name: base\nprofile:\n  dev:\n    app:\n      name: dev"), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := vconf.LoadProfile(path, "dev")
	if err != nil {
		t.Fatal(err)
	}
	if got := c.GetByGroup("app", "name"); got != "dev" {
		t.Fatalf("LoadProfile yaml app.name = %q", got)
	}

	watchPath := filepath.Join(dir, "watch.setting")
	if err := os.WriteFile(watchPath, []byte("name=one"), 0o644); err != nil {
		t.Fatal(err)
	}
	ticks := make(chan time.Time, 1)
	changes := make(chan string, 1)
	stop, err := vconf.WatchWithOptions(watchPath, vconf.WatchOptions{
		Interval:       time.Hour,
		CompareContent: true,
		TickerFactory: func(time.Duration) (<-chan time.Time, vconf.WatchTicker) {
			return ticks, facadeWatchTicker{}
		},
	}, func(c *vconf.Conf, err error) {
		if err != nil {
			changes <- "err"
			return
		}
		changes <- c.Get("name")
	})
	if err != nil {
		t.Fatal(err)
	}
	defer stop()
	if err := os.WriteFile(watchPath, []byte("name=two"), 0o644); err != nil {
		t.Fatal(err)
	}
	ticks <- time.Now()
	select {
	case got := <-changes:
		if got != "two" {
			t.Fatalf("watch change = %q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("watch did not report change")
	}
}

type facadeWatchTicker struct{}

func (facadeWatchTicker) Stop() {}

func TestWatchOptionsProviderTypesFacade(t *testing.T) {
	ticks := make(chan time.Time)
	var factory vconf.WatchTickerFactory = func(time.Duration) (<-chan time.Time, vconf.WatchTicker) {
		return ticks, facadeWatchTicker{}
	}
	_ = vconf.WatchOptions{TickerFactory: factory}
}

func TestAdvancedLoadAndSchemaFacade(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "base.setting")
	main := filepath.Join(dir, "main.setting")
	secret := base64.StdEncoding.EncodeToString([]byte("token"))
	if err := os.WriteFile(base, []byte("name=base\n[server]\nhost=localhost"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(main, []byte("import=base.setting\nname=main\nsecret=ENC(base64:"+secret+")\n[server]\nport=8080"), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := vconf.LoadWithOptions(main, vconf.LoadOptions{AllowInclude: true})
	if err != nil {
		t.Fatal(err)
	}
	if got := c.Get("name"); got != "main" {
		t.Fatalf("LoadWithOptions name = %q", got)
	}
	if got := c.Get("secret"); got != "token" {
		t.Fatalf("LoadWithOptions secret = %q", got)
	}
	if err := c.ValidateSchema(vconf.Schema{Fields: []vconf.FieldRule{
		{Key: "name", Required: true},
		{Group: "server", Key: "port", Required: true, Type: vconf.TypeInt},
	}}); err != nil {
		t.Fatalf("ValidateSchema() error = %v", err)
	}
	merged, err := vconf.LoadFiles(base, main)
	if err != nil {
		t.Fatal(err)
	}
	if got := merged.Get("name"); got != "main" {
		t.Fatalf("LoadFiles name = %q", got)
	}
	type cfg struct {
		Name string `conf:"name,required"`
	}
	schema, err := vconf.SchemaFromStruct(cfg{})
	if err != nil {
		t.Fatal(err)
	}
	if len(schema.Fields) != 1 {
		t.Fatalf("SchemaFromStruct fields = %d", len(schema.Fields))
	}
}

func TestLoadWithOptionsIncludeRootFacade(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "root")
	serviceDir := filepath.Join(root, "service")
	commonDir := filepath.Join(root, "common")
	if err := os.MkdirAll(serviceDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(commonDir, 0o755); err != nil {
		t.Fatal(err)
	}
	common := filepath.Join(commonDir, "base.setting")
	main := filepath.Join(serviceDir, "main.setting")
	if err := os.WriteFile(common, []byte("name=common"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(main, []byte("include=../common/base.setting\nmode=service"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := vconf.LoadWithOptions(main, vconf.LoadOptions{AllowInclude: true})
	if err == nil {
		t.Fatal("LoadWithOptions path traversal error = nil")
	}
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("errors.Is(err, ErrCodeInvalidInput) = false: %v", err)
	}

	c, err := vconf.LoadWithOptions(main, vconf.LoadOptions{AllowInclude: true, IncludeRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	if got := c.Get("name"); got != "common" {
		t.Fatalf("included name = %q", got)
	}
	if got := c.Get("mode"); got != "service" {
		t.Fatalf("main mode = %q", got)
	}
}

func TestLoadFilesAndRemoteWithOptionsFacade(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "base.setting")
	main := filepath.Join(dir, "main.setting")
	if err := os.WriteFile(base, []byte("name=base\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(main, []byte("name=main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	merged, err := vconf.LoadFilesWithOptions(vconf.LoadOptions{MaxBytes: 64}, base, main)
	if err != nil {
		t.Fatalf("LoadFilesWithOptions() error = %v", err)
	}
	if got := merged.Get("name"); got != "main" {
		t.Fatalf("LoadFilesWithOptions() name = %q, want main", got)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Config-Token"); got != "secret" {
			t.Fatalf("remote header X-Config-Token = %q, want secret", got)
		}
		_, _ = w.Write([]byte("remote: true\n"))
	}))
	defer server.Close()
	remote, err := vconf.LoadRemoteWithOptions(server.URL+"/app.yaml", vconf.LoadOptions{
		Headers:  http.Header{"X-Config-Token": []string{"secret"}},
		Timeout:  time.Second,
		MaxBytes: 64,
	})
	if err != nil {
		t.Fatalf("LoadRemoteWithOptions() error = %v", err)
	}
	if got := remote.Get("remote"); got != "true" {
		t.Fatalf("LoadRemoteWithOptions() remote = %q, want true", got)
	}
}

func TestWatchWithOptionsFacade(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.setting")
	if err := os.WriteFile(path, []byte("name=one"), 0o644); err != nil {
		t.Fatal(err)
	}
	changes := make(chan string, 1)
	stop, err := vconf.WatchWithOptions(path, vconf.WatchOptions{Interval: 10 * time.Millisecond, CompareContent: true}, func(c *vconf.Conf, err error) {
		if err != nil {
			changes <- "err"
			return
		}
		changes <- c.Get("name")
	})
	if err != nil {
		t.Fatal(err)
	}
	defer stop()
	time.Sleep(20 * time.Millisecond)
	if err := os.WriteFile(path, []byte("name=two"), 0o644); err != nil {
		t.Fatal(err)
	}
	select {
	case got := <-changes:
		if got != "two" {
			t.Fatalf("watch change = %q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("watch did not report change")
	}
}

func TestFacadeRemoteSafeAndParseWrappers(t *testing.T) {
	trustedServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("trusted=ok\n"))
	}))
	defer trustedServer.Close()
	trusted, err := vconf.LoadRemote(trustedServer.URL + "/app.setting")
	if err != nil {
		t.Fatalf("LoadRemote() error = %v", err)
	}
	if got := trusted.Get("trusted"); got != "ok" {
		t.Fatalf("LoadRemote trusted = %q", got)
	}

	client := &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if got := req.Header.Get("X-Remote-Token"); got != "token" {
			t.Fatalf("remote header = %q, want token", got)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/plain"}},
			Body:       io.NopCloser(strings.NewReader("remote=ok\n")),
			Request:    req,
		}, nil
	})}
	lookupPublic := func(context.Context, string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("8.8.8.8")}, nil
	}
	opts := vconf.LoadOptions{
		RemoteClient:       client,
		Headers:            http.Header{"X-Remote-Token": []string{"token"}},
		RemoteAllowedHosts: []string{"config.example"},
		LookupIP:           lookupPublic,
		Timeout:            time.Second,
		MaxBytes:           64,
	}

	remote, err := vconf.LoadRemoteWithOptions("http://config.example/app.setting", opts)
	if err != nil {
		t.Fatalf("LoadRemoteWithOptions() error = %v", err)
	}
	if got := remote.Get("remote"); got != "ok" {
		t.Fatalf("LoadRemoteWithOptions remote = %q", got)
	}
	safe, err := vconf.LoadRemoteSafeWithOptions("http://config.example/app.setting", opts)
	if err != nil {
		t.Fatalf("LoadRemoteSafeWithOptions() error = %v", err)
	}
	if got := safe.Get("remote"); got != "ok" {
		t.Fatalf("LoadRemoteSafeWithOptions remote = %q", got)
	}
	if _, err := vconf.LoadRemoteSafe("http://127.0.0.1/app.setting"); err == nil {
		t.Fatal("LoadRemoteSafe private host error = nil")
	}

	parsed, err := vconf.ParseByExt("app.setting", []byte("name=parse"))
	if err != nil || parsed.Get("name") != "parse" {
		t.Fatalf("ParseByExt = %#v, %v", parsed, err)
	}
	yaml, err := vconf.ParseYAMLFull("server:\n  port: 8080\n")
	if err != nil || yaml.GetByGroup("server", "port") != "8080" {
		t.Fatalf("ParseYAMLFull = %#v, %v", yaml, err)
	}
	decoded, err := vconf.Base64Decrypt(base64.StdEncoding.EncodeToString([]byte("secret")))
	if err != nil || decoded != "secret" {
		t.Fatalf("Base64Decrypt = %q, %v", decoded, err)
	}
}

func TestFacadeBindAndSchemaParserOptions(t *testing.T) {
	s, err := vconf.Parse(`
flag=yes
count=custom-int
amount=custom-uint
ratio=custom-float
items=1,2,3
schema_bool=yes
schema_float=custom-float
choice=blue
`)
	if err != nil {
		t.Fatal(err)
	}

	type bindConfig struct {
		Flag   bool    `conf:"flag"`
		Count  int     `conf:"count"`
		Amount uint    `conf:"amount"`
		Ratio  float64 `conf:"ratio"`
		Items  []int   `conf:"items"`
	}
	var cfg bindConfig
	if err := s.BindWithOptions(&cfg,
		vconf.WithBindBoolParser(func(value string) (bool, error) {
			return value == "yes", nil
		}),
		vconf.WithBindIntParser(func(value string, base int, bitSize int) (int64, error) {
			if value == "custom-int" {
				return 42, nil
			}
			return 7, nil
		}),
		vconf.WithBindUintParser(func(value string, base int, bitSize int) (uint64, error) {
			if value == "custom-uint" {
				return 9, nil
			}
			return 3, nil
		}),
		vconf.WithBindFloatParser(func(value string, bitSize int) (float64, error) {
			if value == "custom-float" {
				return 1.5, nil
			}
			return 0, errors.New("unexpected float")
		}),
	); err != nil {
		t.Fatalf("BindWithOptions() error = %v", err)
	}
	if !cfg.Flag || cfg.Count != 42 || cfg.Amount != 9 || cfg.Ratio != 1.5 || !reflect.DeepEqual(cfg.Items, []int{7, 7, 7}) {
		t.Fatalf("BindWithOptions cfg = %#v", cfg)
	}

	err = s.ValidateSchemaWithOptions(vconf.Schema{Fields: []vconf.FieldRule{
		{Key: "schema_bool", Required: true, Type: vconf.TypeBool},
		{Key: "schema_float", Required: true, Type: vconf.TypeFloat},
		{Key: "choice", Required: true, Choices: []string{"red", "blue"}},
	}},
		vconf.WithSchemaBoolParser(func(value string) (bool, error) {
			return value == "yes", nil
		}),
		vconf.WithSchemaFloatParser(func(value string, bitSize int) (float64, error) {
			if value == "custom-float" {
				return 2.5, nil
			}
			return 0, errors.New("unexpected schema float")
		}),
	)
	if err != nil {
		t.Fatalf("ValidateSchemaWithOptions() error = %v", err)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }
