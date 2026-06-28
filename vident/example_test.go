package vident_test

import (
	"fmt"
	"time"

	"github.com/imajinyun/knifer-go/vident"
)

func ExampleParseIDCard() {
	info, ok := vident.ParseIDCard("11010519491231002X")
	fmt.Println(ok, info.Province, info.Gender == vident.GenderFemale)
	// Output: true 北京 true
}

func ExampleHide() {
	fmt.Println(vident.Hide("11010519491231002X", 6, 14))
	// Output: 110105********002X
}

func ExampleAgeAt() {
	age, ok := vident.AgeAt("11010519491231002X", time.Date(2024, 12, 31, 0, 0, 0, 0, time.Local))
	fmt.Println(age, ok)
	// Output: 75 true
}

func ExampleConvert15To18() {
	converted, ok := vident.Convert15To18("130503670401001")
	fmt.Println(converted, ok)
	// Output: 130503196704010016 true
}

func ExampleBirthString() {
	birth, ok := vident.BirthString("11010519491231002X")
	fmt.Println(birth, ok)
	// Output: 19491231 true
}

func ExampleParseRegionCard() {
	info, ok := vident.ParseRegionCard("A123456(3)")
	fmt.Println(ok, info.Region, info.Valid)
	// Output: true 香港 true
}

func ExampleIsValidCreditCode() {
	fmt.Println(vident.IsValidCreditCode("91350211M000100Y46"))
	fmt.Println(vident.IsValidCreditCode("91350211M000100Y44"))
	// Output:
	// true
	// false
}

func ExampleParseCreditCode() {
	info, err := vident.ParseCreditCode("91350211M000100Y46")
	if err != nil {
		panic(err)
	}

	fmt.Println(info.AdminDept, info.OrgCategory, info.RegionCode)
	fmt.Println(info.OrgCode, info.CheckDigit)
	// Output:
	// 9 1 350211
	// M000100Y4 6
}

func ExampleAgeWithOptions() {
	age, ok := vident.AgeWithOptions(
		"11010519491231002X",
		vident.WithAgeTime(time.Date(2024, 12, 30, 0, 0, 0, 0, time.Local)),
	)
	fmt.Println(age, ok)
	// Output: 74 true
}

func ExampleGenderOf() {
	gender, ok := vident.GenderOf("11010519491231002X")
	fmt.Println(gender == vident.GenderFemale, ok)
	// Output: true true
}

func ExampleProvince() {
	province, ok := vident.Province("11010519491231002X")
	code, _ := vident.ProvinceCode("11010519491231002X")
	fmt.Println(province, code, ok)
	// Output: 北京 11 true
}

func ExampleCheckCode18() {
	fmt.Printf("%c\n", vident.CheckCode18("11010519491231002"))
	// Output: X
}

func ExampleConvert18To15() {
	converted, ok := vident.Convert18To15("130503196704010016")
	fmt.Println(converted, ok)
	// Output: 130503670401001 true
}

func ExampleAge() {
	age, ok := vident.Age("11010519491231002X")
	fmt.Println(ok, age >= 0)
	// Output: true true
}

func ExampleBirthDate() {
	birth, ok := vident.BirthDate("11010519491231002X")
	fmt.Println(birth.Format("2006-01-02"), ok)
	// Output: 1949-12-31 true
}

func ExampleBirthDateWithOptions() {
	loc := time.FixedZone("facade", 8*3600)
	birth, ok := vident.BirthDateWithOptions("11010519491231002X", vident.WithBirthLocation(loc))
	fmt.Println(birth.Location(), birth.Format("2006-01-02"), ok)
	// Output: facade 1949-12-31 true
}

func ExampleBirthStringWithOptions() {
	birth, ok := vident.BirthStringWithOptions(
		"11010519491231002X",
		vident.WithBirthDigitsMatcher(func(string) bool { return false }),
	)
	fmt.Println(birth, ok)
	// Output: 19491231 false
}

func ExampleCheckCode18WithOptions() {
	accepted := vident.CheckCode18WithOptions("11010519491231002")
	rejected := vident.CheckCode18WithOptions(
		"11010519491231002",
		vident.WithDigitsMatcher(func(string) bool { return false }),
	)
	fmt.Printf("%q %q\n", accepted, rejected)
	// Output: 'X' ' '
}

