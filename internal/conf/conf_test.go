package conf

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	knifer "github.com/imajinyun/go-knifer"
)

func TestParseSetting(t *testing.T) {
	s, err := Parse("name = gokit\n[server]\nport=8080\ndebug=true")
	if err != nil {
		t.Fatal(err)
	}
	if got := s.Get("name"); got != "gokit" {
		t.Fatalf("Get(name) = %q", got)
	}
	if got := s.GetByGroup("server", "port"); got != "8080" {
		t.Fatalf("GetByGroup(server, port) = %q", got)
	}
	if got := s.GetOrDefault("missing", "def"); got != "def" {
		t.Fatalf("GetOrDefault() = %q", got)
	}
}

func TestParseYAML(t *testing.T) {
	s, err := ParseYAML("app: gokit\nserver:\n  port: 8080")
	if err != nil {
		t.Fatal(err)
	}
	if got := s.Get("app"); got != "gokit" {
		t.Fatalf("Get(app) = %q", got)
	}
	if got := s.GetByGroup("server", "port"); got != "8080" {
		t.Fatalf("GetByGroup(server, port) = %q", got)
	}
}

func TestNilConfReadMethodsAreEmptyAndSafe(t *testing.T) {
	var s *Conf

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

func TestGroupsKeysAndToMapKeepStableSemantics(t *testing.T) {
	s := New()
	s.Set("root", "value")
	s.SetByGroup("server", "port", "8080")
	s.SetByGroup("server", "host", "localhost")
	s.SetByGroup("app", "name", "gokit")

	if got, want := s.Groups(), []string{"", "app", "server"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Groups() = %v, want %v", got, want)
	}
	if got, want := s.Keys("server"), []string{"host", "port"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Keys(server) = %v, want %v", got, want)
	}

	m := s.ToMap()
	m["server"]["port"] = "9090"
	if got := s.GetByGroup("server", "port"); got != "8080" {
		t.Fatalf("ToMap() returned shallow copy, source port = %q", got)
	}
}

func TestConfErrorContract(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "missing.setting"))
	assertConfCode(t, err, knifer.ErrCodeNotFound)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Load missing file should preserve os.ErrNotExist: %v", err)
	}

	_, err = Parse("invalid-line")
	assertConfCode(t, err, knifer.ErrCodeInvalidInput)

	_, err = Parse("=empty")
	assertConfCode(t, err, knifer.ErrCodeInvalidInput)

	_, err = ParseYAML("invalid-yaml-line")
	assertConfCode(t, err, knifer.ErrCodeInvalidInput)
}

func TestExpandVariables(t *testing.T) {
	t.Setenv("CONF_ENV_HOST", "envhost")
	s, err := Parse(`
host=localhost
base=http://${host}:8080
env=${ENV:CONF_ENV_HOST}
missing=${missing:fallback}
[db]
host=db.local
url=postgres://${db.host}/${name:app}
`)
	if err != nil {
		t.Fatal(err)
	}

	if got := s.GetExpanded("base"); got != "http://localhost:8080" {
		t.Fatalf("GetExpanded(base) = %q", got)
	}
	if got := s.GetExpanded("env"); got != "envhost" {
		t.Fatalf("GetExpanded(env) = %q", got)
	}
	if got := s.GetExpanded("missing"); got != "fallback" {
		t.Fatalf("GetExpanded(missing) = %q", got)
	}
	if got := s.GetByGroupExpanded("db", "url"); got != "postgres://db.local/app" {
		t.Fatalf("GetByGroupExpanded(db,url) = %q", got)
	}
	if got := s.Expand().Get("base"); got != "http://localhost:8080" {
		t.Fatalf("Expand().Get(base) = %q", got)
	}
}

func TestParseYAMLFullAndBind(t *testing.T) {
	s, err := ParseYAMLFull(`
app: demo
server:
  host: 127.0.0.1
  port: 8080
  debug: true
  tags: [api, admin]
`)
	if err != nil {
		t.Fatal(err)
	}
	if got := s.Get("app"); got != "demo" {
		t.Fatalf("Get(app) = %q", got)
	}
	if got := s.GetByGroup("server", "host"); got != "127.0.0.1" {
		t.Fatalf("GetByGroup(server,host) = %q", got)
	}

	type serverConf struct {
		Host  string   `conf:"host"`
		Port  int      `conf:"port"`
		Debug bool     `conf:"debug"`
		Tags  []string `conf:"tags"`
	}
	var cfg serverConf
	if err := s.BindGroup("server", &cfg); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cfg, serverConf{Host: "127.0.0.1", Port: 8080, Debug: true, Tags: []string{"api", "admin"}}) {
		t.Fatalf("BindGroup() = %#v", cfg)
	}
}

