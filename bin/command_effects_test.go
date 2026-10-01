package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRuntimeArtifactEffects(t *testing.T) {
	for _, name := range []string{"valid", "source_path", "escape", "symlink_escape", "not_ignored", "missing_declaration", "inherited_release"} {
		t.Run(name, func(t *testing.T) {
			fixture := newGovernanceFixture(t)
			cmd := exec.Command("git", "init", "-q", fixture.Root())
			if err := cmd.Run(); err != nil {
				t.Fatal(err)
			}
			fixture.WriteFile(".gitignore", ".aiflow/\n")
			spec := map[string]any{"writes_workspace": true, "workspace_write_scope": "runtime_artifacts", "writes_git_config": false, "writes_files": []string{".aiflow/fuzz/**"}, "creates_artifacts": []string{".aiflow/fuzz/"}}
			commands := map[string]any{"fuzz_smoke": spec}
			switch name {
			case "source_path":
				spec["writes_files"] = []string{"internal/fuzz/**"}
			case "escape":
				spec["writes_files"] = []string{".aiflow/../source.go"}
			case "symlink_escape":
				if err := os.Symlink(t.TempDir(), filepath.Join(fixture.Root(), ".aiflow")); err != nil {
					t.Fatal(err)
				}
			case "not_ignored":
				fixture.WriteFile(".gitignore", "")
			case "missing_declaration":
				delete(spec, "writes_files")
			case "inherited_release":
				commands["release_check"] = map[string]any{"writes_workspace": false}
			}
			fixture.WriteJSON("ai-context.json", map[string]any{"commands": commands})
			cmd = exec.Command("python3", "-B", filepath.Join(repoRoot(t), "bin/command_effects.py"), fixture.Root())
			cmd.Env = append(os.Environ(), "PYTHONDONTWRITEBYTECODE=1")
			output, err := cmd.CombinedOutput()
			if (err == nil) != (name == "valid") {
				t.Fatalf("error=%v output=%s", err, output)
			}
		})
	}
}
