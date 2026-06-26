package vmap_test

import (
	"fmt"
	"sort"

	"github.com/imajinyun/knifer-go/vmap"
)

func ExampleIsEmpty() {
	fmt.Println(vmap.IsEmpty(map[string]int{}))
	fmt.Println(vmap.IsEmpty(map[string]int{"a": 1}))
	// Output:
	// true
	// false
}

func ExampleInverse() {
	inv := vmap.Inverse(map[string]int{"a": 1})
	fmt.Println(inv[1])
	// Output: a
}

func ExampleMerge() {
	merged := vmap.Merge(map[string]int{"a": 1}, map[string]int{"b": 2})
	fmt.Println(merged["a"], merged["b"])
	// Output: 1 2
}

func ExampleMapErr() {
	mapped, err := vmap.MapErr(map[string]int{"a": 1}, func(k string, v int) (string, string, error) {
		return k + k, fmt.Sprint(v), nil
	})
	fmt.Println(mapped["aa"], err)
	// Output: 1 <nil>
}

func ExampleMapKeysErr() {
	mapped, err := vmap.MapKeysErr(map[string]int{"a": 1}, func(k string, v int) (string, error) {
		return k + "!", nil
	})
	fmt.Println(mapped["a!"], err)
	// Output: 1 <nil>
}

func ExampleMapValuesErr() {
	mapped, err := vmap.MapValuesErr(map[string]int{"a": 1}, func(k string, v int) (string, error) {
		return fmt.Sprintf("%s=%d", k, v), nil
	})
	fmt.Println(mapped["a"], err)
	// Output: a=1 <nil>
}

func ExampleFilterErr() {
	filtered, err := vmap.FilterErr(map[string]int{"a": 1, "b": 2}, func(k string, v int) (bool, error) {
		return v%2 == 1, nil
	})
	fmt.Println(filtered["a"], filtered["b"], err)
	// Output: 1 0 <nil>
}

func ExampleReduceErr() {
	sum, err := vmap.ReduceErr(map[string]int{"a": 1}, 0, func(acc int, k string, v int) (int, error) {
		return acc + v, nil
	})
	fmt.Println(sum, err)
	// Output: 1 <nil>
}

func ExampleIter() {
	m := map[string]int{"b": 2, "a": 1}
	items := make([]string, 0, len(m))
	for key, value := range vmap.Iter(m) {
		items = append(items, fmt.Sprintf("%s=%d", key, value))
	}
	sort.Strings(items)
	fmt.Println(items)
	// Output: [a=1 b=2]
}

