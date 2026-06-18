package vcron_test

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vcron"
)

func ExampleNewPattern() {
	p, err := vcron.NewPattern("* * * * *")
	fmt.Println(p != nil)
	fmt.Println(err)
	// Output:
	// true
	// <nil>
}
