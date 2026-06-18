package id

import (
	mathrand "math/rand"
	"testing"
)

func TestWithFallbackRandomSource(t *testing.T) {
	src := mathrand.New(mathrand.NewSource(1))
	opt := WithFallbackRandomSource(src)
	cfg := randomConfig{}
	opt(&cfg)
	if cfg.fallbackSource != src {
		t.Fatal("WithFallbackRandomSource did not set fallbackSource")
	}
}

func TestWithObjectIDFallbackRandomSource(t *testing.T) {
	src := mathrand.New(mathrand.NewSource(1))
	opt := WithObjectIDFallbackRandomSource(src)
	cfg := objectIDConfig{}
	opt(&cfg)
	if cfg.fallbackSource != src {
		t.Fatal("WithObjectIDFallbackRandomSource did not set fallbackSource")
	}
}

func TestWithNanoIDFallbackRandomSource(t *testing.T) {
	src := mathrand.New(mathrand.NewSource(1))
	opt := WithNanoIDFallbackRandomSource(src)
	cfg := nanoIDConfig{}
	opt(&cfg)
	if cfg.fallbackSource != src {
		t.Fatal("WithNanoIDFallbackRandomSource did not set fallbackSource")
	}
}

func TestWithSnowflakeWaitFunc(t *testing.T) {
	fn := func(lastTimestamp int64, now func() int64) int64 { return 0 }
	opt := WithSnowflakeWaitFunc(fn)
	cfg := snowflakeConfig{}
	opt(&cfg)
	if cfg.tilNextMillis == nil || !cfg.runtimeSet {
		t.Fatal("WithSnowflakeWaitFunc did not set wait func")
	}

	// nil should not overwrite
	nilOpt := WithSnowflakeWaitFunc(nil)
	cfg2 := snowflakeConfig{}
	nilOpt(&cfg2)
	if cfg2.tilNextMillis != nil {
		t.Fatal("nil WithSnowflakeWaitFunc should not set")
	}
}

func TestFastUUIDWithOptions(t *testing.T) {
	u := FastUUIDWithOptions()
	if len(u) != 36 {
		t.Fatalf("FastUUIDWithOptions length = %d, want 36", len(u))
	}
}

func TestFastSimpleUUIDWithOptions(t *testing.T) {
	u := FastSimpleUUIDWithOptions()
	if len(u) != 32 {
		t.Fatalf("FastSimpleUUIDWithOptions length = %d, want 32", len(u))
	}
}

func TestWaitNextMillis(t *testing.T) {
	calls := 0
	now := func() int64 {
		calls++
		return int64(100) + int64(calls) // increment each call to exit the loop
	}
	got := waitNextMillis(99, now)
	if got <= 99 {
		t.Fatalf("waitNextMillis = %d, want > 99", got)
	}
	if calls == 0 {
		t.Fatal("waitNextMillis should call now at least once")
	}
}

func TestWaitNextMillisImmediateReturn(t *testing.T) {
	got := waitNextMillis(50, func() int64 { return 100 })
	if got != 100 {
		t.Fatalf("waitNextMillis = %d, want 100", got)
	}
}

func TestWithSnowflakeWaitFuncNilGuard(t *testing.T) {
	opt := WithSnowflakeWaitFunc(nil)
	cfg := snowflakeConfig{}
	opt(&cfg)
	if cfg.tilNextMillis != nil || cfg.runtimeSet {
		t.Fatal("nil WithSnowflakeWaitFunc should not modify config")
	}
}
