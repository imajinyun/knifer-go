package url

import (
	neturl "net/url"
	"strings"
)

type encodeConfig struct {
	queryEscape func(string) string
	pathEscape  func(string) string
}

// EncodeOption customizes URL encoding helpers per call.
type EncodeOption func(*encodeConfig)

// WithQueryEscapeFunc sets the query/form escaping provider.
func WithQueryEscapeFunc(escape func(string) string) EncodeOption {
	return func(c *encodeConfig) { c.queryEscape = escape }
}

// WithPathEscapeFunc sets the path segment escaping provider.
func WithPathEscapeFunc(escape func(string) string) EncodeOption {
	return func(c *encodeConfig) { c.pathEscape = escape }
}

func applyEncodeOptions(opts []EncodeOption) encodeConfig {
	cfg := encodeConfig{queryEscape: neturl.QueryEscape, pathEscape: neturl.PathEscape}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.queryEscape == nil {
		cfg.queryEscape = neturl.QueryEscape
	}
	if cfg.pathEscape == nil {
		cfg.pathEscape = neturl.PathEscape
	}
	return cfg
}

// DecodeForPath unescapes percent-encoded path text without converting plus signs to spaces.
func DecodeForPath(s string) (string, error) { return DecodeWithOptions(s, WithPlusAsSpace(false)) }

// EncodeAll percent-encodes every non-unreserved character.
func EncodeAll(s string) string { return encodeWith(s, isUnreserved, false) }

// EncodeQuery escapes text for query/form usage. Spaces are encoded as '+'.
func EncodeQuery(s string) string { return EncodeQueryWithOptions(s) }

// EncodeQueryWithOptions escapes text for query/form usage with custom providers.
func EncodeQueryWithOptions(s string, opts ...EncodeOption) string {
	return applyEncodeOptions(opts).queryEscape(s)
}

// EncodePathSegment escapes one path segment, including slash characters.
func EncodePathSegment(s string) string { return EncodePathSegmentWithOptions(s) }

// EncodePathSegmentWithOptions escapes one path segment with custom providers.
func EncodePathSegmentWithOptions(s string, opts ...EncodeOption) string {
	return applyEncodeOptions(opts).pathEscape(s)
}

// EncodePath escapes each path segment and keeps slash separators.
func EncodePath(s string) string { return encodePathKeepSlash(s) }

// EncodeFragment escapes URL fragment text.
func EncodeFragment(s string) string { return encodeWith(s, isFragmentSafe, false) }

// FormURLEncode escapes text for application/x-www-form-urlencoded usage.
func FormURLEncode(s string) string { return FormURLEncodeWithOptions(s) }

// FormURLEncodeWithOptions escapes text for application/x-www-form-urlencoded usage with custom providers.
func FormURLEncodeWithOptions(s string, opts ...EncodeOption) string {
	return applyEncodeOptions(opts).queryEscape(s)
}

func encodeWith(s string, safe func(byte) bool, spaceAsPlus bool) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == ' ' && spaceAsPlus {
			b.WriteByte('+')
			continue
		}
		if safe(c) {
			b.WriteByte(c)
			continue
		}
		b.WriteByte('%')
		const hex = "0123456789ABCDEF"
		b.WriteByte(hex[c>>4])
		b.WriteByte(hex[c&0x0f])
	}
	return b.String()
}

func isUnreserved(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '.' || c == '_' || c == '~'
}

func isFragmentSafe(c byte) bool {
	return isUnreserved(c) || strings.ContainsRune("!$&'()*+,;=:@/?", rune(c))
}
