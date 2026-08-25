package verr_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	knifer "github.com/imajinyun/knifer-go"
	"github.com/imajinyun/knifer-go/verr"
	"github.com/sirupsen/logrus"
)

func ExampleErrorIs() {
	inputErr := knifer.WrapError(knifer.ErrCodeInvalidInput, "bad value", errors.New("parse failed"))

	fmt.Println(verr.ErrorIs(inputErr, knifer.ErrCodeInvalidInput))
	fmt.Println(verr.ErrorIs(inputErr, knifer.ErrCodeInternal))
	// Output:
	// true
	// false
}

func ExampleGetStackWithOptions() {
	stack := verr.GetStackWithOptions(errors.New("plain"), verr.WithDebugStackFunc(func() []byte {
		return []byte("captured stack")
	}))

	fmt.Println(stack)
	// Output: captured stack
}

func ExampleNewCollector() {
	c := verr.NewCollector()
	c.Collect(errors.New("first"))
	c.Collect(errors.New("second"))

	err := c.Error()
	fmt.Println(err != nil)
	// Output: true
}

func ExampleRecoverWithoutError() {
	verr.ConfigureDefaultLogFunc(func(context.Context, logrus.Level, error, string, string, ...any) {})
	defer verr.ResetDefaultLogFunc()

	err := verr.RecoverWithoutError(func() {
		panic("boom")
	}, "safe")

	fmt.Println(err != nil)
	// Output: true
}

func ExampleMustExitWithOptions() {
	called := false

	verr.MustExitWithOptions(
		context.Background(),
		errors.New("stop"),
		verr.WithExitLogFunc(func(context.Context, logrus.Level, error, string, string, ...any) {}),
		verr.WithExitPanicFunc(func(error) { called = true }),
	)

	fmt.Println(called)
	// Output: true
}

func ExampleWrap() {
	err := verr.Wrap(func() error {
		panic("boom")
	}).WithLogFunc(func(context.Context, logrus.Level, error, string, string, ...any) {}).WithErrorf("safe call").Exec(context.Background())

	fmt.Println(err != nil)
	// Output: true
}

func ExampleRecover() {
	verr.ConfigureDefaultLogFunc(func(context.Context, logrus.Level, error, string, string, ...any) {})
	defer verr.ResetDefaultLogFunc()

	err := verr.Recover(func() error {
		return errors.New("failed")
	}, "run task")

	fmt.Println(err != nil)
	// Output: true
}

func ExampleConfigureDefaultLogFunc() {
	verr.ConfigureDefaultLogFunc(func(_ context.Context, _ logrus.Level, _ error, _ string, format string, _ ...any) {
		fmt.Println(format)
	})
	defer verr.ResetDefaultLogFunc()

	_ = verr.Recover(func() error { return errors.New("failed") }, "run task")
	// Output: run task
}

func ExampleResetDefaultLogFunc() {
	verr.ConfigureDefaultLogFunc(func(context.Context, logrus.Level, error, string, string, ...any) {})
	verr.ResetDefaultLogFunc()
	fmt.Println("reset")
	// Output: reset
}

func ExampleMustExit() {
	verr.ConfigureDefaultLogFunc(func(context.Context, logrus.Level, error, string, string, ...any) {})
	defer verr.ResetDefaultLogFunc()

	defer func() {
		fmt.Println(recover() != nil)
	}()
	verr.MustExit(context.Background(), errors.New("stop"))
	// Output: true
}

func ExampleGetStack() {
	err := attachedStackError{error: errors.New("plain")}
	fmt.Println(verr.GetStack(err))
	// Output: attached stack
}

func ExampleGetStackTrace() {
	frames := verr.GetStackTrace(0)
	fmt.Println(len(frames) > 0)
	// Output: true
}

func ExampleGetStackTraceWithOptions() {
	verr.ResetStackFrameCache()
	defer verr.ResetStackFrameCache()

	frames := verr.GetStackTraceWithOptions(
		verr.WithStackSkip(0),
		verr.WithStackDepth(1),
		verr.WithStackFrameCache(true),
		verr.WithCallersFunc(func(_ int, pcs []uintptr) int {
			pcs[0] = 2
			return 1
		}),
		verr.WithFuncForPCFunc(func(uintptr) (file string, line int, name string, ok bool) {
			return "example.go", 10, "example.Run", true
		}),
	)
	fmt.Printf("%+v\n", frames)
	// Output:
	//
	// example.Run
	//	example.go:10
}

func ExampleResetStackFrameCache() {
	verr.ResetStackFrameCache()
	fmt.Println("cleared")
	// Output: cleared
}

func ExampleNewCollectorWithOptions() {
	c := verr.NewCollectorWithOptions(
		verr.WithCollectorLogFunc(func(context.Context, logrus.Level, error, string, string, ...any) {}),
		verr.WithCollectorRunner(func(fn func()) { fn() }),
	)
	c.GoRun(func() error { return errors.New("task failed") }, "task")
	fmt.Println(c.Error() != nil)
	// Output: true
}

func ExampleWaitUntilWithOptions() {
	c := verr.NewCollector()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	done, err := verr.WaitUntilWithOptions(c, time.Second,
		verr.WithWaitContext(ctx),
		verr.WithWaitTimerFactory(func(time.Duration) (<-chan time.Time, verr.Timer) {
			return make(chan time.Time), waitExampleTimer{}
		}),
	)
	fmt.Println(done, err == nil)
	// Output: false true
}

type attachedStackError struct{ error }

func (attachedStackError) Stack() string { return "attached stack" }

type waitExampleTimer struct{}

func (waitExampleTimer) Stop() bool { return true }

func ExampleNewIsolatedLogrusWithOptions() {
	var out bytes.Buffer
	logger := verr.NewIsolatedLogrusWithOptions(
		verr.WithLogOutput(&out),
		verr.WithLogFormatter(&logrus.TextFormatter{DisableTimestamp: true, DisableColors: true}),
		verr.WithReportCaller(false),
	)

	logger.Info("ready")
	fmt.Print(out.String())
	// Output: level=info msg=ready
}
