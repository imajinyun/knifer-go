package rand

import "testing"

func TestRandomIntRange(t *testing.T) {
	for i := 0; i < 100; i++ {
		n := RandomIntRange(10, 20)
		if n < 10 || n >= 20 {
			t.Fatalf("RandomIntRange out of bounds: %d", n)
		}
	}
}
