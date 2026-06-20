# vtpl Quickstart

`vtpl` provides Go `html/template` based string rendering facades, with support for template names, function maps, custom delimiters, and parse/execute provider injection. It also exposes an engine-neutral adapter contract so callers can select the standard HTML engine, the standard text engine, or a custom template engine without adding optional dependencies to go-knifer.

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

## Select a template engine explicitly

Use `NewHTMLEngine` when rendered output is HTML and should keep `html/template` escaping. Use `NewTextEngine` for trusted non-HTML text where raw output is expected. `RenderWithEngine` is context-first and returns classified invalid-input errors for missing engines or invalid render requests.

```go
package main

import (
	"context"
	"fmt"

	"github.com/imajinyun/go-knifer/vtpl"
)

func main() {
	textEngine := vtpl.NewTextEngine()
	out, err := vtpl.RenderWithEngine(context.Background(), textEngine, "{{.}}", "<raw>")
	if err != nil {
		panic(err)
	}
	fmt.Println(out)
}
```

Custom adapters implement `Engine`, or use `EngineFunc` for deterministic tests and optional third-party engines:

```go
engine := vtpl.EngineFunc(func(ctx context.Context, req vtpl.RenderRequest) (string, error) {
	return "custom: " + req.Source, ctx.Err()
})
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
