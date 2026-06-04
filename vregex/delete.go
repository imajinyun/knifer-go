package vregex

import (
	"regexp"

	regeximpl "github.com/imajinyun/go-knifer/internal/regex"
)

// DelFirst deletes the first match.
func DelFirst(pattern, content string) string { return regeximpl.DelFirst(pattern, content) }

// DelFirstRe deletes the first match of a compiled expression.
func DelFirstRe(re *regexp.Regexp, content string) string { return regeximpl.DelFirstRe(re, content) }

// DelLast deletes the last match.
func DelLast(pattern, content string) string { return regeximpl.DelLast(pattern, content) }

// DelLastRe deletes the last match of a compiled expression.
func DelLastRe(re *regexp.Regexp, content string) string { return regeximpl.DelLastRe(re, content) }

// DelAll deletes every match.
func DelAll(pattern, content string) string { return regeximpl.DelAll(pattern, content) }

// DelAllRe deletes every match of a compiled expression.
func DelAllRe(re *regexp.Regexp, content string) string { return regeximpl.DelAllRe(re, content) }

// DelPre deletes everything through the first match.
func DelPre(pattern, content string) string { return regeximpl.DelPre(pattern, content) }

// DelPreRe deletes everything through the first match of a compiled expression.
func DelPreRe(re *regexp.Regexp, content string) string { return regeximpl.DelPreRe(re, content) }
