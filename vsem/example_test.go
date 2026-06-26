package vsem_test

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vsem"
)

func ExampleNew() {
	s := vsem.New(3)
	fmt.Println(s.Cap())
	// Output: 3
}

func ExampleNewE() {
	s, err := vsem.NewE(2)

	fmt.Println(s.TryAcquire(1), s.Use())
	s.Release(1)
	fmt.Println(s.Use())
	fmt.Println(err)
	// Output:
	// true 1
	// 0
	// <nil>
}
