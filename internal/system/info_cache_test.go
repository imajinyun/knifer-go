package system

import (
	"sync"
	"testing"
)

func TestResetInfoCacheClearsSingletons(t *testing.T) {
	ResetInfoCache()
	firstHost := GetHostInfo()
	firstOS := GetOsInfo()
	firstUser := GetUserInfo()
	firstGo := GetGoInfo()
	firstRuntime := GetRuntimeInfo()
	ResetInfoCache()
	if got := GetHostInfo(); got == nil || got == firstHost {
		t.Fatalf("GetHostInfo after reset = %p, first %p", got, firstHost)
	}
	if got := GetOsInfo(); got == nil || got == firstOS {
		t.Fatalf("GetOsInfo after reset = %p, first %p", got, firstOS)
	}
	if got := GetUserInfo(); got == nil || got == firstUser {
		t.Fatalf("GetUserInfo after reset = %p, first %p", got, firstUser)
	}
	if got := GetGoInfo(); got == nil || got == firstGo {
		t.Fatalf("GetGoInfo after reset = %p, first %p", got, firstGo)
	}
	if got := GetRuntimeInfo(); got == nil || got == firstRuntime {
		t.Fatalf("GetRuntimeInfo after reset = %p, first %p", got, firstRuntime)
	}
}

func TestInfoCacheConcurrentResetAndRead(t *testing.T) {
	ResetInfoCache()

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				ResetInfoCache()
			}
		}()
	}
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				if GetHostInfo() == nil || GetOsInfo() == nil || GetUserInfo() == nil || GetGoInfo() == nil || GetRuntimeInfo() == nil {
					t.Error("cached system info should not be nil")
				}
			}
		}()
	}
	wg.Wait()
}
