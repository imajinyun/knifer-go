package toolchainpolicy

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	for _, tt := range []struct {
		name, moduleGo, projectGo string
		wantErr                   bool
	}{
		{"valid", "1.26.0", ">=1.26", false},
		{"module_raised", "1.27.0", ">=1.26", true},
		{"stale_metadata", "1.26.0", ">=1.25", true},
		{"missing_directive", "", ">=1.26", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			module := "module example.com/fixture\n"
			if tt.moduleGo != "" {
				module += "go " + tt.moduleGo + "\n"
			}
			if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte(module), 0o600); err != nil {
				t.Fatal(err)
			}
			data, err := json.Marshal(map[string]any{"project": map[string]string{"go_version": tt.projectGo}, "ci_workflows": map[string]any{"tool_versions": validPolicy()}})
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "ai-context.json"), data, 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := Load(root); (err != nil) != tt.wantErr {
				t.Fatalf("Load() = %v, want error %v", err, tt.wantErr)
			}
		})
	}
}
