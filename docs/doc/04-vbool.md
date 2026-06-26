# vbool Quickstart

`vbool` provides lightweight boolean helpers for negation, integer conversion, and batch AND/OR checks.

## Which helper should I use?

Choose `vbool` when a named helper makes a boolean operation easier to pass, test, or reuse than inline operators.

| Need | Use | Notes |
| --- | --- | --- |
| Invert a boolean value | `Negate` | Useful when passing a predicate-like helper into generic or table-driven code. |
| Convert `bool` to `0` or `1` | `ToInt` | Good for counters, flags, metrics labels, and compact examples. |
| Require every value to be true | `And` | Makes variadic aggregation explicit; document empty-input expectations at the call site. |
| Require at least one value to be true | `Or` | Prefer direct `||` for two local operands; use `Or` when values are already collected. |

## Boolean correctness checklist

- Prefer native `!`, `&&`, and `||` for simple local expressions; helper calls should improve readability or composition.
- Keep empty variadic input behavior in mind when using `And` or `Or` with generated slices.
- Avoid hiding authorization or safety decisions behind long helper argument lists; name intermediate predicates first.
- Use `ToInt` only where `0` and `1` semantics are expected by the downstream system.
- Keep batch checks side-effect free. Compute booleans before calling `And` or `Or` so evaluation order is obvious.

## When not to use vbool

- Use Go operators directly for normal control flow and short-circuiting expressions.
- Use named predicate functions when the business meaning matters more than the boolean aggregation itself.
- Use explicit enums or typed states when a value has more than two meaningful states.
- Avoid `ToInt` when the destination protocol expects strings such as `true`/`false`, `yes`/`no`, or domain-specific status values.

## Related packages

- Use `vconv` when boolean parsing is part of a broader loose-conversion workflow.
- Use `vform` when boolean values need validation alongside other request fields.
- Use `vjson` when boolean defaults or checks are driven by JSON payloads.

## Benchmarks and trade-offs

Boolean helpers are tiny wrappers; readability is usually the only reason to choose them over operators. Measure only if a helper appears in a hot generic loop:

```bash
go test -bench=. -benchmem -run=^$ ./internal/boolean ./vbool
```

The main trade-off is short-circuiting: native `&&` and `||` can skip later expressions, while `And` and `Or` receive already-evaluated arguments. Use helpers when the values already exist, not to replace short-circuit guards.

## FAQ

### Why use helpers instead of `!`, `&&`, and `||`?

Use operators for most code. Helpers are useful in table-driven tests, generic pipelines, examples, or when passing a named operation is clearer than embedding an operator expression.

### Does `And` or `Or` short-circuit?

No. Function arguments are evaluated before the call. If later conditions are expensive or unsafe unless earlier checks pass, use native `&&` or `||`.

### What does `ToInt` return?

`ToInt(true)` returns `1`; `ToInt(false)` returns `0`.

## Negate a bool

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vbool"
)

func main() {
	fmt.Println(vbool.Negate(true))
}
```

## Convert bool to int

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vbool"
)

func main() {
	fmt.Println(vbool.ToInt(true))
	fmt.Println(vbool.ToInt(false))
}
```

## Batch logical AND

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vbool"
)

func main() {
	fmt.Println(vbool.And(true, true, false))
}
```

## Batch logical OR

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vbool"
)

func main() {
	fmt.Println(vbool.Or(false, false, true))
}
```
