package vhash

import (
	"testing"

	"github.com/imajinyun/go-knifer/vcrypto"
)

func TestHashFacade(t *testing.T) {
	if MD5Hex("abc") != "900150983cd24fb0d6963f7d28e17f72" {
		t.Fatal("MD5Hex failed")
	}
	if SHA1Hex("abc") != "a9993e364706816aba3e25717850c26c9cd0d89d" {
		t.Fatal("SHA1Hex failed")
	}
	if SHA256Hex("abc") != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
		t.Fatal("SHA256Hex failed")
	}
	if FnvHash("abc") == 0 || AdditiveHash("abc", 31) < 0 {
		t.Fatal("hash helpers failed")
	}
}

func TestDigestShortcutsMatchCryptoFacade_BitsUT(t *testing.T) {
	input := "boundary-doc"
	if got, want := MD5Hex(input), vcrypto.MD5Hex(input); got != want {
		t.Fatalf("MD5Hex mismatch: got %q, want %q", got, want)
	}
	if got, want := SHA1Hex(input), vcrypto.SHA1Hex(input); got != want {
		t.Fatalf("SHA1Hex mismatch: got %q, want %q", got, want)
	}
	if got, want := SHA256Hex(input), vcrypto.SHA256Hex(input); got != want {
		t.Fatalf("SHA256Hex mismatch: got %q, want %q", got, want)
	}
}
