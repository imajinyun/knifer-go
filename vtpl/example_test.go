package vtpl_test

import (
	"fmt"
	"html/template"
	"strings"

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

func ExampleRenderTemplate() {
	result, err := vtpl.RenderTemplate("{{.Greeting}}, {{.Name}}!", map[string]string{
		"Greeting": "Hi",
		"Name":     "Gopher",
	})

	fmt.Println(result)
	fmt.Println(err)
	// Output:
	// Hi, Gopher!
	// <nil>
}

func ExampleRenderWithOptions() {
	result, err := vtpl.RenderWithOptions(
		"Hello [[upper .Name]]",
		map[string]string{"Name": "gopher"},
		vtpl.WithDelims("[[", "]]"),
		vtpl.WithFuncMap(template.FuncMap{"upper": strings.ToUpper}),
	)

	fmt.Println(result)
	fmt.Println(err)
	// Output:
	// Hello GOPHER
	// <nil>
}

func ExampleRender_htmlEscaping() {
	result, err := vtpl.Render("<p>{{.}}</p>", "<go-knifer>")

	fmt.Println(result)
	fmt.Println(err)
	// Output:
	// <p>&lt;go-knifer&gt;</p>
	// <nil>
}

func ExampleRender_parseError() {
	result, err := vtpl.Render("Hello, {{.Name", map[string]string{"Name": "Gopher"})

	fmt.Println(result == "")
	fmt.Println(err != nil)
	// Output:
	// true
	// true
}

func ExampleWithTemplateName() {
	fmt.Println(vtpl.WithTemplateName("email") != nil)
	// Output: true
}

func ExampleWithFuncMap() {
	fmt.Println(vtpl.WithFuncMap(template.FuncMap{"upper": strings.ToUpper}) != nil)
	// Output: true
}
