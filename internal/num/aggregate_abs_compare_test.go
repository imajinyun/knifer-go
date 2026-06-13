package num

import (
	"math"
	"math/big"
	"testing"
)

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

func TestGenericNumberAggregates(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "signed integers",
			run: func(t *testing.T) {
				if got := SumNumber[int](-2, 5, 7); got != 10 {
					t.Fatalf("SumNumber[int] = %v", got)
				}
				if got := AvgNumber[int](-2, 5, 7); got != 10.0/3.0 {
					t.Fatalf("AvgNumber[int] = %v", got)
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
			},
		},
		{
			name: "unsigned integers",
			run: func(t *testing.T) {
				if got := SumNumber[uint](2, 5, 7); got != 14 {
					t.Fatalf("SumNumber[uint] = %v", got)
				}
				if got := AvgNumber[uint](2, 5, 8); got != 5 {
					t.Fatalf("AvgNumber[uint] = %v", got)
				}
				if got := MinIntegers[uint](4, 8, 2); got != 2 {
					t.Fatalf("MinIntegers[uint] = %d", got)
				}
				if got := MaxIntegers[uint](4, 8, 2); got != 8 {
					t.Fatalf("MaxIntegers[uint] = %d", got)
				}
			},
		},
		{
			name: "floats",
			run: func(t *testing.T) {
				if got := SumNumber[float64](1.25, 2.5, -0.75); got != 3 {
					t.Fatalf("SumNumber[float64] = %v", got)
				}
				if got := AvgNumber[float32](1.5, 2.5); got != 2 {
					t.Fatalf("AvgNumber[float32] = %v", got)
				}
				if got := MinFloat64(1.25, -3.5); got != -3.5 {
					t.Fatalf("MinFloat64 = %v", got)
				}
				if got := MaxFloat64(1.25, -3.5); got != 1.25 {
					t.Fatalf("MaxFloat64 = %v", got)
				}
				if got := MinFloat64s(3.5, -1.25, 2); got != -1.25 {
					t.Fatalf("MinFloat64s = %v", got)
				}
				if got := MaxFloat64s(3.5, -1.25, 2); got != 3.5 {
					t.Fatalf("MaxFloat64s = %v", got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestGenericNumberEmptyInputsReturnZero(t *testing.T) {
	if got := AvgNumber[int](); got != 0 {
		t.Fatalf("AvgNumber empty = %v", got)
	}
	if got := MinIntegers[int](); got != 0 {
		t.Fatalf("MinIntegers empty = %d", got)
	}
	if got := MaxIntegers[int](); got != 0 {
		t.Fatalf("MaxIntegers empty = %d", got)
	}
	if got := MinFloat64s(); got != 0 {
		t.Fatalf("MinFloat64s empty = %v", got)
	}
	if got := MaxFloat64s(); got != 0 {
		t.Fatalf("MaxFloat64s empty = %v", got)
	}
}

func TestGenericNumberAbs(t *testing.T) {
	if got := AbsInteger[int](-12); got != 12 {
		t.Fatalf("AbsInteger[int] = %d", got)
	}
	if got := AbsInteger[uint](12); got != 12 {
		t.Fatalf("AbsInteger[uint] = %d", got)
	}
	if got := AbsInteger[int8](math.MinInt8); got != 0 {
		t.Fatalf("AbsInteger overflow = %d", got)
	}
	abs, err := AbsIntegerE[int8](math.MinInt8)
	if err == nil || abs != 0 {
		t.Fatalf("AbsIntegerE overflow = %d, %v", abs, err)
	}
	if got := AbsFloat32(float32(math.Copysign(0, -1))); math.Signbit(float64(got)) || got != 0 {
		t.Fatalf("AbsFloat32(-0) = %v", got)
	}
	if got := AbsFloat32(-3.5); got != 3.5 {
		t.Fatalf("AbsFloat32 = %v", got)
	}
	if got := AbsFloat64(-4.25); got != 4.25 {
		t.Fatalf("AbsFloat64 = %v", got)
	}
	if got := AbsFloat64(math.Inf(-1)); !math.IsInf(got, 1) {
		t.Fatalf("AbsFloat64(-Inf) = %v", got)
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
