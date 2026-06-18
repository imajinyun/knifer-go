package vref_test

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vref"
)

func ExampleTypeOf() {
	t := vref.TypeOf("hello")
	fmt.Println(t.Name())
	// Output: string
}

func ExampleIsNil() {
	var s *string
	fmt.Println(vref.IsNil(s))
	fmt.Println(vref.IsNil("hello"))
	// Output:
	// true
	// false
}
