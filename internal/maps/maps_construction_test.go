package maps

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewAndNewWithCap(t *testing.T) {
	m := New[string, int]()
	assert.NotNil(t, m)
	assert.Empty(t, m)

	m2 := NewWithCap[string, int](128)
	assert.NotNil(t, m2)
	assert.Empty(t, m2)

	// negative hint is normalized to 0
	m3 := NewWithCap[string, int](-5)
	assert.NotNil(t, m3)
}

func TestOf(t *testing.T) {
	m := Of[string, int]("a", 1, "b", 2, "a", 3) // last wins
	assert.Equal(t, map[string]int{"a": 3, "b": 2}, m)

	empty := Of[string, int]()
	assert.NotNil(t, empty)
	assert.Empty(t, empty)
}

func TestOf_PanicOnOddArgs(t *testing.T) {
	t.Run("odd args panics", func(t *testing.T) {
		args := []any{"a", 1, "b"}
		assert.PanicsWithValue(t, "maps.Of: odd number of arguments", func() {
			Of[string, int](args...)
		})
	})
}

func TestOfE(t *testing.T) {
	tests := []struct {
		name    string
		args    []any
		want    map[string]int
		wantErr bool
	}{
		{
			name: "builds map",
			args: []any{"a", 1, "b", 2, "a", 3},
			want: map[string]int{"a": 3, "b": 2},
		},
		{
			name:    "rejects odd args",
			args:    []any{"a", 1, "b"},
			wantErr: true,
		},
		{
			name:    "rejects invalid key type",
			args:    []any{1, 1},
			wantErr: true,
		},
		{
			name:    "rejects invalid value type",
			args:    []any{"a", "bad"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := OfE[string, int](tt.args...)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFromPairs(t *testing.T) {
	got := FromPairs(
		Pair[string, int]{Key: "a", Value: 1},
		Pair[string, int]{Key: "b", Value: 2},
		Pair[string, int]{Key: "a", Value: 3},
	)
	assert.Equal(t, map[string]int{"a": 3, "b": 2}, got)
}

func TestOrEmpty(t *testing.T) {
	var nilMap map[string]int
	got := OrEmpty(nilMap)
	assert.NotNil(t, got)
	assert.Empty(t, got)

	src := map[string]int{"x": 1}
	returned := OrEmpty(src)
	returned["y"] = 2
	assert.Equal(t, 2, src["y"], "OrEmpty should return the original non-nil map")
}

// ---------------------------------------------------------------------------
// Predicates
// ---------------------------------------------------------------------------
