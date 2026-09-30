package main

import (
	"strings"
	"testing"

	"github.com/imajinyun/knifer-go/bin/internal/toolchainpolicy"
)

func TestCheckToolchainWorkflow(t *testing.T) {
	const base = `env:
  GO_RELEASE_VERSION: "1.27.1"
  GOLANGCI_LINT_VERSION: v2.14.0
  GOTOOLCHAIN: local
jobs:
  agent-governance:
    steps:
      - uses: actions/setup-go@v7
        with:
          go-version: ${{ env.GO_RELEASE_VERSION }}
      - run: |
          go version
          go env GOTOOLCHAIN
      - run: make ci-agent-governance
  test:
    strategy:
      matrix:
        go-version: ["1.26.8", "1.27.1"]
    steps:
      - uses: actions/setup-go@v7
        with:
          go-version: ${{ matrix.go-version }}
      - run: |
          go version
          go env GOTOOLCHAIN
      - run: make ci-test
`
	for _, tt := range []struct {
		name, old, replacement, rule string
	}{
		{name: "valid"},
		{name: "minimum_missing", old: `["1.26.8", "1.27.1"]`, replacement: `["1.27.1"]`, rule: "CI_WORKFLOW_VERSION_DRIFT"},
		{name: "release_drift", old: `GO_RELEASE_VERSION: "1.27.1"`, replacement: `GO_RELEASE_VERSION: "1.26.8"`, rule: "CI_WORKFLOW_VERSION_DRIFT"},
		{name: "auto_switch", old: "GOTOOLCHAIN: local", replacement: "GOTOOLCHAIN: auto", rule: "CI_WORKFLOW_AUTO_TOOLCHAIN"},
		{name: "setup_missing", old: "actions/setup-go@v7", replacement: "actions/checkout@v7", rule: "CI_WORKFLOW_GO_SETUP_MISSING"},
		{name: "wrong_matrix_reference", old: "${{ matrix.go-version }}", replacement: "${{ env.GO_RELEASE_VERSION }}", rule: "CI_WORKFLOW_VERSION_DRIFT"},
		{name: "unrecorded_go", old: "go version", replacement: "echo unrecorded", rule: "CI_WORKFLOW_GO_SETUP_MISSING"},
		{name: "conditional_setup", old: "- uses: actions/setup-go@v7", replacement: "- uses: actions/setup-go@v7\n        if: false", rule: "CI_WORKFLOW_VERSION_DRIFT"},
		{name: "step_override", old: "- run: make ci-test", replacement: "- run: make ci-test\n        env:\n          GOTOOLCHAIN: auto", rule: "CI_WORKFLOW_AUTO_TOOLCHAIN"},
		{name: "execution_before_setup", old: "steps:\n", replacement: "steps:\n      - run: make ci-test\n", rule: "CI_WORKFLOW_GO_SETUP_MISSING"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			text := base
			if tt.old != "" {
				text = strings.ReplaceAll(text, tt.old, tt.replacement)
			}
			c := &checker{}
			p := toolchainpolicy.Policy{Minimum: "1.26.0", Release: "1.27.1", Test: []string{"1.26.8", "1.27.1"}, Lint: "v2.14.0"}
			c.checkToolchainWorkflow(text, "fixture.yml", p)
			if tt.rule == "" && len(c.findings) > 0 {
				t.Fatalf("unexpected findings: %+v", c.findings)
			}
			if tt.rule != "" {
				for _, f := range c.findings {
					if f.RuleID == tt.rule {
						return
					}
				}
				t.Fatalf("missing %s: %+v", tt.rule, c.findings)
			}
		})
	}
}
