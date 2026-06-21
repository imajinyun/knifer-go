package rand

import (
	"bytes"
	"errors"
	"io"
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

	b, err = RandomBytesWithOptions(4, WithRandomReader(errReader{}), WithStrictCryptoRandom())
	if err == nil {
		t.Fatal("RandomBytesWithOptions strict mode should return reader error")
	}
	if len(b) != 0 {
		t.Fatalf("RandomBytesWithOptions strict error bytes len = %d, want 0", len(b))
	}

	b, err = RandomBytesWithOptions(4, WithRandomReader(strings.NewReader("xy")), WithStrictCryptoRandom())
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("RandomBytesWithOptions strict short read error = %v, want %v", err, io.ErrUnexpectedEOF)
	}
	if len(b) != 0 {
		t.Fatalf("RandomBytesWithOptions strict short read bytes len = %d, want 0", len(b))
	}

	b, err = RandomBytesWithOptions(4, WithRandomReader(errReader{}), WithRandomSource(mathrand.New(mathrand.NewSource(1))))
	if err != nil || len(b) != 4 {
		t.Fatalf("RandomBytesWithOptions fallback = %v, %v", b, err)
	}

	source := mathrand.New(mathrand.NewSource(1))
	b, err = RandomBytesWithOptions(4,
		WithRandomReader(partialErrReader{data: []byte{0xaa, 0xbb}, err: errors.New("partial entropy failure")}),
		WithRandomSource(mathrand.New(mathrand.NewSource(1))),
	)
	if err != nil {
		t.Fatalf("RandomBytesWithOptions partial fallback error = %v", err)
	}
	want := []byte{byte(source.Intn(256)), byte(source.Intn(256)), byte(source.Intn(256)), byte(source.Intn(256))}
	if !bytes.Equal(b, want) {
		t.Fatalf("RandomBytesWithOptions partial fallback = %#v, want %#v", b, want)
	}
}

func TestSecureRandomBytesFailClosed(t *testing.T) {
	b, err := SecureRandomBytesWithOptions(4, WithRandomReader(strings.NewReader("abcd")))
	if err != nil || string(b) != "abcd" {
		t.Fatalf("SecureRandomBytesWithOptions = %q, %v", b, err)
	}

	b, err = SecureRandomBytesWithOptions(4,
		WithRandomReader(errReader{}),
		WithRandomSource(mathrand.New(mathrand.NewSource(1))),
	)
	if err == nil {
		t.Fatal("SecureRandomBytesWithOptions error = nil, want entropy error")
	}
	if len(b) != 0 {
		t.Fatalf("SecureRandomBytesWithOptions error bytes len = %d, want 0", len(b))
	}

	b, err = SecureRandomBytesWithOptions(4, WithRandomReader(strings.NewReader("xy")))
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("SecureRandomBytesWithOptions short read error = %v, want %v", err, io.ErrUnexpectedEOF)
	}
	if len(b) != 0 {
		t.Fatalf("SecureRandomBytesWithOptions short read bytes len = %d, want 0", len(b))
	}
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("boom") }

type partialErrReader struct {
	data []byte
	err  error
}

func (r partialErrReader) Read(p []byte) (int, error) {
	return copy(p, r.data), r.err
}

func TestFillRandomBytesFallbackKeepsLength(t *testing.T) {
	buf := make([]byte, 8)
	fillRandomBytes(buf)
	if len(buf) != 8 {
		t.Fatalf("fillRandomBytes changed len: %d", len(buf))
	}
}
