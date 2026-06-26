package vconf_test

import (
	"errors"
	"testing"

	"github.com/imajinyun/knifer-go/vconf"
)

func TestFacadeSchemaDefaultsAndValidation(t *testing.T) {
	s, err := vconf.Parse(`
custom_int=custom
[server]
port=8080
`)
	if err != nil {
		t.Fatal(err)
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
