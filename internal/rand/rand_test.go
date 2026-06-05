package rand

import (
	"strings"
	"testing"
)

// Tests cover the utility toolkit-core RandomUtilTest.

func TestRandomIntRange(t *testing.T) {
	for i := 0; i < 100; i++ {
		n := RandomIntRange(10, 20)
		if n < 10 || n >= 20 {
			t.Fatalf("RandomIntRange out of bounds: %d", n)
		}
	}
}

func TestSetSeedMakesPseudoRandomDeterministic(t *testing.T) {
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

func TestRandomString(t *testing.T) {
	s := RandomString(10)
	if len(s) != 10 {
		t.Fatalf("RandomString len: %d", len(s))
	}
	for _, r := range s {
		if !strings.ContainsRune(BaseCharNumber, r) {
			t.Fatalf("RandomString out of charset: %q", s)
		}
	}
	if len(RandomNumbers(8)) != 8 {
		t.Fatalf("RandomNumbers len wrong")
	}
}

func TestRandomBytes(t *testing.T) {
	b := RandomBytes(16)
	if len(b) != 16 {
		t.Fatalf("RandomBytes len: %d", len(b))
	}
}

func TestFillRandomBytesFallbackKeepsLength(t *testing.T) {
	buf := make([]byte, 8)
	fillRandomBytes(buf)
	if len(buf) != 8 {
		t.Fatalf("fillRandomBytes changed len: %d", len(buf))
	}
}

func TestRandomEle(t *testing.T) {
	a := []string{"x", "y", "z"}
	got := RandomEle(a)
	found := false
	for _, v := range a {
		if got == v {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("RandomEle returned non-existing: %q", got)
	}
}
