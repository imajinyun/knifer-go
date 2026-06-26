package vsys_test

import (
	"bytes"
	"os"
	"os/user"
	"runtime"
	"testing"

	"github.com/imajinyun/knifer-go/vsys"
)

func TestFacadeEnv(t *testing.T) {
	_ = os.Setenv("GO_KNIFER_TEST_KEY", "test_value")
	defer os.Unsetenv("GO_KNIFER_TEST_KEY")

	if got := vsys.Env("GO_KNIFER_TEST_KEY"); got != "test_value" {
		t.Fatalf("expected 'test_value', got %q", got)
	}
	if got := vsys.EnvOrDefault("GO_KNIFER_TEST_MISSING", "default"); got != "default" {
		t.Fatalf("expected 'default', got %q", got)
	}
	if got := vsys.EnvInt("GO_KNIFER_TEST_KEY", 0); got != 0 {
		t.Fatalf("expected 0 for non-int env, got %d", got)
	}
	if got := vsys.EnvBool("GO_KNIFER_TEST_KEY", false); got != false {
		t.Fatalf("expected false for non-bool env, got %v", got)
	}

	lookup := vsys.WithEnvLookupFunc(func(key string) (string, bool) {
		switch key {
		case "A":
			return "value", true
		case "N":
			return "13", true
		case "B":
			return "true", true
		default:
			return "", false
		}
	})
	var warning bytes.Buffer
	if got := vsys.EnvWithOptions("A", lookup); got != "value" {
		t.Fatalf("EnvWithOptions = %q", got)
	}
	if got := vsys.GetWithOptions("MISSING", false, lookup, vsys.WithEnvWarningWriter(&warning)); got != "" || warning.Len() == 0 {
		t.Fatalf("GetWithOptions missing = %q warning=%q", got, warning.String())
	}
	if got := vsys.EnvOrDefaultWithOptions("MISSING", "def", lookup); got != "def" {
		t.Fatalf("EnvOrDefaultWithOptions = %q", got)
	}
	if got := vsys.EnvIntWithOptions("N", 0, lookup); got != 13 {
		t.Fatalf("EnvIntWithOptions = %d", got)
	}
	if got := vsys.EnvBoolWithOptions("B", false, lookup); !got {
		t.Fatalf("EnvBoolWithOptions = %v", got)
	}
}

func TestFacadeDumpSystemInfo(t *testing.T) {
	var buf bytes.Buffer
	vsys.DumpSystemInfoTo(&buf)
	if buf.Len() == 0 {
		t.Fatal("expected non-empty system info dump")
	}
}

func TestFacadeDumpSystemInfoStdout(t *testing.T) {
	// DumpSystemInfo writes to stdout; we just verify it doesn't panic.
	vsys.DumpSystemInfo()
}

func TestFacadeGetEnvFunctions(t *testing.T) {
	_ = os.Setenv("GO_KNIFER_GET_TEST", "get_val")
	defer os.Unsetenv("GO_KNIFER_GET_TEST")

	if got := vsys.Get("GO_KNIFER_GET_TEST", true); got != "get_val" {
		t.Fatalf("Get = %q", got)
	}
	if got := vsys.GetOrDefault("GO_KNIFER_GET_MISSING", "def"); got != "def" {
		t.Fatalf("GetOrDefault = %q", got)
	}
	if got := vsys.GetInt("GO_KNIFER_GET_TEST", 0); got != 0 {
		t.Fatalf("GetInt expected 0 for non-int, got %d", got)
	}
	if got := vsys.GetBool("GO_KNIFER_GET_TEST", true); got != true {
		t.Fatalf("GetBool expected default true, got %v", got)
	}

	lookup := vsys.WithEnvLookupFunc(func(key string) (string, bool) {
		switch key {
		case "X":
			return "42", true
		case "Y":
			return "true", true
		case "Z":
			return "zval", true
		default:
			return "", false
		}
	})
	if got := vsys.GetOrDefaultWithOptions("MISSING", "fallback", lookup); got != "fallback" {
		t.Fatalf("GetOrDefaultWithOptions = %q", got)
	}
	if got := vsys.GetIntWithOptions("X", 0, lookup); got != 42 {
		t.Fatalf("GetIntWithOptions = %d", got)
	}
	if got := vsys.GetBoolWithOptions("Y", false, lookup); got != true {
		t.Fatalf("GetBoolWithOptions = %v", got)
	}
}

