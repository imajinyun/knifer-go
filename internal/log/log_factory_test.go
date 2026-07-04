package log

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestLogFactoryCache(t *testing.T) {
	a := Get("cache.same")
	b := Get("cache.same")
	if a != b {
		t.Error("expected cached Log instance to be identical")
	}
	c := Get("cache.different")
	if a == c {
		t.Error("expected different name to produce different instance")
	}
}

func TestSetFactoryReplacesCache(t *testing.T) {
	first := Get("factory.test")
	firstName := first.GetName()

	var lock sync.Mutex
	created := 0
	SetFactory(LogFactoryFunc(func(name string) Log {
		lock.Lock()
		created++
		lock.Unlock()
		// The custom factory returns a ConsoleLog with a prefixed name for easy differentiation.
		return NewConsoleLog("custom:" + name)
	}))
	defer SetFactory(LogFactoryFunc(func(name string) Log { return NewConsoleLog(name) }))

	second := Get("factory.test")
	if first == second {
		t.Error("expected new factory to produce a new instance")
	}
	if firstName == second.GetName() {
		t.Errorf("expected different name from custom factory, both got %q", firstName)
	}
	if !strings.HasPrefix(second.GetName(), "custom:") {
		t.Errorf("expected custom factory output, got name=%q", second.GetName())
	}
	if created == 0 {
		t.Error("expected custom factory to be invoked")
	}
}

func TestGetWithOptionsBypassesGlobalFactoryAndCache(t *testing.T) {
	SetFactory(LogFactoryFunc(func(name string) Log { return NewConsoleLog("global:" + name) }))
	defer SetFactory(LogFactoryFunc(func(name string) Log { return NewConsoleLog(name) }))

	one := GetWithOptions("factory.option", WithLoggerFactory(LogFactoryFunc(func(name string) Log {
		return NewConsoleLog("local:" + name)
	})))
	two := GetWithOptions("factory.option", WithLoggerFactory(LogFactoryFunc(func(name string) Log {
		return NewConsoleLog("local:" + name)
	})))
	if one.GetName() != "local:factory.option" || two.GetName() != "local:factory.option" {
		t.Fatalf("local factory not used: %q %q", one.GetName(), two.GetName())
	}
	if one == two {
		t.Fatal("expected per-call factory lookup to bypass package cache")
	}

	cached := Get("factory.option")
	if cached.GetName() != "global:factory.option" {
		t.Fatalf("global factory/cache should remain isolated, got %q", cached.GetName())
	}
}

func TestNewIsolatedLoggerDoesNotReadGlobalFactory(t *testing.T) {
	SetFactory(LogFactoryFunc(func(name string) Log { return NewConsoleLog("global:" + name) }))
	defer SetFactory(LogFactoryFunc(func(name string) Log { return NewConsoleLog(name) }))

	log := NewIsolatedLogger("isolated")
	if log.GetName() != "isolated" {
		t.Fatalf("isolated logger leaked global factory: %q", log.GetName())
	}
}

func TestGetFactoryCanReenterLoggerCache(t *testing.T) {
	SetFactory(LogFactoryFunc(func(name string) Log {
		if name == "outer" {
			inner := Get("inner")
			if inner == nil {
				t.Fatal("inner logger is nil")
			}
		}
		return NewConsoleLog(name)
	}))
	defer SetFactory(LogFactoryFunc(func(name string) Log { return NewConsoleLog(name) }))

	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = Get("outer")
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Get deadlocked when factory reentered logger cache")
	}
}

func TestGetAndSetFactoryConcurrentAccess(t *testing.T) {
	defer SetFactory(LogFactoryFunc(func(name string) Log { return NewConsoleLog(name) }))

	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			SetFactory(LogFactoryFunc(func(name string) Log { return NewConsoleLog("set:" + name) }))
		}()
		go func(i int) {
			defer wg.Done()
			_ = Get(fmt.Sprintf("concurrent.%d", i%8))
		}(i)
	}
	wg.Wait()
}
