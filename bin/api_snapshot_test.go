package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateSnapshotIncludesCompatibilityDetails(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "go.mod", "module github.com/imajinyun/knifer-go\n\ngo 1.25.0\n")
	writeTestFile(t, root, "vcompat/compat.go", `package vcompat

const DefaultName = "demo"

var DefaultResult Result

type Alias = Result

type Result struct {
	Name   string
	Count  int
	hidden bool
}

func Build(input string) (Result, error) {
	return Result{Name: input}, nil
}

func (r Result) String() string { return r.Name }

func (r *Result) SetName(name string) { r.Name = name }

type Validator interface {
	Validate(Result) error
}
`)
	writeTestFile(t, root, "internal/hidden/hidden.go", `package hidden

func Hidden() {}
`)
	writeTestFile(t, root, "notfacade/notfacade.go", `package notfacade

func Hidden() {}
`)

	lines, err := generateSnapshot(root)
	if err != nil {
		t.Fatalf("generateSnapshot() error = %v", err)
	}
	snapshot := strings.Join(lines, "\n")
	for _, want := range []string{
		"github.com/imajinyun/knifer-go/vcompat",
		`const DefaultName untyped string = "demo"`,
		"var DefaultResult Result",
		"func Build(input string) (Result, error)",
		"type Alias = Result",
		"type Result struct{ Count int; Name string }",
		"method (*Result) SetName(name string)",
		"method (Result) String() string",
		"type Validator interface{ Validate(Result) error }",
	} {
		if !strings.Contains(snapshot, want) {
			t.Fatalf("snapshot missing %q:\n%s", want, snapshot)
		}
	}
	for _, unwanted := range []string{"hidden bool", "internal/hidden", "notfacade"} {
		if strings.Contains(snapshot, unwanted) {
			t.Fatalf("snapshot unexpectedly contains %q:\n%s", unwanted, snapshot)
		}
	}
}

func TestGenerateSnapshotIncludesGenericAndEmbeddedAPIDetails(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "go.mod", "module github.com/imajinyun/knifer-go\n\ngo 1.25.0\n")
	writeTestFile(t, root, "vcompat/compat.go", `package vcompat

import "time"

const Enabled = true

type private struct {
	Value string
}

type Number interface {
	~int | ~int64
}

type DurationMap = map[string]time.Duration

type Box[T any] struct {
	Item T
	time.Time
	private
	hidden string
}

type Handler[T Number] interface {
	Handle(Box[T], ...time.Duration) (private, error)
	reset()
}

func Transform[T Number](box Box[T], labels ...string) (out Box[T], err error) {
	return box, nil
}

func (b Box[T]) Value() T { return b.Item }

func (b *Box[T]) Set(value T) *Box[T] {
	b.Item = value
	return b
}
`)

	lines, err := generateSnapshot(root)
	if err != nil {
		t.Fatalf("generateSnapshot() error = %v", err)
	}
	snapshot := strings.Join(lines, "\n")
	for _, want := range []string{
		"const Enabled untyped bool = true",
		"func Transform[T Number](box Box[T], labels ...string) (out Box[T], err error)",
		"method (*Box[T]) Set(value T) *Box[T]",
		"method (Box[T]) Value() T",
		"type Box[T any] struct{ Item T; time.Time time.Time }",
		"type DurationMap = map[string]time.Duration",
		"type Handler[T Number] interface{ Handle(Box[T], ...time.Duration) (private, error) }",
		"type Number interface{}",
	} {
		if !strings.Contains(snapshot, want) {
			t.Fatalf("snapshot missing %q:\n%s", want, snapshot)
		}
	}
	for _, unwanted := range []string{"hidden string", "reset()", "type private"} {
		if strings.Contains(snapshot, unwanted) {
			t.Fatalf("snapshot unexpectedly contains %q:\n%s", unwanted, snapshot)
		}
	}
}

