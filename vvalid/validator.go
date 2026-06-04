package vvalid

import validatorimpl "github.com/imajinyun/go-knifer/internal/validator"

// IsEmail reports whether s is an email address.
func IsEmail(s string) bool { return validatorimpl.IsEmail(s) }

// IsMobile reports whether s is a mainland China mobile phone number.
func IsMobile(s string) bool { return validatorimpl.IsMobile(s) }

// IsURL reports whether s is an absolute URL with scheme and host.
func IsURL(s string) bool { return validatorimpl.IsURL(s) }

// IsIPv4 reports whether s is an IPv4 address.
func IsIPv4(s string) bool { return validatorimpl.IsIPv4(s) }

// IsIPv6 reports whether s is an IPv6 address.
func IsIPv6(s string) bool { return validatorimpl.IsIPv6(s) }

// IsIDCard reports whether s is a valid identity card number.
func IsIDCard(s string) bool { return validatorimpl.IsIDCard(s) }

// IsChinese reports whether s consists only of Chinese Han characters.
func IsChinese(s string) bool { return validatorimpl.IsChinese(s) }

// IsNumberStr reports whether s is a number string, including decimals and a leading minus sign.
func IsNumberStr(s string) bool { return validatorimpl.IsNumberStr(s) }
