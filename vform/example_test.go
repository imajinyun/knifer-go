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

func ExampleIsIPv4() {
	fmt.Println(vform.IsIPv4("192.0.2.1"))
	fmt.Println(vform.IsIPv4("not-an-ip"))
	// Output:
	// true
	// false
}

func ExampleIsMobile() {
	fmt.Println(vform.IsMobile("13812345678"))
	fmt.Println(vform.IsMobile("12812345678"))
	// Output:
	// true
	// false
}

func ExampleIsURL() {
	fmt.Println(vform.IsURL("https://example.com"))
	fmt.Println(vform.IsURL("/relative/path"))
	// Output:
	// true
	// false
}
