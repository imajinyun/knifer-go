package vsys_test

import (
	"runtime"
	"testing"

	"github.com/imajinyun/go-knifer/vsys"
)

func TestFacadePID(t *testing.T) {
	pid := vsys.CurrentPID()
	if pid <= 0 {
		t.Fatalf("expected positive pid, got %d", pid)
	}
	if got := vsys.CurrentPIDWithOptions(vsys.WithPIDFunc(func() int { return 99 })); got != 99 {
		t.Fatalf("CurrentPIDWithOptions = %d", got)
	}
}

func TestFacadeGetPIDWithOptions(t *testing.T) {
	pid := vsys.GetCurrentPIDWithOptions(vsys.WithPIDFunc(func() int { return 77 }))
	if pid != 77 {
		t.Fatalf("GetCurrentPIDWithOptions = %d", pid)
	}
}

func TestFacadeMemory(t *testing.T) {
	total := vsys.TotalMemory()
	free := vsys.FreeMemory()
	max := vsys.MaxMemory()
	if total == 0 && free == 0 && max == 0 {
		t.Fatal("expected at least one memory metric to be non-zero")
	}
	opt := vsys.WithReadMemStatsFunc(func(stats *runtime.MemStats) {
		stats.Sys = 300
		stats.HeapSys = 200
		stats.HeapIdle = 100
	})
	if vsys.MaxMemoryWithOptions(opt) != 300 || vsys.TotalMemoryWithOptions(opt) != 200 || vsys.FreeMemoryWithOptions(opt) != 100 {
		t.Fatal("expected memory option providers to be used")
	}
}

func TestFacadeGetMemoryScalars(t *testing.T) {
	if got := vsys.GetTotalMemory(); got > 0 {
		t.Logf("GetTotalMemory = %d", got)
	}
	if got := vsys.GetFreeMemory(); got > 0 {
		t.Logf("GetFreeMemory = %d", got)
	}
	if got := vsys.GetMaxMemory(); got > 0 {
		t.Logf("GetMaxMemory = %d", got)
	}
	opt := vsys.WithReadMemStatsFunc(func(stats *runtime.MemStats) {
		stats.Sys = 111
		stats.HeapSys = 222
		stats.HeapIdle = 333
	})
	if vsys.GetTotalMemoryWithOptions(opt) != 222 || vsys.GetFreeMemoryWithOptions(opt) != 333 || vsys.GetMaxMemoryWithOptions(opt) != 111 {
		t.Fatal("expected Get*MemoryWithOptions to use providers")
	}
}

func TestFacadeGoroutineCount(t *testing.T) {
	count := vsys.TotalGoroutineCount()
	if count < 1 {
		t.Fatalf("expected at least 1 goroutine, got %d", count)
	}
	if got := vsys.TotalGoroutineCountWithOptions(vsys.WithProcessNumGoroutineFunc(func() int { return 12 })); got != 12 {
		t.Fatalf("TotalGoroutineCountWithOptions = %d", got)
	}
}

func TestFacadeGetThreadCount(t *testing.T) {
	count := vsys.GetTotalThreadCount()
	if count < 1 {
		t.Fatalf("expected at least 1 thread, got %d", count)
	}
	countWithOpts := vsys.GetTotalThreadCountWithOptions(vsys.WithProcessNumGoroutineFunc(func() int { return 25 }))
	if countWithOpts != 25 {
		t.Fatalf("GetTotalThreadCountWithOptions = %d", countWithOpts)
	}
}

func TestFacadeResetInfoCache(t *testing.T) {
	vsys.ResetInfoCache()
	info := vsys.SystemHostInfo()
	if info == nil {
		t.Fatal("expected non-nil host info after reset")
	}
}
