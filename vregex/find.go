package vregex

import (
	"regexp"

	regeximpl "github.com/imajinyun/knifer-go/internal/regex"
)

// Find returns the first whole-match result.
func Find(pattern, s string) string { return regeximpl.ReFind(pattern, s) }

// FindWithOptions returns the first whole-match result with options.
func FindWithOptions(pattern, s string, opts ...Option) string {
	return regeximpl.ReFindWithOptions(pattern, s, opts...)
}

// FindAll returns all whole-match results.
func FindAll(pattern, s string) []string { return regeximpl.ReFindAll(pattern, s) }

// FindAllWithOptions returns all whole-match results with options.
func FindAllWithOptions(pattern, s string, opts ...Option) []string {
	return regeximpl.ReFindAllWithOptions(pattern, s, opts...)
}

// First calls consumer with the first match of re.
func First(re *regexp.Regexp, content string, consumer func(MatchResult)) {
	regeximpl.First(re, content, consumer)
}

// Each calls consumer for every match.
func Each(re *regexp.Regexp, content string, consumer func(MatchResult)) {
	regeximpl.Each(re, content, consumer)
}

// Count returns the number of matches.
func Count(pattern, content string) int { return regeximpl.Count(pattern, content) }

// CountWithOptions returns the number of matches with options.
func CountWithOptions(pattern, content string, opts ...Option) int {
	return regeximpl.CountWithOptions(pattern, content, opts...)
}

// CountRe returns the number of matches for a compiled expression.
func CountRe(re *regexp.Regexp, content string) int { return regeximpl.CountRe(re, content) }

// IndexOf returns the first match result.
func IndexOf(pattern, content string) *MatchResult { return regeximpl.IndexOf(pattern, content) }

// IndexOfWithOptions returns the first match result with options.
func IndexOfWithOptions(pattern, content string, opts ...Option) *MatchResult {
	return regeximpl.IndexOfWithOptions(pattern, content, opts...)
}

// IndexOfRe returns the first match result for a compiled expression.
func IndexOfRe(re *regexp.Regexp, content string) *MatchResult {
	return regeximpl.IndexOfRe(re, content)
}

// LastIndexOf returns the last match result.
func LastIndexOf(pattern, content string) *MatchResult {
	return regeximpl.LastIndexOf(pattern, content)
}

// LastIndexOfWithOptions returns the last match result with options.
func LastIndexOfWithOptions(pattern, content string, opts ...Option) *MatchResult {
	return regeximpl.LastIndexOfWithOptions(pattern, content, opts...)
}

// LastIndexOfRe returns the last match result for a compiled expression.
func LastIndexOfRe(re *regexp.Regexp, content string) *MatchResult {
	return regeximpl.LastIndexOfRe(re, content)
}

// GetFirstNumber returns the first integer in content.
func GetFirstNumber(content string) (int, bool) { return regeximpl.GetFirstNumber(content) }

// GetFirstNumberWithOptions returns the first integer in content with options.
func GetFirstNumberWithOptions(content string, opts ...Option) (int, bool) {
	return regeximpl.GetFirstNumberWithOptions(content, opts...)
}
