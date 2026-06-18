package vsem_test

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vsem"
)

func ExampleNew() {
	s := vsem.New(3)
	fmt.Println(s.Cap())
	// Output: 3
}
