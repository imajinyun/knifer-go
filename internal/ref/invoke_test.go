package ref

import (
	"reflect"
	"testing"
)

func TestInvokeHelpers(t *testing.T) {
	s := &sample{Name: "alice"}
	got, err := Invoke(s, "Add", int8(1), int8(2))
	if err != nil || got != 3 {
		t.Fatalf("Invoke Add = %v, %v", got, err)
	}
	if _, err := Invoke(s, "Missing"); err == nil {
		t.Fatal("Invoke missing should fail")
	}
	if got, err := InvokeStatic(func(a int, b int) int { return a * b }, 2, 3); err != nil || got != 6 {
		t.Fatalf("InvokeStatic = %v, %v", got, err)
	}
	if got, err := InvokeFunc(func() (int, string) { return 1, "a" }); err != nil || !reflect.DeepEqual(got, []any{1, "a"}) {
		t.Fatalf("InvokeFunc multi return = %#v, %v", got, err)
	}
	if _, err := InvokeRaw(123); err == nil {
		t.Fatal("InvokeRaw non-func should fail")
	}
}

func TestInvokeFuncRejectsUnsafeNumericConversion(t *testing.T) {
	got, err := InvokeFunc(func(v int8) int8 { return v }, int16(128))
	if err != nil {
		t.Fatalf("InvokeFunc overflow argument error = %v", err)
	}
	if got != int8(0) {
		t.Fatalf("InvokeFunc overflow argument = %v, want zero fallback", got)
	}

	got, err = InvokeFunc(func(v int8) int8 { return v }, int16(127))
	if err != nil {
		t.Fatalf("InvokeFunc safe argument error = %v", err)
	}
	if got != int8(127) {
		t.Fatalf("InvokeFunc safe argument = %v, want 127", got)
	}
}
