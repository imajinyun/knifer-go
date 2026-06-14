package rand

import "testing"

func TestRandomEle(t *testing.T) {
	a := []string{"x", "y", "z"}
	got := RandomEle(a)
	found := false
	for _, v := range a {
		if got == v {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("RandomEle returned non-existing: %q", got)
	}
}
