# vlog Quickstart

`vlog` provides console logging facades with package-level static logs, named logger lookup, colored output, log levels, and custom output targets.

## Use static logging functions

```go
package main

import "github.com/imajinyun/go-knifer/vlog"

func main() {
	vlog.SetLogLevel(vlog.LogLevelDebug)
	vlog.Debug("debug message")
	vlog.Infof("hello %s", "go-knifer")
	vlog.Warn("watch this")
}
```

## Create a console logger and customize output

```go
package main

import (
	"bytes"
	"fmt"
	"time"

	"github.com/imajinyun/go-knifer/vlog"
)

func main() {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	fixed := time.Date(2024, 4, 5, 6, 7, 8, 0, time.UTC)

	log := vlog.NewConsoleLogWithOptions("demo",
		vlog.WithLogOutput(out, errOut),
		vlog.WithLogClock(func() time.Time { return fixed }),
		vlog.WithLogTimeLayout(time.RFC3339),
		vlog.WithLogLevel(vlog.LogLevelInfo),
	)
	log.Info("ready")
	log.Error("failed")

	fmt.Println(out.String())
	fmt.Println(errOut.String())
}
```

## Use a colored logger

```go
package main

import (
	"bytes"
	"fmt"

	"github.com/imajinyun/go-knifer/vlog"
)

func main() {
	out := &bytes.Buffer{}
	log := vlog.NewConsoleColorLogWithOptions("color",
		vlog.WithLogOutput(out, &bytes.Buffer{}),
		vlog.WithLogColorFactory(func(level vlog.Level) string { return "\033[36m" }),
	)
	log.Info("colored")
	fmt.Println(out.String())
}
```

## Choose logger configuration for a single call

```go
package main

import (
	"bytes"
	"fmt"

	"github.com/imajinyun/go-knifer/vlog"
)

func main() {
	out := &bytes.Buffer{}
	log := vlog.LoggerWithOptions("request",
		vlog.WithLoggerCache(false),
		vlog.WithLoggerConsoleOptions(vlog.WithLogOutput(out, &bytes.Buffer{})),
	)
	log.Info("request started")

	vlog.InfoWithOptions([]vlog.LoggerOption{
		vlog.WithLoggerCache(false),
		vlog.WithLoggerConsoleOptions(vlog.WithLogOutput(out, &bytes.Buffer{})),
	}, "static call")

	fmt.Println(out.String())
}
```
