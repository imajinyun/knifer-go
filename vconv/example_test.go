package vconv_test

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vconv"
)

func ExampleToInt() {
	fmt.Println(vconv.ToInt("42"))
	fmt.Println(vconv.ToInt(true))
	// Output:
	// 42
	// 1
}

func ExampleToIntDefault() {
	fmt.Println(vconv.ToIntDefault("not-a-number", -1))
	// Output: -1
}

func ExampleToBool() {
	fmt.Println(vconv.ToBool("true"))
	fmt.Println(vconv.ToBool(0))
	// Output:
	// true
	// false
}

func ExampleToString() {
	fmt.Println(vconv.ToString(3.14))
	// Output: 3.14
}

func ExampleToBytes() {
	fmt.Println(string(vconv.ToBytes("go")))
	// Output: go
}
