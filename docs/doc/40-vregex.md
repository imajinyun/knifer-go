# vregex Quickstart

`vregex` provides regex matching, finding, capture-group extraction, replacement, deletion, template-variable extraction, and special-character escaping helpers.

## Which helper should I use?

Choose helpers based on whether you need a boolean, a match list, captured data, replacement, or escaped literal text.

| Need | Use | Notes |
| --- | --- | --- |
| Boolean match check | `Match`, `IsMatch` | Good for validation and branch decisions when the pattern is simple and bounded. |
| Find matches or count occurrences | `FindAll`, `Count`, `FindWithOptions` | Use options such as `WithDotAll` when multiline behavior is intentional. |
| Extract capture groups | `GetGroup1`, `Get`, `GetByName`, `FindAllGroup` | Prefer named groups when the pattern is maintained by humans and group positions may change. |
| Replace or delete matched text | `Replace`, `ReplaceFirst`, `ReplaceAllFunc`, `DelAll` | Use function replacement when the output depends on each match. |
| Extract template variables | `ExtractMulti`, `TemplateVars` | Keep replacement templates close to the regex that defines their group meanings. |
| Treat user text as a literal pattern fragment | `Escape` | Escape user-controlled literals before composing larger regex patterns. |

## Regex safety checklist

- Keep patterns small, reviewed, and close to their expected input shape. Regexes are easy to over-generalize.
- Escape user-controlled literal fragments with `Escape` before combining them into a regex pattern.
- Prefer named capture groups for long-lived patterns so later edits do not silently change group indexes.
- Treat regex validation as one layer, not a complete parser, for structured formats such as URLs, JSON, XML, or programming languages.
- Benchmark or bound inputs for hot paths and large text. Even safe Go regexes can still consume CPU on broad scans.
- Use `WithDotAll` only when matching across newlines is intended; it can greatly widen what a pattern matches.

## Related packages

- Use `vstr` when exact string trimming, splitting, prefix/suffix checks, or naming conversion is sufficient.
- Use `vform` when regex checks are part of a broader field validation workflow.
- Use `vdfa` when text filtering should use dictionary matching instead of pattern matching.

## When not to use vregex

- Use compiled `regexp.Regexp` values directly when the same pattern runs repeatedly in a hot loop.
- Use dedicated parsers for structured formats such as URLs, JSON, XML, HTML, SQL, or programming languages.
- Use exact string helpers from `strings` when the pattern is a simple prefix, suffix, contains, split, or replacement operation.
- Avoid composing unescaped user input into regex patterns; use `Escape` for literal fragments or choose a non-regex API.
- Avoid broad regex scans over unbounded text without input limits, time budgeting, or workload benchmarks.

## Benchmarks and trade-offs

Benchmark with representative pattern complexity, input length, capture groups, and replacement behavior:

```bash
go test -bench=. -benchmem -run=^$ ./internal/regex ./vregex
```

`vregex` helpers are concise for one-off matching and extraction. Repeated hot-path matching should usually compile patterns with `regexp` and reuse the compiled value to avoid repeated parse work.

Capture groups, named groups, DotAll behavior, and replacement callbacks add clarity but can increase CPU and allocation costs. Keep broad scans bounded when inputs are large or user-controlled.

## FAQ

### Does vregex replace regexp?

No. `vregex` provides common helpers around regex workflows. Use `regexp` directly when you need compiled pattern reuse, advanced control, or lower-level APIs.

### Should I compile regexes myself?

Compile patterns yourself with `regexp` when the same pattern is used repeatedly in a hot path. Use `vregex` helpers for concise one-off or low-volume workflows.

### Are regexes enough for validating structured data?

Usually no. Regex can check simple shapes, but structured formats should use dedicated parsers or validators after any lightweight regex precheck.

## Match, find, and count

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vregex"
)

func main() {
	fmt.Println(vregex.Match(`\d+`, "abc123"))
	fmt.Println(vregex.IsMatch(`\d+`, "123"))
	fmt.Println(vregex.FindAll(`\d+`, "a1b22"))
	fmt.Println(vregex.Count(`\d+`, "a1b22"))
}
```

## Extract capture groups and named groups

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vregex"
)

func main() {
	fmt.Println(vregex.GetGroup1(`user:(\w+)`, "user:alice"))
	fmt.Println(vregex.Get(`(a)(b)`, "ab", 2))
	fmt.Println(vregex.GetByName(`(?<word>\w+)-(?<num>\d+)`, "abc-123", "num"))
	fmt.Println(vregex.FindAllGroup(`x(\d+)`, "x1x22", 1))
}
```

## Replace and delete matches

```go
package main

import (
	"fmt"
	"strings"

	"github.com/imajinyun/knifer-go/vregex"
)

func main() {
	fmt.Println(vregex.Replace(`\d+`, "a1b22", "#"))
	fmt.Println(vregex.ReplaceFirst(`\d+`, "a1b22", "#"))
	fmt.Println(vregex.ReplaceAllFunc("a1b22", `\d+`, func(m vregex.MatchResult) string {
		return strings.Repeat("*", len(m.Text))
	}))
	fmt.Println(vregex.DelAll(`\d+`, "a1b22"))
}
```

## Template extraction, DotAll, and escaping

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vregex"
)

func main() {
	fmt.Println(vregex.ExtractMulti(`(\w+)-(\d+)`, "abc-123", "$2:$1"))
	fmt.Println(vregex.TemplateVars("$2:$1"))
	fmt.Println(vregex.FindWithOptions(`a.*c`, "a\nb\nc", vregex.WithDotAll(true)))
	fmt.Println(vregex.Escape("a+b*c?"))
}
```
