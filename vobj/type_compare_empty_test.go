package vobj_test

import (
	"math"
	"strings"
	"testing"

	"github.com/imajinyun/knifer-go/vobj"
)

func TestFacadeTypeNumberCompareAndEmptyHelpers(t *testing.T) {
	if !vobj.IsBasicType("go") || vobj.IsBasicType(record{}) {
		t.Fatal("IsBasicType returned unexpected result")
	}
	if !vobj.IsValidIfNumber(1.25) || vobj.IsValidIfNumber(math.NaN()) || vobj.IsValidIfNumber(math.Inf(1)) {
		t.Fatal("IsValidIfNumber returned unexpected result")
	}
	a, b := 1, 2
	if vobj.Compare(&a, &b) >= 0 || vobj.Compare[int](nil, &a) <= 0 || vobj.CompareNull[int](nil, &a, false) >= 0 {
		t.Fatal("Compare helpers returned unexpected ordering")
	}
	if typ := vobj.TypeOf(record{}); typ == nil || typ.Name() != "record" {
		t.Fatalf("TypeOf = %v", typ)
	}
	if got := vobj.TypeName(record{}); !strings.Contains(got, "record") {
		t.Fatalf("TypeName = %q", got)
	}
	if got := vobj.ToString(nil); got != "null" {
		t.Fatalf("ToString(nil) = %q", got)
	}
	if got := vobj.EmptyCount(nil, "", []int{}, "x"); got != 3 {
		t.Fatalf("EmptyCount = %d", got)
	}
	if !vobj.HasNil("x", (*int)(nil)) || !vobj.HasNull((*int)(nil)) || !vobj.HasEmpty("x", []int{}) {
		t.Fatal("HasNil/HasNull/HasEmpty returned unexpected result")
	}
	if !vobj.IsAllEmpty(nil, "", []int{}) || vobj.IsAllEmpty("", "x") {
		t.Fatal("IsAllEmpty returned unexpected result")
	}
	if !vobj.IsAllNotEmpty("x", []int{1}) || vobj.IsAllNotEmpty("x", []int{}) {
		t.Fatal("IsAllNotEmpty returned unexpected result")
	}
}
