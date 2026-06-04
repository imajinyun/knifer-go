package vmask

import maskimpl "github.com/imajinyun/go-knifer/internal/mask"

// Type identifies a built-in masking strategy.
type Type = maskimpl.Type

const (
	// UserID masks a user ID to 0.
	UserID Type = maskimpl.UserID
	// ChineseNameType masks all but the first character.
	ChineseNameType Type = maskimpl.ChineseNameType
	// IDCard masks identity card numbers.
	IDCard Type = maskimpl.IDCard
	// FixedPhoneType masks fixed-line phone numbers.
	FixedPhoneType Type = maskimpl.FixedPhoneType
	// MobilePhoneType masks mobile phone numbers.
	MobilePhoneType Type = maskimpl.MobilePhoneType
	// AddressType masks address tail content.
	AddressType Type = maskimpl.AddressType
	// EmailType masks email local-part content.
	EmailType Type = maskimpl.EmailType
	// PasswordType masks every password character.
	PasswordType Type = maskimpl.PasswordType
	// CarLicenseType masks license plate middle content.
	CarLicenseType Type = maskimpl.CarLicenseType
	// BankCardType masks bank card middle groups.
	BankCardType Type = maskimpl.BankCardType
	// IPv4Type masks IPv4 host parts.
	IPv4Type Type = maskimpl.IPv4Type
	// IPv6Type masks IPv6 host parts.
	IPv6Type Type = maskimpl.IPv6Type
	// PassportType masks passport middle content.
	PassportType Type = maskimpl.PassportType
	// CreditCodeType masks credit code middle content.
	CreditCodeType Type = maskimpl.CreditCodeType
	// FirstMaskType keeps the first character only.
	FirstMaskType Type = maskimpl.FirstMaskType
	// ClearToNullType clears data to nil in pointer-oriented helpers.
	ClearToNullType Type = maskimpl.ClearToNullType
	// ClearToEmptyType clears data to an empty string.
	ClearToEmptyType Type = maskimpl.ClearToEmptyType
)

// Masked masks str with the built-in strategy represented by typ.
func Masked(str string, typ Type) string { return maskimpl.Masked(str, typ) }

// MaskedPtr masks str and returns nil for ClearToNullType.
func MaskedPtr(str string, typ Type) *string { return maskimpl.MaskedPtr(str, typ) }

// Clear returns an empty string.
func Clear() string { return maskimpl.Clear() }

// ClearToNil returns nil.
func ClearToNil() *string { return maskimpl.ClearToNil() }

// UserIDValue returns the masked user ID value.
func UserIDValue() int64 { return maskimpl.UserIDValue() }

// FirstMask keeps the first character and masks the rest.
func FirstMask(str string) string { return maskimpl.FirstMask(str) }

// ChineseName masks a Chinese name by keeping only the first character.
func ChineseName(fullName string) string { return maskimpl.ChineseName(fullName) }

// IDCardNum masks an identity card number while keeping front and end characters.
func IDCardNum(idCardNum string, front, end int) string {
	return maskimpl.IDCardNum(idCardNum, front, end)
}

// FixedPhone masks a fixed-line phone number by keeping first four and last two characters.
func FixedPhone(num string) string { return maskimpl.FixedPhone(num) }

// MobilePhone masks a mobile phone number by keeping first three and last four characters.
func MobilePhone(num string) string { return maskimpl.MobilePhone(num) }

// Address masks the last sensitiveSize characters of an address.
func Address(address string, sensitiveSize int) string {
	return maskimpl.Address(address, sensitiveSize)
}

// Email masks the email local-part except the first character.
func Email(email string) string { return maskimpl.Email(email) }

// Password masks every password character.
func Password(password string) string { return maskimpl.Password(password) }

// CarLicense masks a regular or new-energy license plate.
func CarLicense(carLicense string) string { return maskimpl.CarLicense(carLicense) }

// BankCard masks a bank card number with four-character groups.
func BankCard(bankCardNo string) string { return maskimpl.BankCard(bankCardNo) }

// IPv4 masks all IPv4 parts except the first one.
func IPv4(ipv4 string) string { return maskimpl.IPv4(ipv4) }

// IPv6 masks all IPv6 parts except the first one.
func IPv6(ipv6 string) string { return maskimpl.IPv6(ipv6) }

// Passport masks a passport number by keeping first two and last two characters.
func Passport(passport string) string { return maskimpl.Passport(passport) }

// CreditCode masks a credit code by keeping first four and last four characters.
func CreditCode(code string) string { return maskimpl.CreditCode(code) }

// Hide replaces runes in [start, end) with '*'. Indexes are rune based.
func Hide(str string, start, end int) string { return maskimpl.Hide(str, start, end) }
