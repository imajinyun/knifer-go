package num

import "testing"

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
