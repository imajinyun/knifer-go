package num

import (
	"errors"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"testing"
)

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
