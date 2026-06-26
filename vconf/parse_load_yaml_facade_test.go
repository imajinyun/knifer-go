package vconf_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/imajinyun/knifer-go/vconf"
)

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
