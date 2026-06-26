# vnum Quickstart

`vnum` provides number parsing, formatting, exact string arithmetic, rounding, expression evaluation, ranges/random numbers, binary helpers, and common numeric predicates.

## Which helper should I use?

| Goal | Start with | Notes |
| --- | --- | --- |
| Do simple floating-point arithmetic | `Add`, `Sub`, `Mul`, `Div` | Convenient for ordinary `float64` work; remember binary floating-point semantics still apply. |
| Keep decimal input exact | `AddStr`, `SubStr`, `MulStr`, `ToBigDecimal` | String helpers return `*big.Rat` so decimal text such as `0.1` is not rounded through `float64`. |
| Control rounding behavior | `Round`, `RoundMode`, `DivWithMode`, `PowWithMode` | Choose an explicit `RoundingMode` for finance or reporting boundaries. |
| Parse with defaults | `ParseIntDefault`, `ParseDoubleDefault`, `ParseLongDefault` | Good for configuration values where invalid input should fall back instead of failing. |
| Parse and validate with custom providers | `ParseNumberWithOptions`, `IsNumberWithOptions`, `CalculateWithOptions` | Use parser injection when tests or callers need deterministic conversion behavior. |
| Format values for display | `DecimalFormat`, `DecimalFormatMoney`, `FormatPercent`, `ToStrStrip` | Presentation helpers should stay at UI/reporting boundaries, not storage boundaries. |
| Aggregate generic numbers | `Sum`, `Avg`, `Min`, `Max`, `SumNumber`, `AvgNumber` | Generic helpers keep call sites concise while preserving numeric type constraints. |
| Generate integer ranges | `Range`, `RangeClosed`, `AppendRange` | Use explicit step values and guard user-provided ranges before allocation. |
| Generate random unique values | `GenRandomNumberWithOptions`, `GenBySetWithOptions` | Inject `WithRandomReader` for deterministic tests or controlled entropy. |
| Work with binary encodings | `GetBinaryStr`, `BinaryToInt`, `ToUnsignedByteArrayLen` | Prefer fixed-length helpers when the byte shape crosses a protocol boundary. |
| Use combinatorics and integer math | `Factorial`, `FactorialBig`, `Divisor`, `Multiple` | Use error-returning or big-number variants when overflow is possible. |

## Numeric correctness checklist

- Use string or `big.Rat` helpers for decimal money, accounting, and exact ratios; `float64` helpers are not decimal-exact.
- Pick an explicit rounding mode at business boundaries. Hidden default rounding can change totals in reports.
- Validate user-provided ranges, counts, and steps before calling range or random generators to avoid large allocations or impossible requests.
- Prefer default-returning parse helpers only when fallback is intended. Use error-returning helpers such as `ParseNumber` when invalid input should be surfaced.
- Inject parser, formatter, or random-reader providers for deterministic tests instead of depending on locale, runtime formatting, or entropy sources.
- Use `AbsIntegerE`, `FactorialBig`, or big-number helpers when overflow is a concern.

## Exact arithmetic and rounding

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vnum"
)

func main() {
	fmt.Println(vnum.Add(1, 2, 3))
	fmt.Println(vnum.AddStr("0.1", "0.2").FloatString(1))
	fmt.Println(vnum.Div(10, 3, 2))
	fmt.Println(vnum.Round(3.14159, 2))
}
```

## Parse, validate, and use defaults

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vnum"
)

func main() {
	fmt.Println(vnum.ParseInt("42"))
	fmt.Println(vnum.ParseDoubleDefault("bad", 3.14))
	fmt.Println(vnum.IsNumber("12.5"), vnum.IsInteger("12.5"))
	fmt.Println(vnum.IsOdd(7), vnum.IsEven(8), vnum.IsPrimes(11))
}
```

## Format amounts, percentages, and strings

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vnum"
)

func main() {
	fmt.Println(vnum.DecimalFormatMoney(12345.6))
	fmt.Println(vnum.FormatPercent(0.1234, 2))
	fmt.Println(vnum.ToStrStrip(12.3400, true))

	value := 0.0
	fmt.Println(vnum.ToStrDefault(&value, "n/a"))
	fmt.Println(vnum.ToStrDefault(nil, "n/a"))
}
```

## Expressions, aggregation, and random numbers

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vnum"
)

func main() {
	result, err := vnum.Calculate("1 + 2 * 3")
	if err != nil {
		panic(err)
	}

	fmt.Println(result)
	fmt.Println(vnum.SumNumber(1, 2, 3), vnum.AvgNumber(2, 4, 6))
	fmt.Println(vnum.MinIntegers(3, 1, 2), vnum.MaxIntegers(3, 1, 2))
	fmt.Println(vnum.GenRandomNumber(1, 10, 3))
}
```

## When not to use vnum

- Use domain-specific decimal or money types when amounts need currency codes, scale enforcement, or audit-grade rounding policy.
- Use `math/big` directly when algorithms need full control over precision, numerator/denominator lifetime, or allocation reuse.
- Use `crypto/rand`, `math/rand/v2`, or a dedicated random library directly when random generation semantics are security-critical or simulation-specific.
- Avoid expression evaluation on untrusted, unbounded input unless the caller constrains expression length and allowed syntax.

## Related packages

- Use `vconv` when numeric parsing is one part of broader loose type conversion.
- Use `vform` when numbers need request validation, ranges, or custom predicates.
- Use `vjson` when numeric behavior is tied to JSON payload parsing or fixture formatting.

## Benchmarks and trade-offs

- Generic aggregate helpers are concise and fast for small slices, but large datasets may benefit from streaming aggregation to avoid holding all values in memory.
- `big.Rat` string arithmetic protects decimal precision at the cost of extra allocation and slower operations than primitive `float64` math.
- Expression evaluation is more flexible than direct arithmetic but requires parsing, so keep `Calculate` away from tight loops when the formula is static.
- Random helpers that guarantee unique values must track selected numbers; cost grows as requested size approaches the range size.
- Formatting helpers allocate strings and should be used at presentation edges rather than repeatedly inside numeric kernels.

## FAQ

### Should I use `Add` or `AddStr` for money?

Use `AddStr` or another exact decimal representation when the input is decimal text and cents must not drift. Use `Add` only when normal binary floating-point behavior is acceptable.

### Why do default parse helpers not return errors?

They encode fallback behavior directly. If invalid input should be observable, use `ParseNumber` or a custom parser via `WithParse*Func` and return the error to the caller.

### How do I test random-number generation?

Use `GenRandomNumberWithOptions` or `GenBySetWithOptions` with `WithRandomReader` so the byte stream is controlled by the test.

### Which range helper includes the stop value?

`RangeClosed` includes the stop value when the step lands on it. `Range` is the half-open helper for `[start, end)` style sequences.
