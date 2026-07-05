# vdfa Quickstart

`vdfa` provides word-tree based sensitive-word matching and filtering, with support for package-level dictionaries, local matchers, JSON object filtering, and custom character filters.

## Golden path APIs

The first-choice API set for this facade is kept in sync with `ai-context.json` and the generated tools catalog.

- `FilterAny`
- `ContainsAnyWithOptions`
- `Contains`
- `ConfigureAsyncRunner`
- `ContainsAny`

## Which helper should I use?

Choose the helper by matcher ownership, input shape, and whether you need detection only or replacement.

| Need | Use | Notes |
| --- | --- | --- |
| Initialize package-level words | `Init`, `InitWithOptions`, `InitString` | Good for application-wide dictionaries loaded at startup. |
| Build an independent matcher | `NewWordTree`, `NewWordTreeWithOptions` | Prefer for tests, tenants, or request flows that should not mutate package state. |
| Detect text matches | `Contains`, `GetFoundFirst`, `GetFoundAll` | Use found-word helpers when offsets or matched words are needed. |
| Override matcher per call | `ContainsWithOptions`, `FilterWithOptions`, `WithMatcher`, `WithMatcherWords` | Keeps local dictionaries explicit without replacing global state. |
| Filter text with custom replacement | `FilterMode`, `FilterModeWithOptions`, `DefaultProcessor` | Use a custom `Processor` when redaction must preserve shape or include audit markers. |
| Scan or filter structured values | `ContainsAnyWithOptions`, `FilterAnyWithOptions`, JSON options | Inject JSON marshal/unmarshal functions for deterministic tests and custom encoders. |
| Initialize asynchronously | `InitAsync`, `InitStringAsync`, `ConfigureAsyncRunner` | Configure the runner in tests so async initialization is deterministic. |

## DFA correctness checklist

- Prefer independent `WordTree` values for tests and multi-tenant flows; package-level initialization is shared process state.
- Complete dictionary initialization before serving requests, or make callers handle the not-yet-initialized window explicitly.
- Keep character filtering policy stable. Changing stop-character rules changes what text matches.
- Decide whether greedy and dense matching modes are required before relying on found-word offsets.
- Treat `FilterAny` as a JSON round trip: only exported fields and marshalable values participate.
- Avoid logging raw sensitive text while debugging match or replacement behavior.

## When not to use vdfa

- Use a dedicated moderation or NLP service when matching depends on language context, morphology, machine learning, or policy review workflows.
- Use exact string or regexp matching when the dictionary is tiny and DFA behavior would make the code harder to understand.
- Use independent matchers instead of package-level helpers when dictionaries differ by tenant, user, request, or test case.
- Avoid `FilterAny` for values that cannot safely round-trip through JSON or that contain fields that should not be rewritten.
- Avoid async initialization unless the application has a clear readiness gate.

## Related packages

- Use `vstr` when the task is general string normalization, trimming, or similarity rather than word filtering.
- Use `vregex` when pattern-based extraction or replacement is more appropriate than dictionary matching.
- Use `vtok` when text should be tokenized or ranked by an injected NLP provider.

## Benchmarks and trade-offs

Use local benchmarks to compare dictionary size, text length, matching mode, and JSON filtering overhead:

```bash
go test -bench=. -benchmem -run=^$ ./internal/dfa ./vdfa
```

Building a word tree is a startup cost; matching is optimized for repeated scans after initialization. Package-level helpers are convenient, but independent `WordTree` values avoid global-state surprises in tests and multi-tenant applications.

`FilterAny` is convenient for structured data, but it pays marshal/unmarshal cost and follows JSON field visibility rules. Prefer text helpers when the caller already has the target string.

## FAQ

### Should I use package-level helpers or a local WordTree?

Use package-level helpers for one process-wide dictionary loaded at startup. Use `NewWordTree` for tests, tenant-specific dictionaries, or code that should not mutate shared state.

### What do greedy and dense matching change?

They control how overlapping or nested words are reported or replaced. Choose the mode based on product policy and test representative overlaps such as `foo` and `foobar`.

### Is FilterAny safe for every struct?

No. It uses JSON marshal/unmarshal behavior, so unexported fields, custom marshalers, unsupported values, and lossy conversions matter. Use text helpers when only one field should be filtered.

## Initialize the default dictionary and filter text

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vdfa"
)

func main() {
	vdfa.Init([]string{"badword", "secret"})

	fmt.Println(vdfa.Contains("a badword appears"))
	fmt.Println(vdfa.Filter("keep the secret"))
}
```

## Use an independent WordTree

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vdfa"
)

func main() {
	tree := vdfa.NewWordTree().AddWords("foo", "foobar")

	found, ok := tree.MatchWord("say foobar now")
	if ok {
		fmt.Println(found.Word, found.Start, found.End)
	}
	fmt.Println(tree.MatchAll("foo and foobar"))
}
```

## Customize character filtering rules

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vdfa"
)

func main() {
	tree := vdfa.NewWordTreeWithOptions(
		vdfa.WithCharFilter(func(r rune) bool { return r != '-' }),
	).AddWord("t-io")

	fmt.Println(tree.IsMatch("tio"))
	fmt.Println(vdfa.ContainsWithOptions("a local word", vdfa.WithMatcherWords([]string{"local"})))
}
```

## Filter JSON content in structs

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vdfa"
)

type Payload struct {
	Text string `json:"text"`
}

func main() {
	vdfa.Init([]string{"secret"})

	got, err := vdfa.FilterAny(Payload{Text: "a secret"}, true, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(got.Text)
}
```
