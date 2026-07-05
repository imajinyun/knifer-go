# vtpl Quickstart

`vtpl` provides Go `html/template` based string rendering facades, with support for template names, function maps, custom delimiters, and parse/execute provider injection. It also exposes an engine-neutral adapter contract so callers can select the standard HTML engine, the standard text engine, or a custom template engine without adding optional dependencies to knifer-go.

## Golden path APIs

The first-choice API set for this facade is kept in sync with `ai-context.json` and the generated tools catalog.

- `Render`
- `NewHTMLEngine`
- `WithFuncMap`
- `NewTextEngine`
- `RenderTemplate`

## Which helper should I use?

| Goal | Start with | Notes |
| --- | --- | --- |
| Render a small HTML-safe string | `Render` | Uses the default HTML template behavior and escapes data for HTML contexts. |
| Add functions, names, or delimiters | `RenderWithOptions` | Use `WithFuncMap`, `WithTemplateName`, and `WithDelims` for one-off render customization. |
| Keep compatibility with older call sites | `RenderTemplate` | Alias-style facade for existing users that do not need new engine selection. |
| Choose HTML escaping explicitly | `NewHTMLEngine` + `RenderWithEngine` | Use for output that will be embedded in HTML. |
| Render trusted plain text | `NewTextEngine` + `RenderWithEngine` | Use only when raw text output is intended and HTML escaping is not desired. |
| Integrate a custom engine | `Engine` or `EngineFunc` | Keeps optional third-party engines outside the base module dependency graph. |
| Test parse or execute errors | `WithTemplateFactory`, `WithTemplateParser`, `WithTemplateExecutor` | Provider injection makes error paths deterministic without brittle template strings. |

## Template safety checklist

- Use `Render` or `NewHTMLEngine` for HTML output so `html/template` escaping is preserved.
- Use `NewTextEngine` only for trusted non-HTML destinations such as config snippets, logs, or plain text emails.
- Treat template source as code. Do not render untrusted templates unless the caller owns the allowed functions and execution context.
- Keep `FuncMap` functions side-effect free where possible; they run during rendering and can affect latency or safety.
- Use context-aware `RenderWithEngine` when a custom engine may block or call external systems.
- Inject parser/executor providers in tests instead of relying on fragile malformed template strings for every error path.

## Render simple templates

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vtpl"
)

func main() {
	out, err := vtpl.Render("hello {{.Name}}", map[string]string{"Name": "tpl"})
	if err != nil {
		panic(err)
	}
	fmt.Println(out)
}
```

## When not to use vtpl

- Use a full web framework renderer when templates are loaded from files, cached, reloaded, localized, or composed with layouts.
- Use `text/template` directly when the project needs full control over parsing, template sets, and execution lifecycle.
- Avoid `NewTextEngine` for browser-bound HTML or attributes; it does not apply `html/template` contextual escaping.
- Avoid executing user-supplied templates unless the application constrains functions, data, and runtime budget.

## Related packages

- Use `vmail` when rendered templates become email bodies or MIME messages.
- Use `vjson` when template data or fixtures need JSON formatting and path inspection.
- Use `vstr` when template inputs need string normalization or HTML escaping before rendering.

## Benchmarks and trade-offs

- `Render` and `RenderWithOptions` parse the source for each call, which is convenient for small dynamic snippets but slower than reusing parsed templates.
- `NewHTMLEngine` protects HTML contexts through escaping; `NewTextEngine` avoids that escaping for trusted text and can preserve raw output.
- Custom engines add abstraction and testability, but their performance and safety depend on the adapter implementation.
- Provider injection for parser and executor paths adds setup but makes failure-mode tests precise.

## FAQ

### Why is HTML escaped by default?

The default facade uses `html/template` semantics because rendered strings often end up in HTML. Use `NewTextEngine` only when the destination is trusted plain text.

### Can templates be cached?

This facade renders strings directly and favors simple call sites. For high-volume repeated templates, build a custom `Engine` that owns parsed-template caching.

### When should I use `RenderWithEngine`?

Use it when callers must choose between HTML, text, or a custom engine, or when a context should control a custom render operation.

### Are `FuncMap` functions safe to call?

They are as safe as the functions you register. Avoid functions with network, filesystem, or mutation side effects unless rendering is explicitly allowed to perform them.

## Use FuncMap

```go
package main

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/imajinyun/knifer-go/vtpl"
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

	"github.com/imajinyun/knifer-go/vtpl"
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

	"github.com/imajinyun/knifer-go/vtpl"
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

	"github.com/imajinyun/knifer-go/vtpl"
)

func main() {
	out, err := vtpl.RenderTemplate("{{.Lang}} quickstart", map[string]string{"Lang": "Go"})
	if err != nil {
		panic(err)
	}
	fmt.Println(out)
}
```
