package vmap

import (
	"reflect"
	"sort"
	"testing"
)

func TestMapTransformFilterAggregateFacades(t *testing.T) {
	m := map[string]int{"b": 2, "a": 1, "c": 3}
	if got := SortedKeys(m); !reflect.DeepEqual(got, []string{"a", "b", "c"}) {
		t.Fatalf("SortedKeys = %#v", got)
	}
	if got := SortedKeysFunc(m, func(a, b string) bool { return a > b }); !reflect.DeepEqual(got, []string{"c", "b", "a"}) {
		t.Fatalf("SortedKeysFunc = %#v", got)
	}
	if got := SortedValues(m); !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("SortedValues = %#v", got)
	}
	keysOf := KeysOf(map[string]int{"a": 1, "b": 2, "c": 1}, 1)
	sort.Strings(keysOf)
	if !reflect.DeepEqual(keysOf, []string{"a", "c"}) {
		t.Fatalf("KeysOf = %#v", keysOf)
	}

	mapped := Map(m, func(k string, v int) (string, string) { return k + k, string(rune('0' + v)) })
	if !reflect.DeepEqual(mapped, map[string]string{"aa": "1", "bb": "2", "cc": "3"}) {
		t.Fatalf("Map = %#v", mapped)
	}
	if got := MapKeys(m, func(k string, _ int) string { return k + "!" }); !reflect.DeepEqual(got, map[string]int{"a!": 1, "b!": 2, "c!": 3}) {
		t.Fatalf("MapKeys = %#v", got)
	}
	if got := MapValues(m, func(_ string, v int) string { return string(rune('A' + v)) }); !reflect.DeepEqual(got, map[string]string{"a": "B", "b": "C", "c": "D"}) {
		t.Fatalf("MapValues = %#v", got)
	}
	toSlice := ToSlice(m, func(k string, v int) string { return k + string(rune('0'+v)) })
	sort.Strings(toSlice)
	if !reflect.DeepEqual(toSlice, []string{"a1", "b2", "c3"}) {
		t.Fatalf("ToSlice = %#v", toSlice)
	}
	if got := ToSlice(map[string]int(nil), func(k string, v int) string { return k + string(rune('0'+v)) }); got == nil || len(got) != 0 {
		t.Fatalf("ToSlice nil = %#v", got)
	}
	if got := Filter(m, func(_ string, v int) bool { return v%2 == 1 }); !reflect.DeepEqual(got, map[string]int{"a": 1, "c": 3}) {
		t.Fatalf("Filter = %#v", got)
	}
	if got := Reject(m, func(_ string, v int) bool { return v%2 == 1 }); !reflect.DeepEqual(got, map[string]int{"b": 2}) {
		t.Fatalf("Reject = %#v", got)
	}
	if got := FilterKeys(m, func(k string) bool { return k != "b" }); !reflect.DeepEqual(got, map[string]int{"a": 1, "c": 3}) {
		t.Fatalf("FilterKeys = %#v", got)
	}
	if got := FilterValues(m, func(v int) bool { return v >= 2 }); !reflect.DeepEqual(got, map[string]int{"b": 2, "c": 3}) {
		t.Fatalf("FilterValues = %#v", got)
	}
	matched, rest := Partition(m, func(_ string, v int) bool { return v >= 2 })
	if !reflect.DeepEqual(matched, map[string]int{"b": 2, "c": 3}) || !reflect.DeepEqual(rest, map[string]int{"a": 1}) {
		t.Fatalf("Partition matched=%#v rest=%#v", matched, rest)
	}

	seen := map[string]int{}
	ForEach(m, func(k string, v int) { seen[k] = v })
	if !reflect.DeepEqual(seen, m) {
		t.Fatalf("ForEach seen = %#v", seen)
	}
	if got := Reduce(m, 0, func(acc int, _ string, v int) int { return acc + v }); got != 6 {
		t.Fatalf("Reduce = %d", got)
	}
	if got := GroupBy([]string{"go", "js", "java"}, func(s string) int { return len(s) }); !reflect.DeepEqual(got, map[int][]string{2: {"go", "js"}, 4: {"java"}}) {
		t.Fatalf("GroupBy = %#v", got)
	}
	if got := CountBy([]string{"go", "js", "java"}, func(s string) int { return len(s) }); !reflect.DeepEqual(got, map[int]int{2: 2, 4: 1}) {
		t.Fatalf("CountBy = %#v", got)
	}
}
