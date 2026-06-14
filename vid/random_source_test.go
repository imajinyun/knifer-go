package vid

import (
	"errors"
	mathrand "math/rand"
	"testing"
)

func TestIDFacadeFallbackRandomSourceProvider(t *testing.T) {
	ResetDefaultFallbackRandomSource()
	t.Cleanup(ResetDefaultFallbackRandomSource)

	ConfigureDefaultFallbackRandomSourceProvider(func() *mathrand.Rand {
		return mathrand.New(mathrand.NewSource(13))
	})
	first := SimpleUUIDWithOptions(WithRandomReader(errReader{}))
	ConfigureDefaultFallbackRandomSourceProvider(func() *mathrand.Rand {
		return mathrand.New(mathrand.NewSource(13))
	})
	if got := SimpleUUIDWithOptions(WithRandomReader(errReader{})); got != first {
		t.Fatalf("SimpleUUIDWithOptions after provider reset = %s, want %s", got, first)
	}

	SetFallbackRandomSeed(14)
	seeded := SimpleUUIDWithOptions(WithRandomReader(errReader{}))
	SetFallbackRandomSeed(14)
	if got := SimpleUUIDWithOptions(WithRandomReader(errReader{})); got != seeded {
		t.Fatalf("SimpleUUIDWithOptions after seed reset = %s, want %s", got, seeded)
	}
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("boom") }
