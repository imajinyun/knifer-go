package rand

import (
	"errors"
	mathrand "math/rand"
	"strings"
	"testing"
)

func TestRandomBytes(t *testing.T) {
	b := RandomBytes(16)
	if len(b) != 16 {
		t.Fatalf("RandomBytes len: %d", len(b))
	}
}

func TestRandomBytesWithOptionsReaderAndStrictMode(t *testing.T) {
	b, err := RandomBytesWithOptions(4, WithRandomReader(strings.NewReader("abcd")), WithStrictCryptoRandom())
	if err != nil || string(b) != "abcd" {
		t.Fatalf("RandomBytesWithOptions = %q, %v", b, err)
	}

	_, err = RandomBytesWithOptions(4, WithRandomReader(errReader{}), WithStrictCryptoRandom())
	if err == nil {
		t.Fatal("RandomBytesWithOptions strict mode should return reader error")
	}

	b, err = RandomBytesWithOptions(4, WithRandomReader(errReader{}), WithRandomSource(mathrand.New(mathrand.NewSource(1))))
	if err != nil || len(b) != 4 {
		t.Fatalf("RandomBytesWithOptions fallback = %v, %v", b, err)
	}
}

func TestSecureRandomBytesFailClosed(t *testing.T) {
	b, err := SecureRandomBytesWithOptions(4, WithRandomReader(strings.NewReader("abcd")))
	if err != nil || string(b) != "abcd" {
		t.Fatalf("SecureRandomBytesWithOptions = %q, %v", b, err)
	}

	_, err = SecureRandomBytesWithOptions(4,
		WithRandomReader(errReader{}),
		WithRandomSource(mathrand.New(mathrand.NewSource(1))),
	)
	if err == nil {
		t.Fatal("SecureRandomBytesWithOptions error = nil, want entropy error")
	}
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("boom") }

func TestFillRandomBytesFallbackKeepsLength(t *testing.T) {
	buf := make([]byte, 8)
	fillRandomBytes(buf)
	if len(buf) != 8 {
		t.Fatalf("fillRandomBytes changed len: %d", len(buf))
	}
}
