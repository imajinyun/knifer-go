package slice

import (
	"errors"
	"reflect"
	"testing"
)

// Tests cover the utility toolkit-core ArrayUtilTest.

func TestSliceBasic(t *testing.T) {
	if !IsEmpty([]int{}) || IsEmpty([]int{1}) {
		t.Fatalf("IsEmpty failed")
	}
	if !IsNotEmpty([]int{1}) {
		t.Fatalf("IsNotEmpty failed")
	}
	if !Contains([]string{"a", "b"}, "b") || Contains([]string{"a"}, "x") {
		t.Fatalf("Contains failed")
	}
	if IndexOf([]int{1, 2, 3}, 2) != 1 {
		t.Fatalf("IndexOf failed")
	}
	if LastIndexOf([]int{1, 2, 1}, 1) != 2 {
		t.Fatalf("LastIndexOf failed")
	}
}

func TestSliceMutation(t *testing.T) {
	a := []int{1, 2, 3, 4}
	Reverse(a)
	if a[0] != 4 || a[3] != 1 {
		t.Fatalf("Reverse failed: %v", a)
	}
	d := Distinct([]int{1, 2, 1, 3, 2})
	if len(d) != 3 || d[0] != 1 || d[1] != 2 || d[2] != 3 {
		t.Fatalf("Distinct failed: %v", d)
	}
	if Join([]int{1, 2, 3}, ",") != "1,2,3" {
		t.Fatalf("Join failed")
	}
	f := Filter([]int{1, 2, 3, 4}, func(v int) bool { return v%2 == 0 })
	if len(f) != 2 || f[0] != 2 || f[1] != 4 {
		t.Fatalf("Filter failed: %v", f)
	}
	m := Map([]int{1, 2, 3}, func(v int) int { return v * v })
	if m[0] != 1 || m[1] != 4 || m[2] != 9 {
		t.Fatalf("Map failed: %v", m)
	}
}

func TestSliceSubAndConcat(t *testing.T) {
	got := Sub([]int{1, 2, 3, 4, 5}, 1, 4)
	if len(got) != 3 || got[0] != 2 || got[2] != 4 {
		t.Fatalf("Sub failed: %v", got)
	}
	got = Sub([]int{1, 2, 3, 4, 5}, -3, -1)
	if len(got) != 2 || got[0] != 3 || got[1] != 4 {
		t.Fatalf("Sub neg failed: %v", got)
	}
	c := Concat([]int{1}, []int{2, 3}, []int{4})
	if len(c) != 4 || c[3] != 4 {
		t.Fatalf("Concat failed: %v", c)
	}
}

func TestSliceLoStyleHelpers(t *testing.T) {
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
	if got := FilterMap(values, func(v int) (int, bool) { return v * 10, v%2 == 0 }); !reflect.DeepEqual(got, []int{20, 20, 40}) {
		t.Fatalf("FilterMap failed: %v", got)
	}
	if got := FlatMap([]int{1, 2}, func(v int) []int { return []int{v, -v} }); !reflect.DeepEqual(got, []int{1, -1, 2, -2}) {
		t.Fatalf("FlatMap failed: %v", got)
	}
	if got := Reduce(values, 0, func(acc, v int) int { return acc + v }); got != 12 {
		t.Fatalf("Reduce failed: %v", got)
	}

	seen := []int{}
	ForEach([]int{1, 2, 3}, func(v int) { seen = append(seen, v) })
	if !reflect.DeepEqual(seen, []int{1, 2, 3}) {
		t.Fatalf("ForEach failed: %v", seen)
	}
	if got, ok := Find(values, func(v int) bool { return v > 2 }); !ok || got != 3 {
		t.Fatalf("Find = %v, %v", got, ok)
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
	if got := Flatten([][]int{{1}, {2, 3}}); !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("Flatten failed: %v", got)
	}
	if got := Compact([]int{0, 1, 0, 2}); !reflect.DeepEqual(got, []int{1, 2}) {
		t.Fatalf("Compact failed: %v", got)
	}
	if got := PartitionBy([]int{1, 3, 2, 4, 5}, func(v int) bool { return v%2 == 0 }); !reflect.DeepEqual(got, [][]int{{1, 3}, {2, 4}, {5}}) {
		t.Fatalf("PartitionBy failed: %v", got)
	}
}

func TestSliceIterators(t *testing.T) {
	values := []string{"go", "js", "rust"}

	seen := []string{}
	for value := range Iter(values) {
		seen = append(seen, value)
	}
	if !reflect.DeepEqual(seen, values) {
		t.Fatalf("Iter = %v", seen)
	}

	indexed := map[int]string{}
	for index, value := range IterIndexed(values) {
		indexed[index] = value
	}
	if !reflect.DeepEqual(indexed, map[int]string{0: "go", 1: "js", 2: "rust"}) {
		t.Fatalf("IterIndexed = %v", indexed)
	}
}

func TestSliceErrorAwareTransforms(t *testing.T) {
	boom := errors.New("boom")

	mapped, err := MapErr([]int{1, 2, 3}, func(v int) (string, error) {
		if v == 3 {
			return "", boom
		}
		return string(rune('a' + v - 1)), nil
	})
	if !errors.Is(err, boom) {
		t.Fatalf("MapErr err = %v, want boom", err)
	}
	if !reflect.DeepEqual(mapped, []string{"a", "b"}) {
		t.Fatalf("MapErr partial = %v", mapped)
	}

	kept, err := FilterErr([]int{1, 2, 3, 4}, func(v int) (bool, error) {
		if v == 4 {
			return false, boom
		}
		return v%2 == 1, nil
	})
	if !errors.Is(err, boom) {
		t.Fatalf("FilterErr err = %v, want boom", err)
	}
	if !reflect.DeepEqual(kept, []int{1, 3}) {
		t.Fatalf("FilterErr partial = %v", kept)
	}

	sum, err := ReduceErr([]int{1, 2, 3}, 0, func(acc, v int) (int, error) {
		if v == 3 {
			return acc, boom
		}
		return acc + v, nil
	})
	if !errors.Is(err, boom) {
		t.Fatalf("ReduceErr err = %v, want boom", err)
	}
	if sum != 3 {
		t.Fatalf("ReduceErr partial = %d", sum)
	}
}

func TestSliceWindowZipHelpers(t *testing.T) {
	if got := Window([]int{1, 2, 3, 4}, 3); !reflect.DeepEqual(got, [][]int{{1, 2, 3}, {2, 3, 4}}) {
		t.Fatalf("Window = %v", got)
	}
	if got := Window([]int{1, 2}, 3); !reflect.DeepEqual(got, [][]int{}) {
		t.Fatalf("Window larger than slice = %v", got)
	}
	if got := Sliding([]int{1, 2, 3, 4, 5}, 2, 2); !reflect.DeepEqual(got, [][]int{{1, 2}, {3, 4}}) {
		t.Fatalf("Sliding = %v", got)
	}
	if got := Sliding([]int{1, 2, 3}, 2, 0); !reflect.DeepEqual(got, [][]int{}) {
		t.Fatalf("Sliding zero step = %v", got)
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
