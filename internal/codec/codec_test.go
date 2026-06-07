package codec

import (
	"encoding/base64"
	"errors"
	"testing"

	knifer "github.com/imajinyun/go-knifer"
)

func TestBase64(t *testing.T) {
	src := "Hello, 世界"
	enc := Base64EncodeStr(src)
	dec, err := Base64DecodeStr(enc)
	if err != nil {
		t.Fatalf("Base64 decode err: %v", err)
	}
	if dec != src {
		t.Fatalf("Base64 mismatch: %q", dec)
	}
}

func TestBase64URL(t *testing.T) {
	data := []byte{0xfa, 0xfb, 0xfc, 0xfd, 0xfe, 0xff}
	enc := Base64URLEncode(data)
	dec, err := Base64URLDecode(enc)
	if err != nil {
		t.Fatalf("Base64URL decode err: %v", err)
	}
	if string(dec) != string(data) {
		t.Fatalf("Base64URL mismatch")
	}
}

func TestBase64CustomEncoding(t *testing.T) {
	data := []byte("custom?")
	enc := Base64EncodeWithEncoding(data, base64.RawURLEncoding)
	if enc != base64.RawURLEncoding.EncodeToString(data) {
		t.Fatalf("Base64EncodeWithEncoding = %q", enc)
	}
	dec, err := Base64DecodeWithEncoding(enc, base64.RawURLEncoding)
	if err != nil || string(dec) != string(data) {
		t.Fatalf("Base64DecodeWithEncoding = %q, %v", dec, err)
	}
	if raw := Base64RawURLEncode(data); raw != enc {
		t.Fatalf("Base64RawURLEncode = %q, want %q", raw, enc)
	}
	dec, err = Base64RawURLDecode(enc)
	if err != nil || string(dec) != string(data) {
		t.Fatalf("Base64RawURLDecode = %q, %v", dec, err)
	}
}

func TestHex(t *testing.T) {
	if HexEncodeStr("AB") != "4142" {
		t.Fatalf("HexEncode: %s", HexEncodeStr("AB"))
	}
	got, err := HexDecodeStr("4142")
	if err != nil || got != "AB" {
		t.Fatalf("HexDecode: %v %q", err, got)
	}
}

func TestCodecErrorContract(t *testing.T) {
	_, err := Base64Decode("invalid!")
	assertCodecInvalidInput(t, err)
	var corrupt base64.CorruptInputError
	if !errors.As(err, &corrupt) {
		t.Fatalf("Base64Decode should preserve corrupt input cause: %v", err)
	}

	_, err = Base64URLDecode("invalid!")
	assertCodecInvalidInput(t, err)

	_, err = HexDecode("xyz")
	assertCodecInvalidInput(t, err)

	_, err = HexDecodeStr("xyz")
	assertCodecInvalidInput(t, err)
}

func assertCodecInvalidInput(t *testing.T, err error) {
	t.Helper()
	const code = knifer.ErrCodeInvalidInput
	if err == nil {
		t.Fatalf("err = nil, want %s", code)
	}
	if !errors.Is(err, code) {
		t.Fatalf("errors.Is(%v, %s) = false", err, code)
	}
	got, ok := knifer.CodeOf(err)
	if !ok || got != code {
		t.Fatalf("CodeOf(%v) = %q, %v; want %q, true", err, got, ok, code)
	}
	var codecErr *CodecError
	if !errors.As(err, &codecErr) {
		t.Fatalf("errors.As(err, *CodecError) = false: %v", err)
	}
}
