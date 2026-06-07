package num

import (
	cryptorand "crypto/rand"
	"errors"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"testing"
)

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("forced random failure") }

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

// Tests for numeric helper functions.

func TestNumberArith(t *testing.T) {
	if !Equals(NumberAdd(0.1, 0.2), 0.3) {
		t.Fatalf("Add failed: %v", NumberAdd(0.1, 0.2))
	}
	if !Equals(NumberSub(1.0, 0.7), 0.3) {
		t.Fatalf("Sub failed: %v", NumberSub(1.0, 0.7))
	}
	if !Equals(NumberMul(0.1, 3), 0.3) {
		t.Fatalf("Mul failed: %v", NumberMul(0.1, 3))
	}
	if got := NumberDiv(10, 3, 2); !Equals(got, 3.33) {
		t.Fatalf("Div failed: %v", got)
	}
}

func TestRound(t *testing.T) {
	if Round(3.14159, 2) != 3.14 {
		t.Fatalf("Round 3.14")
	}
	if Round(3.145, 2) != 3.15 {
		t.Fatalf("Round half up")
	}
	if Round(-3.145, 2) != -3.15 {
		t.Fatalf("Round neg half up")
	}
}

func TestIsNumber(t *testing.T) {
	if !IsNumber("123") || !IsNumber("-3.14") || IsNumber("abc") {
		t.Fatalf("IsNumber failed")
	}
	if !IsInteger("-12") || IsInteger("12.3") {
		t.Fatalf("IsInteger failed")
	}
	if !IsDigits("12345") || IsDigits("-12") {
		t.Fatalf("IsDigits failed")
	}
}

func TestMinMaxSumAvg(t *testing.T) {
	if Min(3, 1, 2) != 1 {
		t.Fatalf("Min failed")
	}
	if Max(3, 1, 2) != 3 {
		t.Fatalf("Max failed")
	}
	if Sum(1, 2, 3, 4) != 10 {
		t.Fatalf("Sum failed")
	}
	if math.Abs(Avg(1, 2, 3, 4)-2.5) > 1e-9 {
		t.Fatalf("Avg failed")
	}
}

func TestRangeFunc(t *testing.T) {
	r := Range(0, 5, 1)
	if len(r) != 5 || r[0] != 0 || r[4] != 4 {
		t.Fatalf("Range asc: %v", r)
	}
	r = Range(5, 0, -1)
	if len(r) != 5 || r[0] != 5 || r[4] != 1 {
		t.Fatalf("Range desc: %v", r)
	}
	r = Range(0, 10, 3)
	if len(r) != 4 || r[3] != 9 {
		t.Fatalf("Range step: %v", r)
	}
}

func TestNumberChecksAndFormat(t *testing.T) {
	if !IsNumber("0x1A") || !IsNumber("123E3") || !IsNumber("123D") || !IsNumber("+123") {
		t.Fatal("IsNumber should support hex, scientific, type suffix and signs")
	}
	if IsNumber("0x") || IsNumber("1E-") || !IsLong("9223372036854775807") || !IsDouble("1.23") || IsDouble("123") {
		t.Fatal("number checks failed")
	}
	if !IsPrimes(97) || IsPrimes(100) {
		t.Fatal("prime check failed")
	}
	if DecimalFormatMoney(12345.6) != "12,345.60" || FormatPercent(0.1234, 2) != "12.34%" {
		t.Fatalf("format failed: %s %s", DecimalFormatMoney(12345.6), FormatPercent(0.1234, 2))
	}
	if CeilDiv(10, 3) != 4 || RoundStr(1.2, 2) != "1.20" || RoundHalfEvenFloat(2.5, 0) != 2 || RoundDownFloat(1.29, 1) != 1.2 {
		t.Fatal("round helpers failed")
	}
}

func TestFormatWithOptionsUsesProviders(t *testing.T) {
	floatCalls := 0
	floatFormatter := func(v float64, fmt byte, prec, bitSize int) string {
		floatCalls++
		if fmt == 'f' && prec == 2 && bitSize == 64 {
			return "custom-float"
		}
		return "fallback-float"
	}
	if got := RoundStrWithOptions(1.2, 2, WithFormatFloatFunc(floatFormatter)); got != "custom-float" {
		t.Fatalf("RoundStrWithOptions = %q", got)
	}
	if got := DecimalFormatWithOptions("0.00", 1.2, WithFormatFloatFunc(floatFormatter)); got != "custom-float" {
		t.Fatalf("DecimalFormatWithOptions = %q", got)
	}
	if got := ToStrStripWithOptions(1.2, false, WithFormatFloatFunc(func(v float64, fmt byte, prec, bitSize int) string {
		if fmt != 'f' || prec != -1 || bitSize != 64 {
			t.Fatalf("format args fmt=%q prec=%d bitSize=%d", fmt, prec, bitSize)
		}
		return "1.200"
	})); got != "1.200" {
		t.Fatalf("ToStrStripWithOptions = %q", got)
	}
	intCalls := 0
	if got := GetBinaryStrWithOptions(int64(5), WithFormatIntFunc(func(v int64, base int) string {
		intCalls++
		if v != 5 || base != 2 {
			t.Fatalf("format int args v=%d base=%d", v, base)
		}
		return "custom-int"
	})); got != "custom-int" || intCalls != 1 {
		t.Fatalf("GetBinaryStrWithOptions = %q intCalls=%d", got, intCalls)
	}
	if floatCalls < 2 {
		t.Fatalf("float formatter calls = %d", floatCalls)
	}
}

