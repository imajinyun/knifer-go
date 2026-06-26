package vsys_test

import (
	"runtime"
	"testing"

	"github.com/imajinyun/knifer-go/vsys"
)

func TestFacadeGoInfo(t *testing.T) {
	info := vsys.SystemGoInfo()
	if info == nil {
		t.Fatal("expected non-nil go info")
	}
	info = vsys.SysGoInfoWithOptions(vsys.WithGoVersionFunc(func() string { return "go-sys" }))
	if info.GetVersion() != "go-sys" {
		t.Fatalf("SysGoInfoWithOptions version = %q", info.GetVersion())
	}
	info = vsys.SystemGoInfoWithOptions(vsys.WithGoVersionFunc(func() string { return "go-system" }))
	if info.GetVersion() != "go-system" {
		t.Fatalf("SystemGoInfoWithOptions version = %q", info.GetVersion())
	}
	info = vsys.NewGoInfoWithOptions(
		vsys.WithGoVersionFunc(func() string { return "go-facade" }),
		vsys.WithGoCompilerFunc(func() string { return "compiler-facade" }),
		vsys.WithGoRootFunc(func() string { return "/go/facade" }),
		vsys.WithGoOSFunc(func() string { return "linux" }),
		vsys.WithGoArchFunc(func() string { return "arm64" }),
		vsys.WithGoNumCPUFunc(func() int { return 8 }),
		vsys.WithGoNumCgoCallFunc(func() int64 { return 11 }),
	)
	if info.GetVersion() != "go-facade" || info.GetCompiler() != "compiler-facade" || info.GetGOROOT() != "/go/facade" || info.GetGOOS() != "linux" || info.GetGOARCH() != "arm64" || info.GetNumCPU() != 8 || info.NumCgoCalls != 11 {
		t.Fatalf("NewGoInfoWithOptions = %#v", info)
	}
}

func TestFacadeNewGoInfo(t *testing.T) {
	info := vsys.NewGoInfo()
	if info == nil {
		t.Fatal("expected non-nil go info from NewGoInfo")
	}
	if info.GetVersion() == "" {
		t.Fatal("expected non-empty Go version")
	}
}

func TestFacadeGetGoInfo(t *testing.T) {
	info := vsys.GetGoInfo()
	if info == nil {
		t.Fatal("expected non-nil go info from GetGoInfo")
	}
	infoWithOpts := vsys.GetGoInfoWithOptions(vsys.WithGoVersionFunc(func() string { return "get-go" }))
	if infoWithOpts.GetVersion() != "get-go" {
		t.Fatalf("GetGoInfoWithOptions version = %q", infoWithOpts.GetVersion())
	}
}

func TestFacadeNewRuntimeInfo(t *testing.T) {
	info := vsys.NewRuntimeInfo()
	if info == nil {
		t.Fatal("expected non-nil runtime info from NewRuntimeInfo")
	}
}

func TestFacadeGetRuntimeInfo(t *testing.T) {
	info := vsys.GetRuntimeInfo()
	if info == nil {
		t.Fatal("expected non-nil runtime info from GetRuntimeInfo")
	}
	infoWithOpts := vsys.GetRuntimeInfoWithOptions(vsys.WithReadMemStatsFunc(func(stats *runtime.MemStats) { stats.Sys = 999 }))
	if infoWithOpts.GetMaxMemory() != 999 {
		t.Fatalf("GetRuntimeInfoWithOptions max = %d", infoWithOpts.GetMaxMemory())
	}
}

func TestFacadeGetGoRuntimeOptions(t *testing.T) {
	opt1 := vsys.WithGoEnvOutputFunc(func(app string, args ...string) ([]byte, error) { return []byte("/custom/goroot"), nil })
	if opt1 == nil {
		t.Fatal("WithGoEnvOutputFunc returned nil")
	}
	opt2 := vsys.WithGoRootEnvLookupFunc(func(key string) string { return "/lookup/goroot" })
	if opt2 == nil {
		t.Fatal("WithGoRootEnvLookupFunc returned nil")
	}
	opt3 := vsys.WithOSEnvLookupFunc(func(key string) string { return "custom-os" })
	if opt3 == nil {
		t.Fatal("WithOSEnvLookupFunc returned nil")
	}
}

func TestFacadeRuntimeInfo(t *testing.T) {
	info := vsys.SystemRuntimeInfo()
	if info == nil {
		t.Fatal("expected non-nil runtime info")
	}
	info = vsys.SysRuntimeInfoWithOptions(
		vsys.WithReadMemStatsFunc(func(stats *runtime.MemStats) {
			stats.Sys = 4096
			stats.HeapSys = 1024
		}),
		vsys.WithNumGoroutineFunc(func() int { return 5 }),
	)
	if info.GetMaxMemory() != 4096 || info.GetTotalMemory() != 1024 || info.GetGoroutineCount() != 5 {
		t.Fatalf("SysRuntimeInfoWithOptions = %#v", info)
	}
	info = vsys.SystemRuntimeInfoWithOptions(vsys.WithReadMemStatsFunc(func(stats *runtime.MemStats) { stats.Sys = 1234 }))
	if info.GetMaxMemory() != 1234 {
		t.Fatalf("SystemRuntimeInfoWithOptions max = %d", info.GetMaxMemory())
	}

	info = vsys.NewRuntimeInfoWithOptions(vsys.WithReadMemStatsFunc(func(stats *runtime.MemStats) { stats.Sys = 8192 }))
	if info.GetMaxMemory() != 8192 {
		t.Fatalf("NewRuntimeInfoWithOptions max = %d", info.GetMaxMemory())
	}
}
