package conv

import (
	"errors"
	"math"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

type namedNumericString string

type namedInt64 int64

type namedFloat64 float64

func TestToInt(t *testing.T) {
	if ToInt("123") != 123 {
		t.Fatalf("string int")
	}
	if ToInt("3.14") != 3 {
		t.Fatalf("string float")
	}
	if ToInt(int64(99)) != 99 {
		t.Fatalf("int64")
	}
	if ToInt(true) != 1 {
		t.Fatalf("bool true")
	}
	if ToIntDefault("abc", 42) != 42 {
		t.Fatalf("default")
	}
}

func TestToInt64AndFloat(t *testing.T) {
	if ToInt64("9999999999") != 9999999999 {
		t.Fatalf("ToInt64")
	}
	if ToFloat64("3.14") != 3.14 {
		t.Fatalf("ToFloat64")
	}
	if ToFloat64Default("x", 1.5) != 1.5 {
		t.Fatalf("ToFloat64 default")
	}
	if ToInt64Default("abc", 42) != 42 {
		t.Fatalf("ToInt64Default")
	}
	if ToInt64Default("123", 0) != 123 {
		t.Fatalf("ToInt64Default valid")
	}
}

func TestErrorReturningNumberConversions(t *testing.T) {
	tests := []struct {
		name        string
		convert     func() (any, error)
		expected    any
		expectedErr bool
	}{
		{
			name:     "int from string",
			convert:  func() (any, error) { return ToIntE("42") },
			expected: 42,
		},
		{
			name:     "int64 from float string truncates",
			convert:  func() (any, error) { return ToInt64E("42.9") },
			expected: int64(42),
		},
		{
			name:     "float64 from string",
			convert:  func() (any, error) { return ToFloat64E("3.5") },
			expected: 3.5,
		},
		{
			name:     "int64 from named string",
			convert:  func() (any, error) { return ToInt64E(namedNumericString("77")) },
			expected: int64(77),
		},
		{
			name:     "int64 from named int64",
			convert:  func() (any, error) { return ToInt64E(namedInt64(88)) },
			expected: int64(88),
		},
		{
			name:     "float64 from named float64",
			convert:  func() (any, error) { return ToFloat64E(namedFloat64(1.25)) },
			expected: 1.25,
		},
		{
			name:        "int invalid string",
			convert:     func() (any, error) { return ToIntE("bad") },
			expectedErr: true,
		},
		{
			name:        "int64 nil",
			convert:     func() (any, error) { return ToInt64E(nil) },
			expectedErr: true,
		},
		{
			name:        "float64 invalid string",
			convert:     func() (any, error) { return ToFloat64E("bad") },
			expectedErr: true,
		},
		{
			name:        "uint64 overflows int64",
			convert:     func() (any, error) { return ToInt64E(uint64(math.MaxInt64) + 1) },
			expectedErr: true,
		},
		{
			name:        "uint64 overflows host int",
			convert:     func() (any, error) { return ToIntE(uint64(math.MaxInt64) + 1) },
			expectedErr: true,
		},
		{
			name:        "float64 overflows int64",
			convert:     func() (any, error) { return ToInt64E(float64(math.MaxInt64)) },
			expectedErr: true,
		},
		{
			name:        "float64 nan cannot narrow to int64",
			convert:     func() (any, error) { return ToInt64E(math.NaN()) },
			expectedErr: true,
		},
		{
			name:        "positive infinity cannot narrow to int64",
			convert:     func() (any, error) { return ToInt64E(math.Inf(1)) },
			expectedErr: true,
		},
		{
			name:        "negative infinity cannot narrow to int64",
			convert:     func() (any, error) { return ToInt64E(math.Inf(-1)) },
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.convert()
			if tt.expectedErr {
				if !errors.Is(err, ErrInvalidConversion) {
					t.Fatalf("error = %v, want ErrInvalidConversion", err)
				}
				if !errors.Is(err, knifer.ErrCodeInvalidInput) {
					t.Fatalf("error = %v, want invalid input code", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error = %v", err)
			}
			if got != tt.expected {
				t.Fatalf("got = %#v, want %#v", got, tt.expected)
			}
		})
	}
}

func TestZeroAndDefaultConversionsKeepLegacyOverflowBehavior(t *testing.T) {
	input := uint64(math.MaxInt64) + 1
	if got := ToInt64(input); got != math.MinInt64 {
		t.Fatalf("ToInt64 overflow legacy value = %d, want %d", got, int64(math.MinInt64))
	}
	if got := ToInt64Default(input, 7); got != math.MinInt64 {
		t.Fatalf("ToInt64Default overflow legacy value = %d, want %d", got, int64(math.MinInt64))
	}
}

func TestToFloat64UsesUnsignedReflectValueWithoutIntWrapping(t *testing.T) {
	type namedUint64 uint64

	input := uint64(math.MaxInt64) + 1
	if got := ToFloat64(input); got != float64(input) {
		t.Fatalf("ToFloat64(uint64 overflow boundary) = %v, want %v", got, float64(input))
	}
	if got := ToFloat64(namedUint64(input)); got != float64(input) {
		t.Fatalf("ToFloat64(named uint64 overflow boundary) = %v, want %v", got, float64(input))
	}
}
