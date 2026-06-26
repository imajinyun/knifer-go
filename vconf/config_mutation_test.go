package vconf_test

import (
	"testing"

	"github.com/imajinyun/knifer-go/vconf"
)

func TestFacadeConfigMutationMethods(t *testing.T) {
	s, err := vconf.Parse(`
name=demo
[server]
enabled=true
`)
	if err != nil {
		t.Fatal(err)
	}

	clone := s.Clone()
	clone.Set("name", "clone")
	if s.Get("name") != "demo" || clone.Get("name") != "clone" {
		t.Fatalf("Clone should not alias source: source=%q clone=%q", s.Get("name"), clone.Get("name"))
	}
	clone.Delete("name")
	if _, ok := clone.Lookup("", "name"); ok {
		t.Fatal("Delete did not remove default key")
	}
	clone.DeleteByGroup("server", "enabled")
	if _, ok := clone.Lookup("server", "enabled"); ok {
		t.Fatal("DeleteByGroup did not remove grouped key")
	}
	merged := clone.Merge(vconf.Merge(s))
	if got := merged.Get("name"); got != "demo" {
		t.Fatalf("Merge method name = %q", got)
	}
}
