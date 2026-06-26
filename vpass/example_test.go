package vpass_test

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vpass"
)

func ExampleScore() {
	fmt.Println(vpass.Score("abc"))
	fmt.Println(vpass.Score("Abc@1234"))
	// Output:
	// 7
	// 65
}

func ExampleIsStrong() {
	fmt.Println(vpass.IsStrong("weak"))
	fmt.Println(vpass.IsStrong("Str0ng!Pass"))
	// Output:
	// false
	// true
}

func ExampleIsWeak() {
	fmt.Println(vpass.IsWeak("password"))
	fmt.Println(vpass.IsWeak("Str0ng!Pass"))
	// Output:
	// true
	// false
}

func ExampleAnalyze() {
	analysis := vpass.Analyze("password")

	fmt.Println(analysis.Score)
	fmt.Println(analysis.Strength)
	fmt.Println(analysis.CommonWeak)
	// Output:
	// 10
	// very weak
	// true
}

func ExampleStrengthOf() {
	fmt.Println(vpass.StrengthOf("Str0ng!Pass"))
	// Output: strong
}
