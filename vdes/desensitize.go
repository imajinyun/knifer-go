package vdes

import desimpl "github.com/imajinyun/go-knifer/internal/desensitize"

// Type identifies a built-in masking strategy.
type Type = desimpl.Type

const (
	// UserID masks a user ID to 0.
	UserID Type = desimpl.UserID
	// ChineseNameType masks all but the first character.
	ChineseNameType Type = desimpl.ChineseNameType
	// IDCard masks identity card numbers.
	IDCard Type = desimpl.IDCard
	// FixedPhoneType masks fixed-line phone numbers.
	FixedPhoneType Type = desimpl.FixedPhoneType
	// MobilePhoneType masks mobile phone numbers.
	MobilePhoneType Type = desimpl.MobilePhoneType
	// AddressType masks address tail content.
	AddressType Type = desimpl.AddressType
	// EmailType masks email local-part content.
	EmailType Type = desimpl.EmailType
	// PasswordType masks every password character.
	PasswordType Type = desimpl.PasswordType
	// CarLicenseType masks license plate middle content.
	CarLicenseType Type = desimpl.CarLicenseType
	// BankCardType masks bank card middle groups.
	BankCardType Type = desimpl.BankCardType
	// IPv4Type masks IPv4 host parts.
	IPv4Type Type = desimpl.IPv4Type
	// IPv6Type masks IPv6 host parts.
	IPv6Type Type = desimpl.IPv6Type
	// PassportType masks passport middle content.
	PassportType Type = desimpl.PassportType
	// CreditCodeType masks credit code middle content.
	CreditCodeType Type = desimpl.CreditCodeType
	// FirstMaskType keeps the first character only.
	FirstMaskType Type = desimpl.FirstMaskType
	// ClearToNullType clears data to nil in pointer-oriented helpers.
	ClearToNullType Type = desimpl.ClearToNullType
	// ClearToEmptyType clears data to an empty string.
	ClearToEmptyType Type = desimpl.ClearToEmptyType
)

// Desensitized masks str with the built-in strategy represented by typ.
func Desensitized(str string, typ Type) string { return desimpl.Desensitized(str, typ) }

// DesensitizedPtr masks str and returns nil for ClearToNullType.
func DesensitizedPtr(str string, typ Type) *string { return desimpl.DesensitizedPtr(str, typ) }

// Clear returns an empty string.
func Clear() string { return desimpl.Clear() }

// ClearToNil returns nil.
func ClearToNil() *string { return desimpl.ClearToNil() }

// UserIDValue returns the masked user ID value.
func UserIDValue() int64 { return desimpl.UserIDValue() }

// FirstMask keeps the first character and masks the rest.
func FirstMask(str string) string { return desimpl.FirstMask(str) }

// ChineseName masks a Chinese name by keeping only the first character.
func ChineseName(fullName string) string { return desimpl.ChineseName(fullName) }

// IDCardNum masks an identity card number while keeping front and end characters.
func IDCardNum(idCardNum string, front, end int) string {
	return desimpl.IDCardNum(idCardNum, front, end)
}

// FixedPhone masks a fixed-line phone number by keeping first four and last two characters.
func FixedPhone(num string) string { return desimpl.FixedPhone(num) }

// MobilePhone masks a mobile phone number by keeping first three and last four characters.
func MobilePhone(num string) string { return desimpl.MobilePhone(num) }

// Address masks the last sensitiveSize characters of an address.
func Address(address string, sensitiveSize int) string {
	return desimpl.Address(address, sensitiveSize)
}

// Email masks the email local-part except the first character.
func Email(email string) string { return desimpl.Email(email) }

// Password masks every password character.
func Password(password string) string { return desimpl.Password(password) }

// CarLicense masks a regular or new-energy license plate.
func CarLicense(carLicense string) string { return desimpl.CarLicense(carLicense) }

// BankCard masks a bank card number with four-character groups.
func BankCard(bankCardNo string) string { return desimpl.BankCard(bankCardNo) }

// IPv4 masks all IPv4 parts except the first one.
func IPv4(ipv4 string) string { return desimpl.IPv4(ipv4) }

// IPv6 masks all IPv6 parts except the first one.
func IPv6(ipv6 string) string { return desimpl.IPv6(ipv6) }

// Passport masks a passport number by keeping first two and last two characters.
func Passport(passport string) string { return desimpl.Passport(passport) }

// CreditCode masks a credit code by keeping first four and last four characters.
func CreditCode(code string) string { return desimpl.CreditCode(code) }

// Hide replaces runes in [start, end) with '*'. Indexes are rune based.
func Hide(str string, start, end int) string { return desimpl.Hide(str, start, end) }
