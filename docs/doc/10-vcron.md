# vcron Quickstart

`vcron` provides cron expression parsing and task scheduling facades, with support for the default scheduler, local schedulers, second-level matching, custom IDs, clocks, and executors.

## Parse and match cron expressions

```go
package main

import (
	"fmt"
	"time"

	"github.com/imajinyun/go-knifer/vcron"
)

func main() {
	p, err := vcron.NewPattern("* * * * *")
	if err != nil {
		panic(err)
	}

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	fmt.Println(p.Match(now, false))
}
```

## Schedule functions with a local scheduler

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vcron"
)

func main() {
	s := vcron.NewScheduler()
	id, err := s.ScheduleFunc("* * * * *", func() {
		fmt.Println("tick")
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("scheduled", id)
	fmt.Println("removed", s.Remove(id))
}
```

## Customize scheduler options

```go
package main

import (
	"fmt"
	"time"

	"github.com/imajinyun/go-knifer/vcron"
)

func main() {
	loc := time.FixedZone("cst", 8*60*60)
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	s := vcron.NewSchedulerWithOptions(
		vcron.WithLocation(loc),
		vcron.WithMatchSecond(true),
		vcron.WithIDGenerator(func() string { return "job-1" }),
		vcron.WithClock(func() time.Time { return now }),
		vcron.WithExecutor(func(fn func()) { fn() }),
	)

	id, err := s.ScheduleFunc("* * * * * *", func() {})
	if err != nil {
		panic(err)
	}
	fmt.Println(id, s.IsMatchSecond(), s.Config().Location == loc)
}
```

## Use the default scheduler facade

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vcron"
)

func main() {
	id, err := vcron.CronScheduleFunc("* * * * *", func() {})
	if err != nil {
		panic(err)
	}

	fmt.Println(vcron.CronRemove(id))
}
```
