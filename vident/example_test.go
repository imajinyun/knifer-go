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
