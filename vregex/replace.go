package vregex

import (
	"regexp"

	regeximpl "github.com/imajinyun/go-knifer/internal/regex"
)

// Replace replaces all matches with replacement.
func Replace(pattern, s, replacement string) string {
	return regeximpl.ReReplace(pattern, s, replacement)
}

// ReplaceFirst replaces the first match.
func ReplaceFirst(pattern, content, replacement string) string {
	return regeximpl.ReplaceFirst(pattern, content, replacement)
}

// ReplaceFirstRe replaces the first match of a compiled expression.
func ReplaceFirstRe(re *regexp.Regexp, content, replacement string) string {
	return regeximpl.ReplaceFirstRe(re, content, replacement)
}

// ReplaceAll replaces all matches using a template with $1, $2, ... placeholders.
func ReplaceAll(content, pattern, replacementTemplate string) string {
	return regeximpl.ReplaceAll(content, pattern, replacementTemplate)
}

// ReplaceAllRe replaces all matches of a compiled expression using a template.
func ReplaceAllRe(content string, re *regexp.Regexp, replacementTemplate string) string {
	return regeximpl.ReplaceAllRe(content, re, replacementTemplate)
}

// ReplaceAllFunc replaces all matches using a custom function.
func ReplaceAllFunc(content, pattern string, replaceFunc func(MatchResult) string) string {
	return regeximpl.ReplaceAllFunc(content, pattern, replaceFunc)
}

// ReplaceAllFuncRe replaces all matches of a compiled expression using a custom function.
func ReplaceAllFuncRe(content string, re *regexp.Regexp, replaceFunc func(MatchResult) string) string {
	return regeximpl.ReplaceAllFuncRe(content, re, replaceFunc)
}
