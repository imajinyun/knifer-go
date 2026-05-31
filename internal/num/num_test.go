package num

import (
	"math"
	"math/big"
	"reflect"
	"testing"
)

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
