package maps

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAndGetOr(t *testing.T) {
	m := map[string]int{"a": 1}
	assert.Equal(t, 1, Get(m, "a"))
	assert.Equal(t, 0, Get(m, "missing"))
	assert.Equal(t, 1, GetOr(m, "a", 99))
	assert.Equal(t, 99, GetOr(m, "missing", 99))
}

func TestGetAny(t *testing.T) {
	headers := map[string]string{"X-Username": "alice"}
	v, ok := GetAny(headers, "X-User", "X-Username", "User")
	assert.True(t, ok)
	assert.Equal(t, "alice", v)

	v2, ok2 := GetAny(headers, "missing-1", "missing-2")
	assert.False(t, ok2)
	assert.Equal(t, "", v2)
}