func TestRangeFactorialAndCombinatorics(t *testing.T) {
	randoms := GenerateRandomNumber(1, 10, 5)
	if len(randoms) != 5 {
		t.Fatalf("GenerateRandomNumber length: %v", randoms)
	}
	seen := map[int]struct{}{}
	for _, v := range randoms {
		if v < 1 || v >= 10 {
			t.Fatalf("GenerateRandomNumber value out of range: %v", randoms)
		}
		if _, ok := seen[v]; ok {
			t.Fatalf("GenerateRandomNumber duplicated value: %v", randoms)
		}
		seen[v] = struct{}{}
	}
	bySet := GenerateBySet(1, 10, 5)
	if len(bySet) != 5 {
		t.Fatalf("GenerateBySet length: %v", bySet)
	}
	if got := RangeClosed(1, 5, 2); !reflect.DeepEqual(got, []int{1, 3, 5}) {
		t.Fatalf("RangeClosed asc: %v", got)
	}
	if got := RangeClosed(5, 1, 2); !reflect.DeepEqual(got, []int{5, 3, 1}) {
		t.Fatalf("RangeClosed desc: %v", got)
	}
	if got := AppendRange(1, 3, 1, []int{0}); !reflect.DeepEqual(got, []int{0, 1, 2, 3}) {
		t.Fatalf("AppendRange: %v", got)
	}
	if got, err := Factorial(5); err != nil || got != 120 {
		t.Fatalf("Factorial: %d %v", got, err)
	}
	if got, err := FactorialRange(5, 2); err != nil || got != 60 {
		t.Fatalf("FactorialRange: %d %v", got, err)
	}
	if FactorialBig(big.NewInt(20)).String() != "2432902008176640000" {
		t.Fatal("FactorialBig failed")
	}
	if Sqrt(81) != 9 || ProcessMultiple(7, 5) != 21 || Divisor(24, 18) != 6 || Multiple(4, 6) != 12 {
		t.Fatal("math helpers failed")
	}
}

func TestBinaryCompareAndConversion(t *testing.T) {
	if GetBinaryStr(int8(-1)) != "11111111" || GetBinaryStr(float32(1)) != "00111111100000000000000000000000" {
		t.Fatal("GetBinaryStr failed")
	}
	if got, err := BinaryToInt("1010"); err != nil || got != 10 {
		t.Fatalf("BinaryToInt: %d %v", got, err)
	}
	if got, err := BinaryToLong("1010"); err != nil || got != 10 {
		t.Fatalf("BinaryToLong: %d %v", got, err)
	}
	if Compare(1, 2) >= 0 || !IsGreater(3, 2) || !IsIn(2, 1, 3) || !EqualsExact(0.0, 0.0) || !EqualsChar('A', 'a', true) {
		t.Fatal("compare helpers failed")
	}
	if ToStr(5.0) != "5" || ToBigDecimal("1,234.50").FloatString(2) != "1234.50" || ToBigInteger("123").String() != "123" {
		t.Fatal("to string/big helpers failed")
	}
	if Count(10, 3) != 4 || Zero2One(0) != 1 || NullToZero[int](nil) != 0 || !IsBeside(1, 2) || PartValue(10, 3) != 4 {
		t.Fatal("small helpers failed")
	}
	bi, ok := NewBigInteger("0x10")
	if !ok || bi.Int64() != 16 {
		t.Fatal("NewBigInteger failed")
	}
}

