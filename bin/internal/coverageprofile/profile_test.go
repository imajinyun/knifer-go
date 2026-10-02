package coverageprofile

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	for _, tt := range []struct {
		name, profile   string
		wantErr         bool
		blocks, covered int
	}{
		{"union", "mode: atomic\np/a.go:1.1,2.2 3 0\np/a.go:1.1,2.2 3 9\np/a.go:3.1,4.2 2 0\n", false, 2, 3},
		{"repeated_hit", "mode: count\np/a.go:1.1,2.2 3 2\np/a.go:1.1,2.2 3 8\n", false, 1, 3},
		{"bad_header", "mode: other\np/a.go:1.1,2.2 3 1\n", true, 0, 0},
		{"bad_line", "mode: atomic\nbroken\n", true, 0, 0},
		{"negative", "mode: atomic\np/a.go:1.1,2.2 3 -1\n", true, 0, 0},
		{"conflicting_blocks", "mode: atomic\np/a.go:1.1,2.2 3 0\np/a.go:1.1,2.2 4 1\n", true, 0, 0},
		{"empty", "mode: atomic\n", true, 0, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			blocks, err := Parse(strings.NewReader(tt.profile))
			if (err != nil) != tt.wantErr {
				t.Fatalf("err=%v", err)
			}
			if err != nil {
				return
			}
			covered := 0
			for _, b := range blocks {
				if b.Covered {
					covered += b.Statements
				}
			}
			if len(blocks) != tt.blocks || covered != tt.covered {
				t.Fatalf("blocks=%v covered=%d", blocks, covered)
			}
		})
	}
}
