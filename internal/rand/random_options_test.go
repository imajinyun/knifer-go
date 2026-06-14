package rand

import (
	mathrand "math/rand"
	"testing"
)

func TestRandomWithOptionsUsesPerCallSource(t *testing.T) {
	src := mathrand.New(mathrand.NewSource(1))
	if got := RandomIntWithOptions(100, WithRandomSource(src)); got != 81 {
		t.Fatalf("RandomIntWithOptions = %d, want 81", got)
	}
	src = mathrand.New(mathrand.NewSource(1))
	if got := RandomStringFromWithOptions("abc", 5, WithRandomSource(src)); got != "caccb" {
		t.Fatalf("RandomStringFromWithOptions = %q", got)
	}
	src = mathrand.New(mathrand.NewSource(1))
	if got := RandomEleWithOptions([]string{"a", "b", "c"}, WithRandomSource(src)); got != "c" {
		t.Fatalf("RandomEleWithOptions = %q", got)
	}
}
