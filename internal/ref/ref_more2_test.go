package ref

import (
	"testing"
)

func TestSetAccessible(t *testing.T) {
	if got := SetAccessible("hello"); got != "hello" {
		t.Fatalf("SetAccessible = %q, want %q", got, "hello")
	}
	if got := SetAccessible(42); got != 42 {
		t.Fatalf("SetAccessible = %d, want %d", got, 42)
	}
}
