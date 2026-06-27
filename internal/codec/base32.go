package codec

import "encoding/base32"

// Base32Encoding identifies a Base32 alphabet.
type Base32Encoding string

const (
	// Base32StdEncoding is RFC 4648 standard Base32.
	Base32StdEncoding Base32Encoding = "std"
	// Base32HexEncoding is RFC 4648 extended hex Base32.
	Base32HexEncoding Base32Encoding = "hex"
)

// Base32Encode encodes bytes with standard Base32 encoding.
func Base32Encode(data []byte) string { return Base32EncodeWithEncoding(data, Base32StdEncoding) }

// Base32EncodeWithEncoding encodes bytes with the requested Base32 alphabet.
func Base32EncodeWithEncoding(data []byte, encoding Base32Encoding) string {
	return base32Encoding(encoding).EncodeToString(data)
}

// Base32Decode decodes a standard Base32 string.
func Base32Decode(s string) ([]byte, error) { return Base32DecodeWithEncoding(s, Base32StdEncoding) }

// Base32DecodeWithEncoding decodes a Base32 string with the requested alphabet.
func Base32DecodeWithEncoding(s string, encoding Base32Encoding) ([]byte, error) {
	b, err := base32Encoding(encoding).DecodeString(s)
	if err != nil {
		return nil, invalidCodecInput("decode base32", err)
	}
	return b, nil
}

func base32Encoding(encoding Base32Encoding) *base32.Encoding {
	if encoding == Base32HexEncoding {
		return base32.HexEncoding
	}
	return base32.StdEncoding
}
