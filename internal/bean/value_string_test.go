package bean

import (
	"reflect"
	"testing"
)

type valueStringSample struct {
	Name string
}

func TestValueString(t *testing.T) {
	// valid string value
	v := reflect.ValueOf("hello")
	if got := valueString(v); got != "hello" {
		t.Fatalf("valueString(string) = %q, want %q", got, "hello")
	}

	// int value (CanInterface)
	v = reflect.ValueOf(42)
	if got := valueString(v); got != "42" {
		t.Fatalf("valueString(int) = %q, want %q", got, "42")
	}

	// invalid value
	v = reflect.Value{}
	if got := valueString(v); got != "" {
		t.Fatalf("valueString(invalid) = %q, want empty", got)
	}

	// pointer to string
	s := "pointer"
	v = reflect.ValueOf(&s)
	if got := valueString(v); got != "pointer" {
		t.Fatalf("valueString(pointer) = %q, want %q", got, "pointer")
	}

	// struct pointer (CanInterface true)
	v = reflect.ValueOf(&valueStringSample{Name: "test"})
	if got := valueString(v); got == "" {
		t.Fatal("valueString(struct) should not be empty")
	}
}
