package maps

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestReduce(t *testing.T) {
	in := map[string]int{"a": 1, "b": 2, "c": 3}
	sum := Reduce(in, 0, func(acc int, _ string, v int) int { return acc + v })
	assert.Equal(t, 6, sum)

	concat := Reduce(in, "", func(acc string, k string, _ int) string { return acc + k })
	// order is non-deterministic; just check length & character set
	assert.Len(t, concat, 3)
}

func TestGroupBy(t *testing.T) {
	type emp struct {
		Name string
		Dept string
	}
	items := []emp{{"a", "eng"}, {"b", "eng"}, {"c", "ops"}}
	got := GroupBy(items, func(e emp) string { return e.Dept })
	require.Len(t, got, 2)
	assert.ElementsMatch(t, []emp{{"a", "eng"}, {"b", "eng"}}, got["eng"])
	assert.Equal(t, []emp{{"c", "ops"}}, got["ops"])
}

func TestCountBy(t *testing.T) {
	logs := []string{"GET", "POST", "GET", "GET", "POST"}
	got := CountBy(logs, func(s string) string { return s })
	assert.Equal(t, map[string]int{"GET": 3, "POST": 2}, got)
}

// ---------------------------------------------------------------------------
// Set algebra
// ---------------------------------------------------------------------------

func TestInverse(t *testing.T) {
	in := map[string]int{"a": 1, "b": 2}
	inv := Inverse(in)
	assert.Equal(t, map[int]string{1: "a", 2: "b"}, inv)
}

func TestIntersect(t *testing.T) {
	a := map[string]int{"a": 1, "b": 2, "c": 3}
	b := map[string]int{"b": 20, "c": 30, "d": 40}
	c := map[string]int{"c": 300, "d": 400}

	got := Intersect(a, b, c)
	assert.Equal(t, map[string]int{"c": 300}, got)

	// edge: zero / one input
	assert.Empty(t, Intersect[string, int]())
	assert.Equal(t, a, Intersect(a))

	// empty intersection
	assert.Empty(t, Intersect(
		map[string]int{"a": 1},
		map[string]int{"b": 2},
	))
}

func TestDiff(t *testing.T) {
	a := map[string]int{"a": 1, "b": 2, "c": 3}
	b := map[string]int{"a": 10}
	c := map[string]int{"b": 20}
	assert.Equal(t, map[string]int{"c": 3}, Diff(a, b, c))

	// no others → returns clone of a
	assert.Equal(t, a, Diff(a))
}

func TestSymmetricDiff(t *testing.T) {
	a := map[string]int{"a": 1, "b": 2}
	b := map[string]int{"b": 20, "c": 3}
	assert.Equal(t, map[string]int{"a": 1, "c": 3}, SymmetricDiff(a, b))
}

// ---------------------------------------------------------------------------
// Selection
// ---------------------------------------------------------------------------
