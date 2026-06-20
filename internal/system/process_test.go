package system

import (
	"os"
	"testing"
)

func TestGetCurrentPID(t *testing.T) {
	if GetCurrentPID() != os.Getpid() {
		t.Errorf("PID 不一致")
	}
	if got := GetCurrentPIDWithOptions(WithPIDFunc(func() int { return 4242 })); got != 4242 {
		t.Fatalf("GetCurrentPIDWithOptions = %d", got)
	}
}

func TestTotalThreadCount(t *testing.T) {
	if GetTotalThreadCount() <= 0 {
		t.Errorf("总协程数应大于 0")
	}
	if got := GetTotalThreadCountWithOptions(WithProcessNumGoroutineFunc(func() int { return 6 })); got != 6 {
		t.Fatalf("GetTotalThreadCountWithOptions = %d", got)
	}
}

func TestProcessNilOptionsFallBackToDefaults(t *testing.T) {
	if got := GetCurrentPIDWithOptions(nil, WithPIDFunc(nil)); got != os.Getpid() {
		t.Fatalf("GetCurrentPIDWithOptions nil fallback = %d", got)
	}
	if got := GetTotalThreadCountWithOptions(nil, WithProcessNumGoroutineFunc(nil)); got <= 0 {
		t.Fatalf("GetTotalThreadCountWithOptions nil fallback = %d", got)
	}
}
