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

func ExampleNewAddress() {
	addr, err := vmail.NewAddress("Alice", "alice@example.com")

	fmt.Println(addr.String())
	fmt.Println(err)
	// Output:
	// "Alice" <alice@example.com>
	// <nil>
}

func ExampleParseAddressList() {
	list, err := vmail.ParseAddressList("bob@example.com, carol@example.com")

	fmt.Println(len(list), list[0].Email, list[1].Email)
	fmt.Println(err)
	// Output:
	// 2 bob@example.com carol@example.com
	// <nil>
}

func ExampleNewAttachment() {
	attachment, err := vmail.NewAttachment("report.txt", []byte("report"), vmail.TypeTextPlain)

	fmt.Println(attachment.Name, attachment.Size, attachment.ContentType)
	fmt.Println(err)
	// Output:
	// report.txt 6 text/plain
	// <nil>
}

func ExampleMessage_Sender() {
	message, err := vmail.NewMessage(
		vmail.WithFrom("sender@example.com"),
		vmail.WithEnvelopeFrom("bounce@example.com"),
		vmail.WithTo("recipient@example.com"),
		vmail.WithSubject("Hello"),
		vmail.WithText("World"),
	)

	fmt.Println(message.Sender())
	fmt.Println(err)
	// Output:
	// bounce@example.com
	// <nil>
}

func ExampleParseAddress() {
	addr, err := vmail.ParseAddress("Carol <carol@example.com>")

	fmt.Println(addr.Name, addr.Email)
	fmt.Println(err)
	// Output:
	// Carol carol@example.com
	// <nil>
}
