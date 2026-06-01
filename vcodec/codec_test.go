package vcodec

import (
	"testing"

	"github.com/imajinyun/go-knifer/vurl"
)

func TestCodecFacade(t *testing.T) {
	if Base64EncodeStr("go") != "Z28=" {
		t.Fatal("Base64EncodeStr failed")
	}
	if got, err := Base64DecodeStr("Z28="); err != nil || got != "go" {
		t.Fatalf("Base64DecodeStr = %q, %v", got, err)
	}
	if got, err := Base64URLDecode(Base64URLEncode([]byte("a/b"))); err != nil || string(got) != "a/b" {
		t.Fatalf("Base64URL roundtrip = %q, %v", got, err)
	}
	if HexEncodeStr("go") != "676f" {
		t.Fatal("HexEncodeStr failed")
	}
	if got, err := HexDecodeStr("676f"); err != nil || got != "go" {
		t.Fatalf("HexDecodeStr = %q, %v", got, err)
	}
	if got, err := URLDecode(URLEncode("a b")); err != nil || got != "a b" {
		t.Fatalf("URL roundtrip = %q, %v", got, err)
	}
}

func TestURLQueryEscapingMatchesURLFacade_BitsUT(t *testing.T) {
	input := "a b+c/中文"
	encoded := URLEncode(input)
	if encoded != vurl.Encode(input) {
		t.Fatalf("URLEncode mismatch: got %q, want %q", encoded, vurl.Encode(input))
	}
	decoded, err := vurl.Decode(encoded)
	if err != nil {
		t.Fatalf("vurl.Decode returned error: %v", err)
	}
	if decoded != input {
		t.Fatalf("Decode roundtrip = %q, want %q", decoded, input)
	}
}
