package maps

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAssign(t *testing.T) {
	a := map[string]int{"a": 1, "b": 2}
	b := map[string]int{"b": 20, "c": 3}
	got := Assign(a, b)
	assert.Equal(t, map[string]int{"a": 1, "b": 20, "c": 3}, got)

	// Assign with no inputs
	assert.Empty(t, Assign[string, int]())

	// Assign with single input
	assert.Equal(t, a, Assign(a))
}

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
