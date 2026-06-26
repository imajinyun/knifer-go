package vconf_test

import (
	"errors"
	"testing"

	"github.com/imajinyun/knifer-go/vconf"
)

func TestFacadeTypedGetters(t *testing.T) {
	s, err := vconf.Parse(`
name=demo
custom_int=custom
custom_bool=yes
[server]
port=8080
enabled=true
`)
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := s.Lookup("", "name"); !ok || got != "demo" {
		t.Fatalf("Lookup default = %q, %v", got, ok)
	}
	if got := s.GetIntWithOptions("custom_int", 0, vconf.WithIntParser(func(value string) (int, error) {
		if value == "custom" {
			return 77, nil
		}
		return 0, errors.New("unexpected")
	})); got != 77 {
		t.Fatalf("GetIntWithOptions = %d", got)
	}
	if got := s.GetBoolByGroupWithOptions("", "custom_bool", false, vconf.WithBoolParser(func(value string) (bool, error) {
		return value == "yes", nil
	})); !got {
		t.Fatal("GetBoolByGroupWithOptions = false")
	}
	if got := s.GetIntByGroupWithOptions("server", "port", 0); got != 8080 {
		t.Fatalf("GetIntByGroupWithOptions = %d", got)
	}
	if got := s.GetBoolByGroupWithOptions("server", "enabled", false); !got {
		t.Fatal("GetBoolByGroupWithOptions server.enabled = false")
	}
}
