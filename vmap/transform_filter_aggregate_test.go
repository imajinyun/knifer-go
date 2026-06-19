package vmap

import (
	"cmp"
	"errors"
	"reflect"
	"slices"
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
	slices.Sort(keysOf)
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
	slices.Sort(toSlice)
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

func TestMapLoStyleFacades(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}

	entries := Entries(m)
	slices.SortFunc(entries, func(a, b Pair[string, int]) int { return cmp.Compare(a.Key, b.Key) })
	if !reflect.DeepEqual(entries, []Pair[string, int]{{Key: "a", Value: 1}, {Key: "b", Value: 2}, {Key: "c", Value: 3}}) {
		t.Fatalf("Entries = %#v", entries)
	}
	if got := FromEntries(entries); !reflect.DeepEqual(got, m) {
		t.Fatalf("FromEntries = %#v", got)
	}

	if got := Pick(m, "a", "c", "missing"); !reflect.DeepEqual(got, map[string]int{"a": 1, "c": 3}) {
		t.Fatalf("Pick = %#v", got)
	}
	if got := Omit(m, "b", "missing"); !reflect.DeepEqual(got, map[string]int{"a": 1, "c": 3}) {
		t.Fatalf("Omit = %#v", got)
	}
	if got := PickBy(m, func(_ string, v int) bool { return v%2 == 1 }); !reflect.DeepEqual(got, map[string]int{"a": 1, "c": 3}) {
		t.Fatalf("PickBy = %#v", got)
	}
	if got := OmitBy(m, func(_ string, v int) bool { return v%2 == 1 }); !reflect.DeepEqual(got, map[string]int{"b": 2}) {
		t.Fatalf("OmitBy = %#v", got)
	}

	if got := Assign(map[string]int{"a": 1}, map[string]int{"a": 10, "b": 2}, nil); !reflect.DeepEqual(got, map[string]int{"a": 10, "b": 2}) {
		t.Fatalf("Assign = %#v", got)
	}
	if got := Invert(map[string]int{"a": 1, "b": 2}); !reflect.DeepEqual(got, map[int]string{1: "a", 2: "b"}) {
		t.Fatalf("Invert = %#v", got)
	}
}

func TestMapErrorAwareFacades(t *testing.T) {
	boom := errors.New("boom")
	m := map[string]int{"a": 1}

	mapped, err := MapErr(m, func(k string, v int) (string, string, error) {
		return k + k, string(rune('0' + v)), nil
	})
	if err != nil || !reflect.DeepEqual(mapped, map[string]string{"aa": "1"}) {
		t.Fatalf("MapErr = %#v, %v", mapped, err)
	}

	keys, err := MapKeysErr(m, func(k string, _ int) (string, error) {
		return k + "!", nil
	})
	if err != nil || !reflect.DeepEqual(keys, map[string]int{"a!": 1}) {
		t.Fatalf("MapKeysErr = %#v, %v", keys, err)
	}

	values, err := MapValuesErr(m, func(k string, v int) (string, error) {
		return string(rune('A' + v)), nil
	})
	if err != nil || !reflect.DeepEqual(values, map[string]string{"a": "B"}) {
		t.Fatalf("MapValuesErr = %#v, %v", values, err)
	}

	filtered, err := FilterErr(m, func(k string, v int) (bool, error) {
		return v%2 == 1, nil
	})
	if err != nil || !reflect.DeepEqual(filtered, map[string]int{"a": 1}) {
		t.Fatalf("FilterErr = %#v, %v", filtered, err)
	}

	sum, err := ReduceErr(m, 0, func(acc int, k string, v int) (int, error) {
		return acc + v, nil
	})
	if err != nil || sum != 1 {
		t.Fatalf("ReduceErr = %d, %v", sum, err)
	}

	if _, err := MapErr(map[string]int{"a": 1}, func(k string, v int) (string, string, error) {
		return "", "", boom
	}); !errors.Is(err, boom) {
		t.Fatalf("MapErr error = %v", err)
	}
	if _, err := MapKeysErr(map[string]int{"a": 1}, func(k string, v int) (string, error) {
		return "", boom
	}); !errors.Is(err, boom) {
		t.Fatalf("MapKeysErr error = %v", err)
	}
	if _, err := MapValuesErr(map[string]int{"a": 1}, func(k string, v int) (string, error) {
		return "", boom
	}); !errors.Is(err, boom) {
		t.Fatalf("MapValuesErr error = %v", err)
	}
	if _, err := FilterErr(map[string]int{"a": 1}, func(k string, v int) (bool, error) {
		return false, boom
	}); !errors.Is(err, boom) {
		t.Fatalf("FilterErr error = %v", err)
	}
	if _, err := ReduceErr(map[string]int{"a": 1}, 0, func(acc int, k string, v int) (int, error) {
		return acc, boom
	}); !errors.Is(err, boom) {
		t.Fatalf("ReduceErr error = %v", err)
	}
}

func TestMapIteratorFacades(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}

	entries := map[string]int{}
	for key, value := range Iter(m) {
		entries[key] = value
	}
	if !reflect.DeepEqual(entries, m) {
		t.Fatalf("Iter = %#v", entries)
	}

	keys := []string{}
	for key := range IterKeys(m) {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	if !reflect.DeepEqual(keys, []string{"a", "b"}) {
		t.Fatalf("IterKeys = %#v", keys)
	}

	values := []int{}
	for value := range IterValues(m) {
		values = append(values, value)
	}
	slices.Sort(values)
	if !reflect.DeepEqual(values, []int{1, 2}) {
		t.Fatalf("IterValues = %#v", values)
	}
}
