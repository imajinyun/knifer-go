package bloomfilter

import "testing"

func TestBitSetBloomFilter_AddAndContains(t *testing.T) {
	bf := NewBitSetBloomFilter(10000, 100, 4)
	if !bf.Add("hello") {
		t.Fatal("first add should return true")
	}
	if bf.Add("hello") {
		t.Fatal("repeat add should return false")
	}
	if !bf.Contains("hello") {
		t.Fatal("should contain hello")
	}
	if bf.Contains("absent-token-xyz") {
		t.Fatal("should not contain absent token")
	}
	if p := bf.FalsePositiveProbability(); p < 0 || p > 1 {
		t.Fatalf("invalid probability: %v", p)
	}
}

func TestBitSetBloomFilter_PanicOnInvalidParams(t *testing.T) {
	cases := []func(){
		func() { NewBitSetBloomFilter(0, 1, 1) },
		func() { NewBitSetBloomFilter(1, 0, 1) },
		func() { NewBitSetBloomFilter(1, 1, 0) },
		func() { NewBitSetBloomFilter(1, 1, 9) },
	}
	for i, fn := range cases {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("case %d should panic", i)
				}
			}()
			fn()
		}()
	}
}

func TestBitMapBloomFilter(t *testing.T) {
	bf := NewBitMapBloomFilter(5)
	if !bf.Add("foo") {
		t.Fatal("add foo should return true")
	}
	if !bf.Contains("foo") {
		t.Fatal("should contain foo")
	}
	if bf.Add("foo") {
		t.Fatal("repeat add foo should return false")
	}
	if bf.Contains("not-in-filter-12345") {
		t.Fatal("should not contain unknown token")
	}
}

func TestBitMapBloomFilter_CustomFilters(t *testing.T) {
	bf := NewBitMapBloomFilterWithFilters(
		5,
		NewFNVFilter(1<<20),
		NewRSFilter(1<<20),
	)
	if !bf.Add("bar") {
		t.Fatal()
	}
	if !bf.Contains("bar") {
		t.Fatal()
	}
}

func TestFuncFilter_MachineNum(t *testing.T) {
	f := NewFuncFilterWithMachineNum(1024, Machine64,
		func(s string) int64 { return int64(JavaDefaultHash(s)) })
	if !f.Add("x") {
		t.Fatal()
	}
	if !f.Contains("x") {
		t.Fatal()
	}
}

func TestFuncFilter_PanicOnUnknownMachineNum(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("should panic for unknown machine number")
		}
	}()
	NewFuncFilterWithMachineNum(1024, 16,
		func(s string) int64 { return 0 })
}

func TestIntMapAndLongMap(t *testing.T) {
	im := NewIntMap(8)
	im.Add(0)
	im.Add(31)
	im.Add(64)
	if !im.Contains(0) || !im.Contains(31) || !im.Contains(64) {
		t.Fatal("intmap contains failed")
	}
	if im.Contains(1) {
		t.Fatal("intmap should not contain 1")
	}
	im.Remove(31)
	if im.Contains(31) {
		t.Fatal("intmap remove failed")
	}

	lm := NewLongMap(4)
	lm.Add(0)
	lm.Add(63)
	lm.Add(128)
	if !lm.Contains(0) || !lm.Contains(63) || !lm.Contains(128) {
		t.Fatal("longmap contains failed")
	}
	lm.Remove(128)
	if lm.Contains(128) {
		t.Fatal("longmap remove failed")
	}
}

func TestHashAlgorithms(t *testing.T) {
	s := "hutool-bloomFilter"
	checks := map[string]int32{
		"rs":   RsHash(s),
		"js":   JsHash(s),
		"pjw":  PjwHash(s),
		"elf":  ElfHash(s),
		"bkdr": BkdrHash(s),
		"sdbm": SdbmHash(s),
		"djb":  DjbHash(s),
		"ap":   ApHash(s),
		"fnv":  FnvHashString(s),
	}
	for name, v := range checks {
		// Only verify stability: the same string should produce the same result.
		if v != checkAgain(name, s) {
			t.Fatalf("%s is unstable", name)
		}
	}
	if JavaDefaultHash("a") != 97 {
		t.Fatal("javaDefault hash 'a' should be 97")
	}
	if TianlHash("") != 0 {
		t.Fatal("tianl empty should be 0")
	}
}

// checkAgain runs the same algorithm again for stability tests.
func checkAgain(name, s string) int32 {
	switch name {
	case "rs":
		return RsHash(s)
	case "js":
		return JsHash(s)
	case "pjw":
		return PjwHash(s)
	case "elf":
		return ElfHash(s)
	case "bkdr":
		return BkdrHash(s)
	case "sdbm":
		return SdbmHash(s)
	case "djb":
		return DjbHash(s)
	case "ap":
		return ApHash(s)
	case "fnv":
		return FnvHashString(s)
	}
	return 0
}

func TestUtilFactories(t *testing.T) {
	if CreateBitSet(1024, 100, 3) == nil {
		t.Fatal()
	}
	if CreateBitMap(5) == nil {
		t.Fatal()
	}
}

func TestBloomFilterInterface(t *testing.T) {
	var _ BloomFilter = (*BitSetBloomFilter)(nil)
	var _ BloomFilter = (*BitMapBloomFilter)(nil)
	var _ BloomFilter = (*FuncFilter)(nil)
}