func TestFacadeEnvParserOptions(t *testing.T) {
	opt1 := vsys.WithEnvIntParser(func(s string) (int, error) {
		if s == "one" {
			return 1, nil
		}
		return 0, nil
	})
	if opt1 == nil {
		t.Fatal("WithEnvIntParser returned nil")
	}

	opt2 := vsys.WithEnvBoolParser(func(s string) (bool, error) {
		return s == "yes", nil
	})
	if opt2 == nil {
		t.Fatal("WithEnvBoolParser returned nil")
	}
}

func TestFacadeDumpOptions(t *testing.T) {
	opt1 := vsys.WithDumpHostOptions(vsys.WithHostNameFunc(func() (string, error) { return "dump-host", nil }))
	if opt1 == nil {
		t.Fatal("WithDumpHostOptions returned nil")
	}
	opt2 := vsys.WithDumpOsOptions(vsys.WithOSNameFunc(func() string { return "dump-os" }))
	if opt2 == nil {
		t.Fatal("WithDumpOsOptions returned nil")
	}
	opt3 := vsys.WithDumpUserOptions(vsys.WithCurrentUserFunc(func() (*user.User, error) {
		return &user.User{Username: "dump-user"}, nil
	}))
	if opt3 == nil {
		t.Fatal("WithDumpUserOptions returned nil")
	}
	opt4 := vsys.WithDumpGoOptions(vsys.WithGoVersionFunc(func() string { return "dump-go" }))
	if opt4 == nil {
		t.Fatal("WithDumpGoOptions returned nil")
	}
	opt5 := vsys.WithDumpRuntimeOptions(vsys.WithReadMemStatsFunc(func(stats *runtime.MemStats) { stats.Sys = 500 }))
	if opt5 == nil {
		t.Fatal("WithDumpRuntimeOptions returned nil")
	}

	var buf bytes.Buffer
	vsys.DumpSystemInfoWithOptions(&buf, opt1, opt2, opt3, opt4, opt5)
	if buf.Len() == 0 {
		t.Fatal("expected non-empty dump with options")
	}
}

func TestFacadeSystemInfoWithOptionsAliases(t *testing.T) {
	info := vsys.SystemHostInfoWithOptions(vsys.WithHostNameFunc(func() (string, error) { return "alias-host", nil }))
	if info.GetName() != "alias-host" {
		t.Fatalf("SystemHostInfoWithOptions name = %q", info.GetName())
	}

	userInfo := vsys.SystemUserInfoWithOptions(vsys.WithCurrentUserFunc(func() (*user.User, error) {
		return &user.User{Username: "alias-user"}, nil
	}))
	if userInfo.GetName() != "alias-user" {
		t.Fatalf("SystemUserInfoWithOptions name = %q", userInfo.GetName())
	}

	goInfo := vsys.SystemGoInfoWithOptions(vsys.WithGoVersionFunc(func() string { return "alias-go" }))
	if goInfo.GetVersion() != "alias-go" {
		t.Fatalf("SystemGoInfoWithOptions version = %q", goInfo.GetVersion())
	}

	runtimeInfo := vsys.SystemRuntimeInfoWithOptions(vsys.WithReadMemStatsFunc(func(stats *runtime.MemStats) { stats.Sys = 777 }))
	if runtimeInfo.GetMaxMemory() != 777 {
		t.Fatalf("SystemRuntimeInfoWithOptions max = %d", runtimeInfo.GetMaxMemory())
	}
}
