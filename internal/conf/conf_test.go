package conf

import (
	"context"
	"encoding/base64"
	"errors"
	"io/fs"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
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

func TestCloneReturnsIndependentCopy(t *testing.T) {
	s := New()
	s.Set("root", "value")
	s.SetByGroup("server", "port", "8080")

	clone := s.Clone()
	clone.Set("root", "changed")
	clone.SetByGroup("server", "port", "9090")
	clone.SetByGroup("server", "host", "localhost")

	if got := s.Get("root"); got != "value" {
		t.Fatalf("source root changed to %q", got)
	}
	if got := s.GetByGroup("server", "port"); got != "8080" {
		t.Fatalf("source server.port changed to %q", got)
	}
	if got := s.GetByGroup("server", "host"); got != "" {
		t.Fatalf("source server.host changed to %q", got)
	}
	if got := clone.GetByGroup("server", "host"); got != "localhost" {
		t.Fatalf("clone server.host = %q", got)
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

func TestExpandWithOptionsUsesCustomEnvLookup(t *testing.T) {
	s, err := Parse(`
host=${ENV:CONF_HOST}
base=http://${host}:${port:8080}
[db]
url=postgres://${ENV:CONF_DB_HOST}/${name:app}
`)
	if err != nil {
		t.Fatal(err)
	}
	lookup := func(key string) string {
		switch key {
		case "CONF_HOST":
			return "option.local"
		case "CONF_DB_HOST":
			return "db.option.local"
		default:
			return ""
		}
	}
	if got := s.GetExpandedWithOptions("host", WithEnvLookup(lookup)); got != "option.local" {
		t.Fatalf("GetExpandedWithOptions(host) = %q", got)
	}
	if got := s.GetExpandedWithOptions("base", WithEnvLookup(lookup)); got != "http://option.local:8080" {
		t.Fatalf("GetExpandedWithOptions(base) = %q", got)
	}
	if got := s.GetByGroupExpandedWithOptions("db", "url", WithEnvLookup(lookup)); got != "postgres://db.option.local/app" {
		t.Fatalf("GetByGroupExpandedWithOptions(db.url) = %q", got)
	}
	expanded := s.ExpandWithOptions(WithEnvLookup(lookup))
	if got := expanded.Get("host"); got != "option.local" {
		t.Fatalf("ExpandWithOptions host = %q", got)
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

func TestTypedGettersWithOptionsUseParsers(t *testing.T) {
	s := New()
	s.Set("port", "custom-int")
	s.Set("debug", "custom-bool")
	s.SetByGroup("server", "port", "9090")
	s.SetByGroup("server", "debug", "true")

	intCalled := false
	if got := s.GetIntWithOptions("port", 10, WithIntParser(func(text string) (int, error) {
		intCalled = true
		if text != "custom-int" {
			t.Fatalf("int parser text = %q", text)
		}
		return 8080, nil
	})); got != 8080 || !intCalled {
		t.Fatalf("GetIntWithOptions = %d, called=%v", got, intCalled)
	}

	boolCalled := false
	if got := s.GetBoolWithOptions("debug", false, WithBoolParser(func(text string) (bool, error) {
		boolCalled = true
		if text != "custom-bool" {
			t.Fatalf("bool parser text = %q", text)
		}
		return true, nil
	})); !got || !boolCalled {
		t.Fatalf("GetBoolWithOptions = %v, called=%v", got, boolCalled)
	}
	if got := s.GetIntWithOptions("port", 10, WithIntParser(func(string) (int, error) {
		return 0, errors.New("invalid")
	})); got != 10 {
		t.Fatalf("GetIntWithOptions fallback = %d", got)
	}
	if got, err := s.GetIntEWithOptions("port", WithIntParser(func(string) (int, error) { return 7000, nil })); err != nil || got != 7000 {
		t.Fatalf("GetIntEWithOptions = %d, err=%v", got, err)
	}
	if got, err := s.GetBoolEWithOptions("debug", WithBoolParser(func(string) (bool, error) { return true, nil })); err != nil || !got {
		t.Fatalf("GetBoolEWithOptions = %v, err=%v", got, err)
	}
	if got, err := s.GetIntByGroupE("server", "port"); err != nil || got != 9090 {
		t.Fatalf("GetIntByGroupE = %d, err=%v", got, err)
	}
	if got, err := s.GetBoolByGroupE("server", "debug"); err != nil || !got {
		t.Fatalf("GetBoolByGroupE = %v, err=%v", got, err)
	}

	s.Set("bad-int", "abc")
	if _, err := s.GetIntE("missing"); !errors.Is(err, knifer.ErrCodeNotFound) {
		t.Fatalf("GetIntE missing err = %v, want not found", err)
	}
	if _, err := s.GetIntE("bad-int"); !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("GetIntE invalid err = %v, want invalid input", err)
	}
}

func TestBindWithOptionsUsesParsers(t *testing.T) {
	s := New()
	s.SetByGroup("server", "port", "custom-int")
	s.SetByGroup("server", "debug", "custom-bool")
	s.SetByGroup("server", "ratio", "custom-float")
	s.SetByGroup("server", "ids", "a,b")

	type serverConf struct {
		Port  int     `conf:"port"`
		Debug bool    `conf:"debug"`
		Ratio float64 `conf:"ratio"`
		IDs   []uint  `conf:"ids"`
	}
	var cfg serverConf
	var intCalled, boolCalled, floatCalled, uintCalled int
	err := s.BindGroupWithOptions("server", &cfg,
		WithBindIntParser(func(text string, base, bitSize int) (int64, error) {
			intCalled++
			if text == "custom-int" {
				return 8080, nil
			}
			return strconv.ParseInt(text, base, bitSize)
		}),
		WithBindBoolParser(func(text string) (bool, error) {
			boolCalled++
			return text == "custom-bool", nil
		}),
		WithBindFloatParser(func(text string, bitSize int) (float64, error) {
			floatCalled++
			if text == "custom-float" {
				return 0.75, nil
			}
			return strconv.ParseFloat(text, bitSize)
		}),
		WithBindUintParser(func(text string, base, bitSize int) (uint64, error) {
			uintCalled++
			switch text {
			case "a":
				return 1, nil
			case "b":
				return 2, nil
			default:
				return strconv.ParseUint(text, base, bitSize)
			}
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cfg, serverConf{Port: 8080, Debug: true, Ratio: 0.75, IDs: []uint{1, 2}}) {
		t.Fatalf("BindGroupWithOptions = %#v", cfg)
	}
	if intCalled != 1 || boolCalled != 1 || floatCalled != 1 || uintCalled != 2 {
		t.Fatalf("parser calls int=%d bool=%d float=%d uint=%d", intCalled, boolCalled, floatCalled, uintCalled)
	}
}

func TestParseYAMLFullWithOptionsUsesProvider(t *testing.T) {
	called := false
	s, err := ParseYAMLFullWithOptions("ignored", WithYAMLUnmarshalFunc(func(data []byte, out any) error {
		called = true
		root, ok := out.(*any)
		if !ok {
			t.Fatalf("unmarshal output = %T, want *any", out)
		}
		*root = map[string]any{"app": map[string]any{"name": "provider"}}
		return nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("custom YAML unmarshal provider was not called")
	}
	if got := s.GetByGroup("app", "name"); got != "provider" {
		t.Fatalf("provider app.name = %q", got)
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

	profileYAML, err := ParseYAMLFull(`
app:
  name: base
server:
  port: 8080
profile:
  dev:
    app:
      name: dev
    server:
      port: 9090
`)
	if err != nil {
		t.Fatal(err)
	}
	dev := profileYAML.ApplyProfile("dev")
	if got := dev.GetByGroup("app", "name"); got != "dev" {
		t.Fatalf("YAML profile app.name = %q", got)
	}
	if got := dev.GetByGroup("server", "port"); got != "9090" {
		t.Fatalf("YAML profile server.port = %q", got)
	}

	custom, err := ParseByExtWithOptions("app.custom", []byte("ignored"), WithParserForExt("custom", func([]byte) (*Conf, error) {
		c := New()
		c.Set("name", "custom-parser")
		return c, nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	if got := custom.Get("name"); got != "custom-parser" {
		t.Fatalf("custom parser name = %q", got)
	}
}

func TestLoadWithOptionsPassesParseOptions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.yaml")
	if err := os.WriteFile(path, []byte("ignored"), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := LoadWithOptions(path, LoadOptions{ParseOptions: []ParseOption{WithYAMLUnmarshalFunc(func([]byte, any) error {
		return errors.New("custom yaml error")
	})}})
	if err == nil {
		t.Fatalf("LoadWithOptions = %#v, nil error", c)
	}
	assertConfCode(t, err, knifer.ErrCodeInvalidInput)
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

func TestWatchWithOptionsUsesRunner(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.setting")
	if err := os.WriteFile(path, []byte("name=one"), 0o644); err != nil {
		t.Fatal(err)
	}
	ticks := make(chan time.Time)
	ticker := &watchTestTicker{stopped: make(chan struct{})}
	runnerCalled := make(chan struct{}, 1)
	stop, err := WatchWithOptions(path, WatchOptions{
		Interval: 10 * time.Second,
		TickerFactory: func(delay time.Duration) (<-chan time.Time, WatchTicker) {
			if delay != 10*time.Second {
				t.Fatalf("ticker delay = %s, want 10s", delay)
			}
			return ticks, ticker
		},
		Runner: func(fn func()) {
			runnerCalled <- struct{}{}
			go fn()
		},
	}, func(*Conf, error) {})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-runnerCalled:
	case <-time.After(time.Second):
		t.Fatal("watch runner was not used")
	}
	stop()
	select {
	case <-ticker.stopped:
	case <-time.After(time.Second):
		t.Fatal("watch ticker was not stopped")
	}
	stop()
}

func TestWatchWithOptionsRejectsNilCallback(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.setting")
	if err := os.WriteFile(path, []byte("name=one"), 0o644); err != nil {
		t.Fatal(err)
	}
	if stop, err := WatchWithOptions(path, WatchOptions{}, nil); err == nil || stop != nil {
		t.Fatalf("WatchWithOptions nil callback stop nil=%v err=%v, want error", stop == nil, err)
	}
}

func TestWatchCallbacksPanicsAreIsolated(t *testing.T) {
	tick := make(chan time.Time)
	ticker := &watchTestTicker{stopped: make(chan struct{})}
	info := fakeFileInfo{name: "app.setting", size: int64(len("name=one")), modTime: time.Unix(1, 0)}
	content := []byte("name=one")
	started := make(chan struct{})

	stop, err := WatchWithOptions("app.setting", WatchOptions{
		Interval:       time.Hour,
		CompareContent: true,
		TickerFactory:  func(time.Duration) (<-chan time.Time, WatchTicker) { return tick, ticker },
		Runner: func(fn func()) {
			close(started)
			go fn()
		},
		Stat: func(string) (os.FileInfo, error) { return info, nil },
		ReadFile: func(string, int64) ([]byte, error) {
			return content, nil
		},
		OnEvent: func(WatchEvent) { panic("event") },
	}, func(*Conf, error) { panic("change") })
	if err != nil {
		t.Fatal(err)
	}
	<-started
	info.size = int64(len("name=two"))
	info.modTime = time.Unix(2, 0)
	content = []byte("name=two")
	tick <- time.Unix(2, 0)

	done := make(chan struct{})
	go func() {
		stop()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("stop blocked after callback panic")
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

func TestParseTOMLWithOptionsUsesProvider(t *testing.T) {
	called := false
	c, err := ParseTOMLWithOptions("ignored", WithTOMLUnmarshalFunc(func(data []byte, out any) error {
		called = true
		root, ok := out.(*map[string]any)
		if !ok {
			t.Fatalf("toml unmarshal output = %T, want *map[string]any", out)
		}
		*root = map[string]any{"app": map[string]any{"name": "provider"}}
		return nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("custom TOML unmarshal provider was not called")
	}
	if got := c.GetByGroup("app", "name"); got != "provider" {
		t.Fatalf("provider app.name = %q", got)
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

func TestLoadRemoteSafeRejectsPrivateHostsAndUnsafeRedirects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("app:\n  name: remote"))
	}))
	defer server.Close()

	if _, err := LoadRemoteSafe(server.URL + "/app.yaml"); err == nil {
		t.Fatal("LoadRemoteSafe should reject private hosts by default")
	}
	if _, err := LoadRemoteSafe("http://224.0.0.1/app.yaml"); err == nil {
		t.Fatal("LoadRemoteSafe should reject multicast hosts by default")
	}
	if _, err := LoadRemoteSafe("http://0.0.0.0/app.yaml"); err == nil {
		t.Fatal("LoadRemoteSafe should reject unspecified hosts by default")
	}
	remoteURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	c, err := LoadRemoteSafeWithOptions(server.URL+"/app.yaml", LoadOptions{RemoteAllowedHosts: []string{remoteURL.Hostname()}})
	if err != nil {
		t.Fatalf("LoadRemoteSafeWithOptions allowed host: %v", err)
	}
	if got := c.GetByGroup("app", "name"); got != "remote" {
		t.Fatalf("remote app.name = %q", got)
	}

	redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://127.0.0.1/private.yaml", http.StatusFound)
	}))
	defer redirect.Close()
	redirectURL, err := url.Parse(redirect.URL)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := LoadRemoteSafeWithOptions(redirect.URL+"/app.yaml", LoadOptions{RemoteAllowedHosts: []string{redirectURL.Hostname()}}); err == nil {
		t.Fatal("LoadRemoteSafeWithOptions should reject unsafe redirect target")
	}
}

func TestLoadRemoteSafeRevalidatesHostAtRoundTrip(t *testing.T) {
	lookups := [][]net.IP{{net.ParseIP("93.184.216.34")}, {net.ParseIP("127.0.0.1")}}
	lookupCount := 0
	client := &http.Client{Transport: confRoundTripperFunc(func(*http.Request) (*http.Response, error) {
		t.Fatal("unsafe request reached base transport")
		return nil, nil
	})}
	_, err := LoadRemoteSafeWithOptions("http://example.com/app.yaml", LoadOptions{
		RemoteClient: client,
		LookupIP: func(context.Context, string) ([]net.IP, error) {
			if lookupCount >= len(lookups) {
				return lookups[len(lookups)-1], nil
			}
			ips := lookups[lookupCount]
			lookupCount++
			return ips, nil
		},
	})
	if err == nil {
		t.Fatal("LoadRemoteSafeWithOptions should reject a host that resolves private during RoundTrip")
	}
	if lookupCount != 2 {
		t.Fatalf("lookup count = %d, want 2", lookupCount)
	}
}

type confRoundTripperFunc func(*http.Request) (*http.Response, error)

func (f confRoundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

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

func TestValidateSchemaWithOptionsUsesParsers(t *testing.T) {
	c := New()
	c.Set("debug", "custom-bool")
	c.Set("port", "custom-int")
	c.Set("ratio", "custom-float")

	var boolCalled, intCalled, floatCalled int
	err := c.ValidateSchemaWithOptions(Schema{Fields: []FieldRule{
		{Key: "debug", Required: true, Type: TypeBool},
		{Key: "port", Required: true, Type: TypeInt},
		{Key: "ratio", Required: true, Type: TypeFloat},
	}},
		WithSchemaBoolParser(func(text string) (bool, error) {
			boolCalled++
			if text == "custom-bool" {
				return true, nil
			}
			return strconv.ParseBool(text)
		}),
		WithSchemaIntParser(func(text string, base, bitSize int) (int64, error) {
			intCalled++
			if text == "custom-int" {
				return 8080, nil
			}
			return strconv.ParseInt(text, base, bitSize)
		}),
		WithSchemaFloatParser(func(text string, bitSize int) (float64, error) {
			floatCalled++
			if text == "custom-float" {
				return 0.75, nil
			}
			return strconv.ParseFloat(text, bitSize)
		}),
	)
	if err != nil {
		t.Fatalf("ValidateSchemaWithOptions() error = %v", err)
	}
	if boolCalled != 1 || intCalled != 1 || floatCalled != 1 {
		t.Fatalf("schema parser calls bool=%d int=%d float=%d", boolCalled, intCalled, floatCalled)
	}
}

func TestValidateStructWithOptionsUsesParsers(t *testing.T) {
	type appConfig struct {
		Port int `conf:"port,required,int"`
	}
	c := New()
	c.Set("port", "custom-int")

	called := false
	if err := c.ValidateStructWithOptions(appConfig{}, WithSchemaIntParser(func(text string, base, bitSize int) (int64, error) {
		called = true
		if text == "custom-int" {
			return 8080, nil
		}
		return strconv.ParseInt(text, base, bitSize)
	})); err != nil {
		t.Fatalf("ValidateStructWithOptions() error = %v", err)
	}
	if !called {
		t.Fatal("schema int parser was not called")
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

type watchTestTicker struct {
	stopped chan struct{}
}

func (t *watchTestTicker) Stop() { close(t.stopped) }

func TestWatchWithOptionsProviders(t *testing.T) {
	tick := make(chan time.Time)
	ticker := &watchTestTicker{stopped: make(chan struct{})}
	reads := 0
	changes := make(chan string, 1)
	events := make(chan WatchEvent, 1)
	info := fakeFileInfo{name: "app.setting", size: int64(len("name=one")), modTime: time.Unix(1, 0)}
	content := []byte("name=one")

	stop, err := WatchWithOptions("app.setting", WatchOptions{
		Interval:       time.Hour,
		Debounce:       time.Nanosecond,
		CompareContent: true,
		TickerFactory: func(delay time.Duration) (<-chan time.Time, WatchTicker) {
			if delay != time.Hour {
				t.Fatalf("ticker delay = %s, want %s", delay, time.Hour)
			}
			return tick, ticker
		},
		After: func(delay time.Duration) <-chan time.Time {
			if delay != time.Nanosecond {
				t.Fatalf("debounce delay = %s, want %s", delay, time.Nanosecond)
			}
			ch := make(chan time.Time, 1)
			ch <- time.Unix(3, 0)
			return ch
		},
		Stat: func(path string) (os.FileInfo, error) {
			if path != "app.setting" {
				t.Fatalf("stat path = %q", path)
			}
			return info, nil
		},
		ReadFile: func(path string, maxBytes int64) ([]byte, error) {
			reads++
			if path != "app.setting" || maxBytes != 64 {
				t.Fatalf("read path=%q maxBytes=%d", path, maxBytes)
			}
			return content, nil
		},
		LoadOptions: LoadOptions{MaxBytes: 64},
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

	info.size = int64(len("name=two"))
	info.modTime = time.Unix(2, 0)
	content = []byte("name=two")
	tick <- time.Unix(2, 0)
	select {
	case got := <-changes:
		if got != "two" {
			t.Fatalf("watch change = %q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("watch did not report provider-driven change")
	}
	select {
	case event := <-events:
		if event.Path != "app.setting" || event.Size != int64(len("name=two")) {
			t.Fatalf("watch event = %#v", event)
		}
	case <-time.After(time.Second):
		t.Fatal("watch did not report provider-driven event")
	}
	stop()
	select {
	case <-ticker.stopped:
	case <-time.After(time.Second):
		t.Fatal("watch ticker was not stopped")
	}
	if reads < 3 {
		t.Fatalf("read count = %d, want at least 3", reads)
	}
}

func TestLoadWithOptionsReadFileProvider(t *testing.T) {
	c, err := LoadWithOptions("virtual.setting", LoadOptions{
		MaxBytes: 16,
		ReadFile: func(path string, maxBytes int64) ([]byte, error) {
			if path != "virtual.setting" || maxBytes != 16 {
				t.Fatalf("read path=%q maxBytes=%d", path, maxBytes)
			}
			return []byte("name=fake"), nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := c.Get("name"); got != "fake" {
		t.Fatalf("loaded name = %q", got)
	}
}

func TestLoadWithOptionsReadFileProviderUsesDefaultMaxBytes(t *testing.T) {
	_, err := LoadWithOptions("virtual.setting", LoadOptions{
		ReadFile: func(path string, maxBytes int64) ([]byte, error) {
			if maxBytes != DefaultMaxBytes {
				t.Fatalf("default maxBytes=%d, want %d", maxBytes, DefaultMaxBytes)
			}
			return []byte("name=fake"), nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoadWithOptionsAllowsExplicitUnlimitedMaxBytes(t *testing.T) {
	_, err := LoadWithOptions("virtual.setting", LoadOptions{
		MaxBytes: -1,
		ReadFile: func(path string, maxBytes int64) ([]byte, error) {
			if maxBytes != -1 {
				t.Fatalf("maxBytes=%d, want -1", maxBytes)
			}
			return []byte("name=fake"), nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoadWithOptionsReadFileProviderEnforcesMaxBytes(t *testing.T) {
	_, err := LoadWithOptions("virtual.setting", LoadOptions{
		MaxBytes: 4,
		ReadFile: func(path string, maxBytes int64) ([]byte, error) {
			return []byte("name=fake"), nil
		},
	})
	assertConfCode(t, err, knifer.ErrCodeInvalidInput)
}

type fakeFileInfo struct {
	name    string
	size    int64
	modTime time.Time
}

func (f fakeFileInfo) Name() string       { return f.name }
func (f fakeFileInfo) Size() int64        { return f.size }
func (f fakeFileInfo) Mode() fs.FileMode  { return 0o644 }
func (f fakeFileInfo) ModTime() time.Time { return f.modTime }
func (f fakeFileInfo) IsDir() bool        { return false }
func (f fakeFileInfo) Sys() any           { return nil }

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
