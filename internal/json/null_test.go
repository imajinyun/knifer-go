package json

import "testing"

func TestNullString(t *testing.T) {
	if got := Null.String(); got != "null" {
		t.Fatalf("Null.String() = %q", got)
	}
}

func TestIsNull(t *testing.T) {
	if !IsNull(nil) {
		t.Fatal("IsNull(nil) = false")
	}
	if !IsNull(Null) {
		t.Fatal("IsNull(Null) = false")
	}
	if IsNull("something") {
		t.Fatal("IsNull(string) = true")
	}
	if IsNull(42) {
		t.Fatal("IsNull(int) = true")
	}
}
