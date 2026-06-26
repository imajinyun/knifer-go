package vsys_test

import (
	"os"
	"os/user"
	"testing"

	"github.com/imajinyun/knifer-go/vsys"
)

func TestFacadeUserInfo(t *testing.T) {
	info := vsys.SystemUserInfo()
	if info == nil {
		t.Fatal("expected non-nil user info")
	}
}

func TestFacadeNewUserInfo(t *testing.T) {
	info := vsys.NewUserInfo()
	if info == nil {
		t.Fatal("expected non-nil user info from NewUserInfo")
	}
}

func TestFacadeGetUserInfo(t *testing.T) {
	info := vsys.GetUserInfo()
	if info == nil {
		t.Fatal("expected non-nil user info from GetUserInfo")
	}
	infoWithOpts := vsys.GetUserInfoWithOptions(vsys.WithCurrentUserFunc(func() (*user.User, error) {
		return &user.User{Username: "get-user"}, nil
	}))
	if infoWithOpts.GetName() != "get-user" {
		t.Fatalf("GetUserInfoWithOptions name = %q", infoWithOpts.GetName())
	}
}

func TestFacadeUserInfoOptions(t *testing.T) {
	info := vsys.SysUserInfoWithOptions(
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
		t.Fatalf("SysUserInfoWithOptions locale = %s/%s", info.GetLanguage(), info.GetCountry())
	}

	info = vsys.NewUserInfoWithOptions(vsys.WithCurrentUserFunc(func() (*user.User, error) {
		return &user.User{Username: "new-user", HomeDir: "/home/new"}, nil
	}))
	if info.GetName() != "new-user" {
		t.Fatalf("NewUserInfoWithOptions name = %q", info.GetName())
	}
}
