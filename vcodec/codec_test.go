package vcodec

import (
	"encoding/base64"
	"errors"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
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
}

func TestCodecFacadeErrorContract(t *testing.T) {
	_, err := Base64Decode("invalid!")
	assertFacadeCodecCode(t, err, knifer.ErrCodeInvalidInput)

	_, err = Base64URLDecode("invalid!")
	assertFacadeCodecCode(t, err, knifer.ErrCodeInvalidInput)

	_, err = HexDecode("xyz")
	assertFacadeCodecCode(t, err, knifer.ErrCodeInvalidInput)
}

func TestCodecFacadeWithEncoding(t *testing.T) {
	data := []byte("facade")
	enc := Base64EncodeWithEncoding(data, base64.StdEncoding)
	if enc != "ZmFjYWRl" {
		t.Fatalf("Base64EncodeWithEncoding = %q", enc)
	}
	dec, err := Base64DecodeWithEncoding(enc, base64.StdEncoding)
	if err != nil || string(dec) != "facade" {
		t.Fatalf("Base64DecodeWithEncoding = %q, %v", dec, err)
	}
	raw := Base64RawURLEncode(data)
	if raw != "ZmFjYWRl" {
		t.Fatalf("Base64RawURLEncode = %q", raw)
	}
	rawDec, err := Base64RawURLDecode(raw)
	if err != nil || string(rawDec) != "facade" {
		t.Fatalf("Base64RawURLDecode = %q, %v", rawDec, err)
	}
}

func assertFacadeCodecCode(t *testing.T, err error, code knifer.ErrCode) {
	t.Helper()
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
	var codecErr *Error
	if !errors.As(err, &codecErr) {
		t.Fatalf("errors.As(err, *vcodec.Error) = false: %v", err)
	}
}