func ExampleCityCode() {
	code, ok := vident.CityCode("11010519491231002X")
	fmt.Println(code, ok)
	// Output: 1101 true
}

func ExampleConvert15To18WithOptions() {
	converted, ok := vident.Convert15To18WithOptions("130503670401001")
	fmt.Println(converted, ok)
	// Output: 130503196704010016 true
}

func ExampleConvert18To15WithOptions() {
	converted, ok := vident.Convert18To15WithOptions("130503196704010016")
	fmt.Println(converted, ok)
	// Output: 130503670401001 true
}

func ExampleDay() {
	day, ok := vident.Day("11010519491231002X")
	fmt.Println(day, ok)
	// Output: 31 true
}

func ExampleDistrictCode() {
	code, ok := vident.DistrictCode("11010519491231002X")
	fmt.Println(code, ok)
	// Output: 110105 true
}

func ExampleIsValidBirthday() {
	fmt.Println(vident.IsValidBirthday("19491231"))
	fmt.Println(vident.IsValidBirthday("19490231"))
	// Output:
	// true
	// false
}

func ExampleIsValidBirthdayWithOptions() {
	valid := vident.IsValidBirthdayWithOptions("19491231", vident.WithBirthLocation(time.UTC))
	rejected := vident.IsValidBirthdayWithOptions(
		"19491231",
		vident.WithBirthDigitsMatcher(func(string) bool { return false }),
	)
	fmt.Println(valid, rejected)
	// Output: true false
}

func ExampleIsValidIDCard() {
	fmt.Println(vident.IsValidIDCard("11010519491231002X"))
	fmt.Println(vident.IsValidIDCard("110105194912310021"))
	// Output:
	// true
	// false
}

func ExampleIsValidIDCard15() {
	fmt.Println(vident.IsValidIDCard15("130503670401001"))
	// Output: true
}

func ExampleIsValidIDCard15WithOptions() {
	accepted := vident.IsValidIDCard15WithOptions("130503670401001")
	rejected := vident.IsValidIDCard15WithOptions(
		"130503670401001",
		vident.WithDigitsMatcher(func(string) bool { return false }),
	)
	fmt.Println(accepted, rejected)
	// Output: true false
}

func ExampleIsValidIDCard18() {
	fmt.Println(vident.IsValidIDCard18("11010519491231002X"))
	// Output: true
}

func ExampleIsValidIDCard18WithIgnoreCase() {
	fmt.Println(vident.IsValidIDCard18WithIgnoreCase("11010519491231002x", true))
	fmt.Println(vident.IsValidIDCard18WithIgnoreCase("11010519491231002x", false))
	// Output:
	// true
	// false
}

func ExampleIsValidIDCard18WithIgnoreCaseAndOptions() {
	accepted := vident.IsValidIDCard18WithIgnoreCaseAndOptions("11010519491231002x", true)
	rejected := vident.IsValidIDCard18WithIgnoreCaseAndOptions(
		"11010519491231002x",
		true,
		vident.WithDigitsMatcher(func(string) bool { return false }),
	)
	fmt.Println(accepted, rejected)
	// Output: true false
}

func ExampleIsValidIDCard18WithOptions() {
	accepted := vident.IsValidIDCard18WithOptions("11010519491231002X")
	rejected := vident.IsValidIDCard18WithOptions(
		"11010519491231002X",
		vident.WithDigitsMatcher(func(string) bool { return false }),
	)
	fmt.Println(accepted, rejected)
	// Output: true false
}

func ExampleIsValidIDCardWithOptions() {
	accepted := vident.IsValidIDCardWithOptions("11010519491231002X")
	rejected := vident.IsValidIDCardWithOptions(
		"11010519491231002X",
		vident.WithDigitsMatcher(func(string) bool { return false }),
	)
	fmt.Println(accepted, rejected)
	// Output: true false
}

func ExampleMonth() {
	month, ok := vident.Month("11010519491231002X")
	fmt.Println(month, ok)
	// Output: 12 true
}

func ExampleProvinceCode() {
	code, ok := vident.ProvinceCode("11010519491231002X")
	fmt.Println(code, ok)
	// Output: 11 true
}

func ExampleWithAgeClock() {
	age, ok := vident.AgeWithOptions(
		"11010519491231002X",
		vident.WithAgeClock(func() time.Time {
			return time.Date(2024, 12, 30, 0, 0, 0, 0, time.Local)
		}),
	)
	fmt.Println(age, ok)
	// Output: 74 true
}

