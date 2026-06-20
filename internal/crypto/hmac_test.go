package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"testing"
)

func TestHMACSHA384AndSHA512(t *testing.T) {
	if got := HMACSHA384Hex([]byte("key"), []byte("hello")); got == "" {
		t.Fatal("HMACSHA384Hex() is empty")
	}
	if got := HMACSHA512Hex([]byte("key"), []byte("hello")); got == "" {
		t.Fatal("HMACSHA512Hex() is empty")
	}
}

func TestHMAC(t *testing.T) {
	if got := HMACSHA256Hex([]byte("key"), []byte("hello")); got == "" {
		t.Fatal("HMACSHA256Hex() is empty")
	}
	mac := HMACBytes(sha256.New, []byte("key"), []byte("hello"))
	if !HMACEqual(mac, HMACBytes(sha256.New, []byte("key"), []byte("hello"))) {
		t.Fatal("HMACEqual() returned false for identical MAC values")
	}
	if !ConstantTimeEqual([]byte("same"), []byte("same")) || ConstantTimeEqual([]byte("same"), []byte("diff")) {
		t.Fatal("ConstantTimeEqual() returned unexpected result")
	}
}

func TestHMACNilHashFallbacks(t *testing.T) {
	key := []byte("key")
	data := []byte("hello")
	if got, want := HMACBytes(nil, key, data), HMACBytes(sha256.New, key, data); !hmac.Equal(got, want) {
		t.Fatalf("HMACBytes nil fallback = %x, want %x", got, want)
	}
	if got, want := HMACHex(nil, key, data), HMACHex(sha256.New, key, data); got != want {
		t.Fatalf("HMACHex nil fallback = %s, want %s", got, want)
	}
	if ConstantTimeEqual([]byte("same"), []byte("same!")) {
		t.Fatal("ConstantTimeEqual should reject different lengths")
	}
}
