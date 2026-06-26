# vstr Quickstart

`vstr` provides string helpers for blank checks, trimming, substring extraction, splitting, naming-style conversion, emoji/HTML handling, and text similarity calculation.

## Which helper should I use?

Start with the smallest helper that matches the text task: normalization, extraction, naming conversion, safe rendering, or similarity.

| Need | Use | Notes |
| --- | --- | --- |
| Check blank or provide fallback text | `IsBlank`, `HasBlank`, `DefaultIfBlank`, `Trim` | Useful at request, config, and template boundaries where empty strings need normalization. |
| Extract a substring by rune-aware positions or separators | `Sub`, `SubBefore`, `SubAfter` | Keep index assumptions local to the schema or protocol you are parsing. |
| Split, trim, or pad user-facing text | `SplitTrim`, `PadLeft`, `PadRight` | Prefer helpers that make trimming and padding rules explicit. |
| Convert naming styles | `ToCamelCase`, `ToPascalCase`, `ToUnderlineCase` | Good for config keys, generated names, and UI-friendly rewrites. |
| Add or remove prefixes/suffixes safely | `AddPrefixIfNot`, `RemoveSuffix`, similar helpers | Prefer these over open-coded concatenation when idempotence matters. |
| Escape or clean rendered text | `EscapeHTML`, emoji helpers | Use escaping at output boundaries, not as a substitute for broader input validation. |
| Compare text similarity | `LevenshteinDistance`, `JaccardSimilarity` | Useful for fuzzy matching, ranking, or heuristics, not strict identity checks. |

## Text handling checklist

- Normalize blank input at the boundary so downstream code does not need to repeatedly guess whether `""`, whitespace, or tabs are meaningful.
- Prefer rune-aware substring helpers for human text; byte indexing can split multi-byte characters.
- Escape HTML at render time when content is inserted into HTML output.
- Treat similarity helpers as heuristics. Keep exact authorization, deduplication, and identity checks on strict comparisons.
- Keep naming-style conversion close to the protocol or schema that requires it so casing rules remain reviewable.
- Be explicit about trimming and separator rules; subtle whitespace assumptions are a common source of bugs.

## Related packages

- Use `vregex` when matching, capture groups, or replacement require reviewed regular expressions.
- Use `vdfa` when text filtering should use dictionary-based word matching.
- Use `vtok` or `vhan` when text processing depends on injected NLP, tokenization, or pinyin providers.

## When not to use vstr

- Use the standard `strings`, `unicode`, or `html` packages directly when they express the operation clearly and no facade helper improves readability.
- Use exact comparisons for authentication, authorization, deduplication keys, and correctness-critical identity checks; similarity helpers are only heuristics.
- Use a parser or validator for structured formats instead of substring helpers when syntax, escaping, or nesting matters.
- Use output-context-specific escaping rather than treating `EscapeHTML` as general input validation.
- Avoid naming-style conversion for persisted identifiers unless the casing rules are part of the data contract and covered by tests.

## Benchmarks and trade-offs

Use the string benchmark suite to compare helper overhead on your machine:

```bash
go test -bench=. -benchmem -run=^$ ./vstr
```

The suite covers representative string transformations such as reverse, camel-case conversion, and contains checks. Treat the output as a local baseline rather than a universal performance claim. For hot paths, benchmark the specific helper and input distribution you expect in production.

## FAQ

### Does vstr replace the standard library strings package?

No. `vstr` complements `strings` with convenience helpers for common application workflows. Use the standard library directly when it already expresses the operation clearly.

### Are similarity helpers suitable for security or identity checks?

No. Similarity scores are heuristic ranking signals. Use exact comparisons for authentication, authorization, cache keys, or other correctness-critical checks.

### When should I escape HTML?

Escape when writing text into HTML output or snippets. Do it at the rendering boundary so the context stays obvious and the original text remains available for non-HTML use.

## Blank checks and defaults

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vstr"
)

func main() {
	fmt.Println(vstr.IsBlank("  "))
	fmt.Println(vstr.HasBlank("name", " "))
	fmt.Println(vstr.DefaultIfBlank("\t", "unknown"))
	fmt.Println(vstr.Trim("  go  "))
}
```

## Substrings, splitting, and padding

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vstr"
)

func main() {
	fmt.Println(vstr.Sub("\u4f60\u597d\u4e16\u754c", 1, 3))
	fmt.Println(vstr.SubBefore("a/b/c", "/", false))
	fmt.Println(vstr.SubAfter("a/b/c", "/", true))
	fmt.Println(vstr.SplitTrim(" a, b ,c ", ","))
	fmt.Println(vstr.PadLeft("7", 3, '0'))
}
```

## Naming styles and prefixes/suffixes

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vstr"
)

func main() {
	fmt.Println(vstr.ToCamelCase("hello_world"))
	fmt.Println(vstr.ToPascalCase("hello-world"))
	fmt.Println(vstr.ToUnderlineCase("HelloWorld"))
	fmt.Println(vstr.AddPrefixIfNot("api", "/"))
	fmt.Println(vstr.RemoveSuffix("main.go", ".go"))
}
```

## Emoji, HTML, and text similarity

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vstr"
)

func main() {
	fmt.Println(vstr.ContainsEmoji("go 🚀"))
	fmt.Println(vstr.RemoveEmoji("go 🚀"))
	fmt.Println(vstr.EscapeHTML("<b>go</b>"))
	fmt.Println(vstr.LevenshteinDistance("kitten", "sitting"))
	fmt.Println(vstr.JaccardSimilarity("go fast", "go faster") > 0)
}
```
