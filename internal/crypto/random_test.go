package crypto

import (
	"bytes"
	"errors"
	"testing"
)

func TestRandomBytes(t *testing.T) {
	b, err := RandomBytes(8)
	if err != nil {
		t.Fatalf("RandomBytes() error = %v", err)
	}
	if len(b) != 8 {
		t.Fatalf("RandomBytes() len = %d", len(b))
	}
	_, err = RandomBytes(-1)
	if err == nil {
		t.Fatal("RandomBytes(-1) error = nil")
	}
}

func TestGenAESKey(t *testing.T) {
	key, err := GenAESKey(16)
	if err != nil {
		t.Fatalf("GenAESKey(16) error = %v", err)
	}
	if len(key) != 16 {
		t.Fatalf("GenAESKey(16) len = %d", len(key))
	}
	_, err = GenAESKey(15)
	if err == nil {
		t.Fatal("GenAESKey(15) error = nil")
	}
}

func TestRandomBytesWithOptions(t *testing.T) {
	reader := bytes.NewReader([]byte{1, 2, 3, 4, 5, 6})
	b, err := RandomBytesWithOptions(4, WithRandomReader(reader))
	if err != nil {
		t.Fatalf("RandomBytesWithOptions() error = %v", err)
	}
	if !bytes.Equal(b, []byte{1, 2, 3, 4}) {
		t.Fatalf("RandomBytesWithOptions() = %v", b)
	}
	key, err := GenAESKeyWithOptions(16, WithRandomReader(bytes.NewReader(bytes.Repeat([]byte{0x7f}, 16))))
	if err != nil {
		t.Fatalf("GenAESKeyWithOptions() error = %v", err)
	}
	if !bytes.Equal(key, bytes.Repeat([]byte{0x7f}, 16)) {
		t.Fatalf("GenAESKeyWithOptions() = %x", key)
	}
	if _, err := GenAESKeyWithOptions(15, WithRandomReader(bytes.NewReader(nil))); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("GenAESKeyWithOptions invalid error = %v", err)
	}
}
