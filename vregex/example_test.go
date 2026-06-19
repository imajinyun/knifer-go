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

func ExampleExtractMulti() {
	fmt.Println(vregex.ExtractMulti(`(\d+)年(\d+)月`, "2026年5月", `$1-$2`))
	// Output: 2026-5
}

func ExampleGetByName() {
	fmt.Println(vregex.GetByName(`(?<word>\w+)-(?<num>\d+)`, "item-42", "num"))
	// Output: 42
}
