package vcodec_test

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vcodec"
)

func ExampleBase64Encode() {
	encoded := vcodec.Base64Encode([]byte("knifer-go"))
	fmt.Println(encoded)
	// Output: Z28ta25pZmVy
}

func ExampleBase64Decode() {
	decoded, _ := vcodec.Base64Decode("Z28ta25pZmVy")
	fmt.Println(string(decoded))
	// Output: knifer-go
}

func ExampleHexEncode() {
	encoded := vcodec.HexEncode([]byte{0x47, 0x6f})
	fmt.Println(encoded)
	// Output: 476f
}

func ExampleBase64RawURLEncode() {
	encoded := vcodec.Base64RawURLEncode([]byte("go?"))
	decoded, _ := vcodec.Base64RawURLDecode(encoded)

	fmt.Println(encoded)
	fmt.Println(string(decoded))
	// Output:
	// Z28_
	// go?
}

func ExampleHexDecodeStr() {
	decoded, err := vcodec.HexDecodeStr("676f")

	fmt.Println(decoded)
	fmt.Println(err)
	// Output:
	// go
	// <nil>
}
