package vregex

import (
	"regexp"

	regeximpl "github.com/imajinyun/go-knifer/internal/regex"
)

// DelFirst deletes the first match.
func DelFirst(pattern, content string) string { return regeximpl.DelFirst(pattern, content) }

// DelFirstWithOptions deletes the first match with options.
func DelFirstWithOptions(pattern, content string, opts ...Option) string {
	return regeximpl.DelFirstWithOptions(pattern, content, opts...)
}

// DelFirstRe deletes the first match of a compiled expression.
func DelFirstRe(re *regexp.Regexp, content string) string { return regeximpl.DelFirstRe(re, content) }

// DelLast deletes the last match.
func DelLast(pattern, content string) string { return regeximpl.DelLast(pattern, content) }

// DelLastWithOptions deletes the last match with options.
func DelLastWithOptions(pattern, content string, opts ...Option) string {
	return regeximpl.DelLastWithOptions(pattern, content, opts...)
}

// DelLastRe deletes the last match of a compiled expression.
func DelLastRe(re *regexp.Regexp, content string) string { return regeximpl.DelLastRe(re, content) }

// DelAll deletes every match.
func DelAll(pattern, content string) string { return regeximpl.DelAll(pattern, content) }

// DelAllWithOptions deletes every match with options.
func DelAllWithOptions(pattern, content string, opts ...Option) string {
	return regeximpl.DelAllWithOptions(pattern, content, opts...)
}

// DelAllRe deletes every match of a compiled expression.
func DelAllRe(re *regexp.Regexp, content string) string { return regeximpl.DelAllRe(re, content) }

// DelPre deletes everything through the first match.
func DelPre(pattern, content string) string { return regeximpl.DelPre(pattern, content) }

// DelPreWithOptions deletes everything through the first match with options.
func DelPreWithOptions(pattern, content string, opts ...Option) string {
	return regeximpl.DelPreWithOptions(pattern, content, opts...)
}

// DelPreRe deletes everything through the first match of a compiled expression.
func DelPreRe(re *regexp.Regexp, content string) string { return regeximpl.DelPreRe(re, content) }