func TestParseBytesValidityAndCalculate(t *testing.T) {
	if ParseInt("0x10") != 16 || ParseInt("123.56") != 123 || ParseLong("123.56") != 123 {
		t.Fatal("parse integer helpers failed")
	}
	if ParseFloat(".125") != 0.125 || ParseDouble("1,234.5") != 1234.5 || ParseIntDefault("bad", 7) != 7 {
		t.Fatal("parse float/default helpers failed")
	}
	b := ToBytes(0x01020304)
	if !reflect.DeepEqual(b, []byte{1, 2, 3, 4}) || ToInt(b) != 0x01020304 {
		t.Fatal("byte conversion failed")
	}
	if got := ToInt(ToBytes(-1)); got != -1 {
		t.Fatalf("signed byte round trip failed: %d", got)
	}
	unsigned, err := ToUnsignedByteArrayLen(4, big.NewInt(255))
	if err != nil || !reflect.DeepEqual(unsigned, []byte{0, 0, 0, 255}) || FromUnsignedByteArray(unsigned).Int64() != 255 {
		t.Fatal("unsigned bytes failed")
	}
	if IsValid(math.Inf(1)) || IsValidFloat32(float32(math.NaN())) || !IsValidNumber(1) {
		t.Fatal("valid number helpers failed")
	}
	result, err := Calculate("(0*1--3)-5/-4-(3*(-2.13))")
	if err != nil || math.Abs(result-10.64) > 1e-9 {
		t.Fatalf("Calculate: %v %v", result, err)
	}
	if ToDouble(float32(1.23)) != 1.23 || !IsOdd(3) || !IsEven(4) || !IsPowerOfTwo(1024) {
		t.Fatal("misc helpers failed")
	}
}

func TestParseWithOptionsUsesProviders(t *testing.T) {
	var intCalled, floatCalled int
	parseInt := func(text string, base, bitSize int) (int64, error) {
		intCalled++
		switch text {
		case "custom-int":
			return 42, nil
		case "1010":
			if base != 2 {
				t.Fatalf("binary base = %d", base)
			}
			return 10, nil
		case "ff":
			if base != 16 {
				t.Fatalf("hex base = %d", base)
			}
			return 255, nil
		default:
			return 0, errors.New("custom int error")
		}
	}
	parseFloat := func(text string, bitSize int) (float64, error) {
		floatCalled++
		if text == "custom-float" {
			return 6.5, nil
		}
		return 0, errors.New("custom float error")
	}

	if got := ParseIntWithOptions("custom-int", WithParseIntFunc(parseInt)); got != 42 {
		t.Fatalf("ParseIntWithOptions = %d", got)
	}
	if got := ParseLongWithOptions("0xff", WithParseIntFunc(parseInt)); got != 255 {
		t.Fatalf("ParseLongWithOptions = %d", got)
	}
	if got := ParseDoubleWithOptions("custom-float", WithParseFloatFunc(parseFloat)); got != 6.5 {
		t.Fatalf("ParseDoubleWithOptions = %v", got)
	}
	if got, err := ParseNumberWithOptions("custom-float", WithParseFloatFunc(parseFloat)); err != nil || got != 6.5 {
		t.Fatalf("ParseNumberWithOptions = %v, %v", got, err)
	}
	if got, err := BinaryToIntWithOptions("1010", WithParseIntFunc(parseInt)); err != nil || got != 10 {
		t.Fatalf("BinaryToIntWithOptions = %d, %v", got, err)
	}
	if !IsNumberWithOptions("custom-float", WithParseFloatFunc(parseFloat)) || !IsIntegerWithOptions("custom-int", WithParseIntFunc(parseInt)) {
		t.Fatal("Is*WithOptions should use custom parsers")
	}
	if got := ParseFloatDefaultWithOptions("custom-float", 1, WithParseFloatFunc(parseFloat)); got != 6.5 {
		t.Fatalf("ParseFloatDefaultWithOptions = %v", got)
	}
	if got := ToBigDecimalWithOptions("not-rat", WithParseFloatFunc(parseFloat)); got.Sign() != 0 {
		t.Fatalf("ToBigDecimalWithOptions fallback = %s", got.String())
	}
	if intCalled == 0 || floatCalled == 0 {
		t.Fatalf("providers not called int=%d float=%d", intCalled, floatCalled)
	}
}

func TestStringArithmeticAndDivisionEdges(t *testing.T) {
	if Add() != 0 || !Equals(Add(0.1, 0.2, 0.3), 0.6) {
		t.Fatalf("Add edge cases failed: %v", Add(0.1, 0.2, 0.3))
	}
	if AddStr("0.1", "", "0.2").FloatString(1) != "0.3" {
		t.Fatalf("AddStr should skip blank values: %s", AddStr("0.1", "", "0.2").FloatString(1))
	}
	if Sub() != 0 || !Equals(Sub(10, 1.25, 2.75), 6) {
		t.Fatalf("Sub edge cases failed: %v", Sub(10, 1.25, 2.75))
	}
	if SubStr().Sign() != 0 || SubStr("10.50", "", "0.50").FloatString(2) != "10.00" {
		t.Fatalf("SubStr edge cases failed: %s", SubStr("10.50", "", "0.50").FloatString(2))
	}
	if Mul() != 0 || !Equals(Mul(0.1, 0.2, 10), 0.2) {
		t.Fatalf("Mul edge cases failed: %v", Mul(0.1, 0.2, 10))
	}
	if MulStr().Sign() != 0 || MulStr("2.5", "4").FloatString(1) != "10.0" || MulStr("2", " ").Sign() != 0 {
		t.Fatal("MulStr edge cases failed")
	}
	if Div(1, 0) != 0 || DivWithMode(1, 0, 2, RoundDown) != 0 {
		t.Fatal("division by zero should return 0")
	}
	if got := NumberDiv(5, 2, -1); got != 2.5 {
		t.Fatalf("negative scale should disable rounding: %v", got)
	}
	if got := DivWithMode(10, 4, 1, RoundDown); got != 2.5 {
		t.Fatalf("DivWithMode RoundDown failed: %v", got)
	}
}

