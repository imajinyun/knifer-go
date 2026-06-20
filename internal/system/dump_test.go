package system

import (
	"bytes"
	"os/user"
	"runtime"
	"strings"
	"testing"
)

func TestDumpSystemInfo(t *testing.T) {
	var buf bytes.Buffer
	DumpSystemInfoTo(&buf)
	out := buf.String()
	for _, kw := range []string{"Go Version:", "OS Name:", "User Name:", "Host Name:", "Goroutine Count:"} {
		if !strings.Contains(out, kw) {
			t.Errorf("Dump 输出缺少 %q：\n%s", kw, out)
		}
	}
}

func TestDumpSystemInfoWithDeterministicProviders(t *testing.T) {
	var buf bytes.Buffer
	DumpSystemInfoWithOptions(&buf,
		nil,
		WithDumpGoOptions(
			WithGoVersionFunc(func() string { return "go-dump" }),
			WithGoCompilerFunc(func() string { return "compiler-dump" }),
			WithGoRootFunc(func() string { return "/dump/go" }),
			WithGoOSFunc(func() string { return "linux" }),
			WithGoArchFunc(func() string { return "amd64" }),
			WithGoNumCPUFunc(func() int { return 4 }),
			WithGoNumCgoCallFunc(func() int64 { return 5 }),
		),
		WithDumpOsOptions(
			WithOSNameFunc(func() string { return "linux" }),
			WithOSArchFunc(func() string { return "amd64" }),
			WithOSVersionFunc(func() string { return "dump-os" }),
			WithOSFileSeparatorFunc(func() string { return "/" }),
			WithOSLineSeparatorFunc(func() string { return "\n" }),
			WithOSPathSeparatorFunc(func() string { return ":" }),
		),
		WithDumpUserOptions(
			WithCurrentUserFunc(func() (*user.User, error) { return &user.User{Username: "dump-user", HomeDir: "/home/dump"}, nil }),
			WithWorkingDirFunc(func() (string, error) { return "/work/dump", nil }),
			WithTempDirFunc(func() string { return "/tmp/dump" }),
		),
		WithDumpHostOptions(
			WithHostNameFunc(func() (string, error) { return "dump-host", nil }),
			WithHostAddressFunc(func() string { return "198.51.100.3" }),
		),
		WithDumpRuntimeOptions(
			WithReadMemStatsFunc(func(stats *runtime.MemStats) {
				stats.Sys = 2048
				stats.HeapSys = 1024
				stats.HeapIdle = 512
				stats.HeapInuse = 256
			}),
			WithNumGoroutineFunc(func() int { return 6 }),
		),
	)
	out := buf.String()
	for _, want := range []string{"go-dump", "compiler-dump", "dump-os", "dump-user", "dump-host", "198.51.100.3", "Goroutine Count:   6"} {
		if !strings.Contains(out, want) {
			t.Fatalf("DumpSystemInfoWithOptions missing %q in:\n%s", want, out)
		}
	}

	DumpSystemInfoWithOptions(nil, WithDumpGoOptions(WithGoVersionFunc(func() string { return "discarded" })))
}
