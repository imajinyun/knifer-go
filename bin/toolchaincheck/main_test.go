package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCheck(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fixture executables use POSIX shell")
	}
	for _, tt := range []struct{ name, required, actual, mode, tool, want string }{
		{name: "valid", required: "1.26.0", actual: "go1.26.8", mode: "local", tool: "2.14.0 built with go1.27.1"},
		{name: "new_dependency", required: "1.27.0", actual: "go1.27.1", mode: "local", want: "dependency example.com/dep requires Go 1.27.0"},
		{name: "auto_switch", required: "1.26.0", actual: "go1.27.1", mode: "auto", want: "GOTOOLCHAIN=local"},
		{name: "old_linter", required: "1.26.0", actual: "go1.26.8", mode: "local", tool: "2.14.0 built with go1.25.13", want: "cannot analyze"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			write := func(name, text string, mode os.FileMode) string {
				t.Helper()
				path := filepath.Join(root, name)
				if err := os.WriteFile(path, []byte(text), mode); err != nil {
					t.Fatal(err)
				}
				return path
			}
			write("go.mod", "module example.com/fixture\ngo 1.26.0\n", 0o600)
			write("ai-context.json", `{"project":{"go_version":">=1.26"},"ci_workflows":{"tool_versions":{"go_minimum":"1.26.0","go_release":"1.27.1","go_test":["1.26.8","1.27.1"],"golangci_lint":"v2.14.0"}}}`, 0o600)
			goCmd := write("fake-go", "#!/bin/sh\nif [ \"$1\" = env ]; then\n  echo '{\"GOVERSION\":\""+tt.actual+"\",\"GOTOOLCHAIN\":\""+tt.mode+"\"}'\nelse\n  echo '{\"Path\":\"example.com/dep\",\"GoVersion\":\""+tt.required+"\"}'\nfi\n", 0o700)
			lint := ""
			if tt.tool != "" {
				lint = write("fake-lint", "#!/bin/sh\necho 'golangci-lint has version "+tt.tool+" from fixture'\n", 0o700)
			}
			err := check(root, goCmd, lint)
			if tt.want == "" && err != nil {
				t.Fatal(err)
			}
			if tt.want != "" && (err == nil || !strings.Contains(err.Error(), tt.want)) {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
		})
	}
}
