package vconf_test

import (
	"testing"

	"github.com/imajinyun/knifer-go/vconf"
)

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
