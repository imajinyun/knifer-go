package vnum_test

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vnum"
)

type sequenceReader struct {
	next byte
}

func (r *sequenceReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = r.next
		r.next++
	}
	return len(p), nil
}

func ExampleRound() {
	fmt.Println(vnum.Round(3.14159, 2))
	// Output: 3.14
}

func ExampleAddStr() {
	// AddStr keeps exact precision and avoids float rounding errors.
	fmt.Println(vnum.AddStr("0.1", "0.2").FloatString(1))
	// Output: 0.3
}

func ExampleIsPrimes() {
	fmt.Println(vnum.IsPrimes(7))
	fmt.Println(vnum.IsPrimes(8))
	// Output:
	// true
	// false
}

func ExampleMax() {
	fmt.Println(vnum.Max(3, 7, 1))
	// Output: 7
}

func ExampleCalculate() {
	result, _ := vnum.Calculate("1 + 2 * 3")
	fmt.Println(result)
	// Output: 7
}

func ExampleDecimalFormatMoney() {
	fmt.Println(vnum.DecimalFormatMoney(12345.6))
	// Output: 12,345.60
}

func ExampleRangeClosed() {
	fmt.Println(vnum.RangeClosed(1, 5, 2))
	fmt.Println(vnum.RangeClosed(5, 1, 2))
	// Output:
	// [1 3 5]
	// [5 3 1]
}

func ExampleGenRandomNumberWithOptions() {
	numbers := vnum.GenRandomNumberWithOptions(0, 5, 3, vnum.WithRandomReader(&sequenceReader{}))
	fmt.Println(numbers)
	// Output: [0 1 2]
}
