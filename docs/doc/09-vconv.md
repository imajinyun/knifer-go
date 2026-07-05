# vconv Quickstart

`vconv` provides loose type conversion helpers that convert common inputs to string, int, int64, float64, bool, and []byte. Each scalar family has zero-value helpers, default-value helpers, and explicit-error `E` helpers for code that must distinguish invalid input from a valid zero value.

## Golden path APIs

The first-choice API set for this facade is kept in sync with `ai-context.json` and the generated tools catalog.

- `ToBoolE`
- `ToBoolDefaultWithOptions`
- `ToBool`
- `ToBoolDefault`
- `ToBoolEWithOptions`

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

## Conversion matrix

Use this matrix when choosing between permissive compatibility helpers and explicit-error helpers. The `E` helpers define the reviewed boundary contract; zero/default helpers preserve concise legacy behavior.

| Source kind | String | Int / Int64 | Float64 | Bool | Bytes | Failure contract |
| --- | --- | --- | --- | --- | --- | --- |
| `nil` | `""` | `0` / default; `E` rejects | `0` / default; `E` rejects | `false` / default; `E` rejects | `nil` | `E` helpers return `ErrInvalidConversion` and match `knifer.ErrCodeInvalidInput`. |
| String or named string | returned as-is | `strconv.ParseInt`; float strings accepted by integer helpers and truncated toward zero | `strconv.ParseFloat` | accepted tokens: `true/yes/y/ok/1/on` and `false/no/n/0/off` | raw string bytes | Invalid strings are swallowed by zero/default helpers and rejected by `E` helpers. |
| Signed integer | decimal string | range-checked for destination | exact numeric conversion | nonzero is true | decimal string bytes | `E` helpers reject overflow for the destination integer type. |
| Unsigned integer | decimal string | range-checked for destination | exact numeric conversion when representable | nonzero is true | decimal string bytes | `E` helpers reject unsigned values that cannot fit signed destinations. |
| Float | formatted with `FormatFloat` policy | truncated toward zero when finite and in range | exact numeric conversion | nonzero is true | formatted float bytes | `E` integer helpers reject `NaN`, `Inf`, and out-of-range values. |
| Bool | formatted with bool formatter | `1` or `0` | `1` or `0` | returned as-is | formatted bool bytes | Formatter/parser options are per-call; no global conversion state is used. |
| `[]byte` | string copy of bytes | unsupported by `E`; zero/default helpers fall back | unsupported by `E`; zero/default helpers fall back | unsupported by `E`; zero/default helpers fall back | returned as-is | Clone before mutation when ownership matters; convert to string first when parsing bytes as text. |
| `json.Number` | returned through `fmt.Sprint` | parsed through string rules | parsed through string rules | unsupported unless parser accepts the string form | decimal string bytes | Use `encoding/json.Decoder.UseNumber` plus `E` helpers when numeric precision matters. |
| `time.Duration` | duration string such as `150ms` | nanoseconds as `int64` | nanoseconds as `float64` | nonzero is true | duration string bytes | Use `vdate` or direct `time.ParseDuration` when duration grammar is the main contract. |
| Other values | `fmt.Sprint` | unsupported except through string/default compatibility paths | unsupported except through string/default compatibility paths | unsupported except through string/default compatibility paths | stringified bytes | Prefer explicit typed conversion or domain validation before using generic values. |

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
- `json.Number` and `time.Duration` are covered by the property matrix so their string, numeric, and failure behavior stays stable.
- Bool conversion accepts `true`, `yes`, `y`, `ok`, `1`, `on`, `false`, `no`, `n`, `0`, and `off` case-insensitively after trimming spaces. Non-string numerics convert to `true` when nonzero.
- `ToBytes` returns `nil` for `nil`, returns an existing `[]byte` as-is, converts strings directly, and stringifies other values.
- `internal/conv/conversion_matrix_test.go` is the executable property contract for the table above; update it when adding a new scalar source or target behavior.

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

	"github.com/imajinyun/knifer-go/vconv"
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

	knifer "github.com/imajinyun/knifer-go"
	"github.com/imajinyun/knifer-go/vconv"
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

	"github.com/imajinyun/knifer-go/vconv"
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

	"github.com/imajinyun/knifer-go/vconv"
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

	"github.com/imajinyun/knifer-go/vconv"
)

func main() {
	b := vconv.ToBytes("hello")
	fmt.Println(string(b))
}
```
