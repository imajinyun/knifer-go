package desensitize

import (
	"strings"
	"unicode"
)

// Type identifies a built-in masking strategy.
type Type int

const (
	// UserID masks a user ID to 0.
	UserID Type = iota
	// ChineseName masks all but the first character.
	ChineseNameType
	// IDCard masks identity card numbers.
	IDCard
	// FixedPhone masks fixed-line phone numbers.
	FixedPhoneType
	// MobilePhone masks mobile phone numbers.
	MobilePhoneType
	// AddressType masks address tail content.
	AddressType
	// EmailType masks email local-part content.
	EmailType
	// PasswordType masks every password character.
	PasswordType
	// CarLicenseType masks license plate middle content.
	CarLicenseType
	// BankCardType masks bank card middle groups.
	BankCardType
	// IPv4Type masks IPv4 host parts.
	IPv4Type
	// IPv6Type masks IPv6 host parts.
	IPv6Type
	// PassportType masks passport middle content.
	PassportType
	// CreditCodeType masks credit code middle content.
	CreditCodeType
	// FirstMaskType keeps the first character only.
	FirstMaskType
	// ClearToNullType clears data to nil in pointer-oriented helpers.
	ClearToNullType
	// ClearToEmptyType clears data to an empty string.
	ClearToEmptyType
)

// Desensitized masks str with the built-in strategy represented by typ.
func Desensitized(str string, typ Type) string {
	if strings.TrimSpace(str) == "" {
		return ""
	}
	switch typ {
	case UserID:
		return "0"
	case ChineseNameType:
		return ChineseName(str)
	case IDCard:
		return IDCardNum(str, 1, 2)
	case FixedPhoneType:
		return FixedPhone(str)
	case MobilePhoneType:
		return MobilePhone(str)
	case AddressType:
		return Address(str, 8)
	case EmailType:
		return Email(str)
	case PasswordType:
		return Password(str)
	case CarLicenseType:
		return CarLicense(str)
	case BankCardType:
		return BankCard(str)
	case IPv4Type:
		return IPv4(str)
	case IPv6Type:
		return IPv6(str)
	case PassportType:
		return Passport(str)
	case CreditCodeType:
		return CreditCode(str)
	case FirstMaskType:
		return FirstMask(str)
	case ClearToEmptyType, ClearToNullType:
		return ""
	default:
		return str
	}
}

// DesensitizedPtr masks str and returns nil for ClearToNullType.
func DesensitizedPtr(str string, typ Type) *string {
	if typ == ClearToNullType {
		return nil
	}
	out := Desensitized(str, typ)
	return &out
}

// Clear returns an empty string.
func Clear() string { return "" }

// ClearToNil returns nil.
func ClearToNil() *string { return nil }

// UserIDValue returns the masked user ID value.
func UserIDValue() int64 { return 0 }

// FirstMask keeps the first character and masks the rest.
func FirstMask(str string) string {
	if strings.TrimSpace(str) == "" {
		return ""
	}
	return Hide(str, 1, runeLen(str))
}

// ChineseName masks a Chinese name by keeping only the first character.
func ChineseName(fullName string) string { return FirstMask(fullName) }

// IDCardNum masks an identity card number while keeping front and end characters.
func IDCardNum(idCardNum string, front, end int) string {
	if strings.TrimSpace(idCardNum) == "" {
		return ""
	}
	length := runeLen(idCardNum)
	if front < 0 || end < 0 || front+end > length {
		return ""
	}
	return Hide(idCardNum, front, length-end)
}

// FixedPhone masks a fixed-line phone number by keeping first four and last two characters.
func FixedPhone(num string) string {
	if strings.TrimSpace(num) == "" {
		return ""
	}
	return Hide(num, 4, runeLen(num)-2)
}

// MobilePhone masks a mobile phone number by keeping first three and last four characters.
func MobilePhone(num string) string {
	if strings.TrimSpace(num) == "" {
		return ""
	}
	return Hide(num, 3, runeLen(num)-4)
}

// Address masks the last sensitiveSize characters of an address.
func Address(address string, sensitiveSize int) string {
	if strings.TrimSpace(address) == "" {
		return ""
	}
	length := runeLen(address)
	return Hide(address, length-sensitiveSize, length)
}

// Email masks the email local-part except the first character.
func Email(email string) string {
	if strings.TrimSpace(email) == "" {
		return ""
	}
	idx := strings.IndexRune(email, '@')
	if idx <= 1 {
		return email
	}
	frontRunes := runeLen(email[:idx])
	return Hide(email, 1, frontRunes)
}

// Password masks every password character.
func Password(password string) string {
	if strings.TrimSpace(password) == "" {
		return ""
	}
	return strings.Repeat("*", runeLen(password))
}

// CarLicense masks a regular or new-energy license plate.
func CarLicense(carLicense string) string {
	if strings.TrimSpace(carLicense) == "" {
		return ""
	}
	length := runeLen(carLicense)
	if length == 7 {
		return Hide(carLicense, 3, 6)
	}
	if length == 8 {
		return Hide(carLicense, 3, 7)
	}
	return carLicense
}

// BankCard masks a bank card number with four-character groups.
func BankCard(bankCardNo string) string {
	if strings.TrimSpace(bankCardNo) == "" {
		return bankCardNo
	}
	bankCardNo = cleanBlank(bankCardNo)
	length := runeLen(bankCardNo)
	if length < 9 {
		return bankCardNo
	}
	endLength := length % 4
	if endLength == 0 {
		endLength = 4
	}
	midLength := length - 4 - endLength
	runes := []rune(bankCardNo)
	var b strings.Builder
	b.WriteString(string(runes[:4]))
	for i := 0; i < midLength; i++ {
		if i%4 == 0 {
			b.WriteByte(' ')
		}
		b.WriteByte('*')
	}
	b.WriteByte(' ')
	b.WriteString(string(runes[length-endLength:]))
	return b.String()
}

// IPv4 masks all IPv4 parts except the first one.
func IPv4(ipv4 string) string {
	if idx := strings.IndexRune(ipv4, '.'); idx >= 0 {
		return ipv4[:idx] + ".*.*.*"
	}
	return ipv4 + ".*.*.*"
}

// IPv6 masks all IPv6 parts except the first one.
func IPv6(ipv6 string) string {
	if idx := strings.IndexRune(ipv6, ':'); idx >= 0 {
		return ipv6[:idx] + ":*:*:*:*:*:*:*"
	}
	return ipv6 + ":*:*:*:*:*:*:*"
}

// Passport masks a passport number by keeping first two and last two characters.
func Passport(passport string) string {
	if strings.TrimSpace(passport) == "" {
		return passport
	}
	length := runeLen(passport)
	if length <= 2 {
		return Hide(passport, 0, length)
	}
	return Hide(passport, 2, length-2)
}

// CreditCode masks a credit code by keeping first four and last four characters.
func CreditCode(code string) string {
	if strings.TrimSpace(code) == "" {
		return code
	}
	length := runeLen(code)
	if length <= 4 {
		return Hide(code, 0, length)
	}
	return Hide(code, 4, length-4)
}

// Hide replaces runes in [start, end) with '*'. Indexes are rune based.
func Hide(str string, start, end int) string {
	runes := []rune(str)
	length := len(runes)
	if start < 0 {
		start = 0
	}
	if end > length {
		end = length
	}
	if start > end {
		start = end
	}
	for i := start; i < end; i++ {
		runes[i] = '*'
	}
	return string(runes)
}

func cleanBlank(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, str)
}

func runeLen(str string) int { return len([]rune(str)) }
