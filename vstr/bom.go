package vstr

import strimpl "github.com/imajinyun/knifer-go/internal/str"

// BOMType identifies a Unicode byte order mark.
type BOMType = strimpl.BOMType

const (
	// BOMNone means no supported byte order mark was found.
	BOMNone BOMType = strimpl.BOMNone
	// BOMUTF8 identifies the UTF-8 byte order mark.
	BOMUTF8 BOMType = strimpl.BOMUTF8
	// BOMUTF16LE identifies the UTF-16 little-endian byte order mark.
	BOMUTF16LE BOMType = strimpl.BOMUTF16LE
	// BOMUTF16BE identifies the UTF-16 big-endian byte order mark.
	BOMUTF16BE BOMType = strimpl.BOMUTF16BE
	// BOMUTF32LE identifies the UTF-32 little-endian byte order mark.
	BOMUTF32LE BOMType = strimpl.BOMUTF32LE
	// BOMUTF32BE identifies the UTF-32 big-endian byte order mark.
	BOMUTF32BE BOMType = strimpl.BOMUTF32BE
)

// HasBOM returns the supported byte order mark at the beginning of data.
func HasBOM(data []byte) BOMType { return strimpl.HasBOM(data) }

// StripBOM returns data without a supported leading byte order mark.
func StripBOM(data []byte) []byte { return strimpl.StripBOM(data) }
