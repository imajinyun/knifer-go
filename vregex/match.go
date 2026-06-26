package vregex

import (
	"regexp"

	regeximpl "github.com/imajinyun/knifer-go/internal/regex"
)

// Option customizes pattern-string regex helpers per call.
type Option = regeximpl.Option

// WithCompileFunc sets the compiler used by pattern-string regex helpers.
func WithCompileFunc(compile func(string) (*regexp.Regexp, error)) Option {
	return regeximpl.WithCompileFunc(compile)
}

// WithDotAll controls whether pattern-string helpers wrap patterns with (?s:...).
func WithDotAll(dotAll bool) Option { return regeximpl.WithDotAll(dotAll) }

// WithGroupVarRegexp sets the regexp used by TemplateVarsWithOptions.
func WithGroupVarRegexp(re *regexp.Regexp) Option { return regeximpl.WithGroupVarRegexp(re) }

// WithNumbersRegexp sets the regexp used by GetFirstNumberWithOptions.
func WithNumbersRegexp(re *regexp.Regexp) Option { return regeximpl.WithNumbersRegexp(re) }

// WithNamedGroupRegexp sets the regexp used to normalize (?<name>...) groups before compiling.
func WithNamedGroupRegexp(re *regexp.Regexp) Option { return regeximpl.WithNamedGroupRegexp(re) }

// WithNamedGroupNormalizer sets the normalizer used before compiling pattern strings.
func WithNamedGroupNormalizer(normalize func(string) string) Option {
	return regeximpl.WithNamedGroupNormalizer(normalize)
}

// Match reports whether s contains a match for pattern.
func Match(pattern, s string) bool { return regeximpl.ReMatch(pattern, s) }

// MatchWithOptions reports whether s contains a match for pattern with options.
func MatchWithOptions(pattern, s string, opts ...Option) bool {
	return regeximpl.ReMatchWithOptions(pattern, s, opts...)
}

// Contains reports whether content contains a match.
func Contains(pattern, content string) bool { return regeximpl.Contains(pattern, content) }

// ContainsWithOptions reports whether content contains a match with options.
func ContainsWithOptions(pattern, content string, opts ...Option) bool {
	return regeximpl.ContainsWithOptions(pattern, content, opts...)
}

// ContainsRe reports whether content contains a match for a compiled expression.
func ContainsRe(re *regexp.Regexp, content string) bool { return regeximpl.ContainsRe(re, content) }

// IsMatch reports whether the whole content matches pattern.
func IsMatch(pattern, content string) bool { return regeximpl.IsMatch(pattern, content) }

// IsMatchWithOptions reports whether the whole content matches pattern with options.
func IsMatchWithOptions(pattern, content string, opts ...Option) bool {
	return regeximpl.IsMatchWithOptions(pattern, content, opts...)
}

// IsMatchRe reports whether the whole content matches a compiled expression.
func IsMatchRe(re *regexp.Regexp, content string) bool { return regeximpl.IsMatchRe(re, content) }
