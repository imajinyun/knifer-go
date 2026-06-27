package vcodec_test

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vcodec"
)

func ExampleBase64Encode() {
	encoded := vcodec.Base64Encode([]byte("knifer-go"))
	fmt.Println(encoded)
	// Output: a25pZmVyLWdv
}

func ExampleBase64Decode() {
	decoded, _ := vcodec.Base64Decode("a25pZmVyLWdv")
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

func ExampleBase32Encode() {
	encoded := vcodec.Base32Encode([]byte("go"))
	decoded, _ := vcodec.Base32Decode(encoded)

	fmt.Println(encoded)
	fmt.Println(string(decoded))
	// Output:
	// M5XQ====
	// go
}

func ExampleBase58Encode() {
	encoded := vcodec.Base58Encode([]byte("hello world"))
	decoded, _ := vcodec.Base58Decode(encoded)

	fmt.Println(encoded)
	fmt.Println(string(decoded))
	// Output:
	// StV1DL6CwTryKyV
	// hello world
}

func ExampleBase62Encode() {
	encoded := vcodec.Base62Encode([]byte("hello"))
	decoded, _ := vcodec.Base62Decode(encoded)

	fmt.Println(encoded)
	fmt.Println(string(decoded))
	// Output:
	// 7tQLFHz
	// hello
}

func ExampleMorseEncode() {
	encoded, _ := vcodec.MorseEncode("SOS 1")
	decoded, _ := vcodec.MorseDecode(encoded)

	fmt.Println(encoded)
	fmt.Println(decoded)
	// Output:
	// ... --- ... / .----
	// SOS 1
}

func ExampleROT13() {
	encoded := vcodec.ROT13("hello")

	fmt.Println(encoded)
	fmt.Println(vcodec.ROT13(encoded))
	// Output:
	// uryyb
	// hello
}
