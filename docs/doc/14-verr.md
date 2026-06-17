# verr Quickstart

`verr` provides error aggregation, panic recovery, stack capture, and logrus/Sentry initialization helpers for centralized handling of errors from synchronous or asynchronous tasks.

## Recover panics and return errors

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/verr"
)

func main() {
	err := verr.Recover(func() error {
		panic("boom")
	}, "running job")

	fmt.Printf("%T\n", err)
	fmt.Println(verr.GetStack(err) != "")
}
```

## Wrap fallible functions with Wrapper

```go
package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/imajinyun/go-knifer/verr"
)

func main() {
	want := errors.New("write failed")
	err := verr.Wrap(func() error { return want }).WithWarnf("save user").Exec(context.Background())

	fmt.Println(verr.ErrorIs(err, want))
}
```

## Aggregate errors from multiple asynchronous tasks

```go
package main

import (
	"errors"
	"fmt"

	"github.com/imajinyun/go-knifer/verr"
)

func main() {
	c := verr.NewCollector()
	want := errors.New("task failed")

	c.GoRun(func() error { return nil }, "task one")
	c.GoRun(func() error { return want }, "task two")

	err := c.Error()
	fmt.Println(verr.ErrorIs(err, want))
}
```

## Capture stack frames

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/verr"
)

func main() {
	stack := verr.GetStackTraceWithOptions(
		verr.WithStackSkip(0),
		verr.WithStackDepth(4),
	)
	fmt.Println(len(stack) > 0)
}
```

## Initialize Sentry without global side effects in tests

`InitWithOptions` accepts Sentry factories so production code can use
`sentry-go` while tests inject isolated clients and hook registration. Prefer
`WithSentryClientOptions`, `WithSentryClient`, or `WithSentryClientFactory`.

```go
package main

import (
	"fmt"

	"github.com/getsentry/sentry-go"
	"github.com/imajinyun/go-knifer/verr"
	"github.com/sirupsen/logrus"
)

type memoryHook struct{}

func (memoryHook) Levels() []logrus.Level { return []logrus.Level{logrus.ErrorLevel} }

func (memoryHook) Fire(*logrus.Entry) error { return nil }

func main() {
	var registered bool

	verr.InitWithOptions(
		verr.WithSentryDSN("https://public@example.invalid/1"),
		verr.WithSentryClientOptions(sentry.ClientOptions{Environment: "test"}),
		verr.WithSentryClientFactory(func(options sentry.ClientOptions) (*sentry.Client, error) {
			return sentry.NewClient(options)
		}),
		verr.WithSentryHookFactory(func(*sentry.Client, []logrus.Level) (logrus.Hook, error) {
			return memoryHook{}, nil
		}),
		verr.WithLogHookAdder(func(logrus.Hook) { registered = true }),
	)

	fmt.Println(registered)
}
```
