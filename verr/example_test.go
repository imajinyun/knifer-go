package verr_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"

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
