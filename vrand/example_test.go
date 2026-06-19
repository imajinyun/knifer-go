package vrand_test

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vrand"
)

func ExampleSecureBytes() {
	b, err := vrand.SecureBytes(16)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(b))
	// Output: 16
}

func ExampleIntRange() {
	n := vrand.IntRange(10, 20)
	fmt.Println(n >= 10 && n < 20)
	// Output: true
}

func ExampleString() {
	s := vrand.String(8)
	fmt.Println(len(s))
	// Output: 8
}

func ExampleStringFrom() {
	fmt.Println(vrand.StringFrom("A", 4))
	// Output: AAAA
}

func ExampleEle() {
	fmt.Println(vrand.Ele([]string{"only"}))
	// Output: only
}
