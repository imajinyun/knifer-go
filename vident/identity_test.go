package vident_test

import (
	"testing"
	"time"

	"github.com/imajinyun/knifer-go/vident"
)

func TestIDCardFacade(t *testing.T) {
	converted, ok := vident.Convert15To18("130503670401001")
	if !ok || converted != "130503196704010016" {
		t.Fatalf("Convert15To18() = %q, %v", converted, ok)
	}
	if !vident.IsValidIDCard("11010519491231002X") {
		t.Fatal("expected valid ID card")
	}
	if !vident.IsValidIDCard18("11010519491231002X") {
		t.Fatal("expected valid 18-digit ID card")
	}
	if !vident.IsValidIDCard18WithIgnoreCase("11010519491231002x", true) {
		t.Fatal("expected valid 18-digit card with ignore case")
	}
	if vident.IsValidIDCard18WithIgnoreCase("11010519491231002x", false) {
		t.Fatal("expected invalid 18-digit card without ignore case")
	}
	if !vident.IsValidIDCard15("130503670401001") {
		t.Fatal("expected valid 15-digit ID card")
	}
	if age, ok := vident.Age("11010519491231002X"); !ok || age < 0 {
		t.Fatalf("Age() = %d, %v", age, ok)
	}
	birth, ok := vident.BirthDate("11010519491231002X")
	if !ok || birth.Format("2006-01-02") != "1949-12-31" {
		t.Fatalf("BirthDate() = %v, %v", birth, ok)
	}
	loc := time.FixedZone("facade", 8*3600)
	birth, ok = vident.BirthDateWithOptions("11010519491231002X", vident.WithBirthLocation(loc))
	if !ok || birth.Location() != loc || birth.Format("2006-01-02") != "1949-12-31" {
		t.Fatalf("BirthDateWithOptions() = %v, %v", birth, ok)
	}
	if !vident.IsValidBirthdayWithOptions("19491231", vident.WithBirthLocation(loc)) {
		t.Fatal("IsValidBirthdayWithOptions should accept valid birthday")
	}
	age, ok := vident.AgeAt("11010519491231002X", time.Date(2024, 12, 31, 0, 0, 0, 0, time.Local))
	if !ok || age != 75 {
		t.Fatalf("AgeAt() = %d, %v", age, ok)
	}
	age, ok = vident.AgeWithOptions(
		"11010519491231002X",
		vident.WithAgeClock(func() time.Time { return time.Date(2024, 12, 30, 0, 0, 0, 0, time.Local) }),
	)
	if !ok || age != 74 {
		t.Fatalf("AgeWithOptions() = %d, %v", age, ok)
	}
	info, ok := vident.ParseIDCard("11010519491231002X")
	if !ok || info.Province != "北京" || info.Gender != vident.GenderFemale {
		t.Fatalf("ParseIDCard() = %+v, %v", info, ok)
	}
}

func TestRegionCardFacade(t *testing.T) {
	info, ok := vident.ParseRegionCard("A123456(3)")
	if !ok || info.Region != "香港" || !info.Valid {
		t.Fatalf("ParseRegionCard() = %+v, %v", info, ok)
	}
	if !vident.IsValidTWIDCard("A123456789") {
		t.Fatal("expected valid Taiwan card")
	}
}

func TestIDCardFacadeOptionWrappers(t *testing.T) {
	if got, ok := vident.Convert15To18WithOptions("130503670401001"); !ok || got != "130503196704010016" {
		t.Fatalf("Convert15To18WithOptions() = %q, %v", got, ok)
	}
	if got, ok := vident.Convert18To15("130503196704010016"); !ok || got != "130503670401001" {
		t.Fatalf("Convert18To15() = %q, %v", got, ok)
	}
	if got, ok := vident.Convert18To15WithOptions("130503196704010016"); !ok || got != "130503670401001" {
		t.Fatalf("Convert18To15WithOptions() = %q, %v", got, ok)
	}

	if !vident.IsValidIDCardWithOptions("11010519491231002X") {
		t.Fatal("IsValidIDCardWithOptions should accept valid card")
	}
	if !vident.IsValidIDCard18WithOptions("11010519491231002X") {
		t.Fatal("IsValidIDCard18WithOptions should accept valid card")
	}
	if vident.IsValidIDCard18WithOptions("11010519491231002X", vident.WithDigitsMatcher(func(string) bool { return false })) {
		t.Fatal("custom digits matcher should reject 18-digit card")
	}
	if !vident.IsValidIDCard18WithIgnoreCaseAndOptions("11010519491231002x", true) {
		t.Fatal("ignore-case validation should accept lowercase check code")
	}
	if !vident.IsValidIDCard15WithOptions("130503670401001") {
		t.Fatal("IsValidIDCard15WithOptions should accept valid 15-digit card")
	}
	if vident.IsValidIDCard15WithOptions("130503670401001", vident.WithDigitsMatcher(func(string) bool { return false })) {
		t.Fatal("custom digits matcher should reject 15-digit card")
	}

	if got := vident.CheckCode18("11010519491231002"); got != 'X' {
		t.Fatalf("CheckCode18() = %q", got)
	}
	if got := vident.CheckCode18WithOptions("11010519491231002", vident.WithDigitsMatcher(func(string) bool { return false })); got != ' ' {
		t.Fatalf("CheckCode18WithOptions reject = %q", got)
	}
}

