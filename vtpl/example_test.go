package vtpl_test

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vtpl"
)

func ExampleRender() {
	result, err := vtpl.Render("Hello, {{.Name}}!", map[string]any{"Name": "World"})
	fmt.Println(result)
	fmt.Println(err)
	// Output:
	// Hello, World!
	// <nil>
}
