package vsys_test

import (
	"bytes"
	"os"
	"os/user"
	"testing"

	"github.com/imajinyun/go-knifer/vsys"
)

func TestFacadeHostInfo(t *testing.T) {
	info := vsys.SystemHostInfo()
	if info == nil {
		t.Fatal("expected non-nil host info")
	}
}

func TestFacadeOsInfo(t *testing.T) {
	info := vsys.SystemOsInfo()
	if info == nil {
		t.Fatal("expected non-nil os info")
	}
}

func TestFacadeUserInfo(t *testing.T) {
	info := vsys.SystemUserInfo()
	if info == nil {
		t.Fatal("expected non-nil user info")
	}
}

func TestFacadeUserInfoOptions(t *testing.T) {
	info := vsys.SystemUserInfoWithOptions(
		vsys.WithCurrentUserFunc(func() (*user.User, error) {
			return &user.User{Username: "facade-user", HomeDir: "/home/facade"}, nil
		}),
		vsys.WithWorkingDirFunc(func() (string, error) { return "/work/facade", nil }),
		vsys.WithTempDirFunc(func() string { return "/tmp/facade" }),
		vsys.WithUserEnvLookup(func(key string) string {
			if key == "LANG" {
				return "zh_CN.UTF-8"
			}
			return ""
		}),
	)
	sep := string(os.PathSeparator)
	if info.GetName() != "facade-user" || info.GetHomeDir() != "/home/facade"+sep || info.GetCurrentDir() != "/work/facade"+sep || info.GetTempDir() != "/tmp/facade"+sep {
		t.Fatalf("SystemUserInfoWithOptions = %#v", info)
	}
	if info.GetLanguage() != "zh" || info.GetCountry() != "CN" {
		t.Fatalf("SystemUserInfoWithOptions locale = %s/%s", info.GetLanguage(), info.GetCountry())
	}

	info = vsys.NewUserInfoWithOptions(vsys.WithCurrentUserFunc(func() (*user.User, error) {
		return &user.User{Username: "new-user", HomeDir: "/home/new"}, nil
	}))
	if info.GetName() != "new-user" {
		t.Fatalf("NewUserInfoWithOptions name = %q", info.GetName())
	}
}

func TestFacadeGoInfo(t *testing.T) {
	info := vsys.SystemGoInfo()
	if info == nil {
		t.Fatal("expected non-nil go info")
	}
}

func TestFacadeRuntimeInfo(t *testing.T) {
	info := vsys.SystemRuntimeInfo()
	if info == nil {
		t.Fatal("expected non-nil runtime info")
	}
}

func TestFacadePID(t *testing.T) {
	pid := vsys.CurrentPID()
	if pid <= 0 {
		t.Fatalf("expected positive pid, got %d", pid)
	}
}

func TestFacadeMemory(t *testing.T) {
	total := vsys.TotalMemory()
	free := vsys.FreeMemory()
	max := vsys.MaxMemory()
	if total == 0 && free == 0 && max == 0 {
		t.Fatal("expected at least one memory metric to be non-zero")
	}
}

func TestFacadeGoroutineCount(t *testing.T) {
	count := vsys.TotalGoroutineCount()
	if count < 1 {
		t.Fatalf("expected at least 1 goroutine, got %d", count)
	}
}

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
}

func TestFacadeDumpSystemInfo(t *testing.T) {
	var buf bytes.Buffer
	vsys.DumpSystemInfoTo(&buf)
	if buf.Len() == 0 {
		t.Fatal("expected non-empty system info dump")
	}
}
