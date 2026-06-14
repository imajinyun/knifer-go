package obj

import (
	"testing"
	"time"
)

func TestEqualLengthContainsAndEmpty(t *testing.T) {
	if !Equal(1, int64(1)) || NotEqual("a", "a") {
		t.Fatal("numeric or string equality failed")
	}
	utc := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	sameInstant := time.Date(2024, 1, 1, 8, 0, 0, 0, time.FixedZone("UTC+8", 8*60*60))
	if !Equals(utc, sameInstant) {
		t.Fatal("time equality should compare instants")
	}
	if Equals(utc, "2024-01-01T00:00:00Z") {
		t.Fatal("time equality should reject non-time values")
	}
	if Length([]int{1, 2, 3}) != 3 || Length(10) != -1 {
		t.Fatal("length failed")
	}
	if !Contains([]int{1, 2, 3}, int64(2)) || !Contains("hello", "ell") {
		t.Fatal("contains failed")
	}
	if !IsEmpty(map[string]int{}) || IsEmpty(1) || !IsNotEmpty([]int{1}) {
		t.Fatal("empty checks failed")
	}
}
