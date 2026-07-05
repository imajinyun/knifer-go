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

func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	return filepath.Dir(wd)
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

func TestCoverageCheckRequiresChangedSecuritySensitivePackageData(t *testing.T) {
	root := repoRoot(t)
	coverage := filepath.Join(t.TempDir(), "coverage.out")
	module := "github.com/imajinyun/knifer-go"
	if err := os.WriteFile(coverage, []byte("mode: set\n"+module+"/vurl/url.go:1.1,1.2 1 1\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	runCoverage := func() ([]byte, error) {
		cmd := exec.Command("bash", filepath.Join(root, "bin/check_coverage.sh"), coverage)
		cmd.Dir = root
		cmd.Env = append(os.Environ(),
			"COVERAGE_THRESHOLD=0",
			"PACKAGE_COVERAGE_THRESHOLDS= ",
			"SECURITY_SENSITIVE_COVERAGE_PATHS= ",
			"CHANGED_SECURITY_SENSITIVE_COVERAGE_PATHS="+module+"/vhttp",
			"SECURITY_SENSITIVE_MIN_COVERAGE_THRESHOLD=0",
		)
		return cmd.CombinedOutput()
	}
	output, err := runCoverage()
	if err == nil {
		t.Fatalf("check_coverage.sh unexpectedly passed:\n%s", output)
	}
	if !strings.Contains(string(output), "changed security-sensitive package(s) have no coverage data") {
		t.Fatalf("coverage check output missing changed package error:\n%s", output)
	}

	if err := os.WriteFile(coverage, []byte("mode: set\n"+module+"/vhttp/request.go:1.1,1.2 1 1\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	output, err = runCoverage()
	if err != nil {
		t.Fatalf("check_coverage.sh with changed package coverage failed: %v\n%s", err, output)
	}
	if !strings.Contains(string(output), "changed security-sensitive coverage data present for 1 package path") {
		t.Fatalf("coverage check output missing changed package success:\n%s", output)
	}
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
	evidence["security_review"] = readySecurityReviewEvidence()
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

func TestAgentEvidenceCheckRejectsMissingSecurityReviewEvidence(t *testing.T) {
	evidence := baseAgentEvidence()
	delete(evidence, "security_review")

	output, err := runAgentEvidenceCheck(t, evidence)
	if err == nil {
		t.Fatalf("agent evidence check unexpectedly passed:\n%s", output)
	}
	if !strings.Contains(output, "security_review must be an object") {
		t.Fatalf("agent evidence output missing security_review error:\n%s", output)
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

func TestAPIFreezeCheckRequiresStatusDecisionCards(t *testing.T) {
	root := t.TempDir()
	toolsPath := filepath.Join(root, "tools.json")
	writeJSONFile(t, toolsPath, map[string]any{
		"packages": []any{
			map[string]any{
				"name": "vcompat",
				"functions": []any{
					map[string]any{
						"name":     "Copy",
						"status":   "compatibility",
						"synopsis": "Copy is a compatibility alias.",
					},
				},
			},
		},
	})
	context := minimalAPIFreezeContext()
	context["api_freeze"].(map[string]any)["api_status_decision_cards"] = map[string]any{
		"recommended":   []string{"recommended-card"},
		"compatibility": []string{"recommended-card"},
		"experimental":  []string{"experimental-card"},
		"deprecated":    []string{"deprecated-card"},
	}
	contextPath := filepath.Join(root, "ai-context.json")
	writeJSONFile(t, contextPath, context)

	output, err := runAPIFreezeCheck(t, contextPath, toolsPath)
	if err == nil {
		t.Fatalf("api freeze check unexpectedly passed:\n%s", output)
	}
	if !strings.Contains(output, "api_freeze.api_status_decision_cards.compatibility references recommended-card with status 'recommended'") {
		t.Fatalf("api freeze output missing decision-card status error:\n%s", output)
	}

	context["api_freeze"].(map[string]any)["api_status_decision_cards"] = map[string]any{
		"recommended":   []string{"recommended-card"},
		"compatibility": []string{"compatibility-card"},
		"experimental":  []string{"experimental-card"},
		"deprecated":    []string{"deprecated-card"},
	}
	writeJSONFile(t, contextPath, context)
	output, err = runAPIFreezeCheck(t, contextPath, toolsPath)
	if err != nil {
		t.Fatalf("api freeze check failed: %v\n%s", err, output)
	}
	if !strings.Contains(output, "api freeze metadata is valid") {
		t.Fatalf("api freeze output missing success message:\n%s", output)
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
		"security_review": map[string]any{
			"audit_conclusion": "Security-sensitive change is blocked until agent_full_check and agent_security_check are attested.",
			"command_attestations": map[string]any{
				"agent_full_check":     attestations["agent_full_check"],
				"agent_security_check": attestations["agent_security_check"],
			},
			"paths":                    []string{"internal/db/scan.go"},
			"required":                 true,
			"required_commands":        []string{"agent_full_check", "agent_security_check"},
			"security_review_required": true,
			"status":                   "blocked",
		},
		"security_sensitive_paths": []string{"internal/db/scan.go"},
		"worktree_status":          " M internal/db/scan.go",
	}
}

func readySecurityReviewEvidence() map[string]any {
	return map[string]any{
		"audit_conclusion": "Security-sensitive change has full and security validation attestations.",
		"command_attestations": map[string]any{
			"agent_full_check": map[string]any{
				"cmd":       "make agent-full-check COVERAGE_FILE=/tmp/knifer-go-coverage.out",
				"exit_code": 0,
				"source":    "agent_run",
				"status":    "passed",
			},
			"agent_security_check": map[string]any{
				"cmd":       "make agent-security-check",
				"exit_code": 0,
				"source":    "agent_run",
				"status":    "passed",
			},
		},
		"paths":                    []string{"internal/db/scan.go"},
		"required":                 true,
		"required_commands":        []string{"agent_full_check", "agent_security_check"},
		"security_review_required": true,
		"status":                   "ready",
	}
}

func minimalAPIFreezeContext() map[string]any {
	return map[string]any{
		"public_facades": []any{
			map[string]any{"package": "vcompat"},
		},
		"api_freeze": map[string]any{
			"v1_candidate":                         true,
			"decision_card_required":               true,
			"replacement_required_for_deprecation": true,
			"allowed_statuses":                     []string{"recommended", "compatibility", "experimental", "deprecated"},
			"decision_cards": []any{
				apiDecisionCard("recommended-card", "vcompat", "recommended"),
				apiDecisionCard("compatibility-card", "vcompat", "compatibility"),
				apiDecisionCard("experimental-card", "vcompat", "experimental"),
				apiDecisionCard("deprecated-card", "vcompat", "deprecated"),
				apiDecisionCard("v1-public-api-entry-budget", "all", "recommended"),
				apiDecisionCard("v1-dynamic-contract-matrix", "vcompat", "recommended"),
				apiDecisionCard("v1-heavy-dependency-isolation", "vcompat", "recommended"),
				apiDecisionCard("v1-error-taxonomy", "all", "recommended"),
				apiDecisionCard("v1-security-threat-model", "vcompat", "recommended"),
			},
			"api_status_decision_cards": map[string]any{
				"recommended":   []string{"recommended-card"},
				"compatibility": []string{"compatibility-card"},
				"experimental":  []string{"experimental-card"},
				"deprecated":    []string{"deprecated-card"},
			},
			"freeze_checks": []string{
				"Every public API requires a decision card.",
				"Every deprecated API requires a replacement.",
				"Snapshot files must stay current.",
				"The tools catalog must classify API status.",
			},
			"deprecations": []any{},
		},
	}
}

func apiDecisionCard(id, pkg, status string) map[string]any {
	return map[string]any{
		"id":         id,
		"packages":   []string{pkg},
		"status":     status,
		"decision":   id + " decision",
		"rationale":  id + " rationale",
		"validation": []string{"make api-freeze-check", "make tools-check"},
	}
}

func runAPIFreezeCheck(t *testing.T, contextPath, toolsPath string) (string, error) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	root := filepath.Dir(wd)
	cmd := exec.Command("bash", filepath.Join(root, "bin/check_api_freeze.sh"))
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "AI_CONTEXT_FILE="+contextPath, "TOOLS_JSON_FILE="+toolsPath)
	combined, err := cmd.CombinedOutput()
	return string(combined), err
}

func writeJSONFile(t *testing.T, path string, value any) {
	t.Helper()
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent() error = %v", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
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