func TestIDCardFacadeFieldWrappers(t *testing.T) {
	const id = "11010519491231002X"
	if birth, ok := vident.BirthString(id); !ok || birth != "19491231" {
		t.Fatalf("BirthString() = %q, %v", birth, ok)
	}
	if birth, ok := vident.BirthStringWithOptions(id, vident.WithBirthDigitsMatcher(func(string) bool { return false })); ok || birth != "19491231" {
		t.Fatalf("BirthStringWithOptions rejected = %q, %v", birth, ok)
	}
	parsedWithCustomParser := false
	if birth, ok := vident.BirthDateWithOptions(id, vident.WithBirthParser(func(layout, value string, location *time.Location) (time.Time, error) {
		parsedWithCustomParser = true
		return time.ParseInLocation(layout, value, location)
	})); !ok || !parsedWithCustomParser || birth.Format("2006-01-02") != "1949-12-31" {
		t.Fatalf("BirthDateWithOptions custom parser = %v, %v called=%v", birth, ok, parsedWithCustomParser)
	}
	if age, ok := vident.AgeWithOptions(id, vident.WithAgeTime(time.Date(2024, 12, 31, 0, 0, 0, 0, time.Local))); !ok || age != 75 {
		t.Fatalf("AgeWithOptions WithAgeTime = %d, %v", age, ok)
	}
	if year, ok := vident.Year(id); !ok || year != 1949 {
		t.Fatalf("Year() = %d, %v", year, ok)
	}
	if month, ok := vident.Month(id); !ok || month != 12 {
		t.Fatalf("Month() = %d, %v", month, ok)
	}
	if day, ok := vident.Day(id); !ok || day != 31 {
		t.Fatalf("Day() = %d, %v", day, ok)
	}
	if gender, ok := vident.GenderOf(id); !ok || gender != vident.GenderFemale {
		t.Fatalf("GenderOf() = %d, %v", gender, ok)
	}
	if provinceCode, ok := vident.ProvinceCode(id); !ok || provinceCode != "11" {
		t.Fatalf("ProvinceCode() = %q, %v", provinceCode, ok)
	}
	if province, ok := vident.Province(id); !ok || province != "北京" {
		t.Fatalf("Province() = %q, %v", province, ok)
	}
	if cityCode, ok := vident.CityCode(id); !ok || cityCode != "1101" {
		t.Fatalf("CityCode() = %q, %v", cityCode, ok)
	}
	if districtCode, ok := vident.DistrictCode(id); !ok || districtCode != "110105" {
		t.Fatalf("DistrictCode() = %q, %v", districtCode, ok)
	}
	if got := vident.Hide(id, 6, 14); got != "110105********002X" {
		t.Fatalf("Hide() = %q", got)
	}
	if !vident.IsValidBirthday("19491231") || !vident.IsValidBirthdayWithOptions("19491231") || vident.IsValidBirthday("19490231") {
		t.Fatal("birthday validators returned unexpected results")
	}
}

func TestRegionCardFacadeOptionWrappers(t *testing.T) {
	if info, ok := vident.ParseRegionCardWithOptions("A123456789"); !ok || info.Region != "台湾" || info.Gender != "M" || !info.Valid {
		t.Fatalf("ParseRegionCardWithOptions Taiwan = %+v, %v", info, ok)
	}
	if _, ok := vident.ParseRegionCardWithOptions("A123456789", vident.WithTWCardMatcher(func(string) bool { return false })); ok {
		t.Fatal("custom Taiwan matcher should reject card")
	}
	if info, ok := vident.ParseRegionCardWithOptions("1571234(5)"); !ok || info.Region != "澳门" || !info.Valid {
		t.Fatalf("ParseRegionCardWithOptions Macau = %+v, %v", info, ok)
	}
	if _, ok := vident.ParseRegionCardWithOptions("1571234(5)", vident.WithMacauCardMatcher(func(string) bool { return false })); ok {
		t.Fatal("custom Macau matcher should reject card")
	}
	if !vident.IsValidTWIDCardWithOptions("A123456789") {
		t.Fatal("IsValidTWIDCardWithOptions should accept valid Taiwan card")
	}
	if vident.IsValidTWIDCardWithOptions("A123456789", vident.WithTWCardMatcher(func(string) bool { return false })) {
		t.Fatal("custom Taiwan matcher should reject validator")
	}
	if !vident.IsValidHKIDCard("A123456(3)") || !vident.IsValidHKIDCardWithOptions("A123456(3)") {
		t.Fatal("Hong Kong validators should accept valid card")
	}
	if vident.IsValidHKIDCardWithOptions("A123456(3)", vident.WithHKCardMatcher(func(string) bool { return false })) {
		t.Fatal("custom Hong Kong matcher should reject validator")
	}
}
