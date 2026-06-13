package maps

import (
	"github.com/stretchr/testify/assert"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func sortedStrings(s []string) []string {
	out := append([]string(nil), s...)
	sort.Strings(out)
	return out
}

// ---------------------------------------------------------------------------
// Construction
// ---------------------------------------------------------------------------

func TestKeysAndValues(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	assert.ElementsMatch(t, []string{"a", "b", "c"}, Keys(m))
	assert.ElementsMatch(t, []int{1, 2, 3}, Values(m))

	// nil-safe
	assert.Empty(t, Keys[string, int](nil))
	assert.Empty(t, Values[string, int](nil))
}

func TestSortedKeysAndValues(t *testing.T) {
	m := map[string]int{"c": 3, "a": 1, "b": 2}
	assert.Equal(t, []string{"a", "b", "c"}, SortedKeys(m))
	assert.Equal(t, []int{1, 2, 3}, SortedValues(m))

	descending := SortedKeysFunc(m, func(a, b string) bool { return a > b })
	assert.Equal(t, []string{"c", "b", "a"}, descending)
}

func TestKeysOf(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 1}
	got := sortedStrings(KeysOf(m, 1))
	assert.Equal(t, []string{"a", "c"}, got)

	assert.Empty(t, KeysOf(m, 99))
}

// ---------------------------------------------------------------------------
// Transformation
// ---------------------------------------------------------------------------

func TestMap(t *testing.T) {
	in := map[string]int{"a": 1, "b": 2}
	out := Map(in, func(k string, v int) (string, string) {
		return strings.ToUpper(k), strconv.Itoa(v)
	})
	assert.Equal(t, map[string]string{"A": "1", "B": "2"}, out)
}

func TestMapKeysAndMapValues(t *testing.T) {
	in := map[string]int{"a": 1, "b": 2}

	mk := MapKeys(in, func(k string, _ int) string { return strings.ToUpper(k) })
	assert.Equal(t, map[string]int{"A": 1, "B": 2}, mk)

	mv := MapValues(in, func(_ string, v int) int { return v * 10 })
	assert.Equal(t, map[string]int{"a": 10, "b": 20}, mv)
}

func TestToSlice(t *testing.T) {
	in := map[string]int{"a": 1, "b": 2}

	got := ToSlice(in, func(k string, v int) string {
		return k + strconv.Itoa(v)
	})
	sort.Strings(got)
	assert.Equal(t, []string{"a1", "b2"}, got)

	empty := ToSlice(map[string]int{}, func(k string, v int) string {
		return k + strconv.Itoa(v)
	})
	assert.NotNil(t, empty)
	assert.Empty(t, empty)

	nilMap := ToSlice(map[string]int(nil), func(k string, v int) string {
		return k + strconv.Itoa(v)
	})
	assert.NotNil(t, nilMap)
	assert.Empty(t, nilMap)
}

func TestFilterAndReject(t *testing.T) {
	in := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}
	keep := Filter(in, func(_ string, v int) bool { return v%2 == 0 })
	drop := Reject(in, func(_ string, v int) bool { return v%2 == 0 })

	assert.Equal(t, map[string]int{"b": 2, "d": 4}, keep)
	assert.Equal(t, map[string]int{"a": 1, "c": 3}, drop)
}

func TestFilterKeysAndFilterValues(t *testing.T) {
	in := map[string]int{"alpha": 1, "beta": 2, "gamma": 3}
	fk := FilterKeys(in, func(k string) bool { return strings.HasPrefix(k, "a") })
	assert.Equal(t, map[string]int{"alpha": 1}, fk)

	fv := FilterValues(in, func(v int) bool { return v > 1 })
	assert.Equal(t, map[string]int{"beta": 2, "gamma": 3}, fv)
}

func TestPartition(t *testing.T) {
	in := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}
	yes, no := Partition(in, func(_ string, v int) bool { return v >= 3 })
	assert.Equal(t, map[string]int{"c": 3, "d": 4}, yes)
	assert.Equal(t, map[string]int{"a": 1, "b": 2}, no)
}

func TestForEach(t *testing.T) {
	in := map[string]int{"a": 1, "b": 2}
	sum := 0
	keys := make([]string, 0, 2)
	ForEach(in, func(k string, v int) {
		sum += v
		keys = append(keys, k)
	})
	assert.Equal(t, 3, sum)
	assert.ElementsMatch(t, []string{"a", "b"}, keys)
}

// ---------------------------------------------------------------------------
// Aggregation
// ---------------------------------------------------------------------------
