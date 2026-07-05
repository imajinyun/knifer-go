package rand

import (
	mathrand "math/rand"
	"sync"
	"testing"
)

func TestSetSeedMakesPseudoRandomDeterministic(t *testing.T) {
	ResetDefaultRandomSource()
	t.Cleanup(ResetDefaultRandomSource)
	SetSeed(42)
	firstInt := RandomInt(1000)
	firstString := RandomString(8)
	SetSeed(42)
	if got := RandomInt(1000); got != firstInt {
		t.Fatalf("RandomInt after SetSeed = %d, want %d", got, firstInt)
	}
	if got := RandomString(8); got != firstString {
		t.Fatalf("RandomString after SetSeed = %q, want %q", got, firstString)
	}
}

func TestDefaultRandomSourceProviderCanBeConfiguredAndReset(t *testing.T) {
	ResetDefaultRandomSource()
	t.Cleanup(ResetDefaultRandomSource)

	ConfigureDefaultRandomSourceProvider(func() *mathrand.Rand {
		return mathrand.New(mathrand.NewSource(7))
	})
	first := RandomInt(1000)
	second := RandomInt(1000)
	ConfigureDefaultRandomSourceProvider(func() *mathrand.Rand {
		return mathrand.New(mathrand.NewSource(7))
	})
	if got := RandomInt(1000); got != first {
		t.Fatalf("RandomInt after provider reset = %d, want %d", got, first)
	}
	if got := RandomInt(1000); got != second {
		t.Fatalf("second RandomInt after provider reset = %d, want %d", got, second)
	}

	ConfigureDefaultRandomSourceProvider(func() *mathrand.Rand { return nil })
	if got := RandomInt(10); got < 0 || got >= 10 {
		t.Fatalf("nil provider fallback RandomInt = %d, want [0,10)", got)
	}

	ResetDefaultRandomSource()
	if got := RandomInt(10); got < 0 || got >= 10 {
		t.Fatalf("RandomInt after reset = %d, want [0,10)", got)
	}
}

func TestDefaultRandomSourceConcurrentConfigureAndUse(t *testing.T) {
	ResetDefaultRandomSource()
	t.Cleanup(ResetDefaultRandomSource)

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		seed := int64(i + 1)
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				ConfigureDefaultRandomSourceProvider(func() *mathrand.Rand {
					return mathrand.New(mathrand.NewSource(seed))
				})
				if got := RandomInt(32); got < 0 || got >= 32 {
					t.Errorf("RandomInt = %d, want [0,32)", got)
				}
			}
		}()
	}
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				ResetDefaultRandomSource()
				_ = RandomLong()
				_ = RandomFloat()
			}
		}()
	}
	wg.Wait()
}
