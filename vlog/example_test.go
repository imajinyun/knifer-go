package vlog_test

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	"github.com/imajinyun/knifer-go/vlog"
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

func ExampleNewConsoleLog() {
	log := vlog.NewConsoleLog("plain")

	fmt.Println(log.GetName())
	// Output: plain
}

func ExampleDefaultLoggerWithOptions() {
	log := vlog.DefaultLoggerWithOptions(vlog.WithLoggerCache(false))

	fmt.Println(log.GetName())
	// Output: default
}

func ExampleLogAtEWithOptions() {
	var out bytes.Buffer
	fixed := time.Date(2024, 4, 5, 6, 7, 8, 0, time.UTC)
	opts := []vlog.LoggerOption{vlog.WithLoggerConsoleOptions(
		vlog.WithLogClock(func() time.Time { return fixed }),
		vlog.WithLogTimeLayout(time.RFC3339),
		vlog.WithLogOutput(&out, &bytes.Buffer{}),
	)}

	vlog.LogAtEWithOptions(opts, vlog.LogLevelInfo, errors.New("closed"), "queue {}", "jobs")
	fmt.Print(out.String())
	// Output: [2024-04-05T06:07:08Z] [INFO ] static: queue jobs | error: closed
}

func ExampleWarnfWithOptions() {
	var errOut bytes.Buffer
	fixed := time.Date(2024, 4, 5, 6, 7, 8, 0, time.UTC)
	opts := []vlog.LoggerOption{vlog.WithLoggerConsoleOptions(
		vlog.WithLogClock(func() time.Time { return fixed }),
		vlog.WithLogTimeLayout(time.RFC3339),
		vlog.WithLogOutput(&bytes.Buffer{}, &errOut),
	)}

	vlog.WarnfWithOptions(opts, "slow request %d", 3)
	fmt.Print(errOut.String())
	// Output: [2024-04-05T06:07:08Z] [WARN ] static: slow request 3
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

func ExampleLoggerWithOptions() {
	var out bytes.Buffer
	fixed := time.Date(2024, 4, 5, 6, 7, 8, 0, time.UTC)
	log := vlog.LoggerWithOptions("request", vlog.WithLoggerCache(false), vlog.WithLoggerConsoleOptions(
		vlog.WithLogClock(func() time.Time { return fixed }),
		vlog.WithLogTimeLayout(time.RFC3339),
		vlog.WithLogOutput(&out, &bytes.Buffer{}),
	))

	log.Infof("handled {}", "GET /healthz")
	fmt.Print(out.String())
	// Output: [2024-04-05T06:07:08Z] [INFO ] request: handled GET /healthz
}

func ExampleErrorfWithOptions() {
	var errOut bytes.Buffer
	fixed := time.Date(2024, 4, 5, 6, 7, 8, 0, time.UTC)
	opts := []vlog.LoggerOption{vlog.WithLoggerConsoleOptions(
		vlog.WithLogClock(func() time.Time { return fixed }),
		vlog.WithLogTimeLayout(time.RFC3339),
		vlog.WithLogOutput(&bytes.Buffer{}, &errOut),
	)}

	vlog.ErrorfWithOptions(opts, "connect {}", "failed")
	fmt.Print(errOut.String())
	// Output: [2024-04-05T06:07:08Z] [ERROR] static: connect failed
}

func ExampleLogAtWithOptions() {
	var out bytes.Buffer
	fixed := time.Date(2024, 4, 5, 6, 7, 8, 0, time.UTC)
	opts := []vlog.LoggerOption{vlog.WithLoggerConsoleOptions(
		vlog.WithLogClock(func() time.Time { return fixed }),
		vlog.WithLogTimeLayout(time.RFC3339),
		vlog.WithLogOutput(&out, &bytes.Buffer{}),
	)}

	vlog.LogAtWithOptions(opts, vlog.LogLevelInfo, "queue depth {}", 3)
	fmt.Print(out.String())
	// Output: [2024-04-05T06:07:08Z] [INFO ] static: queue depth 3
}
