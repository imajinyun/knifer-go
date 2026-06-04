package vregex

import (
	"regexp"

	regeximpl "github.com/imajinyun/go-knifer/internal/regex"
)

// Match reports whether s contains a match for pattern.
func Match(pattern, s string) bool { return regeximpl.ReMatch(pattern, s) }

// Contains reports whether content contains a match.
func Contains(pattern, content string) bool { return regeximpl.Contains(pattern, content) }

// ContainsRe reports whether content contains a match for a compiled expression.
func ContainsRe(re *regexp.Regexp, content string) bool { return regeximpl.ContainsRe(re, content) }

// IsMatch reports whether the whole content matches pattern.
func IsMatch(pattern, content string) bool { return regeximpl.IsMatch(pattern, content) }

// IsMatchRe reports whether the whole content matches a compiled expression.
func IsMatchRe(re *regexp.Regexp, content string) bool { return regeximpl.IsMatchRe(re, content) }
