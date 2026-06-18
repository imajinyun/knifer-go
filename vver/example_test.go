package vver_test

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vver"
)

func ExampleCompareVersion() {
	fmt.Println(vver.CompareVersion("1.0.0", "2.0.0"))
	fmt.Println(vver.CompareVersion("2.0.0", "1.0.0"))
	fmt.Println(vver.CompareVersion("1.0.0", "1.0.0"))
	// Output:
	// -1
	// 1
	// 0
}

func ExampleIsGreaterThan() {
	fmt.Println(vver.IsGreaterThan("2.0.0", "1.0.0"))
	fmt.Println(vver.IsGreaterThan("1.0.0", "2.0.0"))
	// Output:
	// true
	// false
}
