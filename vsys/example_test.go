package vsys_test

import (
	"bytes"
	"fmt"
	"net"
	"os/user"
	"runtime"
	"strings"

	"github.com/imajinyun/go-knifer/vsys"
)

func ExampleGetCurrentPID() {
	pid := vsys.GetCurrentPID()
	fmt.Println(pid > 0)
	// Output: true
}

func ExampleEnvWithOptions() {
	lookup := vsys.WithEnvLookupFunc(func(key string) (string, bool) {
		if key == "APP_MODE" {
			return "test", true
		}
		return "", false
	})

	fmt.Println(vsys.EnvWithOptions("APP_MODE", lookup))
	fmt.Println(vsys.EnvWithOptions("MISSING", lookup))
	// Output:
	// test
	//
}

func ExampleEnvOrDefaultWithOptions() {
	lookup := vsys.WithEnvLookupFunc(func(key string) (string, bool) {
		if key == "PORT" {
			return "8080", true
		}
		return "", false
	})

	fmt.Println(vsys.EnvOrDefaultWithOptions("MISSING", "fallback", lookup))
	fmt.Println(vsys.EnvIntWithOptions("PORT", 0, lookup))
	// Output:
	// fallback
	// 8080
}

func ExampleGetCurrentPIDWithOptions() {
	pid := vsys.GetCurrentPIDWithOptions(vsys.WithPIDFunc(func() int {
		return 4242
	}))

	fmt.Println(pid)
	// Output: 4242
}

func ExampleCurrentPIDWithOptions() {
	pid := vsys.CurrentPIDWithOptions(vsys.WithPIDFunc(func() int {
		return 5150
	}))

	fmt.Println(pid)
	// Output: 5150
}

func ExampleEnvIntWithOptions() {
	lookup := vsys.WithEnvLookupFunc(func(key string) (string, bool) {
		if key == "WORKERS" {
			return "12", true
		}
		return "", false
	})

	fmt.Println(vsys.EnvIntWithOptions("WORKERS", 1, lookup))
	fmt.Println(vsys.EnvIntWithOptions("MISSING", 1, lookup))
	// Output:
	// 12
	// 1
}

func ExampleEnvBoolWithOptions() {
	lookup := vsys.WithEnvLookupFunc(func(key string) (string, bool) {
		if key == "FEATURE_ENABLED" {
			return "true", true
		}
		return "", false
	})

	fmt.Println(vsys.EnvBoolWithOptions("FEATURE_ENABLED", false, lookup))
	fmt.Println(vsys.EnvBoolWithOptions("MISSING", true, lookup))
	// Output:
	// true
	// true
}

func ExampleGetWithOptions() {
	var warning bytes.Buffer
	lookup := vsys.WithEnvLookupFunc(func(key string) (string, bool) {
		if key == "APP_NAME" {
			return "knifer", true
		}
		return "", false
	})

	fmt.Println(vsys.GetWithOptions("APP_NAME", false, lookup, vsys.WithEnvWarningWriter(&warning)))
	fmt.Println(vsys.GetWithOptions("MISSING", false, lookup, vsys.WithEnvWarningWriter(&warning)))
	fmt.Print(warning.String())
	// Output:
	// knifer
	//
	// [gksystem] env "MISSING" not found
}

func ExampleGetOrDefaultWithOptions() {
	lookup := vsys.WithEnvLookupFunc(func(key string) (string, bool) {
		return "", false
	})

	fmt.Println(vsys.GetOrDefaultWithOptions("APP_REGION", "cn", lookup))
	// Output: cn
}

func ExampleGetIntWithOptions() {
	lookup := vsys.WithEnvLookupFunc(func(key string) (string, bool) {
		if key == "LIMIT" {
			return "64", true
		}
		return "", false
	})

	fmt.Println(vsys.GetIntWithOptions("LIMIT", 10, lookup))
	// Output: 64
}

func ExampleGetBoolWithOptions() {
	lookup := vsys.WithEnvLookupFunc(func(key string) (string, bool) {
		if key == "DEBUG" {
			return "false", true
		}
		return "", false
	})

	fmt.Println(vsys.GetBoolWithOptions("DEBUG", true, lookup))
	// Output: false
}

