package coverageprofile

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/imajinyun/knifer-go/bin/internal/cievidence"
	"github.com/imajinyun/knifer-go/bin/internal/sourceidentity"
)

func TestValidateEvidence(t *testing.T) {
	for _, name := range []string{"valid", "missing", "failed_tests", "partial_scope", "wrong_profile", "changed_source", "changed_dependencies", "wrong_commit", "old_run", "no_toolchain"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			out := filepath.Join(t.TempDir(), "coverage.out")
			for file, data := range map[string]string{"go.mod": "module example.com/fixture\ngo 1.26.0\n", "go.sum": "", "code.go": "package fixture\n"} {
				if err := os.WriteFile(filepath.Join(root, file), []byte(data), 0o600); err != nil {
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
			t.Setenv("GITHUB_RUN_ID", "")
			t.Setenv("GITHUB_RUN_ATTEMPT", "")
			if err := os.WriteFile(out, []byte("mode: atomic\nexample.com/fixture/code.go:1.1,2.1 1 1\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			source, err := sourceidentity.Capture(root)
			if err != nil {
				t.Fatal(err)
			}
			hash, err := HashFile(out)
			if err != nil {
				t.Fatal(err)
			}
			e := Evidence{Schema: 1, Before: source, After: source, ProfileSHA256: hash, GoVersion: "go1.27.1", Command: []string{"go", "test", "-race", "-coverpkg=./..."}, Packages: []string{"./..."}, CreatedAt: time.Now().UTC(), Complete: true}
			switch name {
			case "failed_tests":
				e.TestsExitCode = 1
			case "partial_scope":
				e.Packages = []string{"./one"}
			case "wrong_profile":
				e.ProfileSHA256 = "wrong"
			case "changed_source":
				if err := os.WriteFile(filepath.Join(root, "code.go"), []byte("package fixture\n//changed\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			case "changed_dependencies":
				if err := os.WriteFile(filepath.Join(root, "go.sum"), []byte("changed"), 0o600); err != nil {
					t.Fatal(err)
				}
			case "wrong_commit":
				e.Before.Commit = "other"
			case "old_run":
				t.Setenv("GITHUB_RUN_ID", "2")
				e.RunID = "1"
			case "no_toolchain":
				e.GoVersion = ""
			}
			if name != "missing" {
				if err := cievidence.WriteJSON(out+".meta.json", e); err != nil {
					t.Fatal(err)
				}
			}
			_, err = ValidateEvidence(root, out)
			if (err == nil) != (name == "valid") {
				t.Fatalf("validation=%v", err)
			}
		})
	}
}
