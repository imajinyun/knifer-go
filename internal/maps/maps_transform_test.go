package maps

import (
	"errors"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMap(t *testing.T) {
	in := map[string]int{"a": 1, "b": 2}
	out := Map(in, func(k string, v int) (string, string) {
		return strings.ToUpper(k), strconv.Itoa(v)
	})
	assert.Equal(t, map[string]string{"A": "1", "B": "2"}, out)
}

func TestFromEntries(t *testing.T) {
	got := FromEntries([]Pair[string, int]{{"a", 1}, {"b", 2}})
	assert.Equal(t, map[string]int{"a": 1, "b": 2}, got)

	// duplicate keys — last wins
	got2 := FromEntries([]Pair[string, int]{{"a", 1}, {"a", 99}})
	assert.Equal(t, map[string]int{"a": 99}, got2)

	// empty
	assert.Empty(t, FromEntries[string, int](nil))
}

func TestInverseAndInvert(t *testing.T) {
	in := map[string]int{"a": 1, "b": 2}
	got := Inverse(in)
	assert.Equal(t, map[int]string{1: "a", 2: "b"}, got)

	// Invert is alias of Inverse
	got2 := Invert(in)
	assert.Equal(t, got, got2)
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
	slices.Sort(got)
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

func TestErrorAwareTransforms(t *testing.T) {
	boom := errors.New("boom")
	m := map[string]int{"a": 1}

	mapped, err := MapErr(m, func(k string, v int) (string, string, error) {
		return strings.ToUpper(k), strconv.Itoa(v), nil
	})
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"A": "1"}, mapped)

	keys, err := MapKeysErr(m, func(k string, v int) (string, error) {
		return strings.ToUpper(k), nil
	})
	assert.NoError(t, err)
	assert.Equal(t, map[string]int{"A": 1}, keys)

	values, err := MapValuesErr(m, func(k string, v int) (string, error) {
		return strconv.Itoa(v), nil
	})
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"a": "1"}, values)

	filtered, err := FilterErr(m, func(k string, v int) (bool, error) {
		return v%2 == 1, nil
	})
	assert.NoError(t, err)
	assert.Equal(t, map[string]int{"a": 1}, filtered)

	sum, err := ReduceErr(m, 0, func(acc int, k string, v int) (int, error) {
		return acc + v, nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, sum)

	_, err = MapErr(map[string]int{"a": 1}, func(k string, v int) (string, string, error) {
		return "", "", boom
	})
	assert.True(t, errors.Is(err, boom))
	_, err = MapKeysErr(map[string]int{"a": 1}, func(k string, v int) (string, error) {
		return "", boom
	})
	assert.True(t, errors.Is(err, boom))
	_, err = MapValuesErr(map[string]int{"a": 1}, func(k string, v int) (string, error) {
		return "", boom
	})
	assert.True(t, errors.Is(err, boom))
	_, err = FilterErr(map[string]int{"a": 1}, func(k string, v int) (bool, error) {
		return false, boom
	})
	assert.True(t, errors.Is(err, boom))
	_, err = ReduceErr(map[string]int{"a": 1}, 0, func(acc int, k string, v int) (int, error) {
		return acc, boom
	})
	assert.True(t, errors.Is(err, boom))
}
