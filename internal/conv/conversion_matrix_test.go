package conv

import (
	"encoding/json"
	"errors"
	"math"
	"testing"
	"time"

	knifer "github.com/imajinyun/knifer-go"
)

type matrixNamedString string

type matrixNamedBool bool

type matrixNamedInt int

func TestConversionMatrixPropertyContract(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		wantStr   string
		wantBytes string
		wantBool  bool
		boolOK    bool
		wantInt64 int64
		intOK     bool
		wantFloat float64
		floatOK   bool
	}{
		{name: "nil", value: nil, wantStr: "", wantBytes: "", wantBool: false, boolOK: false, wantInt64: 0, intOK: false, wantFloat: 0, floatOK: false},
		{name: "string integer", value: "42", wantStr: "42", wantBytes: "42", wantBool: false, boolOK: false, wantInt64: 42, intOK: true, wantFloat: 42, floatOK: true},
		{name: "string float", value: "42.9", wantStr: "42.9", wantBytes: "42.9", wantBool: false, boolOK: false, wantInt64: 42, intOK: true, wantFloat: 42.9, floatOK: true},
		{name: "bytes", value: []byte("7"), wantStr: "7", wantBytes: "7", wantBool: false, boolOK: false, wantInt64: 0, intOK: false, wantFloat: 0, floatOK: false},
		{name: "bool true", value: true, wantStr: "true", wantBytes: "true", wantBool: true, boolOK: true, wantInt64: 1, intOK: true, wantFloat: 1, floatOK: true},
		{name: "bool false", value: false, wantStr: "false", wantBytes: "false", wantBool: false, boolOK: true, wantInt64: 0, intOK: true, wantFloat: 0, floatOK: true},
		{name: "int", value: -12, wantStr: "-12", wantBytes: "-12", wantBool: true, boolOK: true, wantInt64: -12, intOK: true, wantFloat: -12, floatOK: true},
		{name: "uint", value: uint(12), wantStr: "12", wantBytes: "12", wantBool: true, boolOK: true, wantInt64: 12, intOK: true, wantFloat: 12, floatOK: true},
		{name: "float", value: 3.5, wantStr: "3.5", wantBytes: "3.5", wantBool: true, boolOK: true, wantInt64: 3, intOK: true, wantFloat: 3.5, floatOK: true},
		{name: "named string", value: matrixNamedString("yes"), wantStr: "yes", wantBytes: "yes", wantBool: true, boolOK: true, wantInt64: 0, intOK: false, wantFloat: 0, floatOK: false},
		{name: "named bool", value: matrixNamedBool(true), wantStr: "true", wantBytes: "true", wantBool: true, boolOK: true, wantInt64: 1, intOK: true, wantFloat: 1, floatOK: true},
		{name: "named int", value: matrixNamedInt(9), wantStr: "9", wantBytes: "9", wantBool: true, boolOK: true, wantInt64: 9, intOK: true, wantFloat: 9, floatOK: true},
		{name: "json number", value: json.Number("17"), wantStr: "17", wantBytes: "17", wantBool: false, boolOK: false, wantInt64: 17, intOK: true, wantFloat: 17, floatOK: true},
		{name: "duration", value: 150 * time.Millisecond, wantStr: "150ms", wantBytes: "150ms", wantBool: true, boolOK: true, wantInt64: int64(150 * time.Millisecond), intOK: true, wantFloat: float64(150 * time.Millisecond), floatOK: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToString(tt.value); got != tt.wantStr {
				t.Fatalf("ToString(%#v) = %q, want %q", tt.value, got, tt.wantStr)
			}
			if got := string(ToBytes(tt.value)); got != tt.wantBytes {
				t.Fatalf("ToBytes(%#v) = %q, want %q", tt.value, got, tt.wantBytes)
			}
			gotBool, err := ToBoolE(tt.value)
			assertConversionMatrixResult(t, "ToBoolE", gotBool, err, tt.wantBool, tt.boolOK)
			gotInt, err := ToInt64E(tt.value)
			assertConversionMatrixResult(t, "ToInt64E", gotInt, err, tt.wantInt64, tt.intOK)
			gotFloat, err := ToFloat64E(tt.value)
			assertConversionMatrixResult(t, "ToFloat64E", gotFloat, err, tt.wantFloat, tt.floatOK)
		})
	}
}

func TestConversionMatrixFailureContract(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{name: "bad int", err: errorOnly(ToInt64E("not-int"))},
		{name: "bad bool", err: errorOnly(ToBoolE("maybe"))},
		{name: "bad float", err: errorOnly(ToFloat64E("not-float"))},
		{name: "uint overflow", err: errorOnly(ToInt64E(uint64(math.MaxInt64) + 1))},
		{name: "nan to int", err: errorOnly(ToInt64E(math.NaN()))},
		{name: "inf to int", err: errorOnly(ToInt64E(math.Inf(1)))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !errors.Is(tt.err, ErrInvalidConversion) {
				t.Fatalf("error = %v, want ErrInvalidConversion", tt.err)
			}
			if !errors.Is(tt.err, knifer.ErrCodeInvalidInput) {
				t.Fatalf("error = %v, want ErrCodeInvalidInput", tt.err)
			}
		})
	}
}

func FuzzConversionMatrixStringScalars(f *testing.F) {
	for _, seed := range []string{"", "0", "1", "true", "false", "42", "3.14", "not-number"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input string) {
		if got := ToString([]byte(input)); got != input {
			t.Fatalf("ToString([]byte(%q)) = %q", input, got)
		}
		if got := string(ToBytes(input)); got != input {
			t.Fatalf("ToBytes(%q) = %q", input, got)
		}
		if i, err := ToInt64E(input); err == nil {
			if got := ToInt64Default(input, -1); got != i {
				t.Fatalf("ToInt64Default(%q) = %d, want %d", input, got, i)
			}
		}
		if b, err := ToBoolE(input); err == nil {
			if got := ToBoolDefault(input, !b); got != b {
				t.Fatalf("ToBoolDefault(%q) = %v, want %v", input, got, b)
			}
		}
	})
}

func assertConversionMatrixResult[T comparable](t *testing.T, name string, got T, err error, want T, ok bool) {
	t.Helper()
	if !ok {
		if !errors.Is(err, ErrInvalidConversion) {
			t.Fatalf("%s error = %v, want ErrInvalidConversion", name, err)
		}
		return
	}
	if err != nil {
		t.Fatalf("%s error = %v", name, err)
	}
	if got != want {
		t.Fatalf("%s = %v, want %v", name, got, want)
	}
}

func errorOnly[T any](_ T, err error) error { return err }
