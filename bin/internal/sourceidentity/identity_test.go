package sourceidentity

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCaptureDeletionAndNestedModule(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "integration/db")
	if err := os.MkdirAll(nested, 0o700); err != nil {
		t.Fatal(err)
	}
	for name, value := range map[string]string{"go.mod": "module fixture\ngo 1.26.0\n", "marker": "<deleted>", "integration/db/go.mod": "module fixture/integration\ngo 1.26.0\n"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(value), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{{"init", "-q"}, {"add", "."}, {"-c", "user.name=Fixture", "-c", "user.email=fixture@example.com", "-c", "commit.gpgsign=false", "-c", "core.hooksPath=/dev/null", "commit", "-qm", "fixture"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git: %v %s", err, out)
		}
	}
	before, err := Capture(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, "marker")); err != nil {
		t.Fatal(err)
	}
	after, err := Capture(root)
	if err != nil {
		t.Fatal(err)
	}
	if before.SourceSHA256 == after.SourceSHA256 || before.ModulesSHA256 != after.ModulesSHA256 {
		t.Fatal("file deletion identity is incorrect")
	}
	if err := os.WriteFile(filepath.Join(nested, "go.mod"), []byte("module fixture/integration\ngo 1.27.0\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	next, err := Capture(root)
	if err != nil {
		t.Fatal(err)
	}
	if next.ModulesSHA256 == after.ModulesSHA256 {
		t.Fatal("nested module change was not fingerprinted")
	}
}
