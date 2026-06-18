package vhash

import (
	"hash/fnv"
	"testing"
)

func TestHashFacade(t *testing.T) {
	if FnvHash("abc") == 0 || AdditiveHash("abc", 31) < 0 {
		t.Fatal("hash helpers failed")
	}
}

func TestHash32Facade(t *testing.T) {
	if h := Hash32("abc", nil); h == 0 {
		t.Fatal("Hash32 with nil newHash returned 0")
	}
	if h := Hash32("abc", fnv.New32a); h == 0 {
		t.Fatal("Hash32 with fnv.New32a returned 0")
	}
}

func TestStringHashFacade(t *testing.T) {
	const s = "go-knifer"
	if RsHash(s) < 0 || JsHash(s) < 0 || PjwHash(s) < 0 || ElfHash(s) < 0 {
		t.Fatal("masked hashes must be non-negative")
	}
	if BkdrHash(s) < 0 || SdbmHash(s) < 0 || DjbHash(s) < 0 || FnvHashString(s) < 0 {
		t.Fatal("masked hashes must be non-negative")
	}
	if HfIpHash(s) == 0 {
		t.Fatal("HfIpHash unexpectedly zero")
	}
	_ = ApHash(s)
	_ = HfHash(s)
	if JavaDefaultHash("a") != 97 {
		t.Fatalf("JavaDefaultHash(a) = %d, want 97", JavaDefaultHash("a"))
	}
	if TianlHash("") != 0 {
		t.Fatalf("TianlHash(empty) = %d, want 0", TianlHash(""))
	}
}
