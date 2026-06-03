// Package validator provides value validation helpers.
package validator

import (
	"regexp"

	netimpl "github.com/imajinyun/go-knifer/internal/net"
	urlimpl "github.com/imajinyun/go-knifer/internal/url"
)

var (
	rxEmail   = regexp.MustCompile(`^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}$`)
	rxMobile  = regexp.MustCompile(`^1[3-9]\d{9}$`)
	rxChinese = regexp.MustCompile(`^[\p{Han}]+$`)
	rxNumber  = regexp.MustCompile(`^-?\d+(\.\d+)?$`)
)

// IsEmail reports whether s is an email address.
func IsEmail(s string) bool { return rxEmail.MatchString(s) }

// IsMobile reports whether s is a mainland China mobile phone number.
func IsMobile(s string) bool { return rxMobile.MatchString(s) }

// IsURL reports whether s is an absolute URL with scheme and host.
func IsURL(s string) bool { return urlimpl.IsAbsoluteURL(s) }

// IsIPv4 reports whether s is an IPv4 address.
func IsIPv4(s string) bool { return netimpl.IsIPv4(s) }

// IsChinese reports whether s consists only of Chinese Han characters.
func IsChinese(s string) bool { return s != "" && rxChinese.MatchString(s) }

// IsNumberStr reports whether s is a number string, including decimals and a leading minus sign.
func IsNumberStr(s string) bool { return rxNumber.MatchString(s) }
