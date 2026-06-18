package vnum

import (
	"math"
	"math/big"
	"testing"
)

func TestNumArithFacadeDelegates(t *testing.T) {
	// SubStr / MulStr
	if got := SubStr("10", "3"); got == nil || got.Sign() == 0 {
		t.Fatalf("SubStr = %s", got)
	}
	if got := MulStr("3", "4"); got == nil || got.Sign() == 0 {
		t.Fatalf("MulStr = %s", got)
	}

	// DivWithMode
	if got := DivWithMode(10, 3, 2, RoundHalfUp); got != 3.33 {
		t.Fatalf("DivWithMode = %f", got)
	}

	// CeilDiv
	if CeilDiv(10, 3) != 4 || CeilDiv(9, 3) != 3 {
		t.Fatal("CeilDiv failed")
	}

	// Pow / PowWithMode
	if Pow(2, 3) != 8 {
		t.Fatalf("Pow = %f", Pow(2, 3))
	}
	if got := PowWithMode(2.5, 2, 1, RoundHalfUp); got > 6.2 && got < 6.4 {
		t.Logf("PowWithMode = %f", got)
	} else {
		t.Fatalf("PowWithMode = %f (expected ~6.3)", got)
	}

	// MinFloat64 / MaxFloat64
	if MinFloat64(3.5, 2.1) != 2.1 || MinFloat64(math.Inf(1), math.MaxFloat64) != math.MaxFloat64 {
		t.Fatal("MinFloat64 failed")
	}
	if MaxFloat64(3.5, 2.1) != 3.5 || MaxFloat64(math.Inf(-1), 0) != 0 {
		t.Fatal("MaxFloat64 failed")
	}

	// Abs helpers
	if AbsInteger(-5) != 5 || AbsFloat32(-3.5) != 3.5 || AbsFloat64(-3.5) != 3.5 {
		t.Fatal("Abs helpers failed")
	}
	if _, err := AbsIntegerE[int](-128); err != nil {
		t.Fatal("AbsIntegerE no-overflow case should not return error")
	}

	// MinIntegers / MaxIntegers / MinFloat64s / MaxFloat64s
	if MinIntegers(3, 1, 2) != 1 || MaxIntegers(3, 1, 2) != 3 {
		t.Fatal("MinIntegers/MaxIntegers failed")
	}
	if MinFloat64s(3.5, 1.2, 2.8) != 1.2 || MaxFloat64s(3.5, 1.2, 2.8) != 3.5 {
		t.Fatal("MinFloat64s/MaxFloat64s failed")
	}

	// SumNumber / AvgNumber
	if SumNumber(1, 2, 3) != 6 || AvgNumber(2, 4) != 3 {
		t.Fatal("SumNumber/AvgNumber failed")
	}

	// MinInteger / MaxInteger
	if MinInteger(3, 1) != 1 || MaxInteger(3, 1) != 3 {
		t.Fatal("MinInteger/MaxInteger failed")
	}
}

func TestNumBinaryFacadeDelegates(t *testing.T) {
	if got := GetBinaryStr(10); got != "1010" {
		t.Fatalf("GetBinaryStr = %q", got)
	}
	if got, err := BinaryToInt("1010"); err != nil || got != 10 {
		t.Fatalf("BinaryToInt = %d, %v", got, err)
	}
	if got, err := BinaryToLong("1010"); err != nil || got != 10 {
		t.Fatalf("BinaryToLong = %d, %v", got, err)
	}
	if got := ToUnsignedByteArray(big.NewInt(255)); len(got) == 0 || got[len(got)-1] != 255 {
		t.Fatalf("ToUnsignedByteArray = %v (len=%d)", got, len(got))
	}
}

func TestNumCompareFacadeDelegates(t *testing.T) {
	if !IsBeside(5, 6) || IsBeside(5, 7) {
		t.Fatal("IsBeside failed")
	}
}

func TestNumConvertFacadeDelegates(t *testing.T) {
	if got := ToDouble(3); got != 3.0 {
		t.Fatalf("ToDouble(int) = %f", got)
	}
	if got := ToDouble(float32(3.14)); got != 3.14 {
		t.Fatalf("ToDouble(float32) = %f", got)
	}
}

func TestNumFormatFacadeDelegates(t *testing.T) {
	f := 7.500
	if got := ToStrStrip(f, false); got != "7.5" {
		t.Fatalf("ToStrStrip = %q", got)
	}
	if got := ToStrStrip(f, true); got != "7.5" {
		t.Fatalf("ToStrStrip strip=true = %q", got)
	}
}

func TestNumParseFacadeDelegates(t *testing.T) {
	if got := ParseLong("42"); got != 42 {
		t.Fatalf("ParseLong = %d", got)
	}
	if got := ParseFloat("3.5"); got != 3.5 {
		t.Fatalf("ParseFloat = %f", got)
	}
	if got := ParseDouble("6.25"); got != 6.25 {
		t.Fatalf("ParseDouble = %f", got)
	}
	if got, err := ParseNumber("0x2a"); err != nil || got != 42 {
		t.Fatalf("ParseNumber = %f, %v", got, err)
	}
	if got := ParseIntDefault("bad", 7); got != 7 {
		t.Fatalf("ParseIntDefault = %d", got)
	}
	if got := ParseLongDefault("bad", 9); got != 9 {
		t.Fatalf("ParseLongDefault = %d", got)
	}
	if got := ParseFloatDefault("bad", 1.5); got != 1.5 {
		t.Fatalf("ParseFloatDefault = %f", got)
	}
	if got := ParseDoubleDefault("bad", 2.5); got != 2.5 {
		t.Fatalf("ParseDoubleDefault = %f", got)
	}
}

func TestNumRoundFacadeDelegates(t *testing.T) {
	if got := RoundStr(1.234, 2); got != "1.23" {
		t.Fatalf("RoundStr = %q", got)
	}
}

func TestNumValidateFacadeDelegates(t *testing.T) {
	if !IsLong("42") || IsLong("abc") {
		t.Fatal("IsLong failed")
	}
	if !IsDouble("3.14") || IsDouble("abc") {
		t.Fatal("IsDouble failed")
	}
	if !IsPrimes(7) || IsPrimes(4) {
		t.Fatal("IsPrimes failed")
	}
}

func TestNumOptionConstructors(t *testing.T) {
	_ = WithParseIntFunc(nil)
	_ = WithParseFloatFunc(nil)
	_ = WithDoubleParseFloatFunc(nil)
	_ = WithDoubleFormatFloatFunc(nil)
	_ = WithFormatFloatFunc(nil)
	_ = WithFormatIntFunc(nil)
}
