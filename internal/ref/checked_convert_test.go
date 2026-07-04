package ref

import (
	"math"
	"reflect"
	"strings"
	"testing"
)

func TestCheckedConvertNumericBoundaries(t *testing.T) {
	tests := []struct {
		name    string
		src     any
		dst     any
		want    any
		wantErr string
	}{
		{name: "int16 fits int8", src: int16(127), dst: int8(0), want: int8(127)},
		{name: "int16 overflows int8", src: int16(128), dst: int8(0), wantErr: "integer overflow"},
		{name: "negative to uint8", src: int8(-1), dst: uint8(0), wantErr: "negative value"},
		{name: "uint16 overflows uint8", src: uint16(256), dst: uint8(0), wantErr: "unsigned integer overflow"},
		{name: "float64 overflows float32", src: math.MaxFloat64, dst: float32(0), wantErr: "float overflow"},
		{name: "float64 nan converts to float32", src: math.NaN(), dst: float32(0), want: float32(math.NaN())},
		{name: "uint64 overflows int64", src: uint64(math.MaxInt64) + 1, dst: int64(0), wantErr: "integer overflow"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckedConvert(reflect.ValueOf(tt.src), reflect.TypeOf(tt.dst))
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("CheckedConvert() error = %v, want substring %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("CheckedConvert() error = %v", err)
			}
			if f, ok := tt.want.(float32); ok && math.IsNaN(float64(f)) {
				if got.Kind() != reflect.Float32 || !math.IsNaN(float64(got.Interface().(float32))) {
					t.Fatalf("CheckedConvert() = %#v, want float32 NaN", got.Interface())
				}
				return
			}
			if got.Interface() != tt.want {
				t.Fatalf("CheckedConvert() = %#v, want %#v", got.Interface(), tt.want)
			}
		})
	}
}

func TestCheckedConvertRejectsInvalidInput(t *testing.T) {
	if _, err := CheckedConvert(reflect.Value{}, reflect.TypeOf(int(0))); err == nil || !strings.Contains(err.Error(), "invalid value") {
		t.Fatalf("CheckedConvert(invalid) error = %v, want invalid value", err)
	}
	if _, err := CheckedConvert(reflect.ValueOf(1), nil); err == nil || !strings.Contains(err.Error(), "nil target type") {
		t.Fatalf("CheckedConvert(nil target) error = %v, want nil target type", err)
	}
	if _, err := CheckedConvert(reflect.ValueOf("x"), reflect.TypeOf(0)); err == nil || !strings.Contains(err.Error(), "cannot convert") {
		t.Fatalf("CheckedConvert(incompatible) error = %v, want cannot convert", err)
	}
}
