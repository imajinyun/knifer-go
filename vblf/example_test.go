package vblf_test

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vblf"
)

func ExampleBitSetBloomFilter() {
	bf := vblf.NewBitSetBloomFilter(1000, 100, 3)
	bf.Add("hello")
	bf.Add("world")

	fmt.Println(bf.Contains("hello"))
	fmt.Println(bf.Contains("go"))
	// Output:
	// true
	// false
}

func ExampleIntMap() {
	m := vblf.NewIntMap(1000)
	m.Add(42)

	fmt.Println(m.Contains(42))
	fmt.Println(m.Contains(100))
	// Output:
	// true
	// false
}