func TestRoundingAndDecimalFormatEdges(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		scale int
		mode  RoundingMode
		want  float64
	}{
		{"半入正数", 2.345, 2, RoundHalfUp, 2.35},
		{"半入负数", -2.345, 2, RoundHalfUp, -2.35},
		{"银行家舍入到偶数", 3.5, 0, RoundHalfEven, 4},
		{"银行家舍入保持偶数", 2.5, 0, RoundHalfEven, 2},
		{"向零截断正数", 1.29, 1, RoundDown, 1.2},
		{"向零截断负数", -1.29, 1, RoundDown, -1.2},
		{"负精度按整数处理", 1.51, -2, RoundHalfUp, 2},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := RoundMode(tt.value, tt.scale, tt.mode); got != tt.want {
				t.Fatalf("RoundMode() = %v, want %v", got, tt.want)
			}
		})
	}
	if got := RoundStr(1.2, -1); got != "1" {
		t.Fatalf("RoundStr negative scale = %q", got)
	}
	formatCases := map[string]string{
		DecimalFormat("", 12.8):           "13",
		DecimalFormat("0", 12.5):          "13",
		DecimalFormat("0.###", 1.2349):    "1.235",
		DecimalFormat(",##0.00", -1234.5): "-1,234.50",
		DecimalFormat("0.0%", 0.126):      "12.6%",
		FormatPercent(0.1, -3):            "10%",
	}
	for got, want := range formatCases {
		if got != want {
			t.Fatalf("decimal format = %q, want %q", got, want)
		}
	}
}

func TestNumberPredicatesComprehensive(t *testing.T) {
	validNumbers := []string{"  +12.30 ", "-0x1f", "6.02e23", "1F", "2d", "3L"}
	for _, s := range validNumbers {
		if !IsNumber(s) {
			t.Fatalf("IsNumber(%q) should be true", s)
		}
	}
	invalidNumbers := []string{"", "  ", "0x", "0xz", "1e-", "abc", "1.2.3"}
	for _, s := range invalidNumbers {
		if IsNumber(s) {
			t.Fatalf("IsNumber(%q) should be false", s)
		}
	}
	if IsInteger("") || IsInteger("12.0") || !IsInteger("-12") {
		t.Fatal("IsInteger edge cases failed")
	}
	if IsLong("9223372036854775808") || !IsLong("-9223372036854775808") {
		t.Fatal("IsLong boundary cases failed")
	}
	if IsDouble("1") || IsDouble("") || !IsDouble("1.0") || !IsDouble("-0.25") {
		t.Fatal("IsDouble edge cases failed")
	}
	if IsDigits("") || IsDigits("12a") || !IsDigits("00123") {
		t.Fatal("IsDigits edge cases failed")
	}
	primeCases := map[int]bool{-1: false, 0: false, 1: false, 2: true, 3: true, 4: false, 25: false, 7919: true}
	for n, want := range primeCases {
		if got := IsPrimes(n); got != want {
			t.Fatalf("IsPrimes(%d) = %v, want %v", n, got, want)
		}
	}
}

