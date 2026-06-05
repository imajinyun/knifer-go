package vrand

import (
	mathrand "math/rand"
	"strings"
	"testing"
)

func TestRandFacade(t *testing.T) {
	for i := 0; i < 50; i++ {
		if got := IntRange(10, 20); got < 10 || got >= 20 {
			t.Fatalf("IntRange out of bounds: %d", got)
		}
	}
	if Int(0) != 0 || Long() < 0 || Float() < 0 || Float() >= 1 {
		t.Fatal("basic random helpers failed")
	}
	_ = Bool()
	if len(Bytes(8)) != 8 {
		t.Fatal("Bytes length failed")
	}
	if s := String(8); len(s) != 8 {
		t.Fatalf("String length failed: %q", s)
	}
	if s := Numbers(6); len(s) != 6 {
		t.Fatalf("Numbers length failed: %q", s)
	}
	upper := StringUpper(8)
	for _, r := range upper {
		if !strings.ContainsRune(BaseCharNumberUC, r) {
			t.Fatalf("StringUpper charset failed: %q", upper)
		}
	}
	if s := StringFrom("ab", 4); len(s) != 4 {
		t.Fatalf("StringFrom failed: %q", s)
	}
	if got := Ele([]string{"x"}); got != "x" {
		t.Fatalf("Ele failed: %q", got)
	}
}

func TestRandFacadeOptions(t *testing.T) {
	src := mathrand.New(mathrand.NewSource(1))
	if got := IntWithOptions(100, WithRandomSource(src)); got != 81 {
		t.Fatalf("IntWithOptions = %d", got)
	}
	src = mathrand.New(mathrand.NewSource(1))
	if got := IntRangeWithOptions(10, 20, WithRandomSource(src)); got != 11 {
		t.Fatalf("IntRangeWithOptions = %d", got)
	}
	if b, err := BytesWithOptions(3, WithRandomReader(strings.NewReader("abc")), WithStrictCryptoRandom()); err != nil || string(b) != "abc" {
		t.Fatalf("BytesWithOptions = %q, %v", b, err)
	}
	src = mathrand.New(mathrand.NewSource(1))
	if got := StringFromWithOptions("abc", 5, WithRandomSource(src)); got != "caccb" {
		t.Fatalf("StringFromWithOptions = %q", got)
	}
	src = mathrand.New(mathrand.NewSource(1))
	if got := EleWithOptions([]string{"a", "b", "c"}, WithRandomSource(src)); got != "c" {
		t.Fatalf("EleWithOptions = %q", got)
	}
	_ = LongWithOptions(WithRandomSource(mathrand.New(mathrand.NewSource(1))))
	_ = FloatWithOptions(WithRandomSource(mathrand.New(mathrand.NewSource(1))))
	_ = BoolWithOptions(WithRandomSource(mathrand.New(mathrand.NewSource(1))))
	_ = StringWithOptions(3, WithRandomSource(mathrand.New(mathrand.NewSource(1))))
	_ = NumbersWithOptions(3, WithRandomSource(mathrand.New(mathrand.NewSource(1))))
	_ = StringUpperWithOptions(3, WithRandomSource(mathrand.New(mathrand.NewSource(1))))
}
