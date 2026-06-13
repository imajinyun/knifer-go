package vnum

import (
	"math"
	"math/big"
	"reflect"
	"strconv"
	"testing"
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

func TestNumFacade(t *testing.T) {
	if Add(1.2, 2.3) != 3.5 || Sub(5, 2) != 3 || Mul(2, 3) != 6 || Div(10, 4, 1) != 2.5 {
		t.Fatal("arithmetic helpers failed")
	}
	if Round(1.234, 2) != 1.23 || !IsNumber("3.14") || !IsInteger("42") || !IsDigits("123") {
		t.Fatal("format/check helpers failed")
	}
	if Min(3, 1, 2) != 1 || Max(3, 1, 2) != 3 || Sum(1, 2, 3) != 6 || Avg(2, 4) != 3 {
		t.Fatal("aggregate helpers failed")
	}
	seq := Range(1, 5, 2)
	if len(seq) != 2 || seq[0] != 1 || seq[1] != 3 {
		t.Fatalf("Range failed: %v", seq)
	}
	if !Equals(0.1+0.2, 0.3) || DecimalFormat("0.00", 1.2) != "1.20" {
		t.Fatal("equals/format helpers failed")
	}
}

func TestGenericNumberFacade(t *testing.T) {
	if got := SumNumber[int](-2, 5, 7); got != 10 {
		t.Fatalf("SumNumber[int] = %v", got)
	}
	if got := SumNumber[float64](1.25, 2.5, -0.75); got != 3 {
		t.Fatalf("SumNumber[float64] = %v", got)
	}
	if got := AvgNumber[uint](2, 5, 8); got != 5 {
		t.Fatalf("AvgNumber[uint] = %v", got)
	}
	if got := MinInteger[int](-2, 5); got != -2 {
		t.Fatalf("MinInteger[int] = %d", got)
	}
	if got := MinIntegers[int](4, -8, 2); got != -8 {
		t.Fatalf("MinIntegers[int] = %d", got)
	}
	if got := MaxInteger[int](-2, 5); got != 5 {
		t.Fatalf("MaxInteger[int] = %d", got)
	}
	if got := MaxIntegers[int](4, -8, 2); got != 4 {
		t.Fatalf("MaxIntegers[int] = %d", got)
	}
	if got := MinFloat64s(3.5, -1.25, 2); got != -1.25 {
		t.Fatalf("MinFloat64s = %v", got)
	}
	if got := MaxFloat64s(3.5, -1.25, 2); got != 3.5 {
		t.Fatalf("MaxFloat64s = %v", got)
	}
	if got := AvgNumber[int](); got != 0 {
		t.Fatalf("AvgNumber empty = %v", got)
	}
	if got := MinIntegers[int](); got != 0 {
		t.Fatalf("MinIntegers empty = %d", got)
	}
	if got := MaxIntegers[int](); got != 0 {
		t.Fatalf("MaxIntegers empty = %d", got)
	}
	if got := AbsInteger[int](-12); got != 12 {
		t.Fatalf("AbsInteger[int] = %d", got)
	}
	if got := AbsInteger[int8](math.MinInt8); got != 0 {
		t.Fatalf("AbsInteger overflow = %d", got)
	}
	abs, err := AbsIntegerE[int8](math.MinInt8)
	if err == nil || abs != 0 {
		t.Fatalf("AbsIntegerE overflow = %d, %v", abs, err)
	}
	if got := AbsFloat32(-3.5); got != 3.5 {
		t.Fatalf("AbsFloat32 = %v", got)
	}
	if got := AbsFloat64(math.Inf(-1)); !math.IsInf(got, 1) {
		t.Fatalf("AbsFloat64(-Inf) = %v", got)
	}
}

func TestNumRandomOptionsFacade(t *testing.T) {
	seed := []int{10, 20, 30, 40}
	got := GenRandomNumberWithSeedWithOptions(0, 4, 3, seed, WithRandomReader(&sequenceReader{}))
	if !reflect.DeepEqual(got, []int{10, 20, 40}) {
		t.Fatalf("GenRandomNumberWithSeedWithOptions = %v", got)
	}

	got = GenRandomNumberWithOptions(0, 5, 3, WithRandomReader(&sequenceReader{}))
	if !reflect.DeepEqual(got, []int{0, 1, 2}) {
		t.Fatalf("GenRandomNumberWithOptions = %v", got)
	}

	got = GenBySetWithOptions(0, 5, 3, WithRandomReader(&sequenceReader{}))
	if len(got) != 3 {
		t.Fatalf("GenBySetWithOptions = %v", got)
	}
	seen := map[int]bool{}
	for _, v := range got {
		seen[v] = true
	}
	for _, want := range []int{0, 1, 2} {
		if !seen[want] {
			t.Fatalf("GenBySetWithOptions missing %d in %v", want, got)
		}
	}
}

func TestNumParseFormatAndValidateFacades(t *testing.T) {
	parseIntCalls := 0
	parseFloatCalls := 0
	parseInt := func(s string, base int, bitSize int) (int64, error) {
		parseIntCalls++
		return strconv.ParseInt(s, base, bitSize)
	}
	parseFloat := func(s string, bitSize int) (float64, error) {
		parseFloatCalls++
		return strconv.ParseFloat(s, bitSize)
	}
	opts := []ParseOption{WithParseIntFunc(parseInt), WithParseFloatFunc(parseFloat)}

	if got := ParseInt("0x10"); got != 16 {
		t.Fatalf("ParseInt = %d", got)
	}
	if got := ParseIntWithOptions("42", opts...); got != 42 {
		t.Fatalf("ParseIntWithOptions = %d", got)
	}
	if got := ParseLongWithOptions("1,234", opts...); got != 1234 {
		t.Fatalf("ParseLongWithOptions = %d", got)
	}
	if got := ParseFloatWithOptions("3.5", opts...); got != 3.5 {
		t.Fatalf("ParseFloatWithOptions = %f", got)
	}
	if got := ParseDoubleWithOptions("6.25", opts...); got != 6.25 {
		t.Fatalf("ParseDoubleWithOptions = %f", got)
	}
	if got, err := ParseNumberWithOptions("0x2a", opts...); err != nil || got != 42 {
		t.Fatalf("ParseNumberWithOptions = %f, %v", got, err)
	}
	if parseIntCalls == 0 || parseFloatCalls == 0 {
		t.Fatalf("custom parsers not called: int=%d float=%d", parseIntCalls, parseFloatCalls)
	}

	if got := ParseIntDefaultWithOptions("bad", 7, opts...); got != 7 {
		t.Fatalf("ParseIntDefaultWithOptions = %d", got)
	}
	if got := ParseLongDefaultWithOptions("", 9, opts...); got != 9 {
		t.Fatalf("ParseLongDefaultWithOptions = %d", got)
	}
	if got := ParseFloatDefaultWithOptions("bad", 1.5, opts...); got != 1.5 {
		t.Fatalf("ParseFloatDefaultWithOptions = %f", got)
	}
	if got := ParseDoubleDefaultWithOptions("bad", 2.5, opts...); got != 2.5 {
		t.Fatalf("ParseDoubleDefaultWithOptions = %f", got)
	}
	if !IsNumberWithOptions("0x10", opts...) || !IsIntegerWithOptions("42", opts...) || !IsLongWithOptions("42", opts...) || !IsDoubleWithOptions("3.14", opts...) {
		t.Fatal("numeric validation with options failed")
	}
	if !IsValidNumber(1.25) || IsValid(math.Inf(1)) || IsValidFloat32(float32(math.NaN())) || !IsOdd(3) || !IsEven(4) || !IsPowerOfTwo(64) {
		t.Fatal("validation helpers failed")
	}
}

func TestNumFormatRoundAndConversionFacades(t *testing.T) {
	formatFloatCalls := 0
	formatIntCalls := 0
	formatFloat := func(f float64, fmtByte byte, prec int, bitSize int) string {
		formatFloatCalls++
		return strconv.FormatFloat(f, fmtByte, prec, bitSize)
	}
	formatInt := func(i int64, base int) string {
		formatIntCalls++
		return strconv.FormatInt(i, base)
	}
	formatOpts := []FormatOption{WithFormatFloatFunc(formatFloat), WithFormatIntFunc(formatInt)}

	if got := DecimalFormatWithOptions(",##0.00", 1234.5, formatOpts...); got != "1,234.50" {
		t.Fatalf("DecimalFormatWithOptions = %q", got)
	}
	if got := DecimalFormatMoney(1234.5); got != "1,234.50" {
		t.Fatalf("DecimalFormatMoney = %q", got)
	}
	if got := DecimalFormatMoneyWithOptions(12.3, formatOpts...); got != "12.30" {
		t.Fatalf("DecimalFormatMoneyWithOptions = %q", got)
	}
	if got := FormatPercent(0.125, 1); got != "12.5%" {
		t.Fatalf("FormatPercent = %q", got)
	}
	if got := FormatPercentWithOptions(0.5, 0, formatOpts...); got != "50%" {
		t.Fatalf("FormatPercentWithOptions = %q", got)
	}
	if got := ToStr(12.3400); got != "12.34" {
		t.Fatalf("ToStr = %q", got)
	}
	if got := ToStrWithOptions(10, formatOpts...); got != "10" {
		t.Fatalf("ToStrWithOptions = %q", got)
	}
	if got := ToStrDefault(nil, "empty"); got != "empty" {
		t.Fatalf("ToStrDefault nil = %q", got)
	}
	value := 7.5
	if got := ToStrDefaultWithOptions(&value, "empty", formatOpts...); got != "7.5" {
		t.Fatalf("ToStrDefaultWithOptions = %q", got)
	}
	if got := ToStrStripWithOptions(7.500, false, formatOpts...); got != "7.5" {
		t.Fatalf("ToStrStripWithOptions = %q", got)
	}
	if got := GetBinaryStrWithOptions(10, formatOpts...); got != "1010" {
		t.Fatalf("GetBinaryStrWithOptions = %q", got)
	}
	if _, err := BinaryToIntWithOptions("1010", WithParseIntFunc(strconv.ParseInt)); err != nil {
		t.Fatalf("BinaryToIntWithOptions: %v", err)
	}
	if got, err := BinaryToLongWithOptions("1010", WithParseIntFunc(strconv.ParseInt)); err != nil || got != 10 {
		t.Fatalf("BinaryToLongWithOptions = %d, %v", got, err)
	}
	if formatFloatCalls == 0 || formatIntCalls == 0 {
		t.Fatalf("custom formatters not called: float=%d int=%d", formatFloatCalls, formatIntCalls)
	}

	if got := RoundMode(1.25, 1, RoundHalfEven); got != 1.2 {
		t.Fatalf("RoundMode half even = %f", got)
	}
	if got := RoundStrWithOptions(1.234, 2, formatOpts...); got != "1.23" {
		t.Fatalf("RoundStrWithOptions = %q", got)
	}
	if got := RoundHalfEvenFloat(1.25, 1); got != 1.2 {
		t.Fatalf("RoundHalfEvenFloat = %f", got)
	}
	if got := RoundDownFloat(1.29, 1); got != 1.2 {
		t.Fatalf("RoundDownFloat = %f", got)
	}

	doubleCalls := 0
	if got := ToDoubleWithOptions(float32(1.25),
		WithDoubleFormatFloatFunc(strconv.FormatFloat),
		WithDoubleParseFloatFunc(func(s string, bitSize int) (float64, error) {
			doubleCalls++
			return strconv.ParseFloat(s, bitSize)
		}),
	); got != 1.25 || doubleCalls == 0 {
		t.Fatalf("ToDoubleWithOptions = %f calls=%d", got, doubleCalls)
	}
}

func TestNumBigRangeFactorialAndBinaryFacades(t *testing.T) {
	if got := ToBigDecimal("1,234.50"); got.String() != "2469/2" {
		t.Fatalf("ToBigDecimal = %s", got)
	}
	if got := ToBigDecimalWithOptions("bad", WithParseFloatFunc(func(string, int) (float64, error) { return 2.5, nil })); got.String() != "5/2" {
		t.Fatalf("ToBigDecimalWithOptions fallback = %s", got)
	}
	if got := ToBigInteger("42"); got.String() != "42" {
		t.Fatalf("ToBigInteger = %s", got)
	}
	if got := Null2Zero(nil); got.Sign() != 0 {
		t.Fatalf("Null2Zero = %s", got)
	}
	if Zero2One(0) != 1 || Count(10, 3) != 4 {
		t.Fatal("Zero2One/Count failed")
	}
	var n int
	if NullToZero(&n) != 0 || NullToZero[int](nil) != 0 {
		t.Fatal("NullToZero failed")
	}
	if NullBigIntToZero(nil).Sign() != 0 || NullBigDecimalToZero(nil).Sign() != 0 {
		t.Fatal("null big conversions failed")
	}
	if got, ok := NewBigInteger("0x10"); !ok || got.Int64() != 16 {
		t.Fatalf("NewBigInteger = %v, %v", got, ok)
	}
	if got := PartValue(10, 3); got != 4 {
		t.Fatalf("PartValue = %d", got)
	}
	if got := PartValueWithMode(10, 3, false); got != 3 {
		t.Fatalf("PartValueWithMode = %d", got)
	}

	if got, err := Factorial(5); err != nil || got != 120 {
		t.Fatalf("Factorial = %d, %v", got, err)
	}
	if _, err := Factorial(21); err == nil {
		t.Fatal("Factorial overflow error = nil")
	}
	if got, err := FactorialRange(5, 2); err != nil || got != 60 {
		t.Fatalf("FactorialRange = %d, %v", got, err)
	}
	if got := FactorialBig(big.NewInt(5)); got.String() != "120" {
		t.Fatalf("FactorialBig = %s", got)
	}
	if got := FactorialBigRange(big.NewInt(5), big.NewInt(2)); got.String() != "60" {
		t.Fatalf("FactorialBigRange = %s", got)
	}
	if Sqrt(16) != 4 || ProcessMultiple(5, 2) != 10 || Divisor(24, 18) != 6 || Multiple(4, 6) != 12 {
		t.Fatal("factorial helpers failed")
	}
	if got := RangeClosed(1, 3, 0); !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("RangeClosed = %v", got)
	}
	if got := AppendRange(3, 1, 1, []int{0}); !reflect.DeepEqual(got, []int{0, 3, 2, 1}) {
		t.Fatalf("AppendRange = %v", got)
	}
	if got := GenerateRandomNumberWithSeed(0, 4, 2, []int{1, 2, 3}); len(got) != 2 {
		t.Fatalf("GenerateRandomNumberWithSeed = %v", got)
	}
	if got := GenerateRandomNumber(0, 3, 2); len(got) != 2 {
		t.Fatalf("GenerateRandomNumber = %v", got)
	}
	if got := GenerateBySet(0, 3, 2); len(got) != 2 {
		t.Fatalf("GenerateBySet = %v", got)
	}

	if !EqualsExact(1, 1) || !EqualsFloat32Exact(1, 1) || !EqualsInt64(1, 1) || !EqualsBigDecimal(big.NewRat(1, 2), big.NewRat(2, 4)) || !EqualsChar('A', 'a', true) {
		t.Fatal("equals helpers failed")
	}
	if Compare(1, 2) != -1 || !IsGreater(2, 1) || !IsGreaterOrEqual(2, 2) || !IsLess(1, 2) || !IsLessOrEqual(2, 2) || !IsIn(2, 1, 3) {
		t.Fatal("comparison helpers failed")
	}
	if got, err := ToUnsignedByteArrayLen(2, big.NewInt(255)); err != nil || !reflect.DeepEqual(got, []byte{0, 255}) {
		t.Fatalf("ToUnsignedByteArrayLen = %v, %v", got, err)
	}
	if _, err := ToUnsignedByteArrayLen(1, big.NewInt(256)); err == nil {
		t.Fatal("ToUnsignedByteArrayLen overflow error = nil")
	}
	if got := FromUnsignedByteArray([]byte{1, 0}); got.Int64() != 256 {
		t.Fatalf("FromUnsignedByteArray = %s", got)
	}
	if got := FromUnsignedByteArrayRange([]byte{0, 1, 0}, 1, 2); got.Int64() != 256 {
		t.Fatalf("FromUnsignedByteArrayRange = %s", got)
	}
	if ToInt(ToBytes(12345)) != 12345 {
		t.Fatal("ToBytes/ToInt round trip failed")
	}
}

func TestNumCalculateFacadeWithCustomParser(t *testing.T) {
	calls := 0
	got, err := CalculateWithOptions("1 + 2 * 3", WithParseFloatFunc(func(s string, bitSize int) (float64, error) {
		calls++
		return strconv.ParseFloat(s, bitSize)
	}))
	if err != nil || got != 7 || calls == 0 {
		t.Fatalf("CalculateWithOptions = %f, %v calls=%d", got, err, calls)
	}
	if _, err := Calculate("1 + * 2"); err == nil {
		t.Fatal("Calculate invalid expression error = nil")
	}
}
