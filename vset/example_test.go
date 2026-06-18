package vset_test

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vset"
)

func ExampleNewString() {
	s := vset.NewString("a", "b", "a")
	fmt.Println(len(s.Members()))
	fmt.Println(s.Contains("a"))
	fmt.Println(s.Contains("c"))
	// Output:
	// 2
	// true
	// false
}

func ExampleNewInt() {
	s := vset.NewInt(1, 2, 2, 3)
	fmt.Println(len(s.Members()))
	fmt.Println(s.Contains(2))
	fmt.Println(s.Contains(4))
	// Output:
	// 3
	// true
	// false
}
