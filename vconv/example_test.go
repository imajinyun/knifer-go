package vconv_test

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	knifer "github.com/imajinyun/knifer-go"
	"github.com/imajinyun/knifer-go/vconv"
	"github.com/imajinyun/knifer-go/vmap"
	"github.com/imajinyun/knifer-go/vslice"
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

func ExampleWithBoolParser() {
	parser := func(s string) (bool, error) {
		return s == "enabled", nil
	}

	fmt.Println(vconv.ToBoolWithOptions("enabled", vconv.WithBoolParser(parser)))
	// Output: true
}

func ExampleWithParseIntFunc() {
	parser := func(s string, base, bitSize int) (int64, error) {
		if s == "dozen" {
			return 12, nil
		}
		return strconv.ParseInt(s, base, bitSize)
	}

	fmt.Println(vconv.ToIntWithOptions("dozen", vconv.WithParseIntFunc(parser)))
	// Output: 12
}

func ExampleWithParseFloatFunc() {
	parser := func(s string, bitSize int) (float64, error) {
		if s == "pi" {
			return 3.14, nil
		}
		return strconv.ParseFloat(s, bitSize)
	}

	fmt.Printf("%.2f\n", vconv.ToFloat64WithOptions("pi", vconv.WithParseFloatFunc(parser)))
	// Output: 3.14
}

func ExampleWithFormatBoolFunc() {
	formatter := func(v bool) string {
		if v {
			return "yes"
		}
		return "no"
	}

	fmt.Println(vconv.ToStringWithOptions(false, vconv.WithFormatBoolFunc(formatter)))
	// Output: no
}

func ExampleWithFormatFloatFunc() {
	formatter := func(v float64, _ byte, _ int, _ int) string {
		return fmt.Sprintf("%.1f units", v)
	}

	fmt.Println(vconv.ToStringWithOptions(2.25, vconv.WithFormatFloatFunc(formatter)))
	// Output: 2.2 units
}

func ExampleToStringDefaultWithOptions() {
	formatter := func(v bool) string {
		if v {
			return "on"
		}
		return "off"
	}

	fmt.Println(vconv.ToStringDefaultWithOptions(true, "missing", vconv.WithFormatBoolFunc(formatter)))
	// Output: on
}

func ExampleToIntWithOptions() {
	parser := func(string, int, int) (int64, error) { return 7, nil }

	fmt.Println(vconv.ToIntWithOptions("seven", vconv.WithParseIntFunc(parser)))
	// Output: 7
}

func ExampleToIntDefaultWithOptions() {
	parser := func(string, int, int) (int64, error) { return 0, errors.New("bad int") }

	fmt.Println(vconv.ToIntDefaultWithOptions("bad", -1, vconv.WithParseIntFunc(parser)))
	// Output: -1
}

func ExampleToIntE() {
	value, err := vconv.ToIntE("42")

	fmt.Println(value)
	fmt.Println(err)
	// Output:
	// 42
	// <nil>
}

func ExampleToInt64() {
	fmt.Println(vconv.ToInt64("42"))
	// Output: 42
}

func ExampleToInt64WithOptions() {
	parser := func(string, int, int) (int64, error) { return 64, nil }

	fmt.Println(vconv.ToInt64WithOptions("sixty-four", vconv.WithParseIntFunc(parser)))
	// Output: 64
}

func ExampleToInt64Default() {
	fmt.Println(vconv.ToInt64Default("bad", -64))
	// Output: -64
}

func ExampleToInt64DefaultWithOptions() {
	parser := func(string, int, int) (int64, error) { return 0, errors.New("bad int64") }

	fmt.Println(vconv.ToInt64DefaultWithOptions("bad", -64, vconv.WithParseIntFunc(parser)))
	// Output: -64
}

func ExampleToInt64EWithOptions() {
	parser := func(string, int, int) (int64, error) { return 128, nil }
	value, err := vconv.ToInt64EWithOptions("one-two-eight", vconv.WithParseIntFunc(parser))

	fmt.Println(value)
	fmt.Println(err)
	// Output:
	// 128
	// <nil>
}

