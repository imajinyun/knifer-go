package vrand

import (
	"errors"
	"io"
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
	if b, err := BytesWithOptions(8); err != nil || len(b) != 8 {
		t.Fatalf("BytesWithOptions length failed: len=%d err=%v", len(b), err)
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

func TestSecureBytes(t *testing.T) {
	b, err := SecureBytesWithOptions(3, WithRandomReader(strings.NewReader("abc")))
	if err != nil || string(b) != "abc" {
		t.Fatalf("SecureBytesWithOptions = %q, %v", b, err)
	}

	b, err = SecureBytesWithOptions(3, WithRandomReader(strings.NewReader("x")), WithRandomSource(mathrand.New(mathrand.NewSource(1))))
	if err == nil {
		t.Fatal("SecureBytesWithOptions error = nil, want entropy error")
	}
	if len(b) != 0 {
		t.Fatalf("SecureBytesWithOptions error bytes len = %d, want 0", len(b))
	}
}

func TestRandFacadeBytesFailureBoundaries(t *testing.T) {
	readerErr := errors.New("entropy failed")
	if b, err := SecureBytesWithOptions(4, WithRandomReader(failingReader{err: readerErr})); !errors.Is(err, readerErr) || len(b) != 0 {
		t.Fatalf("SecureBytesWithOptions reader error = %v, want %v", err, readerErr)
	}
	if b, err := SecureBytesWithOptions(4, WithRandomReader(strings.NewReader("x"))); !errors.Is(err, io.ErrUnexpectedEOF) || len(b) != 0 {
		t.Fatalf("SecureBytesWithOptions short read = (%#v, %v), want no bytes and unexpected EOF", b, err)
	}

	if b, err := BytesWithOptions(4, WithRandomReader(failingReader{err: readerErr}), WithStrictCryptoRandom()); !errors.Is(err, readerErr) || len(b) != 0 {
		t.Fatalf("BytesWithOptions strict reader error = %v, want %v", err, readerErr)
	}
	if b, err := BytesWithOptions(4, WithRandomReader(strings.NewReader("xy")), WithStrictCryptoRandom()); !errors.Is(err, io.ErrUnexpectedEOF) || len(b) != 0 {
		t.Fatalf("BytesWithOptions strict short read = (%#v, %v), want no bytes and unexpected EOF", b, err)
	}
	fallback, err := BytesWithOptions(4,
		WithRandomReader(failingReader{err: readerErr}),
		WithRandomSource(mathrand.New(mathrand.NewSource(7))),
	)
	if err != nil || len(fallback) != 4 {
		t.Fatalf("BytesWithOptions fallback len=%d err=%v", len(fallback), err)
	}
}

type failingReader struct{ err error }

func (r failingReader) Read([]byte) (int, error) { return 0, r.err }

func TestRandFacadeDefaultSourceProvider(t *testing.T) {
	ResetDefaultRandomSource()
	t.Cleanup(ResetDefaultRandomSource)

	ConfigureDefaultRandomSourceProvider(func() *mathrand.Rand {
		return mathrand.New(mathrand.NewSource(9))
	})
	first := Int(1000)
	ConfigureDefaultRandomSourceProvider(func() *mathrand.Rand {
		return mathrand.New(mathrand.NewSource(9))
	})
	if got := Int(1000); got != first {
		t.Fatalf("Int after provider reset = %d, want %d", got, first)
	}
}

func TestFacadeSetSeed(t *testing.T) {
	ResetDefaultRandomSource()
	t.Cleanup(ResetDefaultRandomSource)

	// Deterministic sequence after seeding
	SetSeed(42)
	a1 := Int(10000)
	b1 := Int(10000)

	SetSeed(42)
	a2 := Int(10000)
	b2 := Int(10000)

	if a1 != a2 || b1 != b2 {
		t.Fatalf("SetSeed(42) not deterministic: (%d,%d) vs (%d,%d)", a1, b1, a2, b2)
	}
}
