package vconv

import "testing"

func TestConvFacade(t *testing.T) {
	if ToString(12) != "12" || ToStringDefault(nil, "x") != "x" {
		t.Fatal("string conversion failed")
	}
	if ToInt("12") != 12 || ToIntDefault("bad", 7) != 7 {
		t.Fatal("int conversion failed")
	}
	if ToInt64(true) != 1 || ToFloat64("3.5") != 3.5 {
		t.Fatal("number conversion failed")
	}
	if !ToBool("yes") || ToBoolDefault("bad", true) != true {
		t.Fatal("bool conversion failed")
	}
	if string(ToBytes("go")) != "go" {
		t.Fatal("bytes conversion failed")
	}
}

func TestConvFacadeDefaultValues(t *testing.T) {
	if got := ToInt64Default("bad", 42); got != 42 {
		t.Fatalf("ToInt64Default = %d, want 42", got)
	}
	if got := ToInt64Default("123", 42); got != 123 {
		t.Fatalf("ToInt64Default = %d, want 123", got)
	}
	if got := ToFloat64Default("bad", 3.14); got != 3.14 {
		t.Fatalf("ToFloat64Default = %v, want 3.14", got)
	}
	if got := ToFloat64Default("2.5", 1.0); got != 2.5 {
		t.Fatalf("ToFloat64Default = %v, want 2.5", got)
	}
}

func TestConvFacadeWithOptions(t *testing.T) {
	if ToStringWithOptions(true, WithFormatBoolFunc(func(bool) string { return "BOOL" })) != "BOOL" {
		t.Fatal("ToStringWithOptions bool formatter failed")
	}
	if ToStringDefaultWithOptions(nil, "fallback", WithFormatBoolFunc(func(bool) string { return "BOOL" })) != "fallback" {
		t.Fatal("ToStringDefaultWithOptions fallback failed")
	}
	parseInt := WithParseIntFunc(func(string, int, int) (int64, error) { return 41, nil })
	if ToIntWithOptions("ignored", parseInt) != 41 || ToIntDefaultWithOptions("ignored", 7, parseInt) != 41 {
		t.Fatal("int option conversion failed")
	}
	if ToInt64WithOptions("ignored", parseInt) != 41 || ToInt64DefaultWithOptions("ignored", 7, parseInt) != 41 {
		t.Fatal("int64 option conversion failed")
	}
	parseFloat := WithParseFloatFunc(func(string, int) (float64, error) { return 6.25, nil })
	if ToFloat64WithOptions("ignored", parseFloat) != 6.25 || ToFloat64DefaultWithOptions("ignored", 7, parseFloat) != 6.25 {
		t.Fatal("float option conversion failed")
	}
	parseBool := WithBoolParser(func(string) (bool, error) { return true, nil })
	if !ToBoolWithOptions("ignored", parseBool) || !ToBoolDefaultWithOptions("ignored", false, parseBool) {
		t.Fatal("bool option conversion failed")
	}
	if got := string(ToBytesWithOptions(3.5, WithFormatFloatFunc(func(float64, byte, int, int) string { return "FLOAT" }))); got != "FLOAT" {
		t.Fatalf("ToBytesWithOptions = %q", got)
	}
}
