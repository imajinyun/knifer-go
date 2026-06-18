package conf

import (
	"testing"
)

func TestGetInt(t *testing.T) {
	s := New()
	s.Set("port", "8080")
	s.Set("invalid", "not-a-number")
	if got := s.GetInt("port", 0); got != 8080 {
		t.Fatalf("GetInt('port') = %d", got)
	}
	if got := s.GetInt("missing", 42); got != 42 {
		t.Fatalf("GetInt('missing', 42) = %d", got)
	}
	if got := s.GetInt("invalid", 99); got != 99 {
		t.Fatalf("GetInt('invalid', 99) = %d", got)
	}
}

func TestGetIntByGroup(t *testing.T) {
	s := New()
	s.SetByGroup("server", "port", "8080")
	if got := s.GetIntByGroup("server", "port", 0); got != 8080 {
		t.Fatalf("GetIntByGroup = %d", got)
	}
	if got := s.GetIntByGroup("server", "missing", 42); got != 42 {
		t.Fatalf("GetIntByGroup missing = %d", got)
	}
	s.SetByGroup("server", "invalid", "bad")
	if got := s.GetIntByGroupWithOptions("server", "invalid", 99); got != 99 {
		t.Fatalf("GetIntByGroupWithOptions invalid = %d", got)
	}
}

func TestGetBool(t *testing.T) {
	s := New()
	s.Set("enabled", "true")
	s.Set("disabled", "false")
	s.Set("invalid", "maybe")
	if !s.GetBool("enabled", false) {
		t.Fatal("GetBool('enabled') = false")
	}
	if s.GetBool("disabled", true) {
		t.Fatal("GetBool('disabled') = true")
	}
	if s.GetBool("missing", true) != true {
		t.Fatal("GetBool('missing', true) = false")
	}
	if s.GetBool("invalid", true) != true {
		t.Fatal("GetBool('invalid', true) = false")
	}
	if _, err := s.GetBoolE("missing"); err == nil {
		t.Fatal("GetBoolE('missing') err = nil")
	}
}

func TestGetBoolByGroup(t *testing.T) {
	s := New()
	s.SetByGroup("app", "debug", "true")
	if !s.GetBoolByGroup("app", "debug", false) {
		t.Fatal("GetBoolByGroup = false")
	}
	if s.GetBoolByGroup("app", "missing", true) != true {
		t.Fatal("GetBoolByGroup missing = false")
	}
	s.SetByGroup("app", "invalid", "bad")
	if s.GetBoolByGroupWithOptions("app", "invalid", false) {
		t.Fatal("GetBoolByGroupWithOptions invalid = true")
	}
}

func TestBind(t *testing.T) {
	type config struct {
		Name string `conf:"name"`
		Port int    `conf:"port"`
	}
	s := New()
	s.Set("name", "test")
	s.Set("port", "8080")

	var cfg config
	if err := s.Bind(&cfg); err != nil {
		t.Fatalf("Bind error = %v", err)
	}
	if cfg.Name != "test" || cfg.Port != 8080 {
		t.Fatalf("Bind = %+v", cfg)
	}
}

func TestBindWithOptions(t *testing.T) {
	type config struct {
		Port int `conf:"port"`
	}
	s := New()
	s.Set("port", "8080")

	var cfg config
	if err := s.BindWithOptions(&cfg); err != nil {
		t.Fatalf("BindWithOptions error = %v", err)
	}
	if cfg.Port != 8080 {
		t.Fatalf("BindWithOptions = %+v", cfg)
	}
}