func TestParseTOMLAndProfile(t *testing.T) {
	s, err := ParseTOML(`
name = "demo"
tags = ["a", "b"]
[server]
port = 8080
[profile.dev]
name = "dev-demo"
[profile.dev.server]
port = 9090
`)
	if err != nil {
		t.Fatal(err)
	}
	if got := s.Get("tags"); got != "a,b" {
		t.Fatalf("Get(tags) = %q", got)
	}
	dev := s.ApplyProfile("dev")
	if got := dev.Get("name"); got != "dev-demo" {
		t.Fatalf("ApplyProfile(dev).Get(name) = %q", got)
	}
	if got := dev.GetByGroup("server", "port"); got != "9090" {
		t.Fatalf("ApplyProfile(dev).server.port = %q", got)
	}
}

func TestLoadProfileAndParseByExt(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.toml")
	if err := os.WriteFile(path, []byte("name='base'\n[profile.test]\nname='test'"), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := LoadProfile(path, "test")
	if err != nil {
		t.Fatal(err)
	}
	if got := c.Get("name"); got != "test" {
		t.Fatalf("LoadProfile name = %q", got)
	}

	yamlConf, err := ParseByExt("app.yaml", []byte("app:\n  name: demo"))
	if err != nil {
		t.Fatal(err)
	}
	if got := yamlConf.GetByGroup("app", "name"); got != "demo" {
		t.Fatalf("ParseByExt yaml app.name = %q", got)
	}
}

func TestWatchReloadsOnChange(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.setting")
	if err := os.WriteFile(path, []byte("name=one"), 0o644); err != nil {
		t.Fatal(err)
	}
	changes := make(chan string, 1)
	stop, err := Watch(path, 10*time.Millisecond, func(c *Conf, err error) {
		if err != nil {
			changes <- "err:" + err.Error()
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

func TestLoadWithOptionsIncludesMergeDecryptAndSchema(t *testing.T) {
	dir := t.TempDir()
	common := filepath.Join(dir, "common.setting")
	main := filepath.Join(dir, "main.setting")
	secret := base64.StdEncoding.EncodeToString([]byte("s3cr3t"))
	if err := os.WriteFile(common, []byte("name=common\n[server]\nhost=127.0.0.1"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(main, []byte("include=common.setting\nname=main\nsecret=ENC(base64:"+secret+")\n[server]\nport=8080"), 0o644); err != nil {
		t.Fatal(err)
	}

	c, err := LoadWithOptions(main, LoadOptions{AllowInclude: true})
	if err != nil {
		t.Fatal(err)
	}
	if got := c.Get("name"); got != "main" {
		t.Fatalf("merged name = %q", got)
	}
	if got := c.Get("secret"); got != "s3cr3t" {
		t.Fatalf("decrypted secret = %q", got)
	}
	if got := c.GetByGroup("server", "host"); got != "127.0.0.1" {
		t.Fatalf("included server.host = %q", got)
	}
	if _, ok := c.Lookup("", "include"); ok {
		t.Fatal("include key should be removed after loading")
	}
	if err := c.ValidateSchema(Schema{Fields: []FieldRule{
		{Key: "name", Required: true, Type: TypeString},
		{Group: "server", Key: "port", Required: true, Type: TypeInt},
		{Group: "server", Key: "host", Required: true},
	}}); err != nil {
		t.Fatalf("ValidateSchema() error = %v", err)
	}
	if err := c.ValidateSchema(Schema{Fields: []FieldRule{{Group: "server", Key: "debug", Required: true}}}); err == nil {
		t.Fatal("ValidateSchema() missing required error = nil")
	}
}

func TestLoadFilesAndApplyDefaults(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "base.setting")
	override := filepath.Join(dir, "override.toml")
	if err := os.WriteFile(base, []byte("name=base\nmode=dev"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(override, []byte("name='override'\n[server]\nport=9090"), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := LoadFiles(base, override)
	if err != nil {
		t.Fatal(err)
	}
	if got := c.Get("name"); got != "override" {
		t.Fatalf("LoadFiles merged name = %q", got)
	}
	withDefaults := c.ApplyDefaults(Schema{Fields: []FieldRule{{Key: "region", Default: "cn"}}})
	if got := withDefaults.Get("region"); got != "cn" {
		t.Fatalf("ApplyDefaults region = %q", got)
	}
}

func TestParseTOMLNestedDottedKeys(t *testing.T) {
	c, err := ParseTOML(`
title = "demo"
[database]
ports = [8000, 8001, 8002]
enabled = true
connection.max = 5000
[servers.alpha]
ip = "10.0.0.1"
`)
	if err != nil {
		t.Fatal(err)
	}
	if got := c.GetByGroup("database", "ports"); got != "8000,8001,8002" {
		t.Fatalf("database.ports = %q", got)
	}
	if got := c.GetByGroup("database.connection", "max"); got != "5000" {
		t.Fatalf("database.connection.max = %q", got)
	}
	if got := c.GetByGroup("servers.alpha", "ip"); got != "10.0.0.1" {
		t.Fatalf("servers.alpha.ip = %q", got)
	}
}

func TestLoadRemoteWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Config-Token") != "secret" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte("app:\n  name: remote"))
	}))
	defer server.Close()
	calledFactory := false
	c, err := LoadRemoteWithOptions(server.URL+"/app.yaml", LoadOptions{
		Timeout: time.Second,
		Headers: http.Header{"X-Config-Token": []string{"secret"}},
		RequestFactory: func(ctx context.Context, rawURL string) (*http.Request, error) {
			calledFactory = true
			return http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !calledFactory {
		t.Fatal("request factory was not called")
	}
	if got := c.GetByGroup("app", "name"); got != "remote" {
		t.Fatalf("remote app.name = %q", got)
	}
	if _, err := LoadRemoteWithOptions(server.URL+"/app.yaml", LoadOptions{MaxBytes: 3, Headers: http.Header{"X-Config-Token": []string{"secret"}}}); err == nil {
		t.Fatal("LoadRemoteWithOptions max bytes error = nil")
	}
}

func TestSchemaFromStructAndValidateStruct(t *testing.T) {
	type appConfig struct {
		Name string `conf:"name,required"`
		Port int    `conf:"port,required,int"`
		Mode string `conf:"mode,default=dev,choices=dev|prod"`
	}
	c := New()
	c.Set("name", "demo")
	c.Set("port", "8080")
	c.Set("mode", "dev")
	if err := c.ValidateStruct(appConfig{}); err != nil {
		t.Fatalf("ValidateStruct() error = %v", err)
	}
	schema, err := SchemaFromStruct(appConfig{})
	if err != nil {
		t.Fatal(err)
	}
	if len(schema.Fields) != 3 {
		t.Fatalf("SchemaFromStruct fields = %d", len(schema.Fields))
	}
}

func TestWatchWithOptionsCompareContentAndEvent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.setting")
	if err := os.WriteFile(path, []byte("name=one"), 0o644); err != nil {
		t.Fatal(err)
	}
	changes := make(chan string, 1)
	events := make(chan WatchEvent, 1)
	stop, err := WatchWithOptions(path, WatchOptions{
		Interval:       10 * time.Millisecond,
		Debounce:       5 * time.Millisecond,
		CompareContent: true,
		OnEvent: func(event WatchEvent) {
			events <- event
		},
	}, func(c *Conf, err error) {
		if err != nil {
			changes <- "err:" + err.Error()
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
	select {
	case event := <-events:
		if event.Path != path || event.Size == 0 {
			t.Fatalf("watch event = %#v", event)
		}
	case <-time.After(time.Second):
		t.Fatal("watch did not report event")
	}
}

func assertConfCode(t *testing.T, err error, code knifer.ErrCode) {
	t.Helper()
	if err == nil {
		t.Fatalf("err = nil, want %s", code)
	}
	if !errors.Is(err, code) {
		t.Fatalf("errors.Is(%v, %s) = false", err, code)
	}
	got, ok := knifer.CodeOf(err)
	if !ok || got != code {
		t.Fatalf("CodeOf(%v) = %q, %v; want %q, true", err, got, ok, code)
	}
}
