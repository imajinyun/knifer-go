package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFullCheckIncludesToolchainAdmission(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(repoRoot(t), "Makefile"))
	if err != nil {
		t.Fatal(err)
	}
	var dependencies []string
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "full-check:") {
			dependencies = strings.Fields(strings.TrimPrefix(line, "full-check:"))
		}
	}
	for _, required := range []string{"toolchain-check", "ai-context-check", "ci-workflow-check", "governance-maturity-check"} {
		t.Run(required, func(t *testing.T) {
			for _, dependency := range dependencies {
				if dependency == required {
					return
				}
			}
			t.Fatalf("full-check is missing required admission gate %s", required)
		})
	}
}
