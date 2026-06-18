package vimg_test

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vimg"
)

func ExampleVerifyIgnoreCase() {
	matched := vimg.VerifyIgnoreCase("AbC4", "abc4")
	fmt.Println(matched)
	// Output: true
}

func ExampleNewMathGeneratorWith() {
	gen := vimg.NewMathGeneratorWith(2, false)
	result := vimg.GenMathGeneratorWithOptions(gen)
	fmt.Println(len(result) > 0)
	// Output: true
}