func ExampleWithEnvIntParser() {
	lookup := vsys.WithEnvLookupFunc(func(key string) (string, bool) {
		return "one", true
	})
	parseWord := vsys.WithEnvIntParser(func(value string) (int, error) {
		if value == "one" {
			return 1, nil
		}
		return 0, nil
	})

	fmt.Println(vsys.EnvIntWithOptions("COUNT", 0, lookup, parseWord))
	// Output: 1
}

func ExampleWithEnvBoolParser() {
	lookup := vsys.WithEnvLookupFunc(func(key string) (string, bool) {
		return "yes", true
	})
	parseYesNo := vsys.WithEnvBoolParser(func(value string) (bool, error) {
		return value == "yes", nil
	})

	fmt.Println(vsys.EnvBoolWithOptions("READY", false, lookup, parseYesNo))
	// Output: true
}

func ExampleNewGoInfoWithOptions() {
	info := vsys.NewGoInfoWithOptions(
		vsys.WithGoVersionFunc(func() string { return "go1.22.0" }),
		vsys.WithGoCompilerFunc(func() string { return "gc" }),
		vsys.WithGoRootFunc(func() string { return "/sdk/go" }),
		vsys.WithGoOSFunc(func() string { return "linux" }),
		vsys.WithGoArchFunc(func() string { return "amd64" }),
		vsys.WithGoNumCPUFunc(func() int { return 8 }),
		vsys.WithGoNumCgoCallFunc(func() int64 { return 0 }),
	)

	fmt.Println(info.Version, info.GOOS, info.NumCPU)
	// Output: go1.22.0 linux 8
}

func ExampleSysGoInfoWithOptions() {
	info := vsys.SysGoInfoWithOptions(vsys.WithGoVersionFunc(func() string { return "go-sys" }))
	fmt.Println(info.GetVersion())
	// Output: go-sys
}

func ExampleSystemGoInfoWithOptions() {
	info := vsys.SystemGoInfoWithOptions(vsys.WithGoVersionFunc(func() string { return "go-system" }))
	fmt.Println(info.GetVersion())
	// Output: go-system
}

func ExampleGetGoInfoWithOptions() {
	info := vsys.GetGoInfoWithOptions(vsys.WithGoArchFunc(func() string { return "arm64" }))
	fmt.Println(info.GetGOARCH())
	// Output: arm64
}

func ExampleWithGoEnvOutputFunc() {
	info := vsys.NewGoInfoWithOptions(vsys.WithGoEnvOutputFunc(func(string, ...string) ([]byte, error) {
		return []byte("/sdk/custom\n"), nil
	}))

	fmt.Println(info.GetGOROOT())
	// Output: /sdk/custom
}

func ExampleWithGoRootEnvLookupFunc() {
	info := vsys.NewGoInfoWithOptions(
		vsys.WithGoEnvOutputFunc(func(string, ...string) ([]byte, error) { return nil, nil }),
		vsys.WithGoRootEnvLookupFunc(func(key string) string {
			if key == "GOROOT" {
				return "/env/goroot"
			}
			return ""
		}),
	)

	fmt.Println(info.GetGOROOT())
	// Output: /env/goroot
}

func ExampleNewHostInfoWithOptions() {
	info := vsys.NewHostInfoWithOptions(
		vsys.WithHostNameFunc(func() (string, error) { return "host-a", nil }),
		vsys.WithHostAddressFunc(func() string { return "198.51.100.10" }),
	)

	fmt.Println(info.GetName(), info.GetAddress())
	// Output: host-a 198.51.100.10
}

func ExampleGetHostInfoWithOptions() {
	info := vsys.GetHostInfoWithOptions(vsys.WithHostNameFunc(func() (string, error) { return "get-host", nil }))
	fmt.Println(info.GetName())
	// Output: get-host
}

func ExampleSysHostInfoWithOptions() {
	info := vsys.SysHostInfoWithOptions(vsys.WithHostAddressFunc(func() string { return "203.0.113.7" }))
	fmt.Println(info.GetAddress())
	// Output: 203.0.113.7
}

func ExampleSystemHostInfoWithOptions() {
	info := vsys.SystemHostInfoWithOptions(vsys.WithHostNameFunc(func() (string, error) { return "system-host", nil }))
	fmt.Println(info.GetName())
	// Output: system-host
}

func ExampleWithHostInterfaceAddrsFunc() {
	_, ipNet, _ := net.ParseCIDR("10.0.0.8/24")
	ipNet.IP = net.ParseIP("10.0.0.8")
	info := vsys.NewHostInfoWithOptions(vsys.WithHostInterfaceAddrsFunc(func() ([]net.Addr, error) {
		return []net.Addr{ipNet}, nil
	}))

	fmt.Println(info.GetAddress())
	// Output: 10.0.0.8
}

