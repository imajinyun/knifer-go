package vpass_test

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vpass"
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