func TestRandomRangeAndFactorialEdges(t *testing.T) {
	if got := GenerateRandomNumber(10, 1, 3); len(got) != 0 {
		t.Fatalf("GenerateRandomNumber reversed bounds should be empty because default seed is empty: %v", got)
	}
	if got := GenerateRandomNumber(1, 3, 5); len(got) != 0 {
		t.Fatalf("GenerateRandomNumber oversize should be empty: %v", got)
	}
	if got := GenerateRandomNumberWithSeed(1, 10, 2, []int{7}); len(got) != 0 {
		t.Fatalf("GenerateRandomNumberWithSeed short seed should be empty: %v", got)
	}
	seed := []int{1, 2, 3, 4}
	got := GenerateRandomNumberWithSeed(1, 5, 2, seed)
	if len(got) != 2 || !reflect.DeepEqual(seed, []int{1, 2, 3, 4}) {
		t.Fatalf("GenerateRandomNumberWithSeed should not mutate seed: got=%v seed=%v", got, seed)
	}
	if got := GenerateBySet(5, 1, 0); len(got) != 0 {
		t.Fatalf("GenerateBySet zero size should be empty: %v", got)
	}
	if got := GenerateBySet(1, 2, 3); len(got) != 0 {
		t.Fatalf("GenerateBySet oversize should be empty: %v", got)
	}
	if got := Range(1, 5, 0); !reflect.DeepEqual(got, []int{1, 2, 3, 4}) {
		t.Fatalf("Range zero positive step: %v", got)
	}
	if got := Range(5, 1, 0); !reflect.DeepEqual(got, []int{5, 4, 3, 2}) {
		t.Fatalf("Range zero negative step: %v", got)
	}
	if got := RangeClosed(3, 3, 0); !reflect.DeepEqual(got, []int{3}) {
		t.Fatalf("RangeClosed equal endpoints: %v", got)
	}
	if got := RangeClosed(1, 5, -2); !reflect.DeepEqual(got, []int{1, 3, 5}) {
		t.Fatalf("RangeClosed should normalize step sign: %v", got)
	}
	if got, err := Factorial(0); err != nil || got != 1 {
		t.Fatalf("Factorial(0) = %d, %v", got, err)
	}
	if got, err := Factorial(21); err == nil || got != 0 {
		t.Fatalf("Factorial overflow should fail: %d, %v", got, err)
	}
	if got, err := FactorialRange(0, 10); err != nil || got != 1 {
		t.Fatalf("FactorialRange start zero = %d, %v", got, err)
	}
	if got, err := FactorialRange(2, 5); err != nil || got != 0 {
		t.Fatalf("FactorialRange start smaller should be zero: %d, %v", got, err)
	}
	if got, err := FactorialRange(21, 0); err == nil || got != 0 {
		t.Fatalf("FactorialRange overflow should fail: %d, %v", got, err)
	}
	if FactorialBig(nil).String() != "1" || FactorialBig(big.NewInt(-1)).String() != "1" {
		t.Fatal("FactorialBig nil/negative should be one")
	}
	if FactorialBigRange(big.NewInt(5), big.NewInt(2)).String() != "60" {
		t.Fatal("FactorialBigRange normal case failed")
	}
	if FactorialBigRange(nil, big.NewInt(0)).String() != "1" || FactorialBigRange(big.NewInt(2), big.NewInt(5)).String() != "1" {
		t.Fatal("FactorialBigRange guard cases failed")
	}
}

