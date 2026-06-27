package vident

import identityimpl "github.com/imajinyun/knifer-go/internal/identity"

// CreditCodeInfo contains the parsed segments of a unified social credit code.
type CreditCodeInfo = identityimpl.CreditCodeInfo

// IsValidCreditCode reports whether code is a valid unified social credit code.
func IsValidCreditCode(code string) bool { return identityimpl.IsValidCreditCode(code) }

// ParseCreditCode validates and splits a unified social credit code.
func ParseCreditCode(code string) (CreditCodeInfo, error) {
	return identityimpl.ParseCreditCode(code)
}
