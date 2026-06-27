package str

import "bytes"

// BOMType identifies a Unicode byte order mark.
type BOMType string

const (
	// BOMNone means no supported byte order mark was found.
	BOMNone BOMType = ""
	// BOMUTF8 identifies the UTF-8 byte order mark.
	BOMUTF8 BOMType = "UTF-8"
	// BOMUTF16LE identifies the UTF-16 little-endian byte order mark.
	BOMUTF16LE BOMType = "UTF-16LE"
	// BOMUTF16BE identifies the UTF-16 big-endian byte order mark.
	BOMUTF16BE BOMType = "UTF-16BE"
	// BOMUTF32LE identifies the UTF-32 little-endian byte order mark.
	BOMUTF32LE BOMType = "UTF-32LE"
	// BOMUTF32BE identifies the UTF-32 big-endian byte order mark.
	BOMUTF32BE BOMType = "UTF-32BE"
)

var bomPrefixes = []struct {
	typ    BOMType
	prefix []byte
}{
	{BOMUTF32LE, []byte{0xFF, 0xFE, 0x00, 0x00}},
	{BOMUTF32BE, []byte{0x00, 0x00, 0xFE, 0xFF}},
	{BOMUTF8, []byte{0xEF, 0xBB, 0xBF}},
	{BOMUTF16LE, []byte{0xFF, 0xFE}},
	{BOMUTF16BE, []byte{0xFE, 0xFF}},
}

// HasBOM returns the supported byte order mark at the beginning of data.
func HasBOM(data []byte) BOMType {
	for _, bom := range bomPrefixes {
		if bytes.HasPrefix(data, bom.prefix) {
			return bom.typ
		}
	}
	return BOMNone
}

// StripBOM returns data without a supported leading byte order mark.
func StripBOM(data []byte) []byte {
	for _, bom := range bomPrefixes {
		if bytes.HasPrefix(data, bom.prefix) {
			return append([]byte(nil), data[len(bom.prefix):]...)
		}
	}
	return append([]byte(nil), data...)
}
