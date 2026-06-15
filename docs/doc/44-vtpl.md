# vtpl Quickstart

`vtpl` provides Go `html/template` based string rendering facades, with support for template names, function maps, custom delimiters, and parse/execute provider injection.

## Render simple templates

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vtpl"
)

func main() {
	out, err := vtpl.Render("hello {{.Name}}", map[string]string{"Name": "tpl"})
	if err != nil {
		panic(err)
	}
	fmt.Println(out)
}
```

## Use FuncMap

```go
package main

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/imajinyun/go-knifer/vtpl"
)

func main() {
	out, err := vtpl.RenderWithOptions(
		"{{upper .Name}}",
		map[string]string{"Name": "go"},
		vtpl.WithFuncMap(template.FuncMap{"upper": strings.ToUpper}),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(out)
}
```

## Use custom delimiters and template names

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vtpl"
)

func main() {
	out, err := vtpl.RenderWithOptions(
		"hi [[.Name]]",
		map[string]string{"Name": "knifer"},
		vtpl.WithTemplateName("greeting"),
		vtpl.WithDelims("[[", "]]"),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(out)
}
```

## Stay compatible with RenderTemplate

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vtpl"
)

func main() {
	out, err := vtpl.RenderTemplate("{{.Lang}} quickstart", map[string]string{"Lang": "Go"})
	if err != nil {
		panic(err)
	}
	fmt.Println(out)
}
```
