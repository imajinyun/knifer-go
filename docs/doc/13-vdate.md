# vdate Quickstart

`vdate` provides common date/time formatting, parsing, boundary calculation, offset, and comparison helpers for concise `time.Time` business logic.

## Which helper should I use?

Choose helpers by the business rule you are expressing: formatting/parsing, calendar boundary, offset, or comparison.

| Need | Use | Notes |
| --- | --- | --- |
| Format a time with common layouts | `FormatNorm`, `FormatDateOnly`, `FormatTimeOnly`, `Format` | Use named helpers for project-wide conventional layouts; use `Format` when a specific layout is part of the protocol. |
| Parse common date/time strings | `Parse`, `ParseLayoutWithOptions` | Use options such as `WithLocation` when local time semantics matter. |
| Calculate calendar boundaries | `BeginOfDay`, `EndOfDay`, `BeginOfMonth`, `EndOfYear` | Boundaries follow the `time.Time` location carried by the input. |
| Move by calendar units | `OffsetDay`, `OffsetMonth`, related offset helpers | Calendar offsets are not always fixed durations because months and daylight-saving transitions vary. |
| Compare business dates | `BetweenDays`, `IsSameDay`, related comparison helpers | Prefer semantic helpers over manual duration division when calendar days are intended. |

## Date/time correctness checklist

- Keep time zone decisions explicit. Parse with `WithLocation` when input strings do not include an offset but the business rule requires a specific location.
- Do not assume a day is always `24*time.Hour` for calendar logic; daylight-saving and location rules can change elapsed duration.
- Use calendar offset helpers for dates, and duration arithmetic for elapsed time. Mixing the two can produce off-by-one-day bugs.
- Preserve the input location when calculating day, month, or year boundaries unless you intentionally normalize to UTC.
- Treat parse errors as validation failures and surface them to the caller instead of falling back to zero time silently.

## When not to use vdate

- Use `time` directly for timers, deadlines, monotonic-clock comparisons, and low-level duration arithmetic.
- Use a domain calendar library when business days, holidays, fiscal calendars, or locale-specific week rules matter.
- Use strict protocol parsers when input must follow RFC3339, Unix timestamps, or another externally defined format exactly.
- Avoid calendar helpers for elapsed-time measurement; use `time.Since`, `Sub`, or monotonic-aware `time.Time` values.

## Related packages

- Use `vconv` when date parsing is one part of broader loose type conversion.
- Use `vjson` when dates need package-level JSON formatting behavior.
- Use `vnum` when date workflows include duration, age, or calendar-derived numeric calculations.

## Benchmarks and trade-offs

Most date helpers are thin wrappers around `time`, but parsing, formatting, and location lookup can still matter in batch jobs:

```bash
go test -bench=. -benchmem -run=^$ ./internal/date ./vdate
```

Convenience helpers make business-date intent clearer than repeated layout strings and manual boundary calculations. The trade-off is that callers still need to decide storage timezone, display location, and whether the rule is calendar-based or duration-based.

Parsing with explicit locations improves correctness for local date strings, but it should be scoped to inputs that truly lack offsets. For cross-service timestamps, prefer explicit offsets or UTC.

## FAQ

### Does vdate replace the time package?

No. `vdate` wraps common business-date operations. Use `time` directly for timers, monotonic-clock behavior, low-level layouts, and duration arithmetic.

### Should I use UTC everywhere?

Use UTC for storage and cross-service timestamps when possible. Use an explicit business location for parsing, display, calendar boundaries, and reports that are defined by local dates.

### Why prefer calendar helpers over duration math?

Calendar operations depend on location, month length, leap days, and daylight-saving rules. Helpers make that intent clearer than dividing elapsed hours by 24.

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
