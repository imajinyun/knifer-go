# vnum Quickstart

`vnum` provides number parsing, formatting, exact string arithmetic, rounding, expression evaluation, ranges/random numbers, binary helpers, and common numeric predicates.

## Exact arithmetic and rounding

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vnum"
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

	"github.com/imajinyun/go-knifer/vnum"
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

	"github.com/imajinyun/go-knifer/vnum"
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

	"github.com/imajinyun/go-knifer/vnum"
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
