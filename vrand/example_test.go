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
