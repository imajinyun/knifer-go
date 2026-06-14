package maps

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsEmptyAndIsNotEmpty(t *testing.T) {
	var nilMap map[int]int
	assert.True(t, IsEmpty(nilMap))
	assert.False(t, IsNotEmpty(nilMap))

	assert.True(t, IsEmpty(map[int]int{}))
	assert.False(t, IsNotEmpty(map[int]int{}))

	assert.False(t, IsEmpty(map[int]int{1: 1}))
	assert.True(t, IsNotEmpty(map[int]int{1: 1}))
}

func TestContainsKey(t *testing.T) {
	m := map[string]int{"a": 1}
	assert.True(t, ContainsKey(m, "a"))
	assert.False(t, ContainsKey(m, "z"))
}

func TestContainsValue(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	assert.True(t, ContainsValue(m, 1))
	assert.False(t, ContainsValue(m, 99))
}

func TestSome(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	assert.True(t, Some(m, func(_ string, v int) bool { return v > 2 }))
	assert.False(t, Some(m, func(_ string, v int) bool { return v > 100 }))
	assert.False(t, Some(map[string]int{}, func(_ string, _ int) bool { return true }))
}

func TestEvery(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	assert.True(t, Every(m, func(_ string, v int) bool { return v > 0 }))
	assert.False(t, Every(m, func(_ string, v int) bool { return v > 1 }))
	// empty map → vacuously true
	assert.True(t, Every(map[string]int{}, func(_ string, _ int) bool { return false }))
}
