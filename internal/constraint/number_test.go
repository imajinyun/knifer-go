package constraint

import "testing"

type (
	namedInt       int
	namedUint      uint
	namedFloat     float64
	namedComplex   complex128
	namedStringKey string
)

func addInteger[T Integer](a, b T) T { return a + b }

func negateSigned[T Signed](v T) T { return -v }

func addUnsigned[T Unsigned](a, b T) T { return a + b }

func addFloat[T Float](a, b T) T { return a + b }

func addComplex[T Complex](a, b T) T { return a + b }

func addNumber[T Number](a, b T) T { return a + b }

func TestNumericConstraintsAcceptBuiltInAndNamedTypes(t *testing.T) {
	if got := addInteger(namedInt(1), namedInt(2)); got != 3 {
		t.Fatalf("addInteger named signed = %d", got)
	}
	if got := addInteger(namedUint(1), namedUint(2)); got != 3 {
		t.Fatalf("addInteger named unsigned = %d", got)
	}
	if got := negateSigned(namedInt(3)); got != -3 {
		t.Fatalf("negateSigned = %d", got)
	}
	if got := addUnsigned(uintptr(1), uintptr(2)); got != 3 {
		t.Fatalf("addUnsigned uintptr = %d", got)
	}
	if got := addFloat(namedFloat(1.25), namedFloat(2.5)); got != 3.75 {
		t.Fatalf("addFloat = %f", got)
	}
	if got := addComplex(namedComplex(1+2i), namedComplex(3+4i)); got != 4+6i {
		t.Fatalf("addComplex = %v", got)
	}
	if got := addNumber(namedInt(4), namedInt(5)); got != 9 {
		t.Fatalf("addNumber integer = %d", got)
	}
	if got := addNumber(namedFloat(1.25), namedFloat(2.5)); got != 3.75 {
		t.Fatalf("addNumber float = %f", got)
	}
}

func TestUseStandardComparableForStringKeys(t *testing.T) {
	values := map[namedStringKey]int{"one": 1}
	if values["one"] != 1 {
		t.Fatalf("named string key lookup = %d", values["one"])
	}
}
