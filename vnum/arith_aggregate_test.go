package vnum

import "testing"

func TestNumFacade(t *testing.T) {
	if Add(1.2, 2.3) != 3.5 || Sub(5, 2) != 3 || Mul(2, 3) != 6 || Div(10, 4, 1) != 2.5 || Div(10, 4) != 2.5 {
		t.Fatal("arithmetic helpers failed")
	}
	if Round(1.234, 2) != 1.23 || !IsNumber("3.14") || !IsInteger("42") || !IsDigits("123") {
		t.Fatal("format/check helpers failed")
	}
	if Min(3, 1, 2) != 1 || Max(3, 1, 2) != 3 || Sum(1, 2, 3) != 6 || Avg(2, 4) != 3 {
		t.Fatal("aggregate helpers failed")
	}
	seq := Range(1, 5, 2)
	if len(seq) != 2 || seq[0] != 1 || seq[1] != 3 {
		t.Fatalf("Range failed: %v", seq)
	}
	if !Equals(0.1+0.2, 0.3) || DecimalFormat("0.00", 1.2) != "1.20" {
		t.Fatal("equals/format helpers failed")
	}
}
