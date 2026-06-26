package vident

import (
	"time"

	identityimpl "github.com/imajinyun/knifer-go/internal/identity"
)

// Gender identifies the gender encoded in an identity card number.
type Gender = identityimpl.Gender

const (
	// GenderUnknown means the gender is not encoded or cannot be determined.
	GenderUnknown Gender = identityimpl.GenderUnknown
	// GenderFemale represents female.
	GenderFemale Gender = identityimpl.GenderFemale
	// GenderMale represents male.
	GenderMale Gender = identityimpl.GenderMale
)

// IDCardInfo contains parsed information from a mainland China identity card.
type IDCardInfo = identityimpl.IDCardInfo

// AgeOption customizes AgeWithOptions.
type AgeOption = identityimpl.AgeOption

// BirthOption customizes birthday parsing helpers.
type BirthOption = identityimpl.BirthOption

// IDCardOption customizes identity-card validation helpers per call.
type IDCardOption = identityimpl.IDCardOption

// RegionCardInfo contains parsed validation information for Hong Kong, Macau or Taiwan cards.
type RegionCardInfo = identityimpl.RegionCardInfo

// Convert15To18 converts a 15-digit mainland China identity card number to 18 digits.
func Convert15To18(idCard string) (string, bool) { return identityimpl.Convert15To18(idCard) }

// Convert15To18WithOptions converts a 15-digit mainland China identity card number to 18 digits with options.
func Convert15To18WithOptions(idCard string, opts ...IDCardOption) (string, bool) {
	return identityimpl.Convert15To18WithOptions(idCard, opts...)
}

// Convert18To15 converts a valid 18-digit mainland China identity card number to 15 digits.
func Convert18To15(idCard string) (string, bool) { return identityimpl.Convert18To15(idCard) }

// Convert18To15WithOptions converts a valid 18-digit mainland China identity card number to 15 digits with options.
func Convert18To15WithOptions(idCard string, opts ...IDCardOption) (string, bool) {
	return identityimpl.Convert18To15WithOptions(idCard, opts...)
}

// IsValidIDCard reports whether idCard is a valid 18-digit, 15-digit, or Hong Kong/Macau/Taiwan card number.
func IsValidIDCard(idCard string) bool { return identityimpl.IsValidIDCard(idCard) }

// IsValidIDCardWithOptions reports whether idCard is valid with options.
func IsValidIDCardWithOptions(idCard string, opts ...IDCardOption) bool {
	return identityimpl.IsValidIDCardWithOptions(idCard, opts...)
}

// IsValidIDCard18 reports whether idCard is a valid 18-digit mainland China identity card number.
func IsValidIDCard18(idCard string) bool { return identityimpl.IsValidIDCard18(idCard) }

// IsValidIDCard18WithOptions reports whether idCard is a valid 18-digit mainland China identity card number with options.
func IsValidIDCard18WithOptions(idCard string, opts ...IDCardOption) bool {
	return identityimpl.IsValidIDCard18WithOptions(idCard, opts...)
}

// IsValidIDCard18WithIgnoreCase validates an 18-digit identity card number and controls X/x comparison.
func IsValidIDCard18WithIgnoreCase(idCard string, ignoreCase bool) bool {
	return identityimpl.IsValidIDCard18WithIgnoreCase(idCard, ignoreCase)
}

// IsValidIDCard18WithIgnoreCaseAndOptions validates an 18-digit identity card number with options.
func IsValidIDCard18WithIgnoreCaseAndOptions(idCard string, ignoreCase bool, opts ...IDCardOption) bool {
	return identityimpl.IsValidIDCard18WithIgnoreCaseAndOptions(idCard, ignoreCase, opts...)
}

// IsValidIDCard15 reports whether idCard is a valid 15-digit mainland China identity card number.
func IsValidIDCard15(idCard string) bool { return identityimpl.IsValidIDCard15(idCard) }

// IsValidIDCard15WithOptions reports whether idCard is a valid 15-digit mainland China identity card number with options.
func IsValidIDCard15WithOptions(idCard string, opts ...IDCardOption) bool {
	return identityimpl.IsValidIDCard15WithOptions(idCard, opts...)
}

// ParseRegionCard validates a Hong Kong, Macau or Taiwan identity card number.
func ParseRegionCard(idCard string) (RegionCardInfo, bool) {
	return identityimpl.ParseRegionCard(idCard)
}

// ParseRegionCardWithOptions validates a Hong Kong, Macau or Taiwan identity card number with options.
func ParseRegionCardWithOptions(idCard string, opts ...IDCardOption) (RegionCardInfo, bool) {
	return identityimpl.ParseRegionCardWithOptions(idCard, opts...)
}

// IsValidTWIDCard reports whether idCard is a valid Taiwan identity card number.
func IsValidTWIDCard(idCard string) bool { return identityimpl.IsValidTWIDCard(idCard) }

// IsValidTWIDCardWithOptions reports whether idCard is a valid Taiwan identity card number with options.
func IsValidTWIDCardWithOptions(idCard string, opts ...IDCardOption) bool {
	return identityimpl.IsValidTWIDCardWithOptions(idCard, opts...)
}

// IsValidHKIDCard reports whether idCard is a valid Hong Kong identity card number.
func IsValidHKIDCard(idCard string) bool { return identityimpl.IsValidHKIDCard(idCard) }

// IsValidHKIDCardWithOptions reports whether idCard is a valid Hong Kong identity card number with options.
func IsValidHKIDCardWithOptions(idCard string, opts ...IDCardOption) bool {
	return identityimpl.IsValidHKIDCardWithOptions(idCard, opts...)
}

