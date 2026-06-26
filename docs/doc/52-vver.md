# vver Quickstart

`vver` provides version comparison and version-expression matching helpers for checking whether a current version satisfies single, multiple, or custom-delimiter expressions.

## Which helper should I use?

Choose the helper by whether you need ordering, a single expression, or a set of acceptable expressions.

| Need | Use | Notes |
| --- | --- | --- |
| Compare two versions | `CompareVersion` | Returns ordering as an integer; useful when callers need custom branching. |
| Readable relational checks | `IsGreaterThan`, `IsGreaterThanOrEqual`, `IsLessThan`, `IsLessThanOrEqual` | Prefer when a single comparison expresses the policy. |
| Match one expression | `MatchEl`, `MatchElWithDelimiter` | Supports relation and range-style expressions used by the package. |
| Surface invalid expression errors | `MatchElWithDelimiterErr` | Use when invalid policy should fail loudly instead of returning false. |
| Match any acceptable expression | `AnyMatch`, `AnyMatchSlice` | Good for allow-lists, compatibility windows, and rollout gates. |
| Use custom expression separators | `MatchElWithDelimiter`, `MatchElByDelimiter`, `DefaultVersionsDelimiter` | Keep delimiter choice explicit when expressions come from config. |

## Version correctness checklist

- Document the version grammar accepted by the caller; do not assume every SemVer feature is supported.
- Use error-returning helpers for configuration, rollout rules, or policy expressions that must be valid.
- Normalize version strings at system boundaries so prefixes, whitespace, and build metadata are handled consistently.
- Add tests for boundary versions such as exact lower bound, exact upper bound, missing segments, and invalid expressions.
- Keep custom delimiters distinct from characters used inside version numbers or ranges.
- Avoid string comparison for versions; always parse through the version helpers or a stricter SemVer library.

## When not to use vver

- Use a dedicated SemVer library when you need strict SemVer parsing, prerelease ordering, build metadata rules, or constraint language compatibility.
- Use package-manager-specific constraint libraries when matching npm, Maven, Go modules, or other ecosystem semantics.
- Use feature flags or server-side rollout systems when version checks are only one part of a deployment decision.
- Avoid loose version matching for security patch enforcement unless the accepted grammar is tested and reviewed.

## Related packages

- Use `vstr` when version strings first need trimming, splitting, or normalization.
- Use `vconv` when version values are mixed with broader loose conversion workflows.
- Use `vconf` when version constraints come from layered configuration or release policy files.

## Benchmarks and trade-offs

Version helpers are usually used at startup, configuration load, or rollout boundaries. Benchmark only if matching is in a hot request path:

```bash
go test -bench=. -benchmem -run=^$ ./internal/version ./vver
```

Loose helpers are convenient for application-specific version strings and range expressions, but strict ecosystem compatibility may require a specialized parser. Error-returning helpers make invalid policy visible and are worth the extra handling in configuration paths.

## FAQ

### Is vver a full SemVer constraint engine?

No. It provides lightweight comparison and expression helpers. Use a SemVer library when prerelease, metadata, and ecosystem-specific constraint syntax matter.

### Why use `MatchElWithDelimiterErr`?

Use it when version expressions come from config or rollout policy and invalid syntax should be reported rather than treated as a non-match.

### Should I compare versions as strings?

No. Lexicographic string comparison gives wrong results for values such as `1.10.0` and `1.2.0`. Use version helpers or a strict parser.

## Compare two versions

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vver"
)

func main() {
	fmt.Println(vver.CompareVersion("1.0.0", "1.0.2") < 0)
	fmt.Println(vver.CompareVersion("1.2.0", "1.2.0") == 0)
	fmt.Println(vver.CompareVersion("2.0.0", "1.9.9") > 0)
}
```

## Use relational predicate helpers

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vver"
)

func main() {
	fmt.Println(vver.IsGreaterThan("1.0.3", "1.0.2"))
	fmt.Println(vver.IsGreaterThanOrEqual("1.0.2", "1.0.2"))
	fmt.Println(vver.IsLessThan("1.0.1", "1.0.2"))
	fmt.Println(vver.IsLessThanOrEqual("1.0.2", "1.0.2"))
}
```

## Match version expressions

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vver"
)

func main() {
	fmt.Println(vver.MatchEl("1.0.2", ">=1.0.0"))
	fmt.Println(vver.MatchEl("1.0.2", "1.0.1-1.1.0"))
	fmt.Println(vver.MatchElWithDelimiter("1.0.2", "<1.0.1,1.0.2", ","))
}
```

## Match any of multiple expressions

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vver"
)

func main() {
	fmt.Println(vver.AnyMatch("1.0.2", "<1.0.1", "1.0.2"))
	fmt.Println(vver.AnyMatchSlice("1.0.2", []string{"<1.0.1", ">=1.0.0"}))
	fmt.Println(vver.MatchElWithDelimiterErr("1.0.2", ">=1.0.0", ";") == nil)
}
```
