package knifer_test

import (
	"errors"
	"fmt"

	"github.com/imajinyun/knifer-go"
)

func ExampleError() {
	err := knifer.NewError(knifer.ErrCodeInvalidInput, "url is empty")
	fmt.Println(errors.Is(err, knifer.ErrCodeInvalidInput))
	fmt.Println(err)
	// Output:
	// true
	// GK_INVALID_INPUT: url is empty
}

func ExampleWrapError() {
	cause := errors.New("connection refused")
	err := knifer.WrapError(knifer.ErrCodeTimeout, "dial failed", cause)
	fmt.Println(errors.Is(err, knifer.ErrCodeTimeout))
	fmt.Println(errors.Is(err, cause))
	// Output:
	// true
	// true
}
