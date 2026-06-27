package vcodec

import (
	"encoding/base64"

	codecimpl "github.com/imajinyun/knifer-go/internal/codec"
)

// Error is the codec module error type.
type Error = codecimpl.CodecError

// Base32Encoding identifies a Base32 alphabet.
type Base32Encoding = codecimpl.Base32Encoding

const (
	// Base32StdEncoding is RFC 4648 standard Base32.
	Base32StdEncoding Base32Encoding = codecimpl.Base32StdEncoding
	// Base32HexEncoding is RFC 4648 extended hex Base32.
	Base32HexEncoding Base32Encoding = codecimpl.Base32HexEncoding
)

// Base58Alphabet identifies a Base58 alphabet.
type Base58Alphabet = codecimpl.Base58Alphabet

const (
	// Base58BitcoinAlphabet is the Bitcoin Base58 alphabet.
	Base58BitcoinAlphabet Base58Alphabet = codecimpl.Base58BitcoinAlphabet
	// Base58FlickrAlphabet is the Flickr Base58 alphabet.
	Base58FlickrAlphabet Base58Alphabet = codecimpl.Base58FlickrAlphabet
)

func Base64Encode(data []byte) string { return codecimpl.Base64Encode(data) }
func Base64EncodeWithEncoding(data []byte, enc *base64.Encoding) string {
	return codecimpl.Base64EncodeWithEncoding(data, enc)
}
func Base64EncodeStr(s string) string       { return codecimpl.Base64EncodeStr(s) }
func Base64Decode(s string) ([]byte, error) { return codecimpl.Base64Decode(s) }
func Base64DecodeWithEncoding(s string, enc *base64.Encoding) ([]byte, error) {
	return codecimpl.Base64DecodeWithEncoding(s, enc)
}
func Base64DecodeStr(s string) (string, error) { return codecimpl.Base64DecodeStr(s) }
func Base64URLEncode(data []byte) string       { return codecimpl.Base64URLEncode(data) }
func Base64URLDecode(s string) ([]byte, error) { return codecimpl.Base64URLDecode(s) }
func Base64RawURLEncode(data []byte) string    { return codecimpl.Base64RawURLEncode(data) }
func Base64RawURLDecode(s string) ([]byte, error) {
	return codecimpl.Base64RawURLDecode(s)
}
func HexEncode(data []byte) string          { return codecimpl.HexEncode(data) }
func HexEncodeStr(s string) string          { return codecimpl.HexEncodeStr(s) }
func HexDecode(s string) ([]byte, error)    { return codecimpl.HexDecode(s) }
func HexDecodeStr(s string) (string, error) { return codecimpl.HexDecodeStr(s) }

// Base32Encode encodes bytes with standard Base32 encoding.
func Base32Encode(data []byte) string { return codecimpl.Base32Encode(data) }

// Base32EncodeWithEncoding encodes bytes with the requested Base32 alphabet.
func Base32EncodeWithEncoding(data []byte, encoding Base32Encoding) string {
	return codecimpl.Base32EncodeWithEncoding(data, encoding)
}

// Base32Decode decodes a standard Base32 string.
func Base32Decode(s string) ([]byte, error) { return codecimpl.Base32Decode(s) }

// Base32DecodeWithEncoding decodes a Base32 string with the requested alphabet.
func Base32DecodeWithEncoding(s string, encoding Base32Encoding) ([]byte, error) {
	return codecimpl.Base32DecodeWithEncoding(s, encoding)
}

// Base58Encode encodes bytes with the Bitcoin Base58 alphabet.
func Base58Encode(data []byte) string { return codecimpl.Base58Encode(data) }

// Base58EncodeWithAlphabet encodes bytes with a supported Base58 alphabet.
func Base58EncodeWithAlphabet(data []byte, alphabet Base58Alphabet) string {
	return codecimpl.Base58EncodeWithAlphabet(data, alphabet)
}

// Base58Decode decodes a Bitcoin Base58 string.
func Base58Decode(s string) ([]byte, error) { return codecimpl.Base58Decode(s) }

// Base58DecodeWithAlphabet decodes a Base58 string with a supported alphabet.
func Base58DecodeWithAlphabet(s string, alphabet Base58Alphabet) ([]byte, error) {
	return codecimpl.Base58DecodeWithAlphabet(s, alphabet)
}

// Base62Encode encodes bytes with a URL-friendly Base62 alphabet.
func Base62Encode(data []byte) string { return codecimpl.Base62Encode(data) }

// Base62Decode decodes a Base62 string.
func Base62Decode(s string) ([]byte, error) { return codecimpl.Base62Decode(s) }

// MorseEncode encodes supported ASCII letters, digits, and punctuation as Morse code.
func MorseEncode(s string) (string, error) { return codecimpl.MorseEncode(s) }

// MorseDecode decodes Morse code using spaces between letters and "/" between words.
func MorseDecode(s string) (string, error) { return codecimpl.MorseDecode(s) }

// ROT13 applies the ROT13 substitution to ASCII letters.
func ROT13(s string) string { return codecimpl.ROT13(s) }

// ROT47 applies the ROT47 substitution to printable ASCII characters.
func ROT47(s string) string { return codecimpl.ROT47(s) }

// ROTN applies a Caesar shift to ASCII letters.
func ROTN(s string, n int) string { return codecimpl.ROTN(s, n) }
