package vregex

import (
	"regexp"

	regeximpl "github.com/imajinyun/knifer-go/internal/regex"
)

// Replace replaces all matches with replacement.
func Replace(pattern, s, replacement string) string {
	return regeximpl.ReReplace(pattern, s, replacement)
}

// ReplaceWithOptions replaces all matches with replacement using options.
func ReplaceWithOptions(pattern, s, replacement string, opts ...Option) string {
	return regeximpl.ReReplaceWithOptions(pattern, s, replacement, opts...)
}

// ReplaceFirst replaces the first match.
func ReplaceFirst(pattern, content, replacement string) string {
	return regeximpl.ReplaceFirst(pattern, content, replacement)
}

// ReplaceFirstWithOptions replaces the first match with options.
func ReplaceFirstWithOptions(pattern, content, replacement string, opts ...Option) string {
	return regeximpl.ReplaceFirstWithOptions(pattern, content, replacement, opts...)
}

// ReplaceFirstRe replaces the first match of a compiled expression.
func ReplaceFirstRe(re *regexp.Regexp, content, replacement string) string {
	return regeximpl.ReplaceFirstRe(re, content, replacement)
}

// ReplaceAll replaces all matches using a template with $1, $2, ... placeholders.
func ReplaceAll(content, pattern, replacementTemplate string) string {
	return regeximpl.ReplaceAll(content, pattern, replacementTemplate)
}

// ReplaceAllWithOptions replaces all matches using a template with options.
func ReplaceAllWithOptions(content, pattern, replacementTemplate string, opts ...Option) string {
	return regeximpl.ReplaceAllWithOptions(content, pattern, replacementTemplate, opts...)
}

// ReplaceAllRe replaces all matches of a compiled expression using a template.
func ReplaceAllRe(content string, re *regexp.Regexp, replacementTemplate string) string {
	return regeximpl.ReplaceAllRe(content, re, replacementTemplate)
}

// ReplaceAllFunc replaces all matches using a custom function.
func ReplaceAllFunc(content, pattern string, replaceFunc func(MatchResult) string) string {
	return regeximpl.ReplaceAllFunc(content, pattern, replaceFunc)
}

// ReplaceAllFuncWithOptions replaces all matches using a custom function with options.
func ReplaceAllFuncWithOptions(content, pattern string, replaceFunc func(MatchResult) string, opts ...Option) string {
	return regeximpl.ReplaceAllFuncWithOptions(content, pattern, replaceFunc, opts...)
}

// ReplaceAllFuncRe replaces all matches of a compiled expression using a custom function.
func ReplaceAllFuncRe(content string, re *regexp.Regexp, replaceFunc func(MatchResult) string) string {
	return regeximpl.ReplaceAllFuncRe(content, re, replaceFunc)
}
