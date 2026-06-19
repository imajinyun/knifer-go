package vcodec_test

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vcodec"
)

func ExampleBase64Encode() {
	encoded := vcodec.Base64Encode([]byte("go-knifer"))
	fmt.Println(encoded)
	// Output: Z28ta25pZmVy
}

func ExampleBase64Decode() {
	decoded, _ := vcodec.Base64Decode("Z28ta25pZmVy")
	fmt.Println(string(decoded))
	// Output: go-knifer
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
