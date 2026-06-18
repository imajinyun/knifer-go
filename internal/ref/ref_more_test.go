package ref

import (
	"reflect"
	"testing"
)

func TestGetMethodNames(t *testing.T) {
	s := &sample{Name: "test"}
	names := GetMethodNames(s)
	if len(names) == 0 {
		t.Fatal("GetMethodNames should return at least one method")
	}
	found := false
	for _, name := range names {
		if name == "GetName" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("GetMethodNames should contain GetName, got %v", names)
	}
}

func TestInvokeMethod(t *testing.T) {
	s := &sample{Name: "alice"}
	method, ok := GetMethodByName(s, "GetName")
	if !ok {
		t.Fatal("GetMethodByName failed")
	}
	got, err := InvokeMethod(s, method)
	if err != nil {
		t.Fatal(err)
	}
	if got != "alice" {
		t.Fatalf("InvokeMethod GetName = %v, want alice", got)
	}

	// InvokeMethod with arguments
	addMethod, ok := GetMethodByName(s, "Add")
	if !ok {
		t.Fatal("GetMethodByName Add failed")
	}
	got, err = InvokeMethod(s, addMethod, 3, 4)
	if err != nil {
		t.Fatal(err)
	}
	if got != 7 {
		t.Fatalf("InvokeMethod Add = %v, want 7", got)
	}
}

func TestRemoveFinalModify(t *testing.T) {
	// no-op, should not panic
	RemoveFinalModify(nil)
	RemoveFinalModify("any value")
	RemoveFinalModify(42)
}

func TestWithAllowUnexported(t *testing.T) {
	// WithAllowUnexported is an alias for WithUnsafeAccess
	cfg := applyFieldAccessOptions([]FieldAccessOption{WithAllowUnexported(true)})
	if !cfg.unsafeAccess {
		t.Fatal("WithAllowUnexported(true) should set unsafeAccess=true")
	}

	cfg2 := applyFieldAccessOptions([]FieldAccessOption{WithAllowUnexported(false)})
	if cfg2.unsafeAccess {
		t.Fatal("WithAllowUnexported(false) should set unsafeAccess=false")
	}
}

func TestApplyFieldAccessOptionsEmpty(t *testing.T) {
	cfg := applyFieldAccessOptions(nil)
	if cfg.unsafeAccess {
		t.Fatal("default unsafeAccess should be false")
	}
}

func TestApplyFieldAccessOptionsNilOption(t *testing.T) {
	cfg := applyFieldAccessOptions([]FieldAccessOption{nil})
	if cfg.unsafeAccess {
		t.Fatal("nil option should not change config")
	}
}

func TestInvokeMethodInvalidMethod(t *testing.T) {
	_, err := InvokeMethod(nil, reflect.Method{})
	if err == nil {
		t.Fatal("InvokeMethod with invalid method should return error")
	}
}
