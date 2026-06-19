package vlog_test

import (
	"bytes"
	"fmt"
	"time"

	"github.com/imajinyun/go-knifer/vlog"
)

func ExampleGetDefault() {
	log := vlog.GetDefault()
	fmt.Println(log != nil)
	// Output: true
}

func ExampleNewConsoleLogWithOptions() {
	var out bytes.Buffer
	fixed := time.Date(2024, 4, 5, 6, 7, 8, 0, time.UTC)
	log := vlog.NewConsoleLogWithOptions(
		"example",
		vlog.WithLogClock(func() time.Time { return fixed }),
		vlog.WithLogTimeLayout(time.RFC3339),
		vlog.WithLogOutput(&out, &bytes.Buffer{}),
		vlog.WithLogLevel(vlog.LogLevelInfo),
	)

	log.Infof("hello {}", "world")
	fmt.Print(out.String())
	// Output: [2024-04-05T06:07:08Z] [INFO ] example: hello world
}

func ExampleInfoWithOptions() {
	var out bytes.Buffer
	fixed := time.Date(2024, 4, 5, 6, 7, 8, 0, time.UTC)
	opts := []vlog.LoggerOption{vlog.WithLoggerConsoleOptions(
		vlog.WithLogClock(func() time.Time { return fixed }),
		vlog.WithLogTimeLayout(time.RFC3339),
		vlog.WithLogOutput(&out, &bytes.Buffer{}),
	)}

	vlog.InfoWithOptions(opts, "status ok")
	fmt.Print(out.String())
	// Output: [2024-04-05T06:07:08Z] [INFO ] static: status ok
}

func ExampleLevel_String() {
	fmt.Println(vlog.LogLevelWarn.String())
	// Output: WARN
}

func ExampleNewIsolatedLogger() {
	var out bytes.Buffer
	fixed := time.Date(2024, 4, 5, 6, 7, 8, 0, time.UTC)

	log := vlog.NewIsolatedLogger("isolated", vlog.WithLoggerConsoleOptions(
		vlog.WithLogClock(func() time.Time { return fixed }),
		vlog.WithLogTimeLayout(time.RFC3339),
		vlog.WithLogOutput(&out, &bytes.Buffer{}),
	))
	log.Infof("retry {}", 2)

	fmt.Print(out.String())
	// Output: [2024-04-05T06:07:08Z] [INFO ] isolated: retry 2
}

func ExampleGetWithOptions() {
	log := vlog.GetWithOptions("named", vlog.WithLoggerCache(false))

	fmt.Println(log != nil)
	// Output: true
}
