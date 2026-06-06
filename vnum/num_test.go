package vnum

import (
	"reflect"
	"testing"
)

type sequenceReader struct {
	next byte
}

func (r *sequenceReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = r.next
		r.next++
	}
	return len(p), nil
}

func TestNumFacade(t *testing.T) {
	if Add(1.2, 2.3) != 3.5 || Sub(5, 2) != 3 || Mul(2, 3) != 6 || Div(10, 4, 1) != 2.5 {
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

func TestNumRandomOptionsFacade(t *testing.T) {
	seed := []int{10, 20, 30, 40}
	got := GenRandomNumberWithSeedWithOptions(0, 4, 3, seed, WithRandomReader(&sequenceReader{}))
	if !reflect.DeepEqual(got, []int{10, 20, 40}) {
		t.Fatalf("GenRandomNumberWithSeedWithOptions = %v", got)
	}

	got = GenRandomNumberWithOptions(0, 5, 3, WithRandomReader(&sequenceReader{}))
	if !reflect.DeepEqual(got, []int{0, 1, 2}) {
		t.Fatalf("GenRandomNumberWithOptions = %v", got)
	}

	got = GenBySetWithOptions(0, 5, 3, WithRandomReader(&sequenceReader{}))
	if len(got) != 3 {
		t.Fatalf("GenBySetWithOptions = %v", got)
	}
	seen := map[int]bool{}
	for _, v := range got {
		seen[v] = true
	}
	for _, want := range []int{0, 1, 2} {
		if !seen[want] {
			t.Fatalf("GenBySetWithOptions missing %d in %v", want, got)
		}
	}
}
