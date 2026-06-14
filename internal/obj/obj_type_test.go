package obj

import (
	"math"
	"testing"
)

func TestCompareTypeAndString(t *testing.T) {
	src := sample{Name: "n", Tags: []string{"a"}}
	a, b := 1, 2
	if Compare(&a, &b) >= 0 || CompareNull[int](nil, &b, true) <= 0 {
		t.Fatal("compare failed")
	}
	if TypeName(src) == "" || ToString(nil) != "null" {
		t.Fatal("type or string failed")
	}
}

func TestBasicAndValidNumber(t *testing.T) {
	if !IsBasicType("x") || IsBasicType(sample{}) {
		t.Fatal("basic type check failed")
	}
	if !IsValidIfNumber(1) || IsValidIfNumber(math.NaN()) || IsValidIfNumber(math.Inf(1)) {
		t.Fatal("valid number check failed")
	}
}
