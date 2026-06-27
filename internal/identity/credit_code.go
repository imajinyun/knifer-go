package identity

import (
	"strings"

	knifer "github.com/imajinyun/knifer-go"
)

const (
	creditCodeLength  = 18
	creditCodeCharset = "0123456789ABCDEFGHJKLMNPQRTUWXY"
)

var creditCodeWeights = [...]int{1, 3, 9, 27, 19, 26, 16, 17, 20, 29, 25, 13, 8, 24, 10, 30, 28}

// CreditCodeInfo contains the parsed segments of a unified social credit code.
type CreditCodeInfo struct {
	Raw         string
	AdminDept   string
	OrgCategory string
	RegionCode  string
	OrgCode     string
	CheckDigit  string
}

// IsValidCreditCode reports whether code is a valid unified social credit code.
func IsValidCreditCode(code string) bool {
	_, err := ParseCreditCode(code)
	return err == nil
}

// ParseCreditCode validates and splits a unified social credit code.
func ParseCreditCode(code string) (CreditCodeInfo, error) {
	normalized := strings.ToUpper(strings.TrimSpace(code))
	if len(normalized) != creditCodeLength {
		return CreditCodeInfo{}, creditCodeInvalidInput("credit code length must be 18")
	}

	sum := 0
	for i := 0; i < creditCodeLength-1; i++ {
		idx := strings.IndexByte(creditCodeCharset, normalized[i])
		if idx < 0 {
			return CreditCodeInfo{}, creditCodeInvalidInput("credit code contains unsupported character")
		}
		sum += idx * creditCodeWeights[i]
	}
	if strings.IndexByte(creditCodeCharset, normalized[creditCodeLength-1]) < 0 {
		return CreditCodeInfo{}, creditCodeInvalidInput("credit code contains unsupported character")
	}

	check := 31 - sum%31
	if check == 31 {
		check = 0
	}
	expected := creditCodeCharset[check]
	if normalized[creditCodeLength-1] != expected {
		return CreditCodeInfo{}, creditCodeInvalidInput("credit code check digit mismatch")
	}

	return CreditCodeInfo{
		Raw:         normalized,
		AdminDept:   normalized[:1],
		OrgCategory: normalized[1:2],
		RegionCode:  normalized[2:8],
		OrgCode:     normalized[8:17],
		CheckDigit:  normalized[17:],
	}, nil
}

func creditCodeInvalidInput(msg string) error {
	return &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: msg}
}