func TestRandomGenerationWithOptions(t *testing.T) {
	seed := []int{10, 20, 30, 40}
	got := GenRandomNumberWithSeedWithOptions(0, 4, 3, seed, WithRandomReader(&sequenceReader{}))
	if !reflect.DeepEqual(got, []int{10, 20, 40}) {
		t.Fatalf("GenRandomNumberWithSeedWithOptions deterministic = %v", got)
	}
	if !reflect.DeepEqual(seed, []int{10, 20, 30, 40}) {
		t.Fatalf("GenRandomNumberWithSeedWithOptions should not mutate seed: %v", seed)
	}

	got = GenRandomNumberWithOptions(0, 5, 3, WithRandomReader(&sequenceReader{}))
	if !reflect.DeepEqual(got, []int{0, 1, 2}) {
		t.Fatalf("GenRandomNumberWithOptions deterministic = %v", got)
	}

	got = GenBySetWithOptions(0, 5, 3, WithRandomReader(&sequenceReader{}))
	if len(got) != 3 {
		t.Fatalf("GenBySetWithOptions length = %v", got)
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

	if got := GenRandomNumberWithOptions(0, 5, 2, WithRandomReader(errReader{})); !reflect.DeepEqual(got, []int{0, 4}) {
		t.Fatalf("random failure should preserve fallback index behavior: %v", got)
	}
}

func TestCombinatoricsGcdLcmAndBinaryEdges(t *testing.T) {
	if ProcessMultiple(3, 5) != 0 || ProcessMultiple(5, -1) != 0 || ProcessMultiple(5, 5) != 1 || ProcessMultiple(5, 2) != 10 {
		t.Fatal("ProcessMultiple edge cases failed")
	}
	if Divisor(-24, 18) != 6 || Divisor(7, 0) != 7 || Multiple(-4, 6) != 12 || Multiple(0, 6) != 0 {
		t.Fatal("gcd/lcm edge cases failed")
	}
	binaryCases := map[any]string{
		int(5):            "101",
		int8(-2):          "11111110",
		int16(-2):         "1111111111111110",
		int32(-2):         "-10",
		int64(9):          "1001",
		uint8(7):          "111",
		float64(1):        "0011111111110000000000000000000000000000000000000000000000000000",
		big.NewInt(10):    "1010",
		(*big.Int)(nil):   "",
		complex64(1 + 2i): "",
	}
	for input, want := range binaryCases {
		if got := GetBinaryStr(input); got != want {
			t.Fatalf("GetBinaryStr(%T %[1]v) = %q, want %q", input, got, want)
		}
	}
	if _, err := BinaryToInt("102"); err == nil {
		t.Fatal("BinaryToInt should reject invalid binary strings")
	}
	if _, err := BinaryToLong("2"); err == nil {
		t.Fatal("BinaryToLong should reject invalid binary strings")
	}
}

func TestComparisonEqualityAndAggregateEdges(t *testing.T) {
	if Compare(2, 1) != 1 || Compare(1, 1) != 0 || Compare("a", "b") != -1 {
		t.Fatal("Compare cases failed")
	}
	if !IsGreaterOrEqual(2, 2) || !IsLess(1, 2) || !IsLessOrEqual(2, 2) || IsIn(0, 1, 3) {
		t.Fatal("ordered helper cases failed")
	}
	if !Equals(0.1+0.2, 0.3) || Equals(0.1, 0.2) {
		t.Fatal("Equals tolerance cases failed")
	}
	if EqualsExact(0.0, math.Copysign(0, -1)) || !EqualsFloat32Exact(float32(1), float32(1)) || !EqualsInt64(9, 9) {
		t.Fatal("exact equality cases failed")
	}
	if !EqualsBigDecimal(big.NewRat(10, 10), big.NewRat(1, 1)) || EqualsBigDecimal(big.NewRat(1, 1), nil) || !EqualsBigDecimal(nil, nil) {
		t.Fatal("big decimal equality cases failed")
	}
	if !EqualsChar('ß', 'ß', false) || EqualsChar('A', 'a', false) || !EqualsChar('A', 'a', true) {
		t.Fatal("char equality cases failed")
	}
	if Min[int]() != 0 || Max[int]() != 0 || Sum[int]() != 0 || Avg[int]() != 0 {
		t.Fatal("aggregate empty cases failed")
	}
	if Min("b", "a") != "a" || Max("b", "a") != "b" || Sum(1.5, 2.5) != 4 || Avg(1.5, 2.5) != 2 {
		t.Fatal("aggregate normal cases failed")
	}
}

func TestStringBigAndNullConversionEdges(t *testing.T) {
	if ToStrStrip(math.NaN(), true) != "" || ToStrStrip(12.3400, false) != "12.34" || ToStrStrip(12.0, true) != "12" {
		t.Fatal("ToStrStrip edge cases failed")
	}
	v := 12.0
	if ToStrDefault(nil, "fallback") != "fallback" || ToStrDefault(&v, "fallback") != "12" {
		t.Fatal("ToStrDefault edge cases failed")
	}
	if ToBigDecimal("").Sign() != 0 || ToBigDecimal("bad").Sign() != 0 || ToBigDecimal("1,234.25").FloatString(2) != "1234.25" {
		t.Fatal("ToBigDecimal edge cases failed")
	}
	if ToBigInteger("").Sign() != 0 || ToBigInteger("123.9").Int64() != 123 {
		t.Fatal("ToBigInteger edge cases failed")
	}
	if Count(0, 3) != 0 || Count(10, 0) != 0 || Count(9, 3) != 3 || Count(10, 3) != 4 {
		t.Fatal("Count edge cases failed")
	}
	if Null2Zero(nil).Sign() != 0 || NullBigIntToZero(nil).Sign() != 0 || NullBigDecimalToZero(nil).Sign() != 0 {
		t.Fatal("null to zero nil cases failed")
	}
	bi := big.NewInt(5)
	bd := big.NewRat(5, 2)
	if NullBigIntToZero(bi) != bi || NullBigDecimalToZero(bd) != bd || Null2Zero(bd) != bd {
		t.Fatal("null to zero should preserve non-nil pointers")
	}
	x := 7
	if NullToZero(&x) != 7 || Zero2One(9) != 9 {
		t.Fatal("small conversion helpers failed")
	}
}

func TestBigIntegerPartPowAndParseEdges(t *testing.T) {
	bigIntCases := map[string]int64{"42": 42, "-0x10": -16, "#10": 16, "010": 8}
	for input, want := range bigIntCases {
		got, ok := NewBigInteger(input)
		if !ok || got.Int64() != want {
			t.Fatalf("NewBigInteger(%q) = %v/%v, want %d", input, got, ok, want)
		}
	}
	if got, ok := NewBigInteger(""); ok || got != nil {
		t.Fatalf("NewBigInteger blank = %v/%v", got, ok)
	}
	if got, ok := NewBigInteger("0xzz"); ok || got != nil {
		t.Fatalf("NewBigInteger invalid = %v/%v", got, ok)
	}
	if !IsBeside(2, 1) || !IsBeside[int64](1, 2) || IsBeside(1, 3) {
		t.Fatal("IsBeside cases failed")
	}
	if PartValueWithMode(10, 0, true) != 0 || PartValueWithMode(10, 3, false) != 3 || PartValueWithMode(10, 3, true) != 4 {
		t.Fatal("PartValueWithMode cases failed")
	}
	if Pow(2, 3) != 8 || Pow(2, -3) != 0.13 || PowWithMode(2, -3, 2, RoundDown) != 0.12 {
		t.Fatal("Pow cases failed")
	}
	if IsPowerOfTwo(0) || IsPowerOfTwo(-2) || !IsPowerOfTwo(1) || !IsPowerOfTwo(1024) || IsPowerOfTwo(1023) {
		t.Fatal("IsPowerOfTwo cases failed")
	}
	if ParseInt("") != 0 || ParseInt(".5") != 0 || ParseInt("1e3") != 0 || ParseInt("1,234.9") != 1234 {
		t.Fatal("ParseInt edge cases failed")
	}
	if ParseLong("") != 0 || ParseLong(".5") != 0 || ParseLong("0x7f") != 127 || ParseLong("1,234.9") != 1234 {
		t.Fatal("ParseLong edge cases failed")
	}
	if ParseDouble("") != 0 || ParseDouble("1,234.5") != 1234.5 || ParseFloat("2.5") != 2.5 {
		t.Fatal("ParseFloat/ParseDouble edge cases failed")
	}
	if got, err := ParseNumber("+1,234.5"); err != nil || got != 1234.5 {
		t.Fatalf("ParseNumber plus/comma failed: %v %v", got, err)
	}
	if got, err := ParseNumber("0x10"); err != nil || got != 16 {
		t.Fatalf("ParseNumber hex failed: %v %v", got, err)
	}
	if _, err := ParseNumber("bad"); err == nil {
		t.Fatal("ParseNumber should reject invalid input")
	}
	if ParseIntDefault("", 7) != 7 || ParseIntDefault("bad", 7) != 7 || ParseIntDefault("1,234", 7) != 1234 {
		t.Fatal("ParseIntDefault cases failed")
	}
	if ParseLongDefault("", 8) != 8 || ParseLongDefault("bad", 8) != 8 || ParseLongDefault("1,234", 8) != 1234 {
		t.Fatal("ParseLongDefault cases failed")
	}
	if ParseFloatDefault("", 1.5) != 1.5 || ParseFloatDefault("bad", 1.5) != 1.5 || ParseFloatDefault("1,234.5", 1.5) != 1234.5 {
		t.Fatal("ParseFloatDefault cases failed")
	}
	if ParseDoubleDefault("", 2.5) != 2.5 || ParseDoubleDefault("bad", 2.5) != 2.5 || ParseDoubleDefault("1,234.5", 2.5) != 1234.5 {
		t.Fatal("ParseDoubleDefault cases failed")
	}
}

func TestByteUnsignedValidityAndExpressionEdges(t *testing.T) {
	if ToInt(nil) != 0 || ToInt([]byte{1, 2, 3}) != 0 {
		t.Fatal("ToInt short input should be zero")
	}
	if ToUnsignedByteArray(nil) != nil {
		t.Fatal("ToUnsignedByteArray nil should be nil")
	}
	if got := ToUnsignedByteArray(big.NewInt(0)); len(got) != 0 {
		t.Fatalf("ToUnsignedByteArray zero should be empty: %v", got)
	}
	if _, err := ToUnsignedByteArrayLen(1, big.NewInt(256)); err == nil {
		t.Fatal("ToUnsignedByteArrayLen should reject values that exceed requested length")
	}
	if got, err := ToUnsignedByteArrayLen(0, big.NewInt(0)); err != nil || len(got) != 0 {
		t.Fatalf("ToUnsignedByteArrayLen zero length/value = %v, %v", got, err)
	}
	if FromUnsignedByteArray(nil).Sign() != 0 || FromUnsignedByteArrayRange([]byte{1, 2, 3, 4}, 1, 2).Int64() != 0x0203 {
		t.Fatal("FromUnsignedByteArray cases failed")
	}
	if FromUnsignedByteArrayRange([]byte{1, 2}, -1, 1).Sign() != 0 || FromUnsignedByteArrayRange([]byte{1, 2}, 1, 3).Sign() != 0 {
		t.Fatal("FromUnsignedByteArrayRange invalid ranges should be zero")
	}
	if IsValidNumber(nil) || IsValidNumber(math.NaN()) || IsValidNumber(float32(math.Inf(-1))) || !IsValidNumber("not-a-number") {
		t.Fatal("IsValidNumber cases failed")
	}
	if IsValid(math.NaN()) || IsValid(math.Inf(1)) || !IsValid(1.23) || IsValidFloat32(float32(math.Inf(1))) || !IsValidFloat32(1.23) {
		t.Fatal("valid finite checks failed")
	}
	toDoubleCases := []struct {
		input any
		want  float64
	}{
		{float32(1.25), 1.25},
		{float64(2.5), 2.5},
		{int(-3), -3},
		{int64(4), 4},
		{uint64(5), 5},
		{"bad", 0},
	}
	for _, tt := range toDoubleCases {
		if got := ToDouble(tt.input); got != tt.want {
			t.Fatalf("ToDouble(%T) = %v, want %v", tt.input, got, tt.want)
		}
	}
	formatCalled := false
	parseCalled := false
	if got := ToDoubleWithOptions(float32(1.25),
		WithDoubleFormatFloatFunc(func(v float64, fmtByte byte, prec, bitSize int) string {
			formatCalled = true
			return strconv.FormatFloat(v*2, fmtByte, prec, bitSize)
		}),
		WithDoubleParseFloatFunc(func(s string, bitSize int) (float64, error) {
			parseCalled = true
			return strconv.ParseFloat(s, bitSize)
		}),
	); got != 2.5 || !formatCalled || !parseCalled {
		t.Fatalf("ToDoubleWithOptions = %v format=%v parse=%v", got, formatCalled, parseCalled)
	}
	calcCases := map[string]float64{
		"1 + 2 * 3":   7,
		"(1 + 2) * 3": 9,
		"10 % 4":      2,
		"--2 + +3":    5,
		" 3.5 / 2 ":   1.75,
	}
	for expr, want := range calcCases {
		got, err := Calculate(expr)
		if err != nil || math.Abs(got-want) > 1e-9 {
			t.Fatalf("Calculate(%q) = %v, %v, want %v", expr, got, err, want)
		}
	}
	calcParseCalled := false
	got, err := CalculateWithOptions("5 + 2", WithParseFloatFunc(func(s string, bitSize int) (float64, error) {
		calcParseCalled = true
		if s == "5" {
			return 5, nil
		}
		return strconv.ParseFloat(s, bitSize)
	}))
	if err != nil || got != 7 || !calcParseCalled {
		t.Fatalf("CalculateWithOptions = %v, %v called=%v", got, err, calcParseCalled)
	}
	invalidExpressions := []string{"", "1+", "(1+2", "1 2", "abc"}
	for _, expr := range invalidExpressions {
		if got, err := Calculate(expr); err == nil {
			t.Fatalf("Calculate(%q) should fail, got %v", expr, got)
		}
	}
	if secureIntn(0) != 0 || secureIntn(-1) != 0 {
		t.Fatal("secureIntn non-positive max should be zero")
	}
	for i := 0; i < 20; i++ {
		if got := secureIntn(3); got < 0 || got >= 3 {
			t.Fatalf("secureIntn result out of range: %d", got)
		}
	}
	if !IsOdd(-3) || IsOdd(-2) || !IsEven(-2) || IsEven(-3) {
		t.Fatal("odd/even negative cases failed")
	}
}

func TestRemainingInternalHelperEdges(t *testing.T) {
	if IsLong("   ") {
		t.Fatal("IsLong blank should be false")
	}
	if got := RangeClosed(5, 1, 0); !reflect.DeepEqual(got, []int{5, 4, 3, 2, 1}) {
		t.Fatalf("RangeClosed descending zero step: %v", got)
	}
	if Max(1, 3, 2) != 3 {
		t.Fatal("Max should update when a later value is larger")
	}
	if ParseInt("123") != 123 || ParseLong("123") != 123 {
		t.Fatal("ParseInt/ParseLong direct integer branch failed")
	}
	if stripTrailingZeros("12") != "12" || stripTrailingZeros("1.2300") != "1.23" || stripTrailingZeros("1.2300e2") != "1.2300e2" {
		t.Fatal("stripTrailingZeros cases failed")
	}
	if addThousands("123") != "123" || addThousands("1234") != "1,234" {
		t.Fatal("addThousands integer cases failed")
	}
	if got, err := Calculate("1+"); err == nil || got != 0 {
		t.Fatalf("Calculate trailing plus should fail: %v %v", got, err)
	}
	if got, err := Calculate("1*"); err == nil || got != 0 {
		t.Fatalf("Calculate trailing multiply should fail: %v %v", got, err)
	}
	oldReader := cryptorand.Reader
	cryptorand.Reader = errReader{}
	t.Cleanup(func() { cryptorand.Reader = oldReader })
	if got := secureIntn(10); got != 0 {
		t.Fatalf("secureIntn should return 0 when crypto random fails: %d", got)
	}
}
