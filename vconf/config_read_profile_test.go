package vconf_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/imajinyun/knifer-go/vconf"
)

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

func TestLoadProfileFacade(t *testing.T) {
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
}
