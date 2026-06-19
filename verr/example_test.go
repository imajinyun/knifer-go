package verr_test

import (
	"context"
	"errors"
	"fmt"

	knifer "github.com/imajinyun/go-knifer"
	"github.com/imajinyun/go-knifer/verr"
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
