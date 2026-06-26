package vver_test

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vver"
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

func ExampleIsLessThan() {
	fmt.Println(vver.IsLessThan("1.0.0", "2.0.0"))
	fmt.Println(vver.IsLessThan("2.0.0", "1.0.0"))
	// Output:
	// true
	// false
}

func ExampleAnyMatch() {
	fmt.Println(vver.AnyMatch("1.2.0", "<1.0.0", "1.2.0-1.3.0"))
	fmt.Println(vver.AnyMatch("2.0.0", "<1.0.0", "1.2.0-1.3.0"))
	// Output:
	// true
	// false
}

func ExampleMatchElWithDelimiter() {
	fmt.Println(vver.MatchElWithDelimiter("2.0.0", "1.0.0|2.0.0", "|"))
	fmt.Println(vver.MatchElWithDelimiter("3.0.0", "1.0.0|2.0.0", "|"))
	// Output:
	// true
	// false
}
