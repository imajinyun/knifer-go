package vobj_test

import (
	"strings"
	"testing"
	"time"

	"github.com/imajinyun/knifer-go/vobj"
)

func TestFacadeNilDefaultAndCollectionHelpers(t *testing.T) {
	var values []int
	if !vobj.IsNil(values) || !vobj.IsNull(values) || vobj.IsNotNil(values) || vobj.IsNotNull(values) {
		t.Fatal("typed nil checks returned unexpected result")
	}
	if got := vobj.Length(42); got != -1 {
		t.Fatalf("Length unsupported = %d", got)
	}
	if !vobj.Contains("knifer-go", "knife") || vobj.Contains(map[string]int{"a": 1}, 2) {
		t.Fatal("Contains returned unexpected result")
	}
	if !vobj.Equals(1, uint(1)) || !vobj.NotEqual("a", "b") {
		t.Fatal("Equals/NotEqual returned unexpected result")
	}
	utc := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	sameInstant := time.Date(2024, 1, 1, 8, 0, 0, 0, time.FixedZone("UTC+8", 8*60*60))
	if !vobj.Equals(utc, sameInstant) || vobj.Equals(utc, "2024-01-01T00:00:00Z") {
		t.Fatal("time Equals returned unexpected result")
	}

	name := "go"
	supplierCalls := 0
	if got := vobj.DefaultIfNilFunc(&name, func() string {
		supplierCalls++
		return "fallback"
	}); got != "go" || supplierCalls != 0 {
		t.Fatalf("DefaultIfNilFunc existing = %q calls=%d", got, supplierCalls)
	}
	if got := vobj.DefaultIfNilFunc[string](nil, func() string {
		supplierCalls++
		return "fallback"
	}); got != "fallback" || supplierCalls != 1 {
		t.Fatalf("DefaultIfNilFunc nil = %q calls=%d", got, supplierCalls)
	}
	if got := vobj.DefaultIfNilApply(&name, strings.ToUpper, "fallback"); got != "GO" {
		t.Fatalf("DefaultIfNilApply existing = %q", got)
	}
	if got := vobj.DefaultIfNilApply[string, string](nil, strings.ToUpper, "fallback"); got != "fallback" {
		t.Fatalf("DefaultIfNilApply nil = %q", got)
	}
	accepted := ""
	vobj.Accept(&name, func(s string) { accepted = s })
	vobj.Accept[string](nil, func(s string) { t.Fatalf("Accept nil called with %q", s) })
	if accepted != "go" {
		t.Fatalf("Accept captured = %q", accepted)
	}
}
