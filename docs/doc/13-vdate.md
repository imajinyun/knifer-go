# vdate Quickstart

`vdate` provides common date/time formatting, parsing, boundary calculation, offset, and comparison helpers for concise `time.Time` business logic.

## Format and parse common dates

```go
package main

import (
	"fmt"
	"time"

	"github.com/imajinyun/go-knifer/vdate"
)

func main() {
	t := time.Date(2024, 5, 6, 7, 8, 9, 0, time.UTC)
	fmt.Println(vdate.FormatNorm(t))
	fmt.Println(vdate.FormatDateOnly(t))
	fmt.Println(vdate.FormatTimeOnly(t))

	parsed, err := vdate.Parse("2024-05-06 07:08:09")
	if err != nil {
		panic(err)
	}
	fmt.Println(vdate.Format(parsed, vdate.NormDatePattern))
}
```

## Parse with a specific time zone

```go
package main

import (
	"fmt"
	"time"

	"github.com/imajinyun/go-knifer/vdate"
)

func main() {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		panic(err)
	}

	t, err := vdate.ParseLayoutWithOptions(
		"2024-05-06 07:08:09",
		vdate.NormPattern,
		vdate.WithLocation(loc),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(t.Location())
}
```

## Get date boundaries

```go
package main

import (
	"fmt"
	"time"

	"github.com/imajinyun/go-knifer/vdate"
)

func main() {
	t := time.Date(2024, 5, 6, 7, 8, 9, 0, time.Local)

	fmt.Println(vdate.BeginOfDay(t))
	fmt.Println(vdate.EndOfDay(t))
	fmt.Println(vdate.BeginOfMonth(t))
	fmt.Println(vdate.EndOfYear(t))
}
```

## Offset and compare dates

```go
package main

import (
	"fmt"
	"time"

	"github.com/imajinyun/go-knifer/vdate"
)

func main() {
	start := time.Date(2024, 5, 6, 7, 8, 9, 0, time.Local)
	nextWeek := vdate.OffsetDay(start, 7)
	nextMonth := vdate.OffsetMonth(start, 1)

	fmt.Println(vdate.BetweenDays(start, nextWeek))
	fmt.Println(vdate.IsSameDay(start, nextMonth))
}
```
