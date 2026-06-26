package vtpl_test

import (
	"context"
	"fmt"
	"html/template"
	"strings"

	"github.com/imajinyun/knifer-go/vtpl"
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
	result, err := vtpl.Render("<p>{{.}}</p>", "<knifer-go>")

	fmt.Println(result)
	fmt.Println(err)
	// Output:
	// <p>&lt;knifer-go&gt;</p>
	// <nil>
}

func ExampleRenderWithEngine_textEngine() {
	engine := vtpl.NewTextEngine()
	result, err := vtpl.RenderWithEngine(context.Background(), engine, "{{.}}", "<knifer-go>")

	fmt.Println(result)
	fmt.Println(err)
	// Output:
	// <knifer-go>
	// <nil>
}

func ExampleRenderWithEngine_htmlEngine() {
	engine := vtpl.NewHTMLEngine()
	result, err := vtpl.RenderWithEngine(context.Background(), engine, "{{.}}", "<knifer-go>")

	fmt.Println(result)
	fmt.Println(err)
	// Output:
	// &lt;knifer-go&gt;
	// <nil>
}

func ExampleRenderWithEngine_customEngine() {
	engine := vtpl.EngineFunc(func(ctx context.Context, req vtpl.RenderRequest) (string, error) {
		return "custom: " + req.Source, ctx.Err()
	})
	result, err := vtpl.RenderWithEngine(context.Background(), engine, "template source", nil)

	fmt.Println(result)
	fmt.Println(err)
	// Output:
	// custom: template source
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
