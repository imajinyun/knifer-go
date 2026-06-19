package vslice

import (
	"errors"
	"reflect"
	"testing"
)

func TestSliceFacade(t *testing.T) {
	if !IsEmpty([]int{}) || !IsNotEmpty([]int{1}) {
		t.Fatal("empty checks failed")
	}
	values := []int{1, 2, 2, 3}
	if !Contains(values, 2) || IndexOf(values, 2) != 1 || LastIndexOf(values, 2) != 2 {
		t.Fatal("contains/index helpers failed")
	}
	reversed := Reverse([]int{1, 2, 3})
	if !reflect.DeepEqual(reversed, []int{3, 2, 1}) {
		t.Fatalf("Reverse failed: %v", reversed)
	}
	if got := Distinct(values); !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("Distinct failed: %v", got)
	}
	if Join([]int{1, 2, 3}, ",") != "1,2,3" {
		t.Fatal("Join failed")
	}
	if got := Filter(values, func(v int) bool { return v%2 == 0 }); !reflect.DeepEqual(got, []int{2, 2}) {
		t.Fatalf("Filter failed: %v", got)
	}
	if got := Map([]int{1, 2}, func(v int) string { return string(rune('a' + v - 1)) }); !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("Map failed: %v", got)
	}
	if got := Sub([]int{1, 2, 3, 4}, -3, -1); !reflect.DeepEqual(got, []int{2, 3}) {
		t.Fatalf("Sub failed: %v", got)
	}
	if got := Concat([]int{1}, []int{2, 3}); !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("Concat failed: %v", got)
	}
	if got := Union([]int{1, 2}, []int{2, 3}); !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("Union failed: %v", got)
	}
	if got := Intersection([]int{1, 2, 3}, []int{2, 3, 4}); !reflect.DeepEqual(got, []int{2, 3}) {
		t.Fatalf("Intersection failed: %v", got)
	}
	if got := Subtract([]int{1, 2, 3}, []int{2}); !reflect.DeepEqual(got, []int{1, 3}) {
		t.Fatalf("Subtract failed: %v", got)
	}
	if got := Page([]int{1, 2, 3, 4, 5}, 2, 2); !reflect.DeepEqual(got, []int{3, 4}) {
		t.Fatalf("Page failed: %v", got)
	}
}

func TestSliceLoStyleFacades(t *testing.T) {
	values := []int{1, 2, 2, 3, 4}

	if got := Uniq(values); !reflect.DeepEqual(got, []int{1, 2, 3, 4}) {
		t.Fatalf("Uniq failed: %v", got)
	}
	if got := UniqBy([]string{"go", "js", "java"}, func(s string) int { return len(s) }); !reflect.DeepEqual(got, []string{"go", "java"}) {
		t.Fatalf("UniqBy failed: %v", got)
	}
	if got := Reject(values, func(v int) bool { return v%2 == 0 }); !reflect.DeepEqual(got, []int{1, 3}) {
		t.Fatalf("Reject failed: %v", got)
	}
	if got := FilterMap(values, func(v int) (string, bool) {
		if v%2 == 0 {
			return string(rune('a' + v)), true
		}
		return "", false
	}); !reflect.DeepEqual(got, []string{"c", "c", "e"}) {
		t.Fatalf("FilterMap failed: %v", got)
	}
	if got := FlatMap([]int{1, 2}, func(v int) []int { return []int{v, -v} }); !reflect.DeepEqual(got, []int{1, -1, 2, -2}) {
		t.Fatalf("FlatMap failed: %v", got)
	}
	if got := Reduce(values, 0, func(acc, v int) int { return acc + v }); got != 12 {
		t.Fatalf("Reduce failed: %v", got)
	}

	seen := []int{}
	ForEach([]int{1, 2, 3}, func(v int) { seen = append(seen, v*2) })
	if !reflect.DeepEqual(seen, []int{2, 4, 6}) {
		t.Fatalf("ForEach failed: %v", seen)
	}
	if got, ok := Find(values, func(v int) bool { return v > 2 }); !ok || got != 3 {
		t.Fatalf("Find = %v, %v", got, ok)
	}
	if got, ok := Find(values, func(v int) bool { return v > 9 }); ok || got != 0 {
		t.Fatalf("Find missing = %v, %v", got, ok)
	}
	if got := FindIndex(values, func(v int) bool { return v == 3 }); got != 3 {
		t.Fatalf("FindIndex = %d", got)
	}

	words := []string{"go", "js", "rust", "java"}
	if got := GroupBy(words, func(s string) int { return len(s) }); !reflect.DeepEqual(got, map[int][]string{2: {"go", "js"}, 4: {"rust", "java"}}) {
		t.Fatalf("GroupBy failed: %v", got)
	}
	if got := CountBy(words, func(s string) int { return len(s) }); !reflect.DeepEqual(got, map[int]int{2: 2, 4: 2}) {
		t.Fatalf("CountBy failed: %v", got)
	}
	if got := KeyBy(words, func(s string) int { return len(s) }); !reflect.DeepEqual(got, map[int]string{2: "js", 4: "java"}) {
		t.Fatalf("KeyBy failed: %v", got)
	}
	if got := Associate(words, func(s string) (string, int) { return s, len(s) }); !reflect.DeepEqual(got, map[string]int{"go": 2, "js": 2, "rust": 4, "java": 4}) {
		t.Fatalf("Associate failed: %v", got)
	}
	if got := SliceToMap(words, func(s string) (int, string) { return len(s), s }); !reflect.DeepEqual(got, map[int]string{2: "js", 4: "java"}) {
		t.Fatalf("SliceToMap failed: %v", got)
	}

	if got := Chunk([]int{1, 2, 3, 4, 5}, 2); !reflect.DeepEqual(got, [][]int{{1, 2}, {3, 4}, {5}}) {
		t.Fatalf("Chunk failed: %v", got)
	}
	if got := Chunk([]int{1, 2}, 0); !reflect.DeepEqual(got, [][]int{}) {
		t.Fatalf("Chunk zero failed: %v", got)
	}
	if got := Flatten([][]int{{1, 2}, {}, {3}}); !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("Flatten failed: %v", got)
	}
	if got := Compact([]int{0, 1, 0, 2}); !reflect.DeepEqual(got, []int{1, 2}) {
		t.Fatalf("Compact failed: %v", got)
	}
	if got := PartitionBy([]int{1, 3, 2, 4, 5}, func(v int) bool { return v%2 == 0 }); !reflect.DeepEqual(got, [][]int{{1, 3}, {2, 4}, {5}}) {
		t.Fatalf("PartitionBy failed: %v", got)
	}
}

