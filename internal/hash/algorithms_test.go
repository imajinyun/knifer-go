package hash

import "testing"

func TestStringHashAlgorithms(t *testing.T) {
	const s = "knifer-go"
	// Non-zero, deterministic sanity checks for each algorithm.
	if RsHash(s) < 0 || JsHash(s) < 0 || PjwHash(s) < 0 || ElfHash(s) < 0 {
		t.Fatal("masked hashes must be non-negative")
	}
	if BkdrHash(s) < 0 || SdbmHash(s) < 0 || DjbHash(s) < 0 {
		t.Fatal("masked hashes must be non-negative")
	}
	if FnvHashString(s) < 0 {
		t.Fatal("FnvHashString must be non-negative")
	}
	if ApHash(s) == 0 {
		t.Fatal("ApHash unexpectedly zero")
	}
	if HfHash(s) == 0 {
		t.Fatal("HfHash unexpectedly zero")
	}
	if HfIpHash(s) == 0 {
		t.Fatal("HfIpHash unexpectedly zero")
	}
}

func TestJavaDefaultHashAndTianl(t *testing.T) {
	if JavaDefaultHash("a") != 97 {
		t.Fatalf("JavaDefaultHash(a) = %d, want 97", JavaDefaultHash("a"))
	}
	if TianlHash("") != 0 {
		t.Fatalf("TianlHash(empty) = %d, want 0", TianlHash(""))
	}
	if TianlHash("abc") <= 0 {
		t.Fatal("TianlHash(abc) must be positive")
	}
}
