package codec

import (
	"bytes"
	"errors"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestExtendedCodecRoundTrip(t *testing.T) {
	data := []byte{0, 0, 1, 2, 3, 255}
	tests := []struct {
		name   string
		encode func([]byte) string
		decode func(string) ([]byte, error)
	}{
		{name: "base32", encode: Base32Encode, decode: Base32Decode},
		{name: "base32 hex", encode: func(b []byte) string { return Base32EncodeWithEncoding(b, Base32HexEncoding) }, decode: func(s string) ([]byte, error) { return Base32DecodeWithEncoding(s, Base32HexEncoding) }},
		{name: "base58", encode: Base58Encode, decode: Base58Decode},
		{name: "base58 flickr", encode: func(b []byte) string { return Base58EncodeWithAlphabet(b, Base58FlickrAlphabet) }, decode: func(s string) ([]byte, error) { return Base58DecodeWithAlphabet(s, Base58FlickrAlphabet) }},
		{name: "base62", encode: Base62Encode, decode: Base62Decode},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := tt.encode(data)
			decoded, err := tt.decode(encoded)
			if err != nil {
				t.Fatalf("decode error = %v", err)
			}
			if !bytes.Equal(decoded, data) {
				t.Fatalf("round trip = %v, want %v", decoded, data)
			}
		})
	}
}

func TestExtendedCodecKnownVectors(t *testing.T) {
	if got := Base58Encode([]byte("hello world")); got != "StV1DL6CwTryKyV" {
		t.Fatalf("Base58Encode = %q", got)
	}
	if got := Base62Encode([]byte("hello")); got != "7tQLFHz" {
		t.Fatalf("Base62Encode = %q", got)
	}
	encoded, err := MorseEncode("SOS 1")
	if err != nil || encoded != "... --- ... / .----" {
		t.Fatalf("MorseEncode = %q, %v", encoded, err)
	}
	decoded, err := MorseDecode(encoded)
	if err != nil || decoded != "SOS 1" {
		t.Fatalf("MorseDecode = %q, %v", decoded, err)
	}
	if ROT13("Hello") != "Uryyb" || ROT13(ROT13("Hello")) != "Hello" {
		t.Fatal("ROT13 failed")
	}
	if ROT47(ROT47("Hello!")) != "Hello!" {
		t.Fatal("ROT47 failed")
	}
	if ROTN("abcXYZ", 2) != "cdeZAB" {
		t.Fatal("ROTN failed")
	}
}

func TestExtendedCodecInvalidInput(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{name: "base32", err: decodeErr(Base32Decode("!"))},
		{name: "base58", err: decodeErr(Base58Decode("0"))},
		{name: "base62", err: decodeErr(Base62Decode("!"))},
		{name: "morse encode", err: encodeErr(MorseEncode("你好"))},
		{name: "morse decode", err: encodeErr(MorseDecode("....-.-"))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !errors.Is(tt.err, knifer.ErrCodeInvalidInput) {
				t.Fatalf("err = %v, want invalid input", tt.err)
			}
		})
	}
}

func decodeErr(_ []byte, err error) error { return err }
func encodeErr(_ string, err error) error { return err }
