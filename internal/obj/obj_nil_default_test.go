package obj

import "testing"

func TestDefaultsApplyAcceptAndAggregates(t *testing.T) {
	value := "go"
	if DefaultIfNil(&value, "x") != "go" || DefaultIfNil[string](nil, "x") != "x" {
		t.Fatal("DefaultIfNil failed")
	}
	if got := Apply(&value, func(s string) int { return len(s) }); got != 2 {
		t.Fatalf("Apply: %d", got)
	}
	called := false
	Accept(&value, func(string) { called = true })
	if !called {
		t.Fatal("Accept not called")
	}
	if EmptyCount(nil, "", []int{}, 1) != 3 || !HasNil(1, nil) || !HasEmpty(1, "") {
		t.Fatal("aggregate checks failed")
	}
	if !IsAllEmpty(nil, "") || !IsAllNotEmpty(1, "x") {
		t.Fatal("all checks failed")
	}
}