func ExampleNewOsInfoWithOptions() {
	info := vsys.NewOsInfoWithOptions(
		vsys.WithOSNameFunc(func() string { return "windows" }),
		vsys.WithOSArchFunc(func() string { return "amd64" }),
		vsys.WithOSVersionFunc(func() string { return "11" }),
		vsys.WithOSFileSeparatorFunc(func() string { return "\\" }),
		vsys.WithOSLineSeparatorFunc(func() string { return "\r\n" }),
		vsys.WithOSPathSeparatorFunc(func() string { return ";" }),
	)

	fmt.Println(info.GetName(), info.GetArch(), info.GetVersion(), info.IsWindows())
	// Output: windows amd64 11 true
}

func ExampleGetOsInfoWithOptions() {
	info := vsys.GetOsInfoWithOptions(vsys.WithOSNameFunc(func() string { return "linux" }))
	fmt.Println(info.GetName(), info.IsLinux())
	// Output: linux true
}

func ExampleSysOsInfoWithOptions() {
	info := vsys.SysOsInfoWithOptions(vsys.WithOSNameFunc(func() string { return "darwin" }))
	fmt.Println(info.GetName(), info.IsMacOsX())
	// Output: darwin true
}

func ExampleSystemOsInfoWithOptions() {
	info := vsys.SystemOsInfoWithOptions(vsys.WithOSNameFunc(func() string { return "solaris" }))
	fmt.Println(info.GetName(), info.IsSolaris())
	// Output: solaris true
}

func ExampleWithOSEnvLookupFunc() {
	info := vsys.NewOsInfoWithOptions(
		vsys.WithOSNameFunc(func() string { return "linux" }),
		vsys.WithOSEnvLookupFunc(func(key string) string {
			if key == "OSVERSION" {
				return "6.8.0"
			}
			return ""
		}),
	)

	fmt.Println(info.GetVersion())
	// Output: 6.8.0
}

func ExampleNewRuntimeInfoWithOptions() {
	info := vsys.NewRuntimeInfoWithOptions(
		vsys.WithReadMemStatsFunc(func(stats *runtime.MemStats) {
			stats.Sys = 4096
			stats.HeapSys = 2048
			stats.HeapIdle = 512
		}),
		vsys.WithNumGoroutineFunc(func() int { return 3 }),
	)

	fmt.Println(info.GetMaxMemory(), info.GetTotalMemory(), info.GetFreeMemory(), info.GetGoroutineCount())
	// Output: 4096 2048 512 3
}

func ExampleGetRuntimeInfoWithOptions() {
	info := vsys.GetRuntimeInfoWithOptions(vsys.WithReadMemStatsFunc(func(stats *runtime.MemStats) {
		stats.Sys = 9000
	}))

	fmt.Println(info.GetMaxMemory())
	// Output: 9000
}

func ExampleSysRuntimeInfoWithOptions() {
	info := vsys.SysRuntimeInfoWithOptions(vsys.WithNumGoroutineFunc(func() int { return 6 }))
	fmt.Println(info.GetGoroutineCount())
	// Output: 6
}

func ExampleSystemRuntimeInfoWithOptions() {
	info := vsys.SystemRuntimeInfoWithOptions(vsys.WithReadMemStatsFunc(func(stats *runtime.MemStats) {
		stats.HeapInuse = 100
		stats.Sys = 500
	}))

	fmt.Println(info.GetUsableMemory())
	// Output: 400
}

func ExampleGetTotalMemoryWithOptions() {
	mem := vsys.GetTotalMemoryWithOptions(vsys.WithReadMemStatsFunc(func(stats *runtime.MemStats) {
		stats.HeapSys = 2048
	}))

	fmt.Println(mem)
	// Output: 2048
}

func ExampleGetFreeMemoryWithOptions() {
	mem := vsys.GetFreeMemoryWithOptions(vsys.WithReadMemStatsFunc(func(stats *runtime.MemStats) {
		stats.HeapIdle = 1024
	}))

	fmt.Println(mem)
	// Output: 1024
}