func writeTestFile(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

func TestCheckReleaseNotesAcceptsVersionedRelease(t *testing.T) {
	changelog := writeGovernanceFixture(t, `# Changelog

## Unreleased

## 1.2.3

### Governance

- Added a release-note gate.
`)

	output, err := runReleaseNotesCheck(t, changelog, "1.2.3")
	if err != nil {
		t.Fatalf("check_release_notes.sh failed: %v\n%s", err, output)
	}
	if !strings.Contains(output, "release notes are valid for 1.2.3") {
		t.Fatalf("unexpected output:\n%s", output)
	}
}

func TestCheckReleaseNotesRejectsUnmovedUnreleasedEntries(t *testing.T) {
	changelog := writeGovernanceFixture(t, `# Changelog

## Unreleased

### Governance

- This still needs to be moved.

## 1.2.3

### Governance

- Added a release-note gate.
`)

	output, err := runReleaseNotesCheck(t, changelog, "1.2.3")
	if err == nil {
		t.Fatalf("check_release_notes.sh unexpectedly succeeded:\n%s", output)
	}
	if !strings.Contains(output, "still has entries under Unreleased") {
		t.Fatalf("expected unmoved Unreleased error, got:\n%s", output)
	}
}

func TestCheckReleaseNotesAcceptsStructureWithoutReleaseVersion(t *testing.T) {
	changelog := writeGovernanceFixture(t, `# Changelog

## Unreleased

### Governance

- Work in progress is allowed before a release version is selected.
`)

	output, err := runReleaseNotesCheck(t, changelog, "")
	if err != nil {
		t.Fatalf("check_release_notes.sh failed: %v\n%s", err, output)
	}
	if !strings.Contains(output, "release notes structure is valid") {
		t.Fatalf("unexpected output:\n%s", output)
	}
}

func runReleaseNotesCheck(t *testing.T, changelog, version string) (string, error) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	root := filepath.Dir(wd)
	args := []string{filepath.Join(root, "bin/check_release_notes.sh")}
	if version != "" {
		args = append(args, version)
	}
	cmd := exec.Command("bash", args...)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "CHANGELOG_FILE="+changelog)
	combined, err := cmd.CombinedOutput()
	return string(combined), err
}

func writeGovernanceFixture(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "CHANGELOG.md")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return path
}

func TestAgentEvidenceCheckAcceptsSecurityMergeReadyEvidence(t *testing.T) {
	evidence := baseAgentEvidence()
	evidence["command_attestations"].(map[string]any)["agent_full_check"] = map[string]any{
		"cmd":       "make agent-full-check COVERAGE_FILE=/tmp/knifer-go-coverage.out",
		"exit_code": 0,
		"source":    "agent_run",
		"status":    "passed",
	}
	evidence["command_attestations"].(map[string]any)["agent_security_check"] = map[string]any{
		"cmd":       "make agent-security-check",
		"exit_code": 0,
		"source":    "agent_run",
		"status":    "passed",
	}
	evidence["merge_ready"] = true
	evidence["merge_blockers"] = []string{}

	output, err := runAgentEvidenceCheck(t, evidence)
	if err != nil {
		t.Fatalf("agent evidence check failed: %v\n%s", err, output)
	}
	if !strings.Contains(output, "merge_ready=true") {
		t.Fatalf("agent evidence output missing merge_ready=true:\n%s", output)
	}
}

func TestAgentEvidenceCheckRejectsIncorrectSecurityMergeReadyEvidence(t *testing.T) {
	evidence := baseAgentEvidence()
	evidence["merge_ready"] = true
	evidence["merge_blockers"] = []string{}

	output, err := runAgentEvidenceCheck(t, evidence)
	if err == nil {
		t.Fatalf("agent evidence check unexpectedly passed:\n%s", output)
	}
	if !strings.Contains(output, "merge_ready must be false") {
		t.Fatalf("agent evidence output missing merge_ready error:\n%s", output)
	}
}

