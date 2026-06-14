package rand

import (
	"strings"
	"testing"
)

func TestRandomString(t *testing.T) {
	s := RandomString(10)
	if len(s) != 10 {
		t.Fatalf("RandomString len: %d", len(s))
	}
	for _, r := range s {
		if !strings.ContainsRune(BaseCharNumber, r) {
			t.Fatalf("RandomString out of charset: %q", s)
		}
	}
	if len(RandomNumbers(8)) != 8 {
		t.Fatalf("RandomNumbers len wrong")
	}
}
