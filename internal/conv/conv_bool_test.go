package conv

import (
	"errors"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

type namedBool bool

func TestToBool(t *testing.T) {
	cases := map[string]bool{
		"true": true, "yes": true, "y": true, "ok": true, "1": true, "on": true,
		"false": false, "no": false, "n": false, "0": false, "off": false,
	}
	for s, want := range cases {
		if ToBool(s) != want {
			t.Fatalf("ToBool(%q)", s)
		}
	}
	if ToBool(1) != true || ToBool(0) != false {
		t.Fatalf("ToBool int")
	}
	if ToBoolDefault("xx", true) != true {
		t.Fatalf("ToBool default")
	}
}

func TestToBoolE(t *testing.T) {
	tests := []struct {
		name        string
		input       any
		expected    bool
		expectedErr bool
	}{
		{name: "string token", input: "yes", expected: true},
		{name: "named string token", input: namedNumericString("yes"), expected: true},
		{name: "int value", input: 0, expected: false},
		{name: "bool value", input: true, expected: true},
		{name: "named bool value", input: namedBool(true), expected: true},
		{name: "nil input", input: nil, expectedErr: true},
		{name: "invalid string", input: "disabled", expectedErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToBoolE(tt.input)
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
				t.Fatalf("got = %v, want %v", got, tt.expected)
			}
		})
	}
}
