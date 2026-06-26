package vconf_test

import (
	"reflect"
	"testing"

	"github.com/imajinyun/knifer-go/vconf"
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
