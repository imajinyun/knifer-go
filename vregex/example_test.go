package vregex_test

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vregex"
)

func ExampleMatch() {
	fmt.Println(vregex.Match(`^\d+$`, "123"))
	fmt.Println(vregex.Match(`^\d+$`, "abc"))
	// Output:
	// true
	// false
}

func ExampleFind() {
	fmt.Println(vregex.Find(`\d+`, "abc123def456"))
	// Output: 123
}

func ExampleReplace() {
	result := vregex.Replace(`\d+`, "abc123def", "X")
	fmt.Println(result)
	// Output: abcXdef
}
