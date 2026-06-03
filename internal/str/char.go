// Package str provides string and character helpers.
package str

import "unicode"

// IsBlankChar reports whether r is a blank character, including non-breaking spaces.
func IsBlankChar(r rune) bool {
	return unicode.IsSpace(r) || r == '\u00A0' || r == '\u2007' || r == '\u202F' || r == '\uFEFF'
}

// IsLetter reports whether r is a Unicode letter.
func IsLetter(r rune) bool { return unicode.IsLetter(r) }

// IsDigit reports whether r is a Unicode digit.
func IsDigit(r rune) bool { return unicode.IsDigit(r) }

// IsAscii reports whether r is an ASCII character.
func IsAscii(r rune) bool { return r < 128 }

// IsLetterOrDigit reports whether r is a Unicode letter or digit.
func IsLetterOrDigit(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) }
