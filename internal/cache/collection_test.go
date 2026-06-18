package cache

import "testing"

func TestKeysValuesClearOnFIFO(t *testing.T) {
	c := NewFIFO[string, string](10)
	c.Put("a", "1")
	c.Put("b", "2")
	c.Put("c", "3")

	keys := c.Keys()
	if len(keys) != 3 || keys[0] != "a" || keys[1] != "b" || keys[2] != "c" {
		t.Fatalf("Keys = %v, want [a b c]", keys)
	}

	vals := c.Values()
	if len(vals) != 3 || vals[0] != "1" || vals[1] != "2" || vals[2] != "3" {
		t.Fatalf("Values = %v, want [1 2 3]", vals)
	}

	if !c.ContainsKey("b") {
		t.Fatal("ContainsKey(b) = false")
	}
	c.Clear()
	if c.Size() != 0 || c.ContainsKey("a") {
		t.Fatal("Clear failed")
	}
}

func TestKeysValuesClearOnLRU(t *testing.T) {
	c := NewLRU[string, int](10)
	c.Put("x", 10)
	c.Put("y", 20)

	keys := c.Keys()
	if len(keys) != 2 || keys[0] != "x" || keys[1] != "y" {
		t.Fatalf("Keys = %v", keys)
	}

	if !c.ContainsKey("x") || !c.ContainsKey("y") {
		t.Fatal("ContainsKey failed")
	}
	c.Clear()
	if c.Size() != 0 {
		t.Fatal("Clear failed")
	}
}

func TestKeysValuesClearOnLFU(t *testing.T) {
	c := NewLFU[string, string](10)
	c.Put("a", "1")
	c.Put("b", "2")

	vals := c.Values()
	if len(vals) != 2 {
		t.Fatalf("Values = %v", vals)
	}

	if !c.ContainsKey("a") || !c.ContainsKey("b") {
		t.Fatal("ContainsKey failed")
	}
	c.Clear()
	if c.Size() != 0 {
		t.Fatal("Clear failed")
	}
}
