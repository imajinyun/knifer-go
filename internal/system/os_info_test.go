package system

import (
	"runtime"
	"strings"
	"testing"
)

func TestOsInfo(t *testing.T) {
	o := GetOsInfo()
	if o == nil {
		t.Fatal("OsInfo 不应为 nil")
	}
	if o.GetName() != runtime.GOOS {
		t.Errorf("OS Name: 期望 %s 实际 %s", runtime.GOOS, o.GetName())
	}
	if o.GetArch() != runtime.GOARCH {
		t.Errorf("OS Arch: 期望 %s 实际 %s", runtime.GOARCH, o.GetArch())
	}
	switch runtime.GOOS {
	case "darwin":
		if !o.IsMac() || !o.IsMacOsX() {
			t.Errorf("darwin 应识别为 Mac")
		}
	case "linux":
		if !o.IsLinux() {
			t.Errorf("linux 应识别为 Linux")
		}
	case "windows":
		if !o.IsWindows() {
			t.Errorf("windows 应识别为 Windows")
		}
	}
	if o.GetFileSeparator() == "" || o.GetPathSeparator() == "" || o.GetLineSeparator() == "" {
		t.Errorf("分隔符不应为空: %+v", o)
	}
}

func TestOsInfoWithOptions(t *testing.T) {
	o := NewOsInfoWithOptions(
		WithOSNameFunc(func() string { return "linux" }),
		WithOSArchFunc(func() string { return "arm64" }),
		WithOSVersionFunc(func() string { return "test-version" }),
		WithOSFileSeparatorFunc(func() string { return "/" }),
		WithOSLineSeparatorFunc(func() string { return "\n" }),
		WithOSPathSeparatorFunc(func() string { return ":" }),
	)
	if o.GetName() != "linux" || o.GetArch() != "arm64" || o.GetVersion() != "test-version" || o.GetFileSeparator() != "/" || o.GetLineSeparator() != "\n" || o.GetPathSeparator() != ":" {
		t.Fatalf("NewOsInfoWithOptions = %#v", o)
	}
	if !o.IsLinux() || o.IsWindows() {
		t.Fatalf("NewOsInfoWithOptions OS helpers = %#v", o)
	}

	o = NewOsInfoWithOptions(
		WithOSNameFunc(func() string { return "windows" }),
		WithOSEnvLookupFunc(func(string) string { return "" }),
	)
	if o.GetVersion() != "windows" || o.GetLineSeparator() != "\r\n" {
		t.Fatalf("OS providers should drive version and line separator: %#v", o)
	}
}

func TestOsInfoNilOptionsAndVersionFallbacks(t *testing.T) {
	o := NewOsInfoWithOptions(
		nil,
		WithOSNameFunc(nil),
		WithOSArchFunc(nil),
		WithOSVersionFunc(nil),
		WithOSEnvLookupFunc(func(key string) string {
			if key == "OSVERSION" {
				return "13.6"
			}
			return ""
		}),
		WithOSFileSeparatorFunc(nil),
		WithOSLineSeparatorFunc(nil),
		WithOSPathSeparatorFunc(nil),
	)
	if o.GetName() != runtime.GOOS || o.GetArch() != runtime.GOARCH || o.GetVersion() != "13.6" {
		t.Fatalf("nil option OS fallback = %#v", o)
	}
	if o.GetFileSeparator() == "" || o.GetLineSeparator() == "" || o.GetPathSeparator() == "" {
		t.Fatalf("nil option separators = %#v", o)
	}

	ostype := NewOsInfoWithOptions(
		WithOSNameFunc(func() string { return "linux" }),
		WithOSEnvLookupFunc(func(key string) string {
			if key == "OSTYPE" {
				return "linux-gnu"
			}
			return ""
		}),
	)
	if ostype.GetVersion() != "linux-gnu" || !ostype.IsLinux() {
		t.Fatalf("OSTYPE fallback = %#v", ostype)
	}
}

func TestOsInfoHelpersForSupportedNames(t *testing.T) {
	cases := []struct {
		name string
		want func(*OsInfo) bool
	}{
		{name: "darwin", want: func(o *OsInfo) bool { return o.IsMac() && o.IsMacOsX() }},
		{name: "windows", want: func(o *OsInfo) bool { return o.IsWindows() }},
		{name: "aix", want: func(o *OsInfo) bool { return o.IsAix() }},
		{name: "solaris", want: func(o *OsInfo) bool { return o.IsSolaris() }},
		{name: "freebsd", want: func(o *OsInfo) bool { return o.IsFreeBSD() }},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			o := NewOsInfoWithOptions(WithOSNameFunc(func() string { return tt.name }))
			if !tt.want(o) {
				t.Fatalf("OS helper mismatch for %#v", o)
			}
		})
	}
}

func TestLineSeparatorAndReadOsVersionHandleNilProviders(t *testing.T) {
	if got := lineSeparator(func() string { return "windows" }); got != "\r\n" {
		t.Fatalf("lineSeparator(windows) = %q", got)
	}
	if got := lineSeparator(func() string { return "linux" }); got != "\n" {
		t.Fatalf("lineSeparator(linux) = %q", got)
	}
	if got := lineSeparator(nil); got == "" {
		t.Fatal("lineSeparator(nil) should use runtime fallback")
	}
	if got := readOsVersion(func(string) string { return "" }, func() string { return "  plan9  " }); got != "plan9" {
		t.Fatalf("readOsVersion trim fallback = %q", got)
	}
	if got := readOsVersion(nil, func() string { return "linux" }); strings.TrimSpace(got) == "" {
		t.Fatalf("readOsVersion nil getenv = %q", got)
	}
}
