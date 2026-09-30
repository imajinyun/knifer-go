package main

import (
	"os"
	"strings"
	"testing"
)

func TestCheckAdmissionWorkflow(t *testing.T) {
	data, err := os.ReadFile("../../.github/workflows/go.yml")
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct{ name, old, replacement string }{
		{name: "valid"},
		{name: "skips_on_failure", old: "  admission:\n    if: always()", replacement: "  admission:\n    if: success()"},
		{name: "missing_dependency", old: "fuzz-smoke, codeql, scorecard]", replacement: "fuzz-smoke, scorecard]"},
		{name: "missing_upload", old: "name: ci-result-${{ github.run_attempt }}-govulncheck", replacement: "name: unused"},
		{name: "unwrapped_command", old: ".aiflow/ci/ciresult run -id lint -out .aiflow/ci/records -- make lint", replacement: "make lint"},
		{name: "wrong_context", old: "${{ toJSON(needs) }}", replacement: "'{}'"},
		{name: "non_strict", old: "run: make ci-admission-check", replacement: "run: make agent-evidence-check"},
		{name: "ignore_failure", old: "run: make ci-admission-check", replacement: "continue-on-error: true\n        run: make ci-admission-check"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			text := string(data)
			if tt.old != "" {
				if !strings.Contains(text, tt.old) {
					t.Fatalf("fixture pattern missing: %s", tt.old)
				}
				text = strings.Replace(text, tt.old, tt.replacement, 1)
			}
			c := &checker{}
			c.checkAdmissionWorkflow(text, "go.yml")
			if (len(c.findings) == 0) != (tt.name == "valid") {
				t.Fatalf("findings=%+v", c.findings)
			}
		})
	}
}
