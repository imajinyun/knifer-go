package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCIGovernanceCoversAttestedChecks(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(repoRoot(t), "Makefile"))
	if err != nil {
		t.Fatal(err)
	}
	var dependencies []string
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "ci-governance-check:") {
			dependencies = strings.Fields(strings.TrimPrefix(line, "ci-governance-check:"))
		}
	}
	for _, required := range []string{"worktree-check", "aiflow-layout-check", "toolchain-check", "mod-verify", "go-module-cache-check", "governance-maturity-check", "ci-workflow-check", "change-policy-check", "api-check", "docs-check", "diff-whitespace"} {
		t.Run(required, func(t *testing.T) {
			for _, dep := range dependencies {
				if dep == required {
					return
				}
			}
			t.Fatalf("CI governance must execute attested check %s", required)
		})
	}
}
