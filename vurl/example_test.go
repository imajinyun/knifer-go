package vurl_test

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vurl"
)

func ExampleEncode() {
	fmt.Println(vurl.Encode("a b&c"))
	// Output: a+b%26c
}

func ExampleEncodeQueryMap() {
	fmt.Println(vurl.EncodeQueryMap(map[string]any{"q": "go"}))
	// Output: q=go
}

func ExampleIsHTTPURL() {
	fmt.Println(vurl.IsHTTPURL("http://example.com"))
	fmt.Println(vurl.IsHTTPURL("ftp://example.com"))
	// Output:
	// true
	// false
}

func ExampleDecode() {
	decoded, err := vurl.Decode("a+b%26c")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(decoded)
	// Output: a b&c
}

func ExampleComplete() {
	full, err := vurl.Complete("https://example.com/docs/", "api/tools.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(full)
	// Output: https://example.com/docs/api/tools.json
}
