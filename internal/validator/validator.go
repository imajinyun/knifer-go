// Package validator provides value validation helpers.
package validator

import (
	"regexp"

	identityimpl "github.com/imajinyun/go-knifer/internal/identity"
	netimpl "github.com/imajinyun/go-knifer/internal/net"
	urlimpl "github.com/imajinyun/go-knifer/internal/url"
)

var (
	rxEmail   = regexp.MustCompile(`^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}$`)
	rxMobile  = regexp.MustCompile(`^1[3-9]\d{9}$`)
	rxChinese = regexp.MustCompile(`^[\p{Han}]+$`)
	rxNumber  = regexp.MustCompile(`^-?\d+(\.\d+)?$`)
)

type config struct {
	email   func(string) bool
	mobile  func(string) bool
	idCard  func(string) bool
	chinese func(string) bool
	number  func(string) bool
}

// Option customizes validator helpers per call.
type Option func(*config)

// WithEmailMatcher sets the matcher used by IsEmailWithOptions.
func WithEmailMatcher(matcher func(string) bool) Option { return func(c *config) { c.email = matcher } }

// WithMobileMatcher sets the matcher used by IsMobileWithOptions.
func WithMobileMatcher(matcher func(string) bool) Option {
	return func(c *config) { c.mobile = matcher }
}

// WithIDCardMatcher sets the matcher used by IsIDCardWithOptions.
func WithIDCardMatcher(matcher func(string) bool) Option {
	return func(c *config) { c.idCard = matcher }
}

// WithChineseMatcher sets the matcher used by IsChineseWithOptions.
func WithChineseMatcher(matcher func(string) bool) Option {
	return func(c *config) { c.chinese = matcher }
}

// WithNumberMatcher sets the matcher used by IsNumberStrWithOptions.
func WithNumberMatcher(matcher func(string) bool) Option {
	return func(c *config) { c.number = matcher }
}

func applyOptions(opts []Option) config {
	cfg := config{
		email:   rxEmail.MatchString,
		mobile:  rxMobile.MatchString,
		idCard:  identityimpl.IsValidIDCard,
		chinese: rxChinese.MatchString,
		number:  rxNumber.MatchString,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.email == nil {
		cfg.email = rxEmail.MatchString
	}
	if cfg.mobile == nil {
		cfg.mobile = rxMobile.MatchString
	}
	if cfg.idCard == nil {
		cfg.idCard = identityimpl.IsValidIDCard
	}
	if cfg.chinese == nil {
		cfg.chinese = rxChinese.MatchString
	}
	if cfg.number == nil {
		cfg.number = rxNumber.MatchString
	}
	return cfg
}

// IsEmail reports whether s is an email address.
func IsEmail(s string) bool { return IsEmailWithOptions(s) }

// IsEmailWithOptions reports whether s is an email address with options.
func IsEmailWithOptions(s string, opts ...Option) bool { return applyOptions(opts).email(s) }

// IsMobile reports whether s is a mainland China mobile phone number.
func IsMobile(s string) bool { return IsMobileWithOptions(s) }

// IsMobileWithOptions reports whether s is a mobile phone number with options.
func IsMobileWithOptions(s string, opts ...Option) bool { return applyOptions(opts).mobile(s) }

// IsURL reports whether s is an absolute URL with scheme and host.
func IsURL(s string) bool { return urlimpl.IsAbsoluteURL(s) }

// IsIPv4 reports whether s is an IPv4 address.
func IsIPv4(s string) bool { return netimpl.IsIPv4(s) }

// IsIPv6 reports whether s is an IPv6 address.
func IsIPv6(s string) bool { return netimpl.IsIPv6(s) }

// IsIDCard reports whether s is a valid identity card number.
func IsIDCard(s string) bool { return IsIDCardWithOptions(s) }

// IsIDCardWithOptions reports whether s is a valid identity card number with options.
func IsIDCardWithOptions(s string, opts ...Option) bool { return applyOptions(opts).idCard(s) }

// IsChinese reports whether s consists only of Chinese Han characters.
func IsChinese(s string) bool { return IsChineseWithOptions(s) }

// IsChineseWithOptions reports whether s consists only of Chinese Han characters with options.
func IsChineseWithOptions(s string, opts ...Option) bool {
	return s != "" && applyOptions(opts).chinese(s)
}

// IsNumberStr reports whether s is a number string, including decimals and a leading minus sign.
func IsNumberStr(s string) bool { return IsNumberStrWithOptions(s) }

// IsNumberStrWithOptions reports whether s is a number string with options.
func IsNumberStrWithOptions(s string, opts ...Option) bool { return applyOptions(opts).number(s) }
