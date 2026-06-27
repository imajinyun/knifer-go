package vcodec

import (
	"errors"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestExtendedCodecFacade(t *testing.T) {
	if got := Base32Encode([]byte("go")); got != "M5XQ====" {
		t.Fatalf("Base32Encode = %q", got)
	}
	if got := Base58Encode([]byte("hello world")); got != "StV1DL6CwTryKyV" {
		t.Fatalf("Base58Encode = %q", got)
	}
	if got := Base62Encode([]byte("hello")); got != "7tQLFHz" {
		t.Fatalf("Base62Encode = %q", got)
	}
	morse, err := MorseEncode("go")
	if err != nil || morse != "--. ---" {
		t.Fatalf("MorseEncode = %q, %v", morse, err)
	}
	if ROT13("go") != "tb" || ROT47(ROT47("go!")) != "go!" || ROTN("az", 1) != "ba" {
		t.Fatal("rot facade failed")
	}
}

func TestExtendedCodecFacadeError(t *testing.T) {
	_, err := Base58Decode("0")
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("Base58Decode error = %v", err)
	}
}
