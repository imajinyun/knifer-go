package main

import (
	"strings"
	"testing"
)

func TestCheckReleaseLint(t *testing.T) {
	const install = "go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${{ env.GOLANGCI_LINT_VERSION }}"
	for _, tt := range []struct {
		name    string
		steps   string
		version string
		wantErr bool
	}{
		{name: "pinned_before_gate", steps: install + "\n      - run: make release-check", version: "v2.12.2"},
		{name: "missing_install", steps: "make release-check", version: "v2.12.2", wantErr: true},
		{name: "late_install", steps: "make release-check\n      - run: " + install, version: "v2.12.2", wantErr: true},
		{name: "version_drift", steps: install + "\n      - run: make release-check", version: "v2.14.0", wantErr: true},
		{name: "unpinned_install", steps: strings.ReplaceAll(install, "${{ env.GOLANGCI_LINT_VERSION }}", "latest") + "\n      - run: make release-check", version: "v2.12.2", wantErr: true},
		{name: "comment_is_not_install", steps: "|\n          # " + install + "\n          make release-check", version: "v2.12.2", wantErr: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			workflow := "env:\n  GOLANGCI_LINT_VERSION: " + tt.version + "\njobs:\n  release:\n    steps:\n      - run: " + tt.steps + "\n"
			c := &checker{}
			c.checkReleaseLint(workflow, ".github/workflows/release.yml", "v2.12.2")
			if got := len(c.findings) > 0; got != tt.wantErr {
				t.Fatalf("has findings = %v, want %v: %+v", got, tt.wantErr, c.findings)
			}
		})
	}
}
