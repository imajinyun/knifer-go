package vconv_test

import (
	"errors"
	"fmt"
	"strconv"

	knifer "github.com/imajinyun/go-knifer"
	"github.com/imajinyun/go-knifer/vconv"
)

func ExampleToInt() {
	fmt.Println(vconv.ToInt("42"))
	fmt.Println(vconv.ToInt(true))
	// Output:
	// 42
	// 1
}

func ExampleToIntDefault() {
	fmt.Println(vconv.ToIntDefault("not-a-number", -1))
	// Output: -1
}

func ExampleToBool() {
	fmt.Println(vconv.ToBool("true"))
	fmt.Println(vconv.ToBool(0))
	// Output:
	// true
	// false
}

func ExampleToString() {
	fmt.Println(vconv.ToString(3.14))
	// Output: 3.14
}

func ExampleToBytes() {
	fmt.Println(string(vconv.ToBytes("go")))
	// Output: go
}

func ExampleToInt64E() {
	value, err := vconv.ToInt64E("42.9")
	fmt.Println(value)
	fmt.Println(err)
	// Output:
	// 42
	// <nil>
}

func ExampleToInt64E_overflow() {
	value, err := vconv.ToInt64E(uint64(1) << 63)
	fmt.Println(value)
	fmt.Println(errors.Is(err, vconv.ErrInvalidConversion))
	// Output:
	// 0
	// true
}

func ExampleToBoolE() {
	value, err := vconv.ToBoolE("maybe")
	fmt.Println(value)
	fmt.Println(errors.Is(err, vconv.ErrInvalidConversion))
	fmt.Println(errors.Is(err, knifer.ErrCodeInvalidInput))
	// Output:
	// false
	// true
	// true
}

func ExampleToStringWithOptions() {
	formatter := func(v bool) string {
		if v {
			return "enabled"
		}
		return "disabled"
	}

	fmt.Println(vconv.ToStringWithOptions(true, vconv.WithFormatBoolFunc(formatter)))
	// Output: enabled
}

func ExampleToStringDefault() {
	fmt.Println(vconv.ToStringDefault(nil, "unknown"))
	// Output: unknown
}

func ExampleToFloat64() {
	fmt.Printf("%.2f\n", vconv.ToFloat64("3.14"))
	fmt.Printf("%.2f\n", vconv.ToFloat64(true))
	// Output:
	// 3.14
	// 1.00
}

func ExampleToFloat64Default() {
	fmt.Println(vconv.ToFloat64Default("not-a-number", -1.5))
	// Output: -1.5
}

func ExampleToBoolDefault() {
	fmt.Println(vconv.ToBoolDefault("maybe", true))
	// Output: true
}

func ExampleToIntEWithOptions() {
	parser := func(s string, base, bitSize int) (int64, error) {
		if s == "max" {
			return 99, nil
		}
		return strconv.ParseInt(s, base, bitSize)
	}

	value, err := vconv.ToIntEWithOptions("max", vconv.WithParseIntFunc(parser))
	fmt.Println(value)
	fmt.Println(err)
	// Output:
	// 99
	// <nil>
}

func ExampleToBoolEWithOptions() {
	parser := func(s string) (bool, error) {
		return s == "YES", nil
	}

	value, err := vconv.ToBoolEWithOptions("YES", vconv.WithBoolParser(parser))
	fmt.Println(value)
	fmt.Println(err)
	// Output:
	// true
	// <nil>
}
