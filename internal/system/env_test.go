package system

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestGetEnv(t *testing.T) {
	t.Setenv("GKSYSTEM_TEST_KEY", "abc")
	if v := Get("GKSYSTEM_TEST_KEY", true); v != "abc" {
		t.Errorf("Get 应返回 abc，实际 %q", v)
	}
	if v := GetOrDefault("GKSYSTEM_TEST_NOT_EXIST", "def"); v != "def" {
		t.Errorf("GetOrDefault 默认值未生效: %q", v)
	}

	t.Setenv("GKSYSTEM_TEST_INT", "42")
	if n := GetInt("GKSYSTEM_TEST_INT", 0); n != 42 {
		t.Errorf("GetInt: 期望 42，实际 %d", n)
	}
	if n := GetInt("GKSYSTEM_TEST_INT_INVALID", 7); n != 7 {
		t.Errorf("GetInt 无效值应返回默认: %d", n)
	}

	t.Setenv("GKSYSTEM_TEST_BOOL", "true")
	if b := GetBool("GKSYSTEM_TEST_BOOL", false); !b {
		t.Errorf("GetBool 应为 true")
	}
}

func TestGetEnvWithOptions(t *testing.T) {
	lookup := func(key string) (string, bool) {
		switch key {
		case "STRING":
			return "value", true
		case "INT":
			return "12", true
		case "BOOL":
			return "true", true
		case "EMPTY":
			return "", true
		default:
			return "", false
		}
	}
	var warning bytes.Buffer
	if got := GetWithOptions("STRING", true, WithEnvLookupFunc(lookup)); got != "value" {
		t.Fatalf("GetWithOptions = %q", got)
	}
	if got := GetWithOptions("MISSING", false, WithEnvLookupFunc(lookup), WithEnvWarningWriter(&warning)); got != "" || !strings.Contains(warning.String(), "MISSING") {
		t.Fatalf("GetWithOptions missing = %q warning=%q", got, warning.String())
	}
	if got := GetOrDefaultWithOptions("EMPTY", "def", WithEnvLookupFunc(lookup)); got != "def" {
		t.Fatalf("GetOrDefaultWithOptions empty = %q", got)
	}
	intCalled := false
	if got := GetIntWithOptions("INT", 0, WithEnvLookupFunc(lookup), WithEnvIntParser(func(text string) (int, error) {
		intCalled = true
		if text != "12" {
			t.Fatalf("env int parser text = %q", text)
		}
		return 21, nil
	})); got != 21 || !intCalled {
		t.Fatalf("GetIntWithOptions = %d", got)
	}
	if got := GetIntWithOptions("INT", 7, WithEnvLookupFunc(lookup), WithEnvIntParser(func(string) (int, error) {
		return 0, errors.New("invalid int")
	})); got != 7 {
		t.Fatalf("GetIntWithOptions fallback = %d", got)
	}
	boolCalled := false
	if got := GetBoolWithOptions("BOOL", false, WithEnvLookupFunc(lookup), WithEnvBoolParser(func(text string) (bool, error) {
		boolCalled = true
		if text != "true" {
			t.Fatalf("env bool parser text = %q", text)
		}
		return true, nil
	})); !got || !boolCalled {
		t.Fatalf("GetBoolWithOptions = %v", got)
	}
	if got := GetBoolWithOptions("BOOL", true, WithEnvLookupFunc(lookup), WithEnvBoolParser(func(string) (bool, error) {
		return false, errors.New("invalid bool")
	})); !got {
		t.Fatalf("GetBoolWithOptions fallback = %v", got)
	}
}

func TestEnvNilOptionsFallBackToDefaults(t *testing.T) {
	t.Setenv("GKSYSTEM_NIL_OPTION_INT", "8")
	t.Setenv("GKSYSTEM_NIL_OPTION_BOOL", "true")
	if got := GetWithOptions("GKSYSTEM_NIL_OPTION_INT", true, nil, WithEnvLookupFunc(nil), WithEnvWarningWriter(nil)); got != "8" {
		t.Fatalf("GetWithOptions nil fallback = %q", got)
	}
	if got := GetIntWithOptions("GKSYSTEM_NIL_OPTION_INT", 0, WithEnvIntParser(nil)); got != 8 {
		t.Fatalf("GetIntWithOptions nil parser fallback = %d", got)
	}
	if got := GetBoolWithOptions("GKSYSTEM_NIL_OPTION_BOOL", false, WithEnvBoolParser(nil)); !got {
		t.Fatalf("GetBoolWithOptions nil parser fallback = %v", got)
	}
	if got := GetOrDefaultWithOptions("GKSYSTEM_NIL_OPTION_MISSING", "fallback", WithEnvLookupFunc(nil)); got != "fallback" {
		t.Fatalf("GetOrDefaultWithOptions nil lookup fallback = %q", got)
	}
}
