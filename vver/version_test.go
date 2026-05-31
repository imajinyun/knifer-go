package vver_test

import (
	"testing"

	"github.com/imajinyun/go-knifer/vver"
)

func TestFacadeVersionHelpers(t *testing.T) {
	if vver.CompareVersion("1.0.0", "1.0.2") >= 0 {
		t.Fatal("CompareVersion failed")
	}
	if !vver.IsGreaterThan("1.0.3", "1.0.2") {
		t.Fatal("IsGreaterThan failed")
	}
	if !vver.IsGreaterThanOrEqual("1.0.2", "1.0.2") {
		t.Fatal("IsGreaterThanOrEqual failed")
	}
	if !vver.IsLessThan("1.0.1", "1.0.2") {
		t.Fatal("IsLessThan failed")
	}
	if !vver.IsLessThanOrEqual("1.0.2", "1.0.2") {
		t.Fatal("IsLessThanOrEqual failed")
	}
	if !vver.MatchEl("1.0.2", ">=1.0.2") {
		t.Fatal("MatchEl failed")
	}
	if !vver.MatchElWithDelimiter("1.0.2", "<1.0.1,1.0.2", ",") {
		t.Fatal("MatchElWithDelimiter failed")
	}
	if !vver.MatchElByDelimiter("1.0.2", "1.0.1,1.0.2-1.1.1", ",") {
		t.Fatal("MatchElByDelimiter failed")
	}
	if err := vver.MatchElWithDelimiterErr("1.0.2", ">=1.0.0", "-"); err == nil {
		t.Fatal("expected invalid delimiter error")
	}
	if !vver.AnyMatch("1.0.2", "<1.0.1", "1.0.2") {
		t.Fatal("AnyMatch failed")
	}
	if !vver.AnyMatchSlice("1.0.2", []string{"<1.0.1", "1.0.2"}) {
		t.Fatal("AnyMatchSlice failed")
	}
}
