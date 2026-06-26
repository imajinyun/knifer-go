package vhash_test

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vhash"
)

func ExampleAdditiveHash() {
	fmt.Println(vhash.AdditiveHash("abc", 31))
	// Output: 18
}

func ExampleJavaDefaultHash() {
	// Equivalent to Java String.hashCode.
	fmt.Println(vhash.JavaDefaultHash("a"))
	// Output: 97
}

func ExampleBkdrHash() {
	fmt.Println(vhash.BkdrHash("a"))
	// Output: 97
}

func ExampleDjbHash() {
	fmt.Println(vhash.DjbHash("a"))
	// Output: 177670
}

func ExampleHfHash() {
	fmt.Println(vhash.HfHash("abc"))
	// Output: 888
}
