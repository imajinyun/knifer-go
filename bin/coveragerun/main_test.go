package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/imajinyun/knifer-go/bin/internal/coverageprofile"
)

func TestCoverageExecutionEvidence(t *testing.T) {
	for _, failure := range []bool{false, true} {
		name := "passing"
		want := "1"
		if failure {
			name = "failing"
			want = "2"
		}
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			for path, data := range map[string]string{".gitignore": ".aiflow/\n", "go.mod": "module example.com/profilefixture\ngo 1.26.0\n", "value.go": "package profilefixture\nfunc Value() int { return 1 }\n", "value_test.go": "package profilefixture\nimport \"testing\"\nfunc TestValue(t *testing.T) { if Value() != " + want + " { t.Fatal(\"intentional fixture failure\") } }\n"} {
				if err := os.WriteFile(filepath.Join(root, path), []byte(data), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			for _, args := range [][]string{{"init", "-q"}, {"add", "."}, {"-c", "user.name=Fixture", "-c", "user.email=fixture@example.com", "-c", "commit.gpgsign=false", "-c", "core.hooksPath=/dev/null", "commit", "-qm", "fixture"}} {
				cmd := exec.Command("git", args...)
				cmd.Dir = root
				if data, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("git: %v %s", err, data)
				}
			}
			out := filepath.Join(root, ".aiflow/coverage.out")
			err := run(context.Background(), root, "go", out, []string{"./..."})
			if (err != nil) != failure {
				t.Fatalf("run error=%v want failure=%v", err, failure)
			}
			_, err = coverageprofile.ValidateEvidence(root, out)
			if (err != nil) != failure {
				t.Fatalf("evidence error=%v want failure=%v", err, failure)
			}
		})
	}
}
