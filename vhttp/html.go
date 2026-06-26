package vhttp

import (
	"regexp"

	httpx "github.com/imajinyun/knifer-go/internal/httpx/http"
)

type (
	HTMLCleanOption  = httpx.HTMLCleanOption
	HTMLFilterOption = httpx.HTMLFilterOption
)

func WithHTMLTagRegexp(re *regexp.Regexp) HTMLCleanOption {
	return httpx.WithHTMLTagRegexp(re)
}

func WithHTMLCommentRegexp(re *regexp.Regexp) HTMLCleanOption { return httpx.WithHTMLCommentRegexp(re) }

func WithHTMLFilterCompileFunc(compile func(string) (*regexp.Regexp, error)) HTMLFilterOption {
	return httpx.WithHTMLFilterCompileFunc(compile)
}

// HTMLEscape delegates to the internal httpx implementation.
func HTMLEscape(s string) string {
	return httpx.HTMLEscape(s)
}

// HTMLUnescape delegates to the internal httpx implementation.
func HTMLUnescape(s string) string {
	return httpx.HTMLUnescape(s)
}

// CleanHTML delegates to the internal httpx implementation.
func CleanHTML(s string) string {
	return httpx.CleanHTML(s)
}

func CleanHTMLWithOptions(s string, opts ...HTMLCleanOption) string {
	return httpx.CleanHTMLWithOptions(s, opts...)
}

// FilterHTMLTag delegates to the internal httpx implementation.
func FilterHTMLTag(s string, tagNames ...string) string {
	return httpx.FilterHTMLTag(s, tagNames...)
}

func FilterHTMLTagWithOptions(s string, tagNames []string, opts ...HTMLFilterOption) string {
	return httpx.FilterHTMLTagWithOptions(s, tagNames, opts...)
}
