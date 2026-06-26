package vnum_test

import (
	"fmt"
	"math/big"

	"github.com/imajinyun/knifer-go/vnum"
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

func ExampleDiv() {
	fmt.Println(vnum.Div(10, 3, 2))
	fmt.Println(vnum.Div(10, 3))
	// Output:
	// 3.33
	// 3.3333333333
}

func ExampleSumNumber() {
	fmt.Println(vnum.SumNumber(1, 2, 3))
	fmt.Println(vnum.SumNumber(1.25, 2.5, -0.75))
	// Output:
	// 6
	// 3
}

func ExampleAvgNumber() {
	fmt.Println(vnum.AvgNumber(2, 4, 6))
	fmt.Println(vnum.AvgNumber[int]())
	// Output:
	// 4
	// 0
}

func ExampleAbsIntegerE() {
	value, err := vnum.AbsIntegerE[int8](-128)
	fmt.Println(value)
	fmt.Println(err != nil)
	// Output:
	// 0
	// true
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

func ExampleParseIntDefault() {
	fmt.Println(vnum.ParseIntDefault("1,234", -1))
	fmt.Println(vnum.ParseIntDefault("bad", -1))
	// Output:
	// 1234
	// -1
}

func ExampleDecimalFormatMoney() {
	fmt.Println(vnum.DecimalFormatMoney(12345.6))
	// Output: 12,345.60
}

func ExampleFormatPercent() {
	fmt.Println(vnum.FormatPercent(0.1234, 2))
	fmt.Println(vnum.FormatPercent(0.1, -3))
	// Output:
	// 12.34%
	// 10%
}

func ExampleRoundMode() {
	fmt.Println(vnum.RoundMode(2.5, 0, vnum.RoundHalfEven))
	fmt.Println(vnum.RoundMode(2.5, 0, vnum.RoundHalfUp))
	// Output:
	// 2
	// 3
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

func ExampleBinaryToInt() {
	value, err := vnum.BinaryToInt("1010")
	fmt.Println(value)
	fmt.Println(err)
	// Output:
	// 10
	// <nil>
}

func ExampleToUnsignedByteArrayLen() {
	bytes, err := vnum.ToUnsignedByteArrayLen(4, big.NewInt(255))
	fmt.Printf("%v\n", bytes)
	fmt.Println(err)
	// Output:
	// [0 0 0 255]
	// <nil>
}

func ExampleAdd() {
	fmt.Println(vnum.Add(1.5, 2.25, -0.75))
	// Output: 3
}

func ExampleSub() {
	fmt.Println(vnum.Sub(10, 2.5, 1.5))
	// Output: 6
}

func ExampleCompare() {
	fmt.Println(vnum.Compare(3, 5))
	fmt.Println(vnum.Compare("go", "go"))
	// Output:
	// -1
	// 0
}

func ExampleAppendRange() {
	values := vnum.AppendRange(1, 5, 2, []int{0})
	fmt.Println(values)
	// Output: [0 1 3 5]
}

func ExampleCeilDiv() {
	fmt.Println(vnum.CeilDiv(10, 3))
	fmt.Println(vnum.CeilDiv(9, 3))
	// Output:
	// 4
	// 3
}

func ExampleBinaryToLong() {
	value, err := vnum.BinaryToLong("100000000")
	fmt.Println(value)
	fmt.Println(err)
	// Output:
	// 256
	// <nil>
}
