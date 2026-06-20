package system

import (
	"os"
	"os/user"
	"strings"
	"testing"
)

func TestUserInfo(t *testing.T) {
	u := GetUserInfo()
	if u == nil {
		t.Fatal("UserInfo 不应为 nil")
	}
	if u.GetCurrentDir() == "" {
		t.Errorf("CurrentDir 不应为空")
	}
	if u.GetTempDir() == "" {
		t.Errorf("TempDir 不应为空")
	}
	if !strings.HasSuffix(u.GetCurrentDir(), string(os.PathSeparator)) {
		t.Errorf("CurrentDir 应以路径分隔符结尾: %q", u.GetCurrentDir())
	}
	if !strings.Contains(u.String(), "User Name:") {
		t.Errorf("UserInfo.String 缺少 caption")
	}
}

func TestUserInfoWithOptions(t *testing.T) {
	u := NewUserInfoWithOptions(
		WithCurrentUserFunc(func() (*user.User, error) {
			return &user.User{Username: "option-user", HomeDir: "/home/option"}, nil
		}),
		WithWorkingDirFunc(func() (string, error) { return "/work/option", nil }),
		WithTempDirFunc(func() string { return "/tmp/option" }),
		WithUserEnvLookup(func(key string) string {
			if key == "LANG" {
				return "zh_CN.UTF-8"
			}
			return ""
		}),
	)
	sep := string(os.PathSeparator)
	if u.GetName() != "option-user" || u.GetHomeDir() != "/home/option"+sep || u.GetCurrentDir() != "/work/option"+sep || u.GetTempDir() != "/tmp/option"+sep {
		t.Fatalf("NewUserInfoWithOptions paths = %#v", u)
	}
	if u.GetLanguage() != "zh" || u.GetCountry() != "CN" {
		t.Fatalf("NewUserInfoWithOptions locale = %s/%s", u.GetLanguage(), u.GetCountry())
	}

	fallback := GetUserInfoWithOptions(
		WithCurrentUserFunc(func() (*user.User, error) { return nil, os.ErrNotExist }),
		WithWorkingDirFunc(func() (string, error) { return "/cwd/fallback", nil }),
		WithTempDirFunc(func() string { return "/tmp/fallback" }),
		WithUserEnvLookup(func(key string) string {
			switch key {
			case "USER":
				return "env-user"
			case "HOME":
				return "/home/env"
			case "LC_ALL":
				return "en_US.UTF-8"
			default:
				return ""
			}
		}),
	)
	if fallback.GetName() != "env-user" || fallback.GetHomeDir() != "/home/env"+sep || fallback.GetLanguage() != "en" || fallback.GetCountry() != "US" {
		t.Fatalf("GetUserInfoWithOptions fallback = %#v", fallback)
	}
}

func TestUserInfoErrorAndEnvironmentFallbackBoundaries(t *testing.T) {
	sep := string(os.PathSeparator)
	u := NewUserInfoWithOptions(
		WithCurrentUserFunc(func() (*user.User, error) { return nil, os.ErrPermission }),
		WithWorkingDirFunc(func() (string, error) { return "", os.ErrNotExist }),
		WithTempDirFunc(func() string { return "" }),
		WithUserEnvLookup(func(key string) string {
			switch key {
			case "USER":
				return ""
			case "USERNAME":
				return "windows-user"
			case "HOME":
				return "/home/windows-user"
			case "LANG":
				return ""
			case "LC_ALL":
				return "ja_JP.UTF-8"
			default:
				return ""
			}
		}),
	)
	if u.GetName() != "windows-user" || u.GetHomeDir() != "/home/windows-user"+sep || u.GetCurrentDir() != "" || u.GetTempDir() != "" {
		t.Fatalf("user error fallback paths = %#v", u)
	}
	if u.GetLanguage() != "ja" || u.GetCountry() != "JP" {
		t.Fatalf("user LC_ALL fallback locale = %s/%s", u.GetLanguage(), u.GetCountry())
	}
}

func TestUserInfoNilOptionsFallBackToDefaults(t *testing.T) {
	u := NewUserInfoWithOptions(nil, WithCurrentUserFunc(nil), WithUserEnvLookup(nil), WithWorkingDirFunc(nil), WithTempDirFunc(nil))
	if u.GetCurrentDir() == "" || u.GetTempDir() == "" {
		t.Fatalf("nil option user fallback = %#v", u)
	}
}

func TestParseLocale(t *testing.T) {
	lang, country := parseLocale("zh_CN.UTF-8")
	if lang != "zh" || country != "CN" {
		t.Errorf("parseLocale(zh_CN.UTF-8) 错误: %s/%s", lang, country)
	}
	lang, country = parseLocale("")
	if lang != "" || country != "" {
		t.Errorf("空 locale 应返回空")
	}
	lang, country = parseLocale("en")
	if lang != "en" || country != "" {
		t.Errorf("parseLocale(en) 错误: %s/%s", lang, country)
	}
	lang, country = parseLocale("pt_BR_POSIX.UTF-8")
	if lang != "pt" || country != "BR" {
		t.Errorf("parseLocale should ignore extra segments after country: %s/%s", lang, country)
	}
}
