package conv

import "testing"

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
