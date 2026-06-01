package vcodec

import codecimpl "github.com/imajinyun/go-knifer/internal/codec"

func Base64Encode(data []byte) string          { return codecimpl.Base64Encode(data) }
func Base64EncodeStr(s string) string          { return codecimpl.Base64EncodeStr(s) }
func Base64Decode(s string) ([]byte, error)    { return codecimpl.Base64Decode(s) }
func Base64DecodeStr(s string) (string, error) { return codecimpl.Base64DecodeStr(s) }
func Base64URLEncode(data []byte) string       { return codecimpl.Base64URLEncode(data) }
func Base64URLDecode(s string) ([]byte, error) { return codecimpl.Base64URLDecode(s) }
func HexEncode(data []byte) string             { return codecimpl.HexEncode(data) }
func HexEncodeStr(s string) string             { return codecimpl.HexEncodeStr(s) }
func HexDecode(s string) ([]byte, error)       { return codecimpl.HexDecode(s) }
func HexDecodeStr(s string) (string, error)    { return codecimpl.HexDecodeStr(s) }

// URLEncode escapes s so it can be safely placed inside a URL query.
// For full URL parsing, normalization, and semantic processing, use vurl.
func URLEncode(s string) string { return codecimpl.URLEncode(s) }

// URLDecode unescapes a URL query component.
// For full URL parsing, normalization, and semantic processing, use vurl.
func URLDecode(s string) (string, error) { return codecimpl.URLDecode(s) }
