package vform_test

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vform"
)

func ExampleIsEmail() {
	fmt.Println(vform.IsEmail("team@example.com"))
	fmt.Println(vform.IsEmail("not-an-email"))

	// Output:
	// true
	// false
}

func ExampleIsNumberStrWithOptions() {
	isHexNumber := func(s string) bool {
		for _, r := range s {
			switch {
			case '0' <= r && r <= '9':
			case 'a' <= r && r <= 'f':
			default:
				return false
			}
		}
		return s != ""
	}

	fmt.Println(vform.IsNumberStrWithOptions("1af", vform.WithNumberMatcher(isHexNumber)))
	fmt.Println(vform.IsNumberStrWithOptions("1xz", vform.WithNumberMatcher(isHexNumber)))

	// Output:
	// true
	// false
}
