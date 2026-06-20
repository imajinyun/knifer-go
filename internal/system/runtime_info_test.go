package system

import (
	"runtime"
	"strings"
	"testing"
)

func TestRuntimeInfo(t *testing.T) {
	r := GetRuntimeInfo()
	if r == nil {
		t.Fatal("RuntimeInfo 不应为 nil")
	}
	if r.GetGoroutineCount() <= 0 {
		t.Errorf("Goroutine 数应大于 0")
	}
	if r.GetMaxMemory() == 0 {
		t.Errorf("MaxMemory 不应为 0")
	}
	if !strings.Contains(r.String(), "Goroutine Count:") {
		t.Errorf("RuntimeInfo.String 缺少 caption")
	}
}

func TestRuntimeInfoWithOptions(t *testing.T) {
	readCalls := 0
	r := NewRuntimeInfoWithOptions(
		WithReadMemStatsFunc(func(stats *runtime.MemStats) {
			readCalls++
			stats.Sys = 1024
			stats.HeapSys = 512
			stats.HeapIdle = 128
			stats.HeapInuse = 256
		}),
		WithNumGoroutineFunc(func() int { return 7 }),
	)
	if readCalls != 1 || r.GetMaxMemory() != 1024 || r.GetTotalMemory() != 512 || r.GetFreeMemory() != 128 || r.GetUsableMemory() != 768 || r.GetGoroutineCount() != 7 {
		t.Fatalf("NewRuntimeInfoWithOptions = %#v calls=%d", r, readCalls)
	}
	r.Refresh()
	if readCalls != 2 {
		t.Fatalf("Refresh read calls = %d", readCalls)
	}

	r = GetRuntimeInfoWithOptions(WithReadMemStatsFunc(func(stats *runtime.MemStats) { stats.Sys = 2048 }))
	if r.GetMaxMemory() != 2048 {
		t.Fatalf("GetRuntimeInfoWithOptions max = %d", r.GetMaxMemory())
	}
}

func TestRuntimeInfoSafetyBoundaries(t *testing.T) {
	r := NewRuntimeInfoWithOptions(
		WithReadMemStatsFunc(func(stats *runtime.MemStats) {
			stats.Sys = 10
			stats.HeapSys = 8
			stats.HeapIdle = 3
			stats.HeapInuse = 20
		}),
		WithNumGoroutineFunc(func() int { return 11 }),
	)
	if got := r.GetUsableMemory(); got != 0 {
		t.Fatalf("GetUsableMemory underflow guard = %d", got)
	}
	if r.GetTotalMemory() != 8 || r.GetFreeMemory() != 3 || r.GetHeapInuse() != 20 || r.GetGoroutineCount() != 11 {
		t.Fatalf("runtime getters = %#v", r.GetMemStats())
	}

	r.readMemStats = nil
	r.Refresh()
	if r.GetMaxMemory() == 0 {
		t.Fatal("Refresh with nil provider should restore runtime.ReadMemStats")
	}
	r.numGoroutine = nil
	if r.GetGoroutineCount() <= 0 {
		t.Fatal("GetGoroutineCount with nil provider should use runtime fallback")
	}
}

func TestRuntimeInfoNilOptionsFallBackToRuntimeProviders(t *testing.T) {
	r := NewRuntimeInfoWithOptions(nil, WithReadMemStatsFunc(nil), WithNumGoroutineFunc(nil))
	if r.GetMaxMemory() == 0 || r.GetGoroutineCount() <= 0 {
		t.Fatalf("nil option runtime fallback = %#v", r)
	}
}
