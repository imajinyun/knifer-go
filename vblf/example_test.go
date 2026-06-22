package vblf_test

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vblf"
)

func ExampleNewBitSetBloomFilter() {
	bf := vblf.NewBitSetBloomFilter(1000, 100, 3)
	bf.Add("hello")
	bf.Add("world")

	fmt.Println(bf.Contains("hello"))
	fmt.Println(bf.Contains("go"))
	// Output:
	// true
	// false
}

func ExampleNewIntMap() {
	m := vblf.NewIntMap(1000)
	m.Add(42)

	fmt.Println(m.Contains(42))
	fmt.Println(m.Contains(100))
	// Output:
	// true
	// false
}

func ExampleNewBitMapBloomFilter() {
	bf := vblf.NewBitMapBloomFilter(5)
	bf.Add("hello")

	fmt.Println(bf.Contains("hello"))
	// Output: true
}

func ExampleNewFuncFilter() {
	filter := vblf.NewFuncFilter(1000, func(s string) int64 { return int64(len(s)) })
	filter.Add("hello")

	fmt.Println(filter.Contains("hello"))
	// Output: true
}

func ExampleNewFuncFilterWithMachineNumE() {
	filter, err := vblf.NewFuncFilterWithMachineNumE(1000, vblf.BloomMachine64, func(s string) int64 {
		return int64(vblf.JavaDefaultHash(s))
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(filter.Add("hello"))
	fmt.Println(filter.Add("hello"))
	fmt.Println(filter.Contains("hello"))
	// Output:
	// true
	// false
	// true
}

func ExampleNewDefaultFilter() {
	filter := vblf.NewDefaultFilter(1000)
	filter.Add("go-knifer")

	fmt.Println(filter.Contains("go-knifer"))
	// Output: true
}

func ExampleNewLongMap() {
	m := vblf.NewLongMap(100)
	m.Add(7)

	fmt.Println(m.Contains(7))
	fmt.Println(m.Contains(99))
	// Output:
	// true
	// false
}