func TestSliceErrorAwareFacades(t *testing.T) {
	boom := errors.New("boom")

	mapped, err := MapErr([]int{1, 2, 3}, func(v int) (string, error) {
		if v == 3 {
			return "", boom
		}
		return string(rune('a' + v - 1)), nil
	})
	if !errors.Is(err, boom) || !reflect.DeepEqual(mapped, []string{"a", "b"}) {
		t.Fatalf("MapErr = %v, %v", mapped, err)
	}

	kept, err := FilterErr([]int{1, 2, 3, 4}, func(v int) (bool, error) {
		if v == 4 {
			return false, boom
		}
		return v%2 == 1, nil
	})
	if !errors.Is(err, boom) || !reflect.DeepEqual(kept, []int{1, 3}) {
		t.Fatalf("FilterErr = %v, %v", kept, err)
	}

	sum, err := ReduceErr([]int{1, 2, 3}, 0, func(acc, v int) (int, error) {
		if v == 3 {
			return acc, boom
		}
		return acc + v, nil
	})
	if !errors.Is(err, boom) || sum != 3 {
		t.Fatalf("ReduceErr = %d, %v", sum, err)
	}
}

func TestSliceWindowZipFacades(t *testing.T) {
	if got := Window([]int{1, 2, 3, 4}, 3); !reflect.DeepEqual(got, [][]int{{1, 2, 3}, {2, 3, 4}}) {
		t.Fatalf("Window = %v", got)
	}
	if got := Sliding([]int{1, 2, 3, 4, 5}, 2, 2); !reflect.DeepEqual(got, [][]int{{1, 2}, {3, 4}}) {
		t.Fatalf("Sliding = %v", got)
	}
	pairs := Zip2([]string{"a", "b", "c"}, []int{1, 2})
	if !reflect.DeepEqual(pairs, []Pair[string, int]{{First: "a", Second: 1}, {First: "b", Second: 2}}) {
		t.Fatalf("Zip2 = %v", pairs)
	}
	left, right := Unzip2(pairs)
	if !reflect.DeepEqual(left, []string{"a", "b"}) || !reflect.DeepEqual(right, []int{1, 2}) {
		t.Fatalf("Unzip2 left=%v right=%v", left, right)
	}
}

func TestSliceIteratorFacades(t *testing.T) {
	values := []int{10, 20, 30}

	seen := []int{}
	for value := range Iter(values) {
		seen = append(seen, value)
	}
	if !reflect.DeepEqual(seen, values) {
		t.Fatalf("Iter failed: %v", seen)
	}

	indexed := map[int]int{}
	for index, value := range IterIndexed(values) {
		indexed[index] = value
	}
	if !reflect.DeepEqual(indexed, map[int]int{0: 10, 1: 20, 2: 30}) {
		t.Fatalf("IterIndexed failed: %v", indexed)
	}
}
