package http

import (
	"html"
	"regexp"
	"strings"
)

// HTMLEscape escapes HTML, aligned with HtmlUtil.escape.
func HTMLEscape(s string) string { return html.EscapeString(s) }

// HTMLUnescape unescapes HTML, aligned with HtmlUtil.unescape.
func HTMLUnescape(s string) string { return html.UnescapeString(s) }

var (
	tagRegex     = regexp.MustCompile(`(?is)<[^>]+>`)
	commentRegex = regexp.MustCompile(`(?is)<!--.*?-->`)
)

type htmlCleanConfig struct {
	tagRe     *regexp.Regexp
	commentRe *regexp.Regexp
}

// HTMLCleanOption customizes HTML cleaning helpers per call.
type HTMLCleanOption func(*htmlCleanConfig)

// WithHTMLTagRegexp sets the tag regexp used by CleanHTMLWithOptions.
func WithHTMLTagRegexp(re *regexp.Regexp) HTMLCleanOption {
	return func(c *htmlCleanConfig) { c.tagRe = re }
}

// WithHTMLCommentRegexp sets the comment regexp used by CleanHTMLWithOptions.
func WithHTMLCommentRegexp(re *regexp.Regexp) HTMLCleanOption {
	return func(c *htmlCleanConfig) { c.commentRe = re }
}

func applyHTMLCleanOptions(opts []HTMLCleanOption) htmlCleanConfig {
	cfg := htmlCleanConfig{tagRe: tagRegex, commentRe: commentRegex}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.tagRe == nil {
		cfg.tagRe = tagRegex
	}
	if cfg.commentRe == nil {
		cfg.commentRe = commentRegex
	}
	return cfg
}

type htmlFilterConfig struct {
	compile func(string) (*regexp.Regexp, error)
}

// HTMLFilterOption customizes HTML tag filtering helpers per call.
type HTMLFilterOption func(*htmlFilterConfig)

// WithHTMLFilterCompileFunc sets the compiler used by FilterHTMLTagWithOptions.
func WithHTMLFilterCompileFunc(compile func(string) (*regexp.Regexp, error)) HTMLFilterOption {
	return func(c *htmlFilterConfig) {
		if compile != nil {
			c.compile = compile
		}
	}
}

func applyHTMLFilterOptions(opts []HTMLFilterOption) htmlFilterConfig {
	cfg := htmlFilterConfig{compile: regexp.Compile}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.compile == nil {
		cfg.compile = regexp.Compile
	}
	return cfg
}

// CleanHTML removes HTML tags and keeps plain text only, aligned with HtmlUtil.cleanHtmlTag.
func CleanHTML(s string) string {
	return CleanHTMLWithOptions(s)
}

// CleanHTMLWithOptions removes HTML tags and keeps plain text with options.
func CleanHTMLWithOptions(s string, opts ...HTMLCleanOption) string {
	cfg := applyHTMLCleanOptions(opts)
	s = cfg.commentRe.ReplaceAllString(s, "")
	s = cfg.tagRe.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}

// FilterHTMLTag removes the specified HTML tags.
func FilterHTMLTag(s string, tagNames ...string) string {
	return FilterHTMLTagWithOptions(s, tagNames, nil)
}

// FilterHTMLTagWithOptions removes the specified HTML tags with options.
func FilterHTMLTagWithOptions(s string, tagNames []string, opts ...HTMLFilterOption) string {
	cfg := applyHTMLFilterOptions(opts)
	for _, tag := range tagNames {
		t := regexp.QuoteMeta(tag)
		// Tags with content: <tag ...>...</tag>.
		if re, err := cfg.compile(`(?is)<` + t + `(\s[^>]*)?>.*?</` + t + `\s*>`); err == nil {
			s = re.ReplaceAllString(s, "")
		}
		// Self-closing or single tags: <tag ... /> or <tag>.
		if re, err := cfg.compile(`(?is)<` + t + `(\s[^>]*)?/?>`); err == nil {
			s = re.ReplaceAllString(s, "")
		}
	}
	return s
}
