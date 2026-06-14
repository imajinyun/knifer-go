package conf

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	knifer "github.com/imajinyun/go-knifer"
)

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
