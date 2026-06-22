package vident_test

import (
	"fmt"
	"time"

	"github.com/imajinyun/go-knifer/vident"
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
