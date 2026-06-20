package vcrypto_test

import (
	"bytes"
	"crypto/sha256"
	"crypto/sha512"
	"hash"
	"testing"

	"github.com/imajinyun/go-knifer/vcrypto"
)

func TestDigestAndHMAC(t *testing.T) {
	if got := vcrypto.SHA256Hex("hello"); got != "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824" {
		t.Fatalf("SHA256Hex() = %s", got)
	}
	if got := vcrypto.SHA512Hex("hello"); got == "" {
		t.Fatal("SHA512Hex() is empty")
	}
	if got := vcrypto.HMACSHA256Hex([]byte("key"), []byte("hello")); got != "9307b3b915efb5171ff14d8cb55fbcc798c6c0ef1456d66ded1a6aa723a58b7b" {
		t.Fatalf("HMACSHA256Hex() = %s", got)
	}
	if got := vcrypto.SHA224Hex([]byte("hello")); got != "ea09ae9cc6768c50fcee903ed054556e5bfc8347907f12598aa24193" {
		t.Fatalf("SHA224Hex() = %s", got)
	}
	if !vcrypto.HMACEqual(vcrypto.HMACBytes(sha256.New, []byte("key"), []byte("hello")), vcrypto.HMACBytes(sha256.New, []byte("key"), []byte("hello"))) {
		t.Fatal("HMACEqual() returned false for identical MAC values")
	}
	if !vcrypto.ConstantTimeEqual([]byte("same"), []byte("same")) || vcrypto.ConstantTimeEqual([]byte("same"), []byte("diff")) {
		t.Fatal("ConstantTimeEqual() returned unexpected result")
	}
}

func TestAdditionalDigestAndHMAC(t *testing.T) {
	payload := []byte("hello")
	if got := vcrypto.DigestHex(payload, sha256.New); got != vcrypto.SHA256HexBytes(payload) {
		t.Fatalf("DigestHex = %q", got)
	}
	if got := vcrypto.SHA384Hex(payload); got == "" || !bytes.Equal(vcrypto.SHA384(payload), vcrypto.Digest(payload, sha512New384)) {
		t.Fatalf("SHA384 helpers returned unexpected values")
	}
	if got := vcrypto.SHA512HexBytes(payload); got != vcrypto.SHA512Hex("hello") {
		t.Fatalf("SHA512HexBytes = %q", got)
	}
	if got := vcrypto.HMACHex(sha256.New, []byte("key"), payload); got != vcrypto.HMACSHA256Hex([]byte("key"), payload) {
		t.Fatalf("HMACHex = %q", got)
	}
	if got := vcrypto.HMACHex(nil, []byte("key"), payload); got != vcrypto.HMACSHA256Hex([]byte("key"), payload) {
		t.Fatalf("HMACHex nil hash = %q", got)
	}
	if got := vcrypto.HMACBytes(nil, []byte("key"), payload); !bytes.Equal(got, vcrypto.HMACBytes(sha256.New, []byte("key"), payload)) {
		t.Fatalf("HMACBytes nil hash = %x", got)
	}
	if got := vcrypto.HMACSHA384Hex([]byte("key"), payload); got == "" {
		t.Fatal("HMACSHA384Hex is empty")
	}
	if got := vcrypto.HMACSHA512Hex([]byte("key"), payload); got == "" {
		t.Fatal("HMACSHA512Hex is empty")
	}
	if vcrypto.ConstantTimeEqual([]byte("short"), []byte("longer")) {
		t.Fatal("ConstantTimeEqual returned true for different lengths")
	}
}

func sha512New384() hash.Hash { return sha512.New384() }

func TestFacadeSHA224(t *testing.T) {
	payload := []byte("hello")
	d := vcrypto.SHA224(payload)
	if len(d) != 28 {
		t.Fatalf("SHA224 len = %d, want 28", len(d))
	}
	if vcrypto.SHA224Hex(payload) != "ea09ae9cc6768c50fcee903ed054556e5bfc8347907f12598aa24193" {
		t.Fatal("SHA224Hex mismatch")
	}
}

func TestFacadeSHA256(t *testing.T) {
	payload := []byte("hello")
	d := vcrypto.SHA256(payload)
	if len(d) != 32 {
		t.Fatalf("SHA256 len = %d, want 32", len(d))
	}
}

func TestFacadeSHA512(t *testing.T) {
	payload := []byte("hello")
	d := vcrypto.SHA512(payload)
	if len(d) != 64 {
		t.Fatalf("SHA512 len = %d, want 64", len(d))
	}
}

func BenchmarkHMACSHA256Hex(b *testing.B) {
	key := []byte("benchmark-key")
	payload := []byte("benchmark payload")
	b.ReportAllocs()
	for b.Loop() {
		_ = vcrypto.HMACSHA256Hex(key, payload)
	}
}
