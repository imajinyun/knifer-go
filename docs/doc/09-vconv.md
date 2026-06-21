# vconv Quickstart

`vconv` provides loose type conversion helpers that convert common inputs to string, int, int64, float64, bool, and []byte. Each scalar family has zero-value helpers, default-value helpers, and explicit-error `E` helpers for code that must distinguish invalid input from a valid zero value.

## Which helper should I use?

Choose the conversion family by how much ambiguity the caller can tolerate.

| Need | Use | Notes |
| --- | --- | --- |
| Best-effort conversion with a zero fallback | `ToString`, `ToInt`, `ToInt64`, `ToFloat64`, `ToBool`, `ToBytes` | Good for display, logs, compatibility shims, and legacy permissive behavior. |
| Best-effort conversion with caller fallback | `ToXxxDefault`, `ToXxxDefaultWithOptions` | Use when zero is a valid value and the fallback should be visible. |
| Conversion that must reject invalid input | `ToIntE`, `ToInt64E`, `ToFloat64E`, `ToBoolE` | Prefer at trust boundaries and when invalid input must be reported. |
| Custom parse or format policy | `WithParseIntFunc`, `WithParseFloatFunc`, `WithBoolParser`, formatter options | Inject deterministic parsers/formatters in tests or domain-specific conversions. |
| Byte conversion | `ToBytes`, `ToBytesWithOptions` | Remember existing `[]byte` values may be returned directly. Clone if caller mutation matters. |

## Conversion correctness checklist

- Use explicit-error `E` helpers at API, configuration, database, and user-input boundaries.
- Avoid zero-value helpers when zero, false, or empty string are valid business values; use default or error-returning helpers instead.
- Review numeric truncation: permissive string-to-int conversion can turn float strings into integers by truncating toward zero.
- Reject or handle `NaN`, `Inf`, overflow, and unsigned-to-signed conversions where correctness matters.
- Treat `ToBytes` results as potentially shared when the input is already `[]byte`; make a defensive copy before mutation.
- Keep custom parser options local to the call site or constructor so conversion policy is easy to audit.

## When not to use vconv

- Use `strconv` directly when a strict grammar, bit size, or parse error is part of the contract.
- Use typed decoding such as JSON, YAML, or database scanning when source structure is known.
- Use domain-specific validators before conversion when input carries business meaning such as currency, age, or identifiers.
- Avoid permissive conversion for authorization, quota, payment, or security policy decisions.

## Related packages

- Use `vbean` when scalar conversions are part of struct-to-struct or map-to-struct binding.
- Use `vnum` when numeric parsing needs decimal precision, rounding, or expression support.
- Use `vdate` when string conversion involves calendar parsing, formatting, or timezone rules.

## Benchmarks and trade-offs

Benchmark conversion paths when they sit in hot deserialization or logging loops:

```bash
go test -bench=. -benchmem -run=^$ ./internal/conv ./vconv
```

Loose conversion saves boilerplate but pays for type switches, reflection-like handling, string formatting, and fallback logic. Direct `strconv` calls are clearer and usually cheaper when the input type and grammar are known.

Error-returning helpers add branches but improve correctness by separating invalid input from valid zero values. Prefer them at boundaries even when the permissive helper would be shorter.

## Conversion contract

- `ToXxx` helpers return the destination type zero value when conversion fails.
- `ToXxxDefault` helpers return the caller-provided fallback when conversion fails.
- `ToIntE`, `ToInt64E`, `ToFloat64E`, and `ToBoolE` return `ErrInvalidConversion` and match `knifer.ErrCodeInvalidInput` when conversion fails.
- String-to-int conversion trims spaces, tries integer parsing first, then accepts float strings by truncating toward zero, so `"42.9"` becomes `42`.
- `E` integer helpers reject uint, float, `NaN`, and `Inf` inputs that cannot fit in the destination integer type; the legacy zero/default helpers keep their permissive conversion behavior for backward compatibility.
- Named string and numeric types are accepted by the same scalar conversion rules as their underlying types.
- Bool conversion accepts `true`, `yes`, `y`, `ok`, `1`, `on`, `false`, `no`, `n`, `0`, and `off` case-insensitively after trimming spaces. Non-string numerics convert to `true` when nonzero.
- `ToBytes` returns `nil` for `nil`, returns an existing `[]byte` as-is, converts strings directly, and stringifies other values.

## FAQ

### Why do plain helpers return zero values on failure?

They preserve the package's permissive compatibility behavior. Use `ToXxxDefault` when the fallback should be explicit, or `ToXxxE` when invalid input must be handled.

### Should I use vconv for request validation?

Only with the `E` helpers and domain validation around them. Do not silently coerce untrusted request values when the caller must know whether input was invalid.

### Does ToBytes copy byte slices?

No. Existing `[]byte` inputs may be returned as-is. Clone the returned slice before mutation if the original owner must not observe changes.

## Convert to numbers

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vconv"
)

func main() {
	fmt.Println(vconv.ToInt("42"))
	fmt.Println(vconv.ToIntDefault("bad", 7))
	fmt.Println(vconv.ToFloat64("3.14"))
}
```

## Return explicit conversion errors

```go
package main

import (
	"errors"
	"fmt"

	knifer "github.com/imajinyun/go-knifer"
	"github.com/imajinyun/go-knifer/vconv"
)

func main() {
	value, err := vconv.ToInt64E("42.9")
	fmt.Println(value, err)

	_, err = vconv.ToBoolE("maybe")
	fmt.Println(errors.Is(err, vconv.ErrInvalidConversion))
	fmt.Println(errors.Is(err, knifer.ErrCodeInvalidInput))
}
```

## Convert to bool

```go
package main

import (
	"fmt"
	"strings"

	"github.com/imajinyun/go-knifer/vconv"
)

func main() {
	fmt.Println(vconv.ToBool("true"))
	fmt.Println(vconv.ToBoolWithOptions("YES", vconv.WithBoolParser(func(s string) (bool, error) {
		return strings.EqualFold(s, "yes"), nil
	})))
}
```

## Convert to strings

```go
package main

import (
	"fmt"
	"strconv"

	"github.com/imajinyun/go-knifer/vconv"
)

func main() {
	fmt.Println(vconv.ToString(123))
	fmt.Println(vconv.ToStringDefault(nil, "fallback"))
	fmt.Println(vconv.ToStringWithOptions(true, vconv.WithFormatBoolFunc(strconv.FormatBool)))
}
```

## Convert to byte slices

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vconv"
)

func main() {
	b := vconv.ToBytes("hello")
	fmt.Println(string(b))
}
```
