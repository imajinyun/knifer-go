package id

import (
	"bytes"
	"errors"
	mathrand "math/rand"
	"strings"
	"sync"
	"testing"
)

func TestSimpleUUID(t *testing.T) {
	u1 := SimpleUUID()
	u2 := SimpleUUID()
	if len(u1) != 32 || len(u2) != 32 {
		t.Fatalf("UUID length wrong")
	}
	if u1 == u2 {
		t.Fatalf("UUID collision")
	}
	// Version 4 marker: the 13th character is '4'.
	if u1[12] != '4' {
		t.Fatalf("UUID version: %s", u1)
	}
}

func TestRandomUUIDAndFastSimpleUUID(t *testing.T) {
	u := RandomUUID()
	if len(u) != 36 || strings.Count(u, "-") != 4 {
		t.Fatalf("RandomUUID format: %s", u)
	}
	s := FastSimpleUUID()
	if len(s) != 32 || strings.Contains(s, "-") || s[12] != '4' {
		t.Fatalf("FastSimpleUUID format: %s", s)
	}
}

func TestFastUUID(t *testing.T) {
	u := FastUUID()
	if len(u) != 36 || strings.Count(u, "-") != 4 {
		t.Fatalf("FastUUID format: %s", u)
	}
}

func TestDefaultFallbackRandomSourceProviderCanBeConfiguredAndReset(t *testing.T) {
	ResetDefaultFallbackRandomSource()
	t.Cleanup(ResetDefaultFallbackRandomSource)

	ConfigureDefaultFallbackRandomSourceProvider(func() *mathrand.Rand {
		return mathrand.New(mathrand.NewSource(11))
	})
	first := SimpleUUIDWithOptions(WithRandomReader(errReader{}))
	second := SimpleUUIDWithOptions(WithRandomReader(errReader{}))
	ConfigureDefaultFallbackRandomSourceProvider(func() *mathrand.Rand {
		return mathrand.New(mathrand.NewSource(11))
	})
	if got := SimpleUUIDWithOptions(WithRandomReader(errReader{})); got != first {
		t.Fatalf("SimpleUUIDWithOptions after provider reset = %s, want %s", got, first)
	}
	if got := SimpleUUIDWithOptions(WithRandomReader(errReader{})); got != second {
		t.Fatalf("second SimpleUUIDWithOptions after provider reset = %s, want %s", got, second)
	}

	SetFallbackRandomSeed(12)
	seeded := SimpleUUIDWithOptions(WithRandomReader(errReader{}))
	SetFallbackRandomSeed(12)
	if got := SimpleUUIDWithOptions(WithRandomReader(errReader{})); got != seeded {
		t.Fatalf("SimpleUUIDWithOptions after seed reset = %s, want %s", got, seeded)
	}

	ConfigureDefaultFallbackRandomSourceProvider(func() *mathrand.Rand { return nil })
	if got := SimpleUUIDWithOptions(WithRandomReader(errReader{})); len(got) != 32 || got[12] != '4' {
		t.Fatalf("nil provider fallback UUID = %s", got)
	}
}

func TestDefaultFallbackRandomSourceConcurrentConfigureAndUse(t *testing.T) {
	ResetDefaultFallbackRandomSource()
	t.Cleanup(ResetDefaultFallbackRandomSource)

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		seed := int64(i + 1)
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				ConfigureDefaultFallbackRandomSourceProvider(func() *mathrand.Rand {
					return mathrand.New(mathrand.NewSource(seed))
				})
				if got := SimpleUUIDWithOptions(WithRandomReader(errReader{})); len(got) != 32 || got[12] != '4' {
					t.Errorf("fallback UUID = %q, want v4 simple UUID", got)
				}
			}
		}()
	}
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				ResetDefaultFallbackRandomSource()
				_ = ObjectIdWithOptions(WithObjectIDRandomReader(errReader{}))
				_ = NanoIdWithOptions(WithNanoIDRandomReader(errReader{}))
			}
		}()
	}
	wg.Wait()
}

func TestUUIDOptions(t *testing.T) {
	reader := bytes.NewReader(bytes.Repeat([]byte{0x11}, 32))
	u := SimpleUUIDWithOptions(WithRandomReader(reader))
	if len(u) != 32 || u[12] != '4' || u[16] != '9' {
		t.Fatalf("SimpleUUIDWithOptions format: %s", u)
	}
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("boom") }