// WithDigitsMatcher sets the decimal-digits matcher used by mainland ID card helpers.
func WithDigitsMatcher(matcher func(string) bool) IDCardOption {
	return identityimpl.WithDigitsMatcher(matcher)
}

// WithTWCardMatcher sets the format matcher used by Taiwan ID card helpers.
func WithTWCardMatcher(matcher func(string) bool) IDCardOption {
	return identityimpl.WithTWCardMatcher(matcher)
}

// WithMacauCardMatcher sets the format matcher used by Macau ID card helpers.
func WithMacauCardMatcher(matcher func(string) bool) IDCardOption {
	return identityimpl.WithMacauCardMatcher(matcher)
}

// WithHKCardMatcher sets the format matcher used by Hong Kong ID card helpers.
func WithHKCardMatcher(matcher func(string) bool) IDCardOption {
	return identityimpl.WithHKCardMatcher(matcher)
}

// BirthString returns the birthday encoded in idCard as yyyyMMdd.
func BirthString(idCard string) (string, bool) { return identityimpl.BirthString(idCard) }

// BirthDate returns the birthday encoded in idCard.
func BirthDate(idCard string) (time.Time, bool) { return identityimpl.BirthDate(idCard) }

// WithBirthLocation sets the location used to parse yyyyMMdd birthdays.
func WithBirthLocation(location *time.Location) BirthOption {
	return identityimpl.WithBirthLocation(location)
}

// WithBirthDigitsMatcher sets the decimal-digits matcher used by birthday helpers.
func WithBirthDigitsMatcher(matcher func(string) bool) BirthOption {
	return identityimpl.WithBirthDigitsMatcher(matcher)
}

// WithBirthParser sets the date parser used by birthday helpers.
func WithBirthParser(parser func(layout, value string, location *time.Location) (time.Time, error)) BirthOption {
	return identityimpl.WithBirthParser(parser)
}

// BirthStringWithOptions returns the birthday encoded in idCard as yyyyMMdd using custom parsing options.
func BirthStringWithOptions(idCard string, opts ...BirthOption) (string, bool) {
	return identityimpl.BirthStringWithOptions(idCard, opts...)
}

// BirthDateWithOptions returns the birthday encoded in idCard using custom parsing options.
func BirthDateWithOptions(idCard string, opts ...BirthOption) (time.Time, bool) {
	return identityimpl.BirthDateWithOptions(idCard, opts...)
}

// Age returns the current age encoded in idCard.
func Age(idCard string) (int, bool) { return identityimpl.Age(idCard) }

// WithAgeTime sets the time used by AgeWithOptions.
func WithAgeTime(at time.Time) AgeOption { return identityimpl.WithAgeTime(at) }

// WithAgeClock sets the clock used by AgeWithOptions.
func WithAgeClock(clock func() time.Time) AgeOption { return identityimpl.WithAgeClock(clock) }

// AgeWithOptions returns the age encoded in idCard using custom time options.
func AgeWithOptions(idCard string, opts ...AgeOption) (int, bool) {
	return identityimpl.AgeWithOptions(idCard, opts...)
}

// AgeAt returns the age encoded in idCard at the specified time.
func AgeAt(idCard string, at time.Time) (int, bool) { return identityimpl.AgeAt(idCard, at) }

// Year returns the birth year encoded in idCard.
func Year(idCard string) (int, bool) { return identityimpl.Year(idCard) }

// Month returns the birth month encoded in idCard.
func Month(idCard string) (int, bool) { return identityimpl.Month(idCard) }

// Day returns the birth day encoded in idCard.
func Day(idCard string) (int, bool) { return identityimpl.Day(idCard) }

// GenderOf returns the gender encoded in a 15- or 18-digit identity card number.
func GenderOf(idCard string) (Gender, bool) { return identityimpl.GenderOf(idCard) }

// ProvinceCode returns the province code encoded in a 15- or 18-digit identity card number.
func ProvinceCode(idCard string) (string, bool) { return identityimpl.ProvinceCode(idCard) }

// Province returns the province name encoded in a 15- or 18-digit identity card number.
func Province(idCard string) (string, bool) { return identityimpl.Province(idCard) }

// CityCode returns the city-level code encoded in a 15- or 18-digit identity card number.
func CityCode(idCard string) (string, bool) { return identityimpl.CityCode(idCard) }

// DistrictCode returns the district-level code encoded in a 15- or 18-digit identity card number.
func DistrictCode(idCard string) (string, bool) { return identityimpl.DistrictCode(idCard) }

// ParseIDCard parses a valid 15- or 18-digit mainland China identity card number.
func ParseIDCard(idCard string) (IDCardInfo, bool) { return identityimpl.ParseIDCard(idCard) }

// Hide replaces runes in [start, end) with '*'. Indexes are rune based.
func Hide(idCard string, start, end int) string { return identityimpl.Hide(idCard, start, end) }

// CheckCode18 returns the 18th check code for a 17-digit identity card body.
func CheckCode18(code17 string) byte { return identityimpl.CheckCode18(code17) }

// CheckCode18WithOptions returns the 18th check code for a 17-digit identity card body with options.
func CheckCode18WithOptions(code17 string, opts ...IDCardOption) byte {
	return identityimpl.CheckCode18WithOptions(code17, opts...)
}

// IsValidBirthday reports whether s is a valid yyyyMMdd date.
func IsValidBirthday(s string) bool { return identityimpl.IsValidBirthday(s) }

// IsValidBirthdayWithOptions reports whether s is a valid yyyyMMdd date using custom parsing options.
func IsValidBirthdayWithOptions(s string, opts ...BirthOption) bool {
	return identityimpl.IsValidBirthdayWithOptions(s, opts...)
}