func ExampleWithAgeTime() {
	age, ok := vident.AgeWithOptions(
		"11010519491231002X",
		vident.WithAgeTime(time.Date(2024, 12, 31, 0, 0, 0, 0, time.Local)),
	)
	fmt.Println(age, ok)
	// Output: 75 true
}

func ExampleWithBirthDigitsMatcher() {
	birth, ok := vident.BirthStringWithOptions(
		"11010519491231002X",
		vident.WithBirthDigitsMatcher(func(string) bool { return false }),
	)
	fmt.Println(birth, ok)
	// Output: 19491231 false
}

func ExampleWithBirthLocation() {
	loc := time.FixedZone("facade", 8*3600)
	birth, ok := vident.BirthDateWithOptions("11010519491231002X", vident.WithBirthLocation(loc))
	fmt.Println(birth.Location(), ok)
	// Output: facade true
}

func ExampleWithBirthParser() {
	called := false
	birth, ok := vident.BirthDateWithOptions(
		"11010519491231002X",
		vident.WithBirthParser(func(layout, value string, location *time.Location) (time.Time, error) {
			called = true
			return time.ParseInLocation(layout, value, location)
		}),
	)
	fmt.Println(called, birth.Format("2006-01-02"), ok)
	// Output: true 1949-12-31 true
}

func ExampleWithDigitsMatcher() {
	rejected := vident.IsValidIDCard18WithOptions(
		"11010519491231002X",
		vident.WithDigitsMatcher(func(string) bool { return false }),
	)
	fmt.Println(rejected)
	// Output: false
}

func ExampleYear() {
	year, ok := vident.Year("11010519491231002X")
	fmt.Println(year, ok)
	// Output: 1949 true
}

func ExampleIsValidHKIDCard() {
	fmt.Println(vident.IsValidHKIDCard("A123456(3)"))
	// Output: true
}

func ExampleIsValidHKIDCardWithOptions() {
	accepted := vident.IsValidHKIDCardWithOptions("A123456(3)")
	rejected := vident.IsValidHKIDCardWithOptions(
		"A123456(3)",
		vident.WithHKCardMatcher(func(string) bool { return false }),
	)
	fmt.Println(accepted, rejected)
	// Output: true false
}

func ExampleIsValidTWIDCard() {
	fmt.Println(vident.IsValidTWIDCard("A123456789"))
	// Output: true
}

func ExampleIsValidTWIDCardWithOptions() {
	accepted := vident.IsValidTWIDCardWithOptions("A123456789")
	rejected := vident.IsValidTWIDCardWithOptions(
		"A123456789",
		vident.WithTWCardMatcher(func(string) bool { return false }),
	)
	fmt.Println(accepted, rejected)
	// Output: true false
}

func ExampleParseRegionCardWithOptions() {
	info, ok := vident.ParseRegionCardWithOptions("1571234(5)")
	rejected, rejectedOK := vident.ParseRegionCardWithOptions(
		"1571234(5)",
		vident.WithMacauCardMatcher(func(string) bool { return false }),
	)
	fmt.Println(info.Region, info.Valid, ok)
	fmt.Println(rejected.Region, rejected.Valid, rejectedOK)
	// Output:
	// 澳门 true true
	//  false false
}

func ExampleWithHKCardMatcher() {
	accepted := vident.IsValidHKIDCardWithOptions("A123456(3)")
	rejected := vident.IsValidHKIDCardWithOptions(
		"A123456(3)",
		vident.WithHKCardMatcher(func(string) bool { return false }),
	)
	fmt.Println(accepted, rejected)
	// Output: true false
}

func ExampleWithMacauCardMatcher() {
	info, ok := vident.ParseRegionCardWithOptions(
		"1571234(5)",
		vident.WithMacauCardMatcher(func(card string) bool { return card == "1571234(5)" }),
	)
	fmt.Println(info.Region, info.Valid, ok)
	// Output: 澳门 true true
}

func ExampleWithTWCardMatcher() {
	accepted := vident.IsValidTWIDCardWithOptions("A123456789")
	rejected := vident.IsValidTWIDCardWithOptions(
		"A123456789",
		vident.WithTWCardMatcher(func(string) bool { return false }),
	)
	fmt.Println(accepted, rejected)
	// Output: true false
}
