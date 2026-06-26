package vmail_test

import (
	"fmt"
	"strings"

	"github.com/imajinyun/knifer-go/vmail"
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

func ExampleNewInline() {
	inline, err := vmail.NewInline("logo.png", "logo", []byte("img"), vmail.TypeApplicationOctetStream)

	fmt.Println(inline.Name, inline.ContentID, inline.Size)
	fmt.Println(err)
	// Output:
	// logo.png logo 3
	// <nil>
}

func ExampleWithHeader() {
	message, err := vmail.NewMessage(
		vmail.WithFrom("sender@example.com"),
		vmail.WithTo("recipient@example.com"),
		vmail.WithSubject("Hello"),
		vmail.WithText("World"),
		vmail.WithHeader("X-Trace", "abc"),
	)
	data, bytesErr := message.Bytes()

	fmt.Println(strings.Contains(string(data), "X-Trace: abc"))
	fmt.Println(err)
	fmt.Println(bytesErr)
	// Output:
	// true
	// <nil>
	// <nil>
}

func ExampleMessage_Recipients() {
	message, err := vmail.NewMessage(
		vmail.WithFrom("sender@example.com"),
		vmail.WithTo("alice@example.com", "bob@example.com"),
		vmail.WithCc("alice@example.com"),
		vmail.WithText("World"),
	)

	fmt.Println(message.Recipients())
	fmt.Println(err)
	// Output:
	// [alice@example.com bob@example.com]
	// <nil>
}
