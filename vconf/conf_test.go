package vconf_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

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
