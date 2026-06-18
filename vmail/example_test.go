package vmail_test

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vmail"
)

func ExampleNewMessage() {
	m, err := vmail.NewMessage(
		vmail.WithFrom("sender@example.com"),
		vmail.WithTo("recipient@example.com"),
		vmail.WithSubject("Hello"),
		vmail.WithText("World"),
	)
	fmt.Println(m != nil)
	fmt.Println(err)
	// Output:
	// true
	// <nil>
}