func ExampleIterKeys() {
	m := map[string]int{"b": 2, "a": 1}
	keys := make([]string, 0, len(m))
	for key := range vmap.IterKeys(m) {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	fmt.Println(keys)
	// Output: [a b]
}

func ExampleIterValues() {
	m := map[string]int{"b": 2, "a": 1}
	values := make([]int, 0, len(m))
	for value := range vmap.IterValues(m) {
		values = append(values, value)
	}
	sort.Ints(values)
	fmt.Println(values)
	// Output: [1 2]
}

func ExampleNew() {
	m := vmap.New[string, int]()
	m["count"] = 2

	fmt.Println(m["count"])
	// Output: 2
}

func ExampleOf() {
	m := vmap.Of[string, int]("a", 1, "b", 2)
	fmt.Println(m["a"], m["b"], len(m))
	// Output: 1 2 2
}

func ExampleContainsKey() {
	m := map[string]int{"a": 1}
	fmt.Println(vmap.ContainsKey(m, "a"))
	fmt.Println(vmap.ContainsKey(m, "b"))
	// Output:
	// true
	// false
}

func ExampleKeys() {
	keys := vmap.Keys(map[string]int{"b": 2, "a": 1})
	sort.Strings(keys)
	fmt.Println(keys)
	// Output: [a b]
}

func ExampleValues() {
	values := vmap.Values(map[string]int{"b": 2, "a": 1})
	sort.Ints(values)
	fmt.Println(values)
	// Output: [1 2]
}

func ExampleFilter() {
	filtered := vmap.Filter(map[string]int{"a": 1, "b": 2}, func(_ string, value int) bool {
		return value%2 == 0
	})
	fmt.Println(filtered)
	// Output: map[b:2]
}

func ExampleGroupBy() {
	grouped := vmap.GroupBy([]string{"go", "js", "ts"}, func(value string) int {
		return len(value)
	})
	sort.Strings(grouped[2])
	fmt.Println(grouped[2])
	// Output: [go js ts]
}

func ExampleCountBy() {
	counts := vmap.CountBy([]string{"go", "js", "rust"}, func(value string) int {
		return len(value)
	})
	fmt.Println(counts[2], counts[4])
	// Output: 2 1
}

func ExampleClone() {
	original := map[string][]int{"a": {1, 2}}
	cloned := vmap.Clone(original)
	cloned["a"] = append(cloned["a"], 3)

	fmt.Println(original["a"])
	fmt.Println(cloned["a"])
	// Output:
	// [1 2]
	// [1 2 3]
}

func ExampleDiff() {
	diff := vmap.Diff(map[string]int{"a": 1, "b": 2}, map[string]int{"b": 20})
	fmt.Println(diff)
	// Output: map[a:1]
}

func ExampleAssign() {
	assigned := vmap.Assign(map[string]int{"a": 1}, map[string]int{"a": 2, "b": 3})
	fmt.Println(assigned["a"], assigned["b"])
	// Output: 2 3
}

func ExampleClear() {
	m := map[string]int{"a": 1, "b": 2}
	vmap.Clear(m)
	fmt.Println(len(m))
	// Output: 0
}

func ExampleContainsValue() {
	m := map[string]int{"a": 1, "b": 2}
	fmt.Println(vmap.ContainsValue(m, 2))
	fmt.Println(vmap.ContainsValue(m, 3))
	// Output:
	// true
	// false
}

func ExampleEntries() {
	entries := vmap.Entries(map[string]int{"b": 2, "a": 1})
	items := make([]string, 0, len(entries))
	for _, entry := range entries {
		items = append(items, fmt.Sprintf("%s=%d", entry.Key, entry.Value))
	}
	sort.Strings(items)
	fmt.Println(items)
	// Output: [a=1 b=2]
}

func ExampleEqual() {
	fmt.Println(vmap.Equal(map[string]int{"a": 1}, map[string]int{"a": 1}))
	fmt.Println(vmap.Equal(map[string]int{"a": 1}, map[string]int{"a": 2}))
	// Output:
	// true
	// false
}

func ExampleEqualFunc() {
	equal := vmap.EqualFunc(
		map[string]string{"go": "GO"},
		map[string]string{"go": "go"},
		func(a, b string) bool { return len(a) == len(b) },
	)
	fmt.Println(equal)
	// Output: true
}

func ExampleEvery() {
	m := map[string]int{"a": 2, "b": 4}
	fmt.Println(vmap.Every(m, func(_ string, value int) bool { return value%2 == 0 }))
	// Output: true
}

func ExampleFilterKeys() {
	filtered := vmap.FilterKeys(map[string]int{"keep": 1, "drop": 2}, func(key string) bool {
		return key == "keep"
	})
	fmt.Println(filtered["keep"], len(filtered))
	// Output: 1 1
}

func ExampleFilterValues() {
	filtered := vmap.FilterValues(map[string]int{"a": 1, "b": 2}, func(value int) bool {
		return value > 1
	})
	fmt.Println(filtered["b"], len(filtered))
	// Output: 2 1
}

func ExampleFind() {
	key, value, ok := vmap.Find(map[string]int{"a": 1}, func(_ string, value int) bool {
		return value == 1
	})
	fmt.Println(key, value, ok)
	// Output: a 1 true
}

func ExampleFindKey() {
	key, ok := vmap.FindKey(map[string]int{"a": 1}, func(value int) bool {
		return value == 1
	})
	fmt.Println(key, ok)
	// Output: a true
}

func ExampleForEach() {
	m := map[string]int{"b": 2, "a": 1}
	items := make([]string, 0, len(m))
	vmap.ForEach(m, func(key string, value int) {
		items = append(items, fmt.Sprintf("%s=%d", key, value))
	})
	sort.Strings(items)
	fmt.Println(items)
	// Output: [a=1 b=2]
}

func ExampleFromEntries() {
	m := vmap.FromEntries([]vmap.Pair[string, int]{
		{Key: "a", Value: 1},
		{Key: "b", Value: 2},
	})
	fmt.Println(m["a"], m["b"])
	// Output: 1 2
}

func ExampleFromPairs() {
	m := vmap.FromPairs(
		vmap.Pair[string, int]{Key: "a", Value: 1},
		vmap.Pair[string, int]{Key: "b", Value: 2},
	)
	fmt.Println(m["a"], m["b"])
	// Output: 1 2
}

func ExampleGet() {
	m := map[string]int{"a": 1}
	fmt.Println(vmap.Get(m, "a"))
	fmt.Println(vmap.Get(m, "b"))
	// Output:
	// 1
	// 0
}

func ExampleGetAny() {
	value, ok := vmap.GetAny(map[string]int{"b": 2}, "a", "b")
	fmt.Println(value, ok)
	// Output: 2 true
}

func ExampleGetOr() {
	m := map[string]int{"a": 1}
	fmt.Println(vmap.GetOr(m, "a", 9))
	fmt.Println(vmap.GetOr(m, "b", 9))
	// Output:
	// 1
	// 9
}

func ExampleIntersect() {
	intersected := vmap.Intersect(
		map[string]int{"a": 1, "b": 2},
		map[string]int{"b": 20, "c": 30},
	)
	fmt.Println(intersected["b"], len(intersected))
	// Output: 20 1
}

func ExampleInvert() {
	inverted := vmap.Invert(map[string]int{"a": 1})
	fmt.Println(inverted[1])
	// Output: a
}

func ExampleIsNotEmpty() {
	fmt.Println(vmap.IsNotEmpty(map[string]int{}))
	fmt.Println(vmap.IsNotEmpty(map[string]int{"a": 1}))
	// Output:
	// false
	// true
}

func ExampleKeysOf() {
	keys := vmap.KeysOf(map[string]int{"b": 2, "a": 1, "c": 1}, 1)
	sort.Strings(keys)
	fmt.Println(keys)
	// Output: [a c]
}

func ExampleMap() {
	mapped := vmap.Map(map[string]int{"a": 1}, func(key string, value int) (string, string) {
		return key + key, fmt.Sprint(value)
	})
	fmt.Println(mapped["aa"])
	// Output: 1
}

func ExampleMapKeys() {
	mapped := vmap.MapKeys(map[string]int{"a": 1}, func(key string, _ int) string {
		return key + "!"
	})
	fmt.Println(mapped["a!"])
	// Output: 1
}

func ExampleMapValues() {
	mapped := vmap.MapValues(map[string]int{"a": 1}, func(key string, value int) string {
		return fmt.Sprintf("%s=%d", key, value)
	})
	fmt.Println(mapped["a"])
	// Output: a=1
}

func ExampleMergeFunc() {
	merged := vmap.MergeFunc(
		func(oldValue, newValue int) int { return oldValue + newValue },
		map[string]int{"a": 1},
		map[string]int{"a": 2, "b": 3},
	)
	fmt.Println(merged["a"], merged["b"])
	// Output: 3 3
}

func ExampleMergeWithOverwrite() {
	dst := map[string]int{"a": 1}
	vmap.MergeWithOverwrite(dst, map[string]int{"a": 2, "b": 3})
	fmt.Println(dst["a"], dst["b"])
	// Output: 2 3
}

func ExampleMergeWithoutOverwrite() {
	dst := map[string]int{"a": 1}
	vmap.MergeWithoutOverwrite(dst, map[string]int{"a": 2, "b": 3})
	fmt.Println(dst["a"], dst["b"])
	// Output: 1 3
}

func ExampleNewWithCap() {
	m := vmap.NewWithCap[string, int](2)
	m["a"] = 1
	fmt.Println(m["a"], len(m))
	// Output: 1 1
}

func ExampleOfE() {
	m, err := vmap.OfE[string, int]("a", 1, "b", 2)
	fmt.Println(m["a"], m["b"], err)
	// Output: 1 2 <nil>
}

func ExampleOmit() {
	omitted := vmap.Omit(map[string]int{"a": 1, "b": 2}, "b")
	fmt.Println(omitted["a"], len(omitted))
	// Output: 1 1
}

func ExampleOmitBy() {
	omitted := vmap.OmitBy(map[string]int{"a": 1, "b": 2}, func(_ string, value int) bool {
		return value%2 == 0
	})
	fmt.Println(omitted["a"], len(omitted))
	// Output: 1 1
}

func ExampleOrEmpty() {
	m := vmap.OrEmpty[string, int](nil)
	fmt.Println(m == nil, len(m))
	// Output: false 0
}

func ExamplePartition() {
	even, odd := vmap.Partition(map[string]int{"a": 1, "b": 2}, func(_ string, value int) bool {
		return value%2 == 0
	})
	fmt.Println(even["b"], odd["a"])
	// Output: 2 1
}

func ExamplePick() {
	picked := vmap.Pick(map[string]int{"a": 1, "b": 2}, "b")
	fmt.Println(picked["b"], len(picked))
	// Output: 2 1
}

func ExamplePickBy() {
	picked := vmap.PickBy(map[string]int{"a": 1, "b": 2}, func(_ string, value int) bool {
		return value%2 == 0
	})
	fmt.Println(picked["b"], len(picked))
	// Output: 2 1
}

func ExampleReduce() {
	sum := vmap.Reduce(map[string]int{"a": 1}, 10, func(acc int, _ string, value int) int {
		return acc + value
	})
	fmt.Println(sum)
	// Output: 11
}

func ExampleReject() {
	rejected := vmap.Reject(map[string]int{"a": 1, "b": 2}, func(_ string, value int) bool {
		return value%2 == 0
	})
	fmt.Println(rejected["a"], len(rejected))
	// Output: 1 1
}

func ExampleSome() {
	m := map[string]int{"a": 1, "b": 2}
	fmt.Println(vmap.Some(m, func(_ string, value int) bool { return value%2 == 0 }))
	// Output: true
}

func ExampleSortedKeys() {
	keys := vmap.SortedKeys(map[string]int{"b": 2, "a": 1})
	fmt.Println(keys)
	// Output: [a b]
}

func ExampleSortedKeysFunc() {
	keys := vmap.SortedKeysFunc(map[string]int{"aa": 1, "b": 2}, func(a, b string) bool {
		return len(a) < len(b)
	})
	fmt.Println(keys)
	// Output: [b aa]
}

func ExampleSortedValues() {
	values := vmap.SortedValues(map[string]int{"b": 2, "a": 1})
	fmt.Println(values)
	// Output: [1 2]
}

func ExampleSymmetricDiff() {
	diff := vmap.SymmetricDiff(map[string]int{"a": 1, "b": 2}, map[string]int{"b": 20, "c": 3})
	fmt.Println(diff["a"], diff["c"], len(diff))
	// Output: 1 3 2
}

func ExampleToSlice() {
	items := vmap.ToSlice(map[string]int{"b": 2, "a": 1}, func(key string, value int) string {
		return fmt.Sprintf("%s=%d", key, value)
	})
	sort.Strings(items)
	fmt.Println(items)
	// Output: [a=1 b=2]
}

func ExampleUpdate() {
	updated := vmap.Update(map[string]int{"a": 1}, map[string]int{"a": 2, "b": 3})
	fmt.Println(updated["a"], updated["b"])
	// Output: 2 3
}
