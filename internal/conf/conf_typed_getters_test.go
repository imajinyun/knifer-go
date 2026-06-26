package conf

import (
	"errors"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestTypedGettersWithOptionsUseParsers(t *testing.T) {
	s := New()
	s.Set("port", "custom-int")
	s.Set("debug", "custom-bool")
	s.SetByGroup("server", "port", "9090")
	s.SetByGroup("server", "debug", "true")

	intCalled := false
	if got := s.GetIntWithOptions("port", 10, WithIntParser(func(text string) (int, error) {
		intCalled = true
		if text != "custom-int" {
			t.Fatalf("int parser text = %q", text)
		}
		return 8080, nil
	})); got != 8080 || !intCalled {
		t.Fatalf("GetIntWithOptions = %d, called=%v", got, intCalled)
	}

	boolCalled := false
	if got := s.GetBoolWithOptions("debug", false, WithBoolParser(func(text string) (bool, error) {
		boolCalled = true
		if text != "custom-bool" {
			t.Fatalf("bool parser text = %q", text)
		}
		return true, nil
	})); !got || !boolCalled {
		t.Fatalf("GetBoolWithOptions = %v, called=%v", got, boolCalled)
	}
	if got := s.GetIntWithOptions("port", 10, WithIntParser(func(string) (int, error) {
		return 0, errors.New("invalid")
	})); got != 10 {
		t.Fatalf("GetIntWithOptions fallback = %d", got)
	}
	if got, err := s.GetIntEWithOptions("port", WithIntParser(func(string) (int, error) { return 7000, nil })); err != nil || got != 7000 {
		t.Fatalf("GetIntEWithOptions = %d, err=%v", got, err)
	}
	if got, err := s.GetBoolEWithOptions("debug", WithBoolParser(func(string) (bool, error) { return true, nil })); err != nil || !got {
		t.Fatalf("GetBoolEWithOptions = %v, err=%v", got, err)
	}
	if got, err := s.GetIntByGroupE("server", "port"); err != nil || got != 9090 {
		t.Fatalf("GetIntByGroupE = %d, err=%v", got, err)
	}
	if got, err := s.GetBoolByGroupE("server", "debug"); err != nil || !got {
		t.Fatalf("GetBoolByGroupE = %v, err=%v", got, err)
	}

	s.Set("bad-int", "abc")
	if _, err := s.GetIntE("missing"); !errors.Is(err, knifer.ErrCodeNotFound) {
		t.Fatalf("GetIntE missing err = %v, want not found", err)
	}
	if _, err := s.GetIntE("bad-int"); !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("GetIntE invalid err = %v, want invalid input", err)
	}
}
