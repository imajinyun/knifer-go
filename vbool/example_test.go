package vbool_test

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vbool"
)

func ExampleAnd() {
	fmt.Println(vbool.And(true, true, true))
	fmt.Println(vbool.And(true, false, true))
	// Output:
	// true
	// false
}

func ExampleOr() {
	fmt.Println(vbool.Or(false, true, false))
	fmt.Println(vbool.Or(false, false))
	// Output:
	// true
	// false
}

func ExampleNegate() {
	fmt.Println(vbool.Negate(true))
	fmt.Println(vbool.Negate(false))
	// Output:
	// false
	// true
}

func ExampleToInt() {
	fmt.Println(vbool.ToInt(true))
	fmt.Println(vbool.ToInt(false))
	// Output:
	// 1
	// 0
}