func ExampleGetMaxMemoryWithOptions() {
	mem := vsys.GetMaxMemoryWithOptions(vsys.WithReadMemStatsFunc(func(stats *runtime.MemStats) {
		stats.Sys = 8192
	}))

	fmt.Println(mem)
	// Output: 8192
}

func ExampleTotalMemoryWithOptions() {
	mem := vsys.TotalMemoryWithOptions(vsys.WithReadMemStatsFunc(func(stats *runtime.MemStats) {
		stats.HeapSys = 333
	}))

	fmt.Println(mem)
	// Output: 333
}

func ExampleFreeMemoryWithOptions() {
	mem := vsys.FreeMemoryWithOptions(vsys.WithReadMemStatsFunc(func(stats *runtime.MemStats) {
		stats.HeapIdle = 444
	}))

	fmt.Println(mem)
	// Output: 444
}

func ExampleMaxMemoryWithOptions() {
	mem := vsys.MaxMemoryWithOptions(vsys.WithReadMemStatsFunc(func(stats *runtime.MemStats) {
		stats.Sys = 555
	}))

	fmt.Println(mem)
	// Output: 555
}

func ExampleGetTotalThreadCountWithOptions() {
	fmt.Println(vsys.GetTotalThreadCountWithOptions(vsys.WithProcessNumGoroutineFunc(func() int { return 9 })))
	// Output: 9
}

func ExampleTotalGoroutineCountWithOptions() {
	fmt.Println(vsys.TotalGoroutineCountWithOptions(vsys.WithProcessNumGoroutineFunc(func() int { return 11 })))
	// Output: 11
}

func ExampleNewUserInfoWithOptions() {
	info := vsys.NewUserInfoWithOptions(
		vsys.WithCurrentUserFunc(func() (*user.User, error) {
			return &user.User{Username: "gopher", HomeDir: "/home/gopher"}, nil
		}),
		vsys.WithWorkingDirFunc(func() (string, error) { return "/workspace", nil }),
		vsys.WithTempDirFunc(func() string { return "/tmp/example" }),
		vsys.WithUserEnvLookup(func(key string) string {
			if key == "LANG" {
				return "en_US.UTF-8"
			}
			return ""
		}),
	)

	fmt.Println(info.GetName(), info.GetLanguage(), info.GetCountry())
	// Output: gopher en US
}

func ExampleGetUserInfoWithOptions() {
	info := vsys.GetUserInfoWithOptions(vsys.WithCurrentUserFunc(func() (*user.User, error) {
		return &user.User{Username: "get-user"}, nil
	}))

	fmt.Println(info.GetName())
	// Output: get-user
}

func ExampleSysUserInfoWithOptions() {
	info := vsys.SysUserInfoWithOptions(vsys.WithTempDirFunc(func() string { return "/tmp/sys" }))
	fmt.Println(strings.HasSuffix(info.GetTempDir(), "/tmp/sys/"))
	// Output: true
}

func ExampleSystemUserInfoWithOptions() {
	info := vsys.SystemUserInfoWithOptions(vsys.WithCurrentUserFunc(func() (*user.User, error) {
		return &user.User{Username: "system-user"}, nil
	}))

	fmt.Println(info.GetName())
	// Output: system-user
}

func ExampleDumpSystemInfoTo() {
	var buf bytes.Buffer
	vsys.DumpSystemInfoTo(&buf)

	fmt.Println(buf.Len() > 0)
	// Output: true
}

func ExampleDumpSystemInfoWithOptions() {
	var buf bytes.Buffer
	vsys.DumpSystemInfoWithOptions(
		&buf,
		vsys.WithDumpGoOptions(vsys.WithGoVersionFunc(func() string { return "go-example" })),
		vsys.WithDumpOsOptions(vsys.WithOSNameFunc(func() string { return "linux" })),
		vsys.WithDumpUserOptions(vsys.WithCurrentUserFunc(func() (*user.User, error) {
			return &user.User{Username: "dump-user"}, nil
		})),
		vsys.WithDumpHostOptions(vsys.WithHostNameFunc(func() (string, error) { return "dump-host", nil })),
		vsys.WithDumpRuntimeOptions(vsys.WithNumGoroutineFunc(func() int { return 2 })),
	)

	out := buf.String()
	fmt.Println(strings.Contains(out, "Go Version:    go-example"))
	fmt.Println(strings.Contains(out, "User Name:        dump-user"))
	fmt.Println(strings.Contains(out, "Host Name:    dump-host"))
	// Output:
	// true
	// true
	// true
}
