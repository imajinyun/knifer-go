package rand

import (
	mathrand "math/rand"
	"testing"
)

func TestRandomLong(t *testing.T) {
	v := RandomLong()
	if v < 0 {
		t.Fatalf("RandomLong = %d, want non-negative", v)
	}
}

func TestRandomLongWithOptions(t *testing.T) {
	// with explicit source
	src := mathrand.New(mathrand.NewSource(1))
	v := RandomLongWithOptions(WithRandomSource(src))
	if v < 0 {
		t.Fatalf("RandomLongWithOptions source = %d, want non-negative", v)
	}
	// without options (uses default source)
	v2 := RandomLongWithOptions()
	if v2 < 0 {
		t.Fatalf("RandomLongWithOptions default = %d", v2)
	}
}

func TestRandomFloat(t *testing.T) {
	v := RandomFloat()
	if v < 0 || v >= 1.0 {
		t.Fatalf("RandomFloat = %f, want [0.0, 1.0)", v)
	}
}

func TestRandomFloatWithOptions(t *testing.T) {
	src := mathrand.New(mathrand.NewSource(1))
	v := RandomFloatWithOptions(WithRandomSource(src))
	if v < 0 || v >= 1.0 {
		t.Fatalf("RandomFloatWithOptions = %f, want [0.0, 1.0)", v)
	}
}

func TestRandomBool(t *testing.T) {
	seenTrue, seenFalse := false, false
	for i := 0; i < 100; i++ {
		if RandomBool() {
			seenTrue = true
		} else {
			seenFalse = true
		}
	}
	if !seenTrue || !seenFalse {
		t.Fatal("RandomBool should produce both true and false over 100 iterations")
	}
}

func TestRandomBoolWithOptions(t *testing.T) {
	src := mathrand.New(mathrand.NewSource(1))
	for i := 0; i < 20; i++ {
		_ = RandomBoolWithOptions(WithRandomSource(src))
	}
}

func TestSecureRandomBytes(t *testing.T) {
	b, err := SecureRandomBytes(16)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) != 16 {
		t.Fatalf("SecureRandomBytes len = %d, want 16", len(b))
	}
}

func TestRandomStringUpperAndFrom(t *testing.T) {
	s := RandomStringUpper(12)
	if len(s) != 12 {
		t.Fatalf("RandomStringUpper len = %d, want 12", len(s))
	}
	for _, r := range s {
		if r < '0' || (r > '9' && r < 'A') || (r > 'Z' && r < 'a') || r > 'z' {
			t.Fatalf("RandomStringUpper has invalid char %q", r)
		}
	}

	// RandomStringFrom
	s2 := RandomStringFrom("ABC", 5)
	if len(s2) != 5 {
		t.Fatalf("RandomStringFrom len = %d, want 5", len(s2))
	}
	for _, r := range s2 {
		if r != 'A' && r != 'B' && r != 'C' {
			t.Fatalf("RandomStringFrom has char out of charset: %q", r)
		}
	}
}

func TestRandomStringUpperZeroOrNegative(t *testing.T) {
	if s := RandomStringUpper(0); s != "" {
		t.Fatalf("RandomStringUpper(0) = %q, want empty", s)
	}
	if s := RandomStringFrom("", 5); s != "" {
		t.Fatalf("RandomStringFrom empty charset = %q, want empty", s)
	}
}
