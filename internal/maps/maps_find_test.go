package maps

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindAndFindKey(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}

	k, v, ok := Find(m, func(_ string, v int) bool { return v == 2 })
	assert.True(t, ok)
	assert.Equal(t, "b", k)
	assert.Equal(t, 2, v)

	_, _, ok = Find(m, func(_ string, v int) bool { return v < 0 })
	assert.False(t, ok)

	fk, ok := FindKey(m, func(v int) bool { return v == 3 })
	assert.True(t, ok)
	assert.Equal(t, "c", fk)

	_, ok = FindKey(m, func(v int) bool { return v > 999 })
	assert.False(t, ok)
}
