package vconf_test

import (
	"encoding/base64"
	"testing"

	"github.com/imajinyun/knifer-go/vconf"
)

func TestFacadeParseWrappers(t *testing.T) {
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
