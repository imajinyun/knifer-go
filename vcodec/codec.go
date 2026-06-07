package vcodec

import (
	"encoding/base64"

	codecimpl "github.com/imajinyun/go-knifer/internal/codec"
)

// Error is the codec module error type.
type Error = codecimpl.CodecError

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
