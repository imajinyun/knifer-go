package hash

import (
	"hash/fnv"
	"testing"
)

func TestHashFunctions(t *testing.T) {
	if FnvHash("abc") == 0 {
		t.Fatalf("FnvHash zero")
	}
	if AdditiveHash("abc", 31) < 0 {
		t.Fatalf("AdditiveHash failed")
	}
	if got, want := Hash32("abc", fnv.New32a), fnv32a("abc"); got != want {
		t.Fatalf("Hash32 = %d, want %d", got, want)
	}
}

func fnv32a(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}
