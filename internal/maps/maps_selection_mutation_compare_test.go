package maps

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

func TestPick(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	got := Pick(m, "a", "c", "missing")
	assert.Equal(t, map[string]int{"a": 1, "c": 3}, got)

	assert.Empty(t, Pick(m))
	assert.Empty(t, Pick[string, int](nil, "a"))
}

func TestOmit(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	got := Omit(m, "b", "missing")
	assert.Equal(t, map[string]int{"a": 1, "c": 3}, got)

	assert.Equal(t, m, Omit(m))
}

// ---------------------------------------------------------------------------
// Mutation helpers
// ---------------------------------------------------------------------------

func TestClear(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	Clear(m)
	assert.Empty(t, m)
	assert.NotNil(t, m) // still the same allocated map
}

func TestUpdate(t *testing.T) {
	dst := map[string]int{"a": 1}
	src := map[string]int{"a": 10, "b": 2}
	got := Update(dst, src)
	assert.Equal(t, map[string]int{"a": 10, "b": 2}, dst)
	got["c"] = 3
	assert.Equal(t, 3, dst["c"], "Update should return dst for chaining")

	// nil dst is allocated
	got2 := Update[string, int](nil, src)
	assert.NotNil(t, got2)
	assert.Equal(t, src, got2)
}

func TestClone(t *testing.T) {
	m := map[string]int{"a": 1}
	c := Clone(m)
	assert.Equal(t, m, c)

	c["a"] = 999
	assert.Equal(t, 1, m["a"], "Clone must not share storage with the input")

	// nil input → empty non-nil
	cn := Clone[string, int](nil)
	assert.NotNil(t, cn)
	assert.Empty(t, cn)
}

// ---------------------------------------------------------------------------
// Comparison
// ---------------------------------------------------------------------------

func TestEqual(t *testing.T) {
	a := map[string]int{"a": 1, "b": 2}
	b := map[string]int{"b": 2, "a": 1}
	c := map[string]int{"a": 1}
	d := map[string]int{"a": 1, "b": 99}

	assert.True(t, Equal(a, b))
	assert.False(t, Equal(a, c))
	assert.False(t, Equal(a, d))
	assert.True(t, Equal[string, int](nil, nil))
	assert.True(t, Equal(map[string]int{}, map[string]int{}))
}

func TestEqualFunc(t *testing.T) {
	a := map[string]int{"a": 1, "b": 2}
	b := map[string]string{"a": "1", "b": "2"}

	got := EqualFunc(a, b, func(x int, y string) bool {
		return strconv.Itoa(x) == y
	})
	assert.True(t, got)

	notMatching := EqualFunc(a, b, func(x int, _ string) bool { return x < 0 })
	assert.False(t, notMatching)
}

// ---------------------------------------------------------------------------
// Property-style sanity: shape consistency between functions
// ---------------------------------------------------------------------------

func TestKeysValuesShape(t *testing.T) {
	m := map[int]int{1: 10, 2: 20, 3: 30}
	keys := SortedKeys(m)
	values := SortedValues(m)
	require.Len(t, keys, len(m))
	require.Len(t, values, len(m))
	for i, k := range keys {
		assert.Equal(t, m[k], values[i])
	}
}

// ---------------------------------------------------------------------------
// Benchmarks
// ---------------------------------------------------------------------------
