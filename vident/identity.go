package vident

import (
	"time"

	identityimpl "github.com/imajinyun/go-knifer/internal/identity"
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

// RegionCardInfo contains parsed validation information for Hong Kong, Macau or Taiwan cards.
type RegionCardInfo = identityimpl.RegionCardInfo

// Convert15To18 converts a 15-digit mainland China identity card number to 18 digits.
func Convert15To18(idCard string) (string, bool) { return identityimpl.Convert15To18(idCard) }

// Convert18To15 converts a valid 18-digit mainland China identity card number to 15 digits.
func Convert18To15(idCard string) (string, bool) { return identityimpl.Convert18To15(idCard) }

// IsValidIDCard reports whether idCard is a valid 18-digit, 15-digit, or Hong Kong/Macau/Taiwan card number.
func IsValidIDCard(idCard string) bool { return identityimpl.IsValidIDCard(idCard) }

// IsValidIDCard18 reports whether idCard is a valid 18-digit mainland China identity card number.
func IsValidIDCard18(idCard string) bool { return identityimpl.IsValidIDCard18(idCard) }

// IsValidIDCard18WithIgnoreCase validates an 18-digit identity card number and controls X/x comparison.
func IsValidIDCard18WithIgnoreCase(idCard string, ignoreCase bool) bool {
	return identityimpl.IsValidIDCard18WithIgnoreCase(idCard, ignoreCase)
}

// IsValidIDCard15 reports whether idCard is a valid 15-digit mainland China identity card number.
func IsValidIDCard15(idCard string) bool { return identityimpl.IsValidIDCard15(idCard) }

// ParseRegionCard validates a Hong Kong, Macau or Taiwan identity card number.
func ParseRegionCard(idCard string) (RegionCardInfo, bool) {
	return identityimpl.ParseRegionCard(idCard)
}

// IsValidTWIDCard reports whether idCard is a valid Taiwan identity card number.
func IsValidTWIDCard(idCard string) bool { return identityimpl.IsValidTWIDCard(idCard) }

// IsValidHKIDCard reports whether idCard is a valid Hong Kong identity card number.
func IsValidHKIDCard(idCard string) bool { return identityimpl.IsValidHKIDCard(idCard) }

// BirthString returns the birthday encoded in idCard as yyyyMMdd.
func BirthString(idCard string) (string, bool) { return identityimpl.BirthString(idCard) }

// BirthDate returns the birthday encoded in idCard.
func BirthDate(idCard string) (time.Time, bool) { return identityimpl.BirthDate(idCard) }

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

// IsValidBirthday reports whether s is a valid yyyyMMdd date.
func IsValidBirthday(s string) bool { return identityimpl.IsValidBirthday(s) }