func baseAgentEvidence() map[string]any {
	checks := map[string]any{
		"ai_context_check": map[string]any{
			"cmd":       "bash bin/check_ai_context.sh",
			"exit_code": 0,
			"status":    "passed",
			"stdout":    "ai-context.json is valid",
			"stderr":    "",
		},
		"change_policy_check": map[string]any{
			"cmd":       "bash bin/check_change_policy.sh",
			"exit_code": 0,
			"status":    "passed",
			"stdout":    "change policy check passed",
			"stderr":    "",
		},
		"security_sensitive_diff": map[string]any{
			"cmd":       "bash bin/check_security_sensitive_diff.sh",
			"exit_code": 1,
			"status":    "failed",
			"stdout":    "",
			"stderr":    "SECURITY DIFF CHECK ERROR: security-sensitive files changed:\n  - internal/db/scan.go\nRun make agent-security-check and document security review evidence before merging.",
		},
	}
	attestations := map[string]any{
		"ai_context_check": map[string]any{
			"cmd":       "bash bin/check_ai_context.sh",
			"exit_code": 0,
			"source":    "embedded_check",
			"status":    "passed",
		},
		"change_policy_check": map[string]any{
			"cmd":       "bash bin/check_change_policy.sh",
			"exit_code": 0,
			"source":    "embedded_check",
			"status":    "passed",
		},
		"security_sensitive_diff": map[string]any{
			"cmd":       "bash bin/check_security_sensitive_diff.sh",
			"exit_code": 1,
			"source":    "embedded_check",
			"status":    "failed",
		},
		"agent_evidence": map[string]any{
			"cmd":       "make agent-evidence",
			"exit_code": 0,
			"source":    "current_process",
			"status":    "passed",
		},
		"agent_evidence_check": map[string]any{
			"cmd":    "make agent-evidence-check",
			"reason": "validated by make agent-evidence-check after evidence generation",
			"source": "post_generation",
			"status": "pending",
		},
		"agent_full_check": map[string]any{
			"cmd":    "make agent-full-check COVERAGE_FILE=/tmp/knifer-go-coverage.out",
			"reason": "required command has not been attested in this evidence",
			"source": "required_by_policy",
			"status": "not_recorded",
		},
		"agent_security_check": map[string]any{
			"cmd":    "make agent-security-check",
			"reason": "required command has not been attested in this evidence",
			"source": "required_by_policy",
			"status": "not_recorded",
		},
	}
	return map[string]any{
		"branch":                        "test",
		"changed_files":                 []string{"internal/db/scan.go"},
		"checks":                        checks,
		"command_attestations":          attestations,
		"commit":                        "0000000000000000000000000000000000000000",
		"detected_change_policies":      []string{"security_sensitive"},
		"diff_filter":                   "ACDMRTUXB",
		"generated_at":                  "2026-07-04T00:00:00Z",
		"highest_required_command_risk": "medium",
		"merge_blockers":                []string{"agent_full_check", "agent_security_check"},
		"merge_ready":                   false,
		"module":                        "github.com/imajinyun/knifer-go",
		"repository":                    "knifer-go",
		"required_commands":             []string{"change_policy_check", "security_sensitive_diff", "agent_full_check", "agent_security_check", "agent_evidence", "agent_evidence_check"},
		"schema_version":                "1.0",
		"security_sensitive_paths":      []string{"internal/db/scan.go"},
		"worktree_status":               " M internal/db/scan.go",
	}
}

func runAgentEvidenceCheck(t *testing.T, evidence map[string]any) (string, error) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	root := filepath.Dir(wd)
	path := filepath.Join(t.TempDir(), "agent-evidence.json")
	data, err := json.MarshalIndent(evidence, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent() error = %v", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	cmd := exec.Command("bash", filepath.Join(root, "bin/check_agent_evidence.sh"))
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "AGENT_EVIDENCE_FILE="+path)
	combined, err := cmd.CombinedOutput()
	return string(combined), err
}
