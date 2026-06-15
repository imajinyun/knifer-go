package vform

import validatorimpl "github.com/imajinyun/go-knifer/internal/validator"

// Option customizes validator helpers per call.
type Option = validatorimpl.Option

// WithEmailMatcher sets the matcher used by IsEmailWithOptions.
func WithEmailMatcher(matcher func(string) bool) Option {
	return validatorimpl.WithEmailMatcher(matcher)
}

// WithMobileMatcher sets the matcher used by IsMobileWithOptions.
func WithMobileMatcher(matcher func(string) bool) Option {
	return validatorimpl.WithMobileMatcher(matcher)
}

// WithIDCardMatcher sets the matcher used by IsIDCardWithOptions.
func WithIDCardMatcher(matcher func(string) bool) Option {
	return validatorimpl.WithIDCardMatcher(matcher)
}

// WithChineseMatcher sets the matcher used by IsChineseWithOptions.
func WithChineseMatcher(matcher func(string) bool) Option {
	return validatorimpl.WithChineseMatcher(matcher)
}

// WithNumberMatcher sets the matcher used by IsNumberStrWithOptions.
func WithNumberMatcher(matcher func(string) bool) Option {
	return validatorimpl.WithNumberMatcher(matcher)
}

// IsEmail reports whether s is an email address.
func IsEmail(s string) bool { return validatorimpl.IsEmail(s) }

// IsEmailWithOptions reports whether s is an email address with options.
func IsEmailWithOptions(s string, opts ...Option) bool {
	return validatorimpl.IsEmailWithOptions(s, opts...)
}

// IsMobile reports whether s is a mainland China mobile phone number.
func IsMobile(s string) bool { return validatorimpl.IsMobile(s) }

// IsMobileWithOptions reports whether s is a mobile phone number with options.
func IsMobileWithOptions(s string, opts ...Option) bool {
	return validatorimpl.IsMobileWithOptions(s, opts...)
}

// IsURL reports whether s is an absolute URL with scheme and host.
func IsURL(s string) bool { return validatorimpl.IsURL(s) }

// IsIPv4 reports whether s is an IPv4 address.
func IsIPv4(s string) bool { return validatorimpl.IsIPv4(s) }

// IsIPv6 reports whether s is an IPv6 address.
func IsIPv6(s string) bool { return validatorimpl.IsIPv6(s) }

// IsIDCard reports whether s is a valid identity card number.
func IsIDCard(s string) bool { return validatorimpl.IsIDCard(s) }

// IsIDCardWithOptions reports whether s is a valid identity card number with options.
func IsIDCardWithOptions(s string, opts ...Option) bool {
	return validatorimpl.IsIDCardWithOptions(s, opts...)
}

// IsChinese reports whether s consists only of Chinese Han characters.
func IsChinese(s string) bool { return validatorimpl.IsChinese(s) }

// IsChineseWithOptions reports whether s consists only of Chinese Han characters with options.
func IsChineseWithOptions(s string, opts ...Option) bool {
	return validatorimpl.IsChineseWithOptions(s, opts...)
}

// IsNumberStr reports whether s is a number string, including decimals and a leading minus sign.
func IsNumberStr(s string) bool { return validatorimpl.IsNumberStr(s) }

// IsNumberStrWithOptions reports whether s is a number string with options.
func IsNumberStrWithOptions(s string, opts ...Option) bool {
	return validatorimpl.IsNumberStrWithOptions(s, opts...)
}
