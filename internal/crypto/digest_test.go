package crypto

import "testing"

func TestDigestHex(t *testing.T) {
	if got := DigestHex([]byte("hello"), nil); got != "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824" {
		t.Fatalf("DigestHex() = %s", got)
	}
}

func TestSHA256AndSHA512(t *testing.T) {
	if got := SHA256([]byte("hello")); len(got) != 32 {
		t.Fatalf("SHA256() len = %d", len(got))
	}
	if got := SHA512([]byte("hello")); len(got) != 64 {
		t.Fatalf("SHA512() len = %d", len(got))
	}
	if got := SHA512Hex([]byte("hello")); len(got) != 128 {
		t.Fatalf("SHA512Hex() len = %d", len(got))
	}
}

func TestDigest(t *testing.T) {
	if got := SHA224Hex([]byte("hello")); got != "ea09ae9cc6768c50fcee903ed054556e5bfc8347907f12598aa24193" {
		t.Fatalf("SHA224Hex() = %s", got)
	}
	if got := SHA256Hex([]byte("hello")); got != "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824" {
		t.Fatalf("SHA256Hex() = %s", got)
	}
	if got := SHA384Hex([]byte("hello")); got != "59e1748777448c69de6b800d7a33bbfb9ff1b463e44354c3553bcdb9c666fa90125a3c79f90397bdf5f6a13de828684f" {
		t.Fatalf("SHA384Hex() = %s", got)
	}
}
