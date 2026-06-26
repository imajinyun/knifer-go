package vconf_test

import (
	"errors"
	"testing"

	"github.com/imajinyun/knifer-go/vconf"
)

func TestFacadeParserProviderOptions(t *testing.T) {
	parsed, err := vconf.ParseByExtWithOptions("app.custom", []byte("ignored"), vconf.WithParserForExt(".custom", func(data []byte) (*vconf.Conf, error) {
		c := vconf.New()
		c.Set("from", string(data))
		return c, nil
	}))
	if err != nil || parsed.Get("from") != "ignored" {
		t.Fatalf("ParseByExtWithOptions = %#v, %v", parsed, err)
	}

	yamlCalled := false
	_, err = vconf.ParseYAMLFullWithOptions("ignored", vconf.WithYAMLUnmarshalFunc(func(data []byte, out any) error {
		yamlCalled = true
		return errors.New("yaml provider failed")
	}))
	if err == nil || !yamlCalled {
		t.Fatalf("ParseYAMLFullWithOptions err=%v called=%v", err, yamlCalled)
	}

	tomlCalled := false
	_, err = vconf.ParseTOMLWithOptions("ignored", vconf.WithTOMLUnmarshalFunc(func(data []byte, out any) error {
		tomlCalled = true
		return errors.New("toml provider failed")
	}))
	if err == nil || !tomlCalled {
		t.Fatalf("ParseTOMLWithOptions err=%v called=%v", err, tomlCalled)
	}
}
