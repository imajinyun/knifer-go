package codec

import (
	"encoding/base64"
	"encoding/hex"
)

// This file provides encoding and decoding helpers for Base64 and Hex text.

// Base64Encode encodes bytes with standard Base64 encoding.
func Base64Encode(data []byte) string { return base64.StdEncoding.EncodeToString(data) }

// Base64EncodeWithEncoding encodes bytes with enc. nil falls back to standard Base64 encoding.
func Base64EncodeWithEncoding(data []byte, enc *base64.Encoding) string {
	if enc == nil {
		enc = base64.StdEncoding
	}
	return enc.EncodeToString(data)
}

// Base64EncodeStr encodes a string with standard Base64 encoding.
func Base64EncodeStr(s string) string { return Base64Encode([]byte(s)) }

// Base64Decode decodes a standard Base64 string.
func Base64Decode(s string) ([]byte, error) {
	return Base64DecodeWithEncoding(s, base64.StdEncoding)
}

// Base64DecodeWithEncoding decodes a Base64 string with enc. nil falls back to standard Base64 encoding.
func Base64DecodeWithEncoding(s string, enc *base64.Encoding) ([]byte, error) {
	if enc == nil {
		enc = base64.StdEncoding
	}
	b, err := enc.DecodeString(s)
	if err != nil {
		return nil, invalidCodecInput("decode base64", err)
	}
	return b, nil
}

// Base64DecodeStr decodes a standard Base64 string and returns text.
func Base64DecodeStr(s string) (string, error) {
	b, err := Base64Decode(s)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Base64URLEncode encodes bytes with URL-safe Base64 encoding.
func Base64URLEncode(data []byte) string { return base64.URLEncoding.EncodeToString(data) }

// Base64RawURLEncode encodes bytes with raw URL-safe Base64 encoding without padding.
func Base64RawURLEncode(data []byte) string { return base64.RawURLEncoding.EncodeToString(data) }

// Base64URLDecode decodes a URL-safe Base64 string.
func Base64URLDecode(s string) ([]byte, error) {
	b, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return nil, invalidCodecInput("decode url-safe base64", err)
	}
	return b, nil
}

// Base64RawURLDecode decodes a raw URL-safe Base64 string without padding.
func Base64RawURLDecode(s string) ([]byte, error) {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, invalidCodecInput("decode raw url-safe base64", err)
	}
	return b, nil
}

// HexEncode encodes bytes as a lowercase hexadecimal string.
func HexEncode(data []byte) string { return hex.EncodeToString(data) }

// HexEncodeStr encodes a string as lowercase hexadecimal text.
func HexEncodeStr(s string) string { return HexEncode([]byte(s)) }

// HexDecode decodes a hexadecimal string.
func HexDecode(s string) ([]byte, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, invalidCodecInput("decode hex", err)
	}
	return b, nil
}

// HexDecodeStr decodes a hexadecimal string and returns text.
func HexDecodeStr(s string) (string, error) {
	b, err := HexDecode(s)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
