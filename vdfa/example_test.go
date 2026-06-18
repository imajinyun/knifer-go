package vdfa_test

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vdfa"
)

func ExampleContains() {
	vdfa.Init([]string{"bad", "word"})

	fmt.Println(vdfa.Contains("this is a bad word"))
	fmt.Println(vdfa.Contains("clean text"))
	// Output:
	// true
	// false
}

func ExampleFilter() {
	vdfa.Init([]string{"bad"})

	result := vdfa.Filter("this has bad content")
	fmt.Println(result)
	// Output: this has *** content
}