func ExampleToFloat64WithOptions() {
	parser := func(string, int) (float64, error) { return 6.28, nil }

	fmt.Printf("%.2f\n", vconv.ToFloat64WithOptions("tau", vconv.WithParseFloatFunc(parser)))
	// Output: 6.28
}

func ExampleToFloat64DefaultWithOptions() {
	parser := func(string, int) (float64, error) { return 0, errors.New("bad float") }

	fmt.Println(vconv.ToFloat64DefaultWithOptions("bad", -1.25, vconv.WithParseFloatFunc(parser)))
	// Output: -1.25
}

func ExampleToFloat64E() {
	value, err := vconv.ToFloat64E("3.5")

	fmt.Printf("%.1f\n", value)
	fmt.Println(err)
	// Output:
	// 3.5
	// <nil>
}

func ExampleToFloat64EWithOptions() {
	parser := func(string, int) (float64, error) { return 9.5, nil }
	value, err := vconv.ToFloat64EWithOptions("nine-point-five", vconv.WithParseFloatFunc(parser))

	fmt.Printf("%.1f\n", value)
	fmt.Println(err)
	// Output:
	// 9.5
	// <nil>
}

func ExampleToBoolWithOptions() {
	parser := func(s string) (bool, error) { return s == "YES", nil }

	fmt.Println(vconv.ToBoolWithOptions("YES", vconv.WithBoolParser(parser)))
	// Output: true
}

func ExampleToBoolDefaultWithOptions() {
	parser := func(string) (bool, error) { return false, errors.New("bad bool") }

	fmt.Println(vconv.ToBoolDefaultWithOptions("bad", true, vconv.WithBoolParser(parser)))
	// Output: true
}

func ExampleToBytesWithOptions() {
	formatter := func(v bool) string {
		if v {
			return "enabled"
		}
		return "disabled"
	}

	fmt.Println(string(vconv.ToBytesWithOptions(true, vconv.WithFormatBoolFunc(formatter))))
	// Output: enabled
}

func Example_castMigration_strictConversion() {
	port, err := vconv.ToIntE("8080")

	fmt.Println(port)
	fmt.Println(err)
	// Output:
	// 8080
	// <nil>
}

func Example_castMigration_weakConversion() {
	count := vconv.ToInt("not-a-number")

	fmt.Println(count)
	// Output: 0
}

func Example_castMigration_defaultFallback() {
	limit := vconv.ToIntDefault("bad", 100)

	fmt.Println(limit)
	// Output: 100
}

func Example_castMigration_customParserPolicy() {
	value, err := vconv.ToIntEWithOptions("max", vconv.WithParseIntFunc(func(s string, base, bitSize int) (int64, error) {
		if s == "max" {
			return 100, nil
		}
		return strconv.ParseInt(s, base, bitSize)
	}))

	fmt.Println(value)
	fmt.Println(err)
	// Output:
	// 100
	// <nil>
}

func Example_castMigration_sliceMapBoundary() {
	numbers := vslice.Map([]string{"1", "2"}, func(s string) int { return vconv.ToInt(s) })
	picked := vmap.Pick(map[string]any{"port": 8080, "debug": true}, "port")

	fmt.Println(numbers)
	fmt.Println(picked)
	// Output:
	// [1 2]
	// map[port:8080]
}

func Example_castMigration_durationTimeBoundary() {
	timeout, err := time.ParseDuration("150ms")

	fmt.Println(timeout)
	fmt.Println(err)
	fmt.Println(vconv.ToInt64(timeout))
	// Output:
	// 150ms
	// <nil>
	// 150000000
}

func Example_castMigration_overflowHandling() {
	_, err := vconv.ToInt64E(uint64(math.MaxInt64) + 1)

	fmt.Println(errors.Is(err, vconv.ErrInvalidConversion))
	// Output: true
}
