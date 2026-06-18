package maps

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestPickBy(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}
	// pick even values
	got := PickBy(m, func(k string, v int) bool { return v%2 == 0 })
	assert.Equal(t, map[string]int{"b": 2, "d": 4}, got)

	// empty result
	assert.Empty(t, PickBy(m, func(k string, v int) bool { return false }))
}

func TestOmitBy(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}
	// omit even values
	got := OmitBy(m, func(k string, v int) bool { return v%2 == 0 })
	assert.Equal(t, map[string]int{"a": 1, "c": 3}, got)

	// omit nothing
	assert.Equal(t, m, OmitBy(m, func(k string, v int) bool { return false }))
}
