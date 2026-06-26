package vconf_test

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"github.com/imajinyun/knifer-go/vconf"
)

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
