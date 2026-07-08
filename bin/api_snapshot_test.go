package main

import (
	"encoding/json"
	"os"
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

func TestReleaseCheckEnforcesFullPackageCoverageMode(t *testing.T) {
	root := repoRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "Makefile"))
	if err != nil {
		t.Fatalf("ReadFile(Makefile) error = %v", err)
	}
	makefile := string(data)
	releaseCheckIndex := strings.Index(makefile, "\nrelease-check:")
	if releaseCheckIndex < 0 {
		t.Fatal("Makefile must define release-check target")
	}
	rest := makefile[releaseCheckIndex+1:]
	nextTargetIndex := strings.Index(rest[len("release-check:"):], "\nagent-check:")
	if nextTargetIndex < 0 {
		t.Fatal("release-check target must appear before agent-check target")
	}
	releaseCheck := rest[:len("release-check:")+nextTargetIndex]
	for _, want := range []string{
		"COVERAGE_CHECK_ALL_PACKAGES=1",
		"full-check",
		"COVERAGE_FILE=$(COVERAGE_FILE)",
	} {
		if !strings.Contains(releaseCheck, want) {
			t.Fatalf("release-check target must contain %q:\n%s", want, releaseCheck)
		}
	}
}

func TestCheckReleaseNotesRejectsMissingGovernanceTemplateFields(t *testing.T) {
	changelog := writeGovernanceFixture(t, `# Changelog

## Unreleased

`)
	template := writeGovernanceTemplateFixture(t, `# Adoption Trust

## Governance Validation Contracts

- Contract changed: make release-check
`)

	output, err := runReleaseNotesCheckWithTemplate(t, changelog, template, "")
	if err == nil {
		t.Fatalf("check_release_notes.sh unexpectedly succeeded:\n%s", output)
	}
	if !strings.Contains(output, "governance release summary template must include 'User impact:'") {
		t.Fatalf("expected governance template field error, got:\n%s", output)
	}
}

func runReleaseNotesCheck(t *testing.T, changelog, version string) (string, error) {
	t.Helper()
	fixture := newGovernanceFixture(t)
	template := fixture.WriteTempFile("adoption-trust.md", `# Adoption Trust

## Governance Validation Contracts

- Contract changed: make release-check
- User impact: release maintainers must include validation evidence.
- Required action: run make release-check.
- Validation evidence: attach command output.
- Compatibility note: public APIs are unchanged.
`)
	return fixture.RunReleaseNotesCheck(changelog, template, version)
}

func runReleaseNotesCheckWithTemplate(t *testing.T, changelog, template, version string) (string, error) {
	t.Helper()
	return newGovernanceFixture(t).RunReleaseNotesCheck(changelog, template, version)
}

func writeGovernanceFixture(t *testing.T, content string) string {
	t.Helper()
	return newGovernanceFixture(t).WriteTempFile("CHANGELOG.md", content)
}

func writeGovernanceTemplateFixture(t *testing.T, content string) string {
	t.Helper()
	return newGovernanceFixture(t).WriteTempFile("adoption-trust.md", content)
}

func TestCoverageCheckRequiresChangedSecuritySensitivePackageData(t *testing.T) {
	fixture := newGovernanceFixture(t)
	module := "github.com/imajinyun/knifer-go"
	coverage := fixture.WriteTempFile("coverage.out", "mode: set\n"+module+"/vurl/url.go:1.1,1.2 1 1\n")

	runCoverage := func() (string, error) {
		return fixture.RunCoverageCheck(
			coverage,
			"COVERAGE_THRESHOLD=0",
			"PACKAGE_COVERAGE_THRESHOLDS= ",
			"CHANGED_PACKAGE_COVERAGE_THRESHOLDS= ",
			"SECURITY_SENSITIVE_COVERAGE_PATHS= ",
			"CHANGED_SECURITY_SENSITIVE_COVERAGE_PATHS="+module+"/vhttp",
			"SECURITY_SENSITIVE_MIN_COVERAGE_THRESHOLD=0",
		)
	}
	output, err := runCoverage()
	if err == nil {
		t.Fatalf("check_coverage.sh unexpectedly passed:\n%s", output)
	}
	if !strings.Contains(output, "COVERAGE_CHANGED_SECURITY_SENSITIVE_MISSING") {
		t.Fatalf("coverage check output missing COVERAGE_CHANGED_SECURITY_SENSITIVE_MISSING rule id:\n%s", output)
	}
	if !strings.Contains(output, "changed security-sensitive package(s) have no coverage data") {
		t.Fatalf("coverage check output missing changed package error:\n%s", output)
	}

	fixture.WriteFile("coverage.out", "mode: set\n"+module+"/vhttp/request.go:1.1,1.2 1 1\n")
	output, err = runCoverage()
	if err != nil {
		t.Fatalf("check_coverage.sh with changed package coverage failed: %v\n%s", err, output)
	}
	if !strings.Contains(output, "changed security-sensitive coverage data present for 1 package path") {
		t.Fatalf("coverage check output missing changed package success:\n%s", output)
	}
}

func TestCoverageCheckEmitsJSONReport(t *testing.T) {
	fixture := newGovernanceFixture(t)
	module := "github.com/imajinyun/knifer-go"
	fixture.WriteJSON("ai-context.json", map[string]any{
		"project": map[string]any{
			"module": module,
		},
		"coverage_gates": map[string]any{
			"repository_threshold":             0,
			"security_sensitive_min_threshold": 0,
			"package_thresholds":               map[string]any{},
		},
		"public_facades":              []any{},
		"security_sensitive_packages": []any{},
	})
	coverage := fixture.WriteTempFile("coverage.out", "mode: set\n"+module+"/vhttp/request.go:1.1,1.2 1 1\n")

	output, err := fixture.RunCoverageCheckJSON(
		coverage,
		"COVERAGE_THRESHOLD=0",
		"PACKAGE_COVERAGE_THRESHOLDS= ",
		"CHANGED_PACKAGE_COVERAGE_THRESHOLDS= ",
		"SECURITY_SENSITIVE_COVERAGE_PATHS= ",
		"CHANGED_SECURITY_SENSITIVE_COVERAGE_PATHS="+module+"/vhttp",
		"SECURITY_SENSITIVE_MIN_COVERAGE_THRESHOLD=0",
	)
	if err != nil {
		t.Fatalf("coveragecheck -json failed: %v\n%s", err, output)
	}
	var result struct {
		Status             string            `json:"status"`
		Findings           []json.RawMessage `json:"findings"`
		RepositoryCoverage float64           `json:"repository_coverage"`
		RequiredCoverage   float64           `json:"required_coverage"`
	}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("coveragecheck -json output is not valid JSON: %v\n%s", err, output)
	}
	if result.Status != "passed" || len(result.Findings) != 0 {
		t.Fatalf("unexpected coverage JSON status/findings:\n%s", output)
	}
	if result.RepositoryCoverage != 100 || result.RequiredCoverage != 0 {
		t.Fatalf("unexpected coverage JSON coverage summary:\n%s", output)
	}
}

func TestChangePolicyClassifiesFacadeTestAndBenchmarkSeparatelyFromPublicAPI(t *testing.T) {
	output, err := newGovernanceFixture(t).RunChangePolicyCheck("vcodec/codec_benchmark_test.go\nvhttp/request_safe_constructors_test.go")
	if err != nil {
		t.Fatalf("check_change_policy.sh failed: %v\n%s", err, output)
	}
	if !strings.Contains(output, "detected policies: bug_fix") {
		t.Fatalf("change policy output missing bug_fix-only policy:\n%s", output)
	}
	if !strings.Contains(output, "rule ids: CHANGE_BUG_FIX") {
		t.Fatalf("change policy output missing CHANGE_BUG_FIX rule id:\n%s", output)
	}
	if strings.Contains(output, "public_api") {
		t.Fatalf("facade test/benchmark files must not be classified as public_api:\n%s", output)
	}
}

func TestChangePolicyClassifiesFacadeProductionAndSnapshotAsPublicAPI(t *testing.T) {
	output, err := newGovernanceFixture(t).RunChangePolicyCheck("vcodec/codec.go\ndocs/api/exports.txt")
	if err != nil {
		t.Fatalf("check_change_policy.sh failed: %v\n%s", err, output)
	}
	if !strings.Contains(output, "public_api") {
		t.Fatalf("change policy output missing public_api policy:\n%s", output)
	}
	if !strings.Contains(output, "CHANGE_PUBLIC_API") {
		t.Fatalf("change policy output missing CHANGE_PUBLIC_API rule id:\n%s", output)
	}
	if !strings.Contains(output, "public_api paths:") || !strings.Contains(output, "vcodec/codec.go") || !strings.Contains(output, "docs/api/exports.txt") {
		t.Fatalf("change policy output missing public_api paths:\n%s", output)
	}
}

func TestChangePolicyReportsSemanticRuleIDs(t *testing.T) {
	output, err := newGovernanceFixture(t).RunChangePolicyCheck("ai-context.schema.json\nMakefile\nbin/check_api_freeze.sh\nbin/check_coverage.sh")
	if err != nil {
		t.Fatalf("check_change_policy.sh failed: %v\n%s", err, output)
	}
	for _, want := range []string{
		"SEMANTIC_SCHEMA_CONTRACT_CHANGE",
		"SEMANTIC_RELEASE_GATE_CHANGE",
		"SEMANTIC_COVERAGE_POLICY_CHANGE",
		"SEMANTIC_API_FREEZE_POLICY_CHANGE",
		"SEMANTIC_SECURITY_POLICY_CHANGE",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("change policy output missing semantic rule id %s:\n%s", want, output)
		}
	}
}

func TestChangePolicyReportsFieldLevelSemanticRuleIDs(t *testing.T) {
	fixture := newGovernanceFixture(t)
	writeSemanticChangePolicyFixture(fixture)

	output, err := fixture.RunChangePolicyCheckWithDiff("ai-context.json\nMakefile", changePolicySemanticDiffFixture())
	if err != nil {
		t.Fatalf("check_change_policy.sh failed: %v\n%s", err, output)
	}
	for _, want := range []string{
		"SEMANTIC_AI_CONTEXT_API_FREEZE_CHANGE",
		"SEMANTIC_AI_CONTEXT_COVERAGE_GATES_CHANGE",
		"SEMANTIC_AI_CONTEXT_SECURITY_SENSITIVE_PACKAGES_CHANGE",
		"SEMANTIC_AI_CONTEXT_RANDOM_SOURCE_POLICY_CHANGE",
		"SEMANTIC_AI_CONTEXT_THREAT_MODEL_CHANGE",
		"SEMANTIC_MAKEFILE_RELEASE_CHECK_CHANGE",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("change policy output missing field-level semantic rule id %s:\n%s", want, output)
		}
	}
	for _, unwanted := range []string{
		"SEMANTIC_AI_CONTEXT_CHANGE",
		"SEMANTIC_MAKEFILE_CHANGE",
		"SEMANTIC_COVERAGE_POLICY_CHANGE",
		"SEMANTIC_API_FREEZE_POLICY_CHANGE",
		"SEMANTIC_SECURITY_POLICY_CHANGE",
	} {
		if strings.Contains(output, unwanted) {
			t.Fatalf("change policy output should prefer field-level semantic id over broad id %s:\n%s", unwanted, output)
		}
	}
}

func TestChangePolicyDoesNotOverclassifyUnrelatedAIContextDiff(t *testing.T) {
	fixture := newGovernanceFixture(t)
	writeSemanticChangePolicyFixture(fixture)
	diffText := `diff --git a/ai-context.json b/ai-context.json
--- a/ai-context.json
+++ b/ai-context.json
@@ -3,1 +3,1 @@
-    "agent_check": {"cmd": "make agent-check"}
+    "agent_check": {"cmd": "USE_ISOLATED_GO_CACHE=1 make agent-check"}
`

	output, err := fixture.RunChangePolicyCheckWithDiff("ai-context.json", diffText)
	if err != nil {
		t.Fatalf("check_change_policy.sh failed: %v\n%s", err, output)
	}
	for _, unwanted := range []string{
		"SEMANTIC_AI_CONTEXT_COVERAGE_GATES_CHANGE",
		"SEMANTIC_AI_CONTEXT_API_FREEZE_CHANGE",
		"SEMANTIC_AI_CONTEXT_SECURITY_SENSITIVE_PACKAGES_CHANGE",
		"SEMANTIC_AI_CONTEXT_RANDOM_SOURCE_POLICY_CHANGE",
		"SEMANTIC_AI_CONTEXT_THREAT_MODEL_CHANGE",
	} {
		if strings.Contains(output, unwanted) {
			t.Fatalf("change policy output overclassified unrelated ai-context diff as %s:\n%s", unwanted, output)
		}
	}
}

func TestChangePolicyCheckEmitsJSONReport(t *testing.T) {
	fixture := newGovernanceFixture(t)
	writeSemanticChangePolicyFixture(fixture)

	output, err := fixture.RunChangePolicyCheckJSON("ai-context.json\nMakefile", changePolicySemanticDiffFixture())
	if err != nil {
		t.Fatalf("changepolicycheck -json failed: %v\n%s", err, output)
	}

	var result struct {
		Status           string              `json:"status"`
		Findings         []json.RawMessage   `json:"findings"`
		DetectedPolicies []string            `json:"detected_policies"`
		RuleIDs          []string            `json:"rule_ids"`
		SemanticRuleIDs  []string            `json:"semantic_rule_ids"`
		RequiredCommands []string            `json:"required_commands"`
		PolicyPaths      map[string][]string `json:"policy_paths"`
	}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("changepolicycheck -json output is not valid JSON: %v\n%s", err, output)
	}
	if result.Status != "passed" {
		t.Fatalf("JSON status = %q, want passed\n%s", result.Status, output)
	}
	if len(result.Findings) != 0 {
		t.Fatalf("JSON findings length = %d, want 0\n%s", len(result.Findings), output)
	}
	if !stringSliceContains(result.DetectedPolicies, "ci_governance") {
		t.Fatalf("JSON detected_policies missing ci_governance:\n%s", output)
	}
	if !stringSliceContains(result.RuleIDs, "CHANGE_CI_GOVERNANCE") {
		t.Fatalf("JSON rule_ids missing CHANGE_CI_GOVERNANCE:\n%s", output)
	}
	if !stringSliceContains(result.SemanticRuleIDs, "SEMANTIC_AI_CONTEXT_API_FREEZE_CHANGE") ||
		!stringSliceContains(result.SemanticRuleIDs, "SEMANTIC_MAKEFILE_RELEASE_CHECK_CHANGE") {
		t.Fatalf("JSON semantic_rule_ids missing field-level ids:\n%s", output)
	}
	if !stringSliceContains(result.RequiredCommands, "agent_check") {
		t.Fatalf("JSON required_commands missing agent_check:\n%s", output)
	}
	if !stringSliceContains(result.PolicyPaths["ci_governance"], "ai-context.json") ||
		!stringSliceContains(result.PolicyPaths["ci_governance"], "Makefile") {
		t.Fatalf("JSON policy_paths missing ci_governance paths:\n%s", output)
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

	output, err := newGovernanceFixture(t).RunAgentEvidenceCheck(evidence)
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

	output, err := newGovernanceFixture(t).RunAgentEvidenceCheck(evidence)
	if err == nil {
		t.Fatalf("agent evidence check unexpectedly passed:\n%s", output)
	}
	if !strings.Contains(output, "AGENT_EVIDENCE_SECURITY_REVIEW_SCHEMA") {
		t.Fatalf("agent evidence output missing AGENT_EVIDENCE_SECURITY_REVIEW_SCHEMA rule id:\n%s", output)
	}
	if !strings.Contains(output, "security_review must be an object") {
		t.Fatalf("agent evidence output missing security_review error:\n%s", output)
	}
}

func TestAgentEvidenceCheckRejectsIncorrectSecurityMergeReadyEvidence(t *testing.T) {
	evidence := baseAgentEvidence()
	evidence["merge_ready"] = true
	evidence["merge_blockers"] = []string{}

	output, err := newGovernanceFixture(t).RunAgentEvidenceCheck(evidence)
	if err == nil {
		t.Fatalf("agent evidence check unexpectedly passed:\n%s", output)
	}
	if !strings.Contains(output, "AGENT_EVIDENCE_MERGE_READY_MISMATCH") {
		t.Fatalf("agent evidence output missing AGENT_EVIDENCE_MERGE_READY_MISMATCH rule id:\n%s", output)
	}
	if !strings.Contains(output, "merge_ready must be false") {
		t.Fatalf("agent evidence output missing merge_ready error:\n%s", output)
	}
}

func TestAgentEvidenceCheckRejectsStructuredChangePolicyMismatch(t *testing.T) {
	evidence := baseAgentEvidence()
	structuredChecks := evidence["structured_checks"].(map[string]any)
	changePolicy := structuredChecks["change_policy_check"].(map[string]any)
	changePolicyJSON := changePolicy["json"].(map[string]any)
	changePolicyJSON["rule_ids"] = []string{"CHANGE_BUG_FIX"}

	output, err := newGovernanceFixture(t).RunAgentEvidenceCheck(evidence)
	if err == nil {
		t.Fatalf("agent evidence check unexpectedly passed:\n%s", output)
	}
	if !strings.Contains(output, "AGENT_EVIDENCE_STRUCTURED_CHANGE_POLICY_MISMATCH") {
		t.Fatalf("agent evidence output missing AGENT_EVIDENCE_STRUCTURED_CHANGE_POLICY_MISMATCH rule id:\n%s", output)
	}
	if !strings.Contains(output, "structured change policy rule_ids must match detected policies") {
		t.Fatalf("agent evidence output missing structured rule id mismatch:\n%s", output)
	}
}

func TestAgentEvidenceCheckRejectsStructuredCIWorkflowFindings(t *testing.T) {
	evidence := baseAgentEvidence()
	structuredChecks := evidence["structured_checks"].(map[string]any)
	ciWorkflow := structuredChecks["ci_workflow_check"].(map[string]any)
	ciWorkflowJSON := ciWorkflow["json"].(map[string]any)
	ciWorkflowJSON["findings"] = []any{
		map[string]any{
			"rule_id":  "CI_WORKFLOW_UNKNOWN_MAKE_TARGET",
			"path":     ".github/workflows/go.yml",
			"message":  "fixture",
			"severity": "error",
		},
	}

	output, err := newGovernanceFixture(t).RunAgentEvidenceCheck(evidence)
	if err == nil {
		t.Fatalf("agent evidence check unexpectedly passed:\n%s", output)
	}
	if !strings.Contains(output, "AGENT_EVIDENCE_STRUCTURED_CHECK_FINDINGS") {
		t.Fatalf("agent evidence output missing AGENT_EVIDENCE_STRUCTURED_CHECK_FINDINGS rule id:\n%s", output)
	}
}

func TestAgentEvidenceCheckEmitsJSONReport(t *testing.T) {
	evidence := baseAgentEvidence()
	output, err := newGovernanceFixture(t).RunAgentEvidenceCheckJSON(evidence)
	if err != nil {
		t.Fatalf("agentevidencecheck -json failed: %v\n%s", err, output)
	}

	var result struct {
		Status               string            `json:"status"`
		Findings             []json.RawMessage `json:"findings"`
		PolicyCount          int               `json:"policy_count"`
		RequiredCommandCount int               `json:"required_command_count"`
		MergeReady           bool              `json:"merge_ready"`
	}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("agentevidencecheck -json output is not valid JSON: %v\n%s", err, output)
	}
	if result.Status != "passed" || len(result.Findings) != 0 {
		t.Fatalf("unexpected agent evidence JSON status/findings:\n%s", output)
	}
	if result.PolicyCount != 1 || result.RequiredCommandCount != 6 || result.MergeReady {
		t.Fatalf("unexpected agent evidence JSON summary:\n%s", output)
	}
}

func TestAPIFreezeCheckRequiresStatusDecisionCards(t *testing.T) {
	fixture := newGovernanceFixture(t)
	fixture.WriteJSON("tools.json", map[string]any{
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
	toolsPath := filepath.Join(fixture.Root(), "tools.json")
	context := minimalAPIFreezeContext()
	context["api_freeze"].(map[string]any)["api_status_decision_cards"] = map[string]any{
		"recommended":   []string{"recommended-card"},
		"compatibility": []string{"recommended-card"},
		"experimental":  []string{"experimental-card"},
		"deprecated":    []string{"deprecated-card"},
	}
	fixture.WriteJSON("ai-context.json", context)
	contextPath := filepath.Join(fixture.Root(), "ai-context.json")

	output, err := fixture.RunAPIFreezeCheck(contextPath, toolsPath)
	if err == nil {
		t.Fatalf("api freeze check unexpectedly passed:\n%s", output)
	}
	if !strings.Contains(output, "API_FREEZE_STATUS_CARD_STATUS_MISMATCH") {
		t.Fatalf("api freeze output missing API_FREEZE_STATUS_CARD_STATUS_MISMATCH rule id:\n%s", output)
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
	fixture.WriteJSON("ai-context.json", context)
	output, err = fixture.RunAPIFreezeCheck(contextPath, toolsPath)
	if err != nil {
		t.Fatalf("api freeze check failed: %v\n%s", err, output)
	}
	if !strings.Contains(output, "api freeze metadata is valid") {
		t.Fatalf("api freeze output missing success message:\n%s", output)
	}
}

func TestAPIFreezeCheckEmitsJSONReport(t *testing.T) {
	fixture := newGovernanceFixture(t)
	fixture.WriteJSON("tools.json", map[string]any{
		"packages": []any{},
	})
	toolsPath := filepath.Join(fixture.Root(), "tools.json")
	context := minimalAPIFreezeContext()
	fixture.WriteJSON("ai-context.json", context)
	contextPath := filepath.Join(fixture.Root(), "ai-context.json")

	output, err := fixture.RunAPIFreezeCheckJSON(contextPath, toolsPath)
	if err != nil {
		t.Fatalf("apifreezecheck -json failed: %v\n%s", err, output)
	}
	var result struct {
		Status            string            `json:"status"`
		Findings          []json.RawMessage `json:"findings"`
		DeprecatedCount   int               `json:"deprecated_count"`
		ExperimentalCount int               `json:"experimental_count"`
	}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("apifreezecheck -json output is not valid JSON: %v\n%s", err, output)
	}
	if result.Status != "passed" || len(result.Findings) != 0 {
		t.Fatalf("unexpected API freeze JSON status/findings:\n%s", output)
	}
	if result.DeprecatedCount != 0 || result.ExperimentalCount != 0 {
		t.Fatalf("unexpected API freeze JSON counts:\n%s", output)
	}
}

func TestRandomSourcePolicyCheckAcceptsValidFixture(t *testing.T) {
	fixture := randomSourcePolicyFixture(t)
	output, err := fixture.RunRandomSourcePolicyCheck()
	if err != nil {
		t.Fatalf("randomsourcepolicycheck failed: %v\n%s", err, output)
	}
	if !strings.Contains(output, "random source policy governance is valid") {
		t.Fatalf("random source policy output missing success:\n%s", output)
	}
}

func TestRandomSourcePolicyCheckRejectsUnknownPolicyName(t *testing.T) {
	fixture := randomSourcePolicyFixture(t)
	context := randomSourcePolicyContext()
	policies := context["random_source_policy"].(map[string]any)["policies"].([]any)
	policies[0].(map[string]any)["name"] = "secure_bytes_fail_clsoed"
	fixture.WriteJSON("ai-context.json", context)

	output, err := fixture.RunRandomSourcePolicyCheckJSON()
	if err == nil {
		t.Fatalf("randomsourcepolicycheck -json unexpectedly passed:\n%s", output)
	}
	var result struct {
		Status   string `json:"status"`
		Findings []struct {
			RuleID string `json:"rule_id"`
		} `json:"findings"`
	}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("randomsourcepolicycheck -json output is not valid JSON: %v\n%s", err, output)
	}
	if result.Status != "failed" {
		t.Fatalf("JSON status = %q, want failed\n%s", result.Status, output)
	}
	var sawUnknown bool
	for _, finding := range result.Findings {
		if finding.RuleID == "RANDOM_SOURCE_POLICY_UNKNOWN_NAME" {
			sawUnknown = true
		}
	}
	if !sawUnknown {
		t.Fatalf("random source JSON findings missing unknown-name rule id:\n%s", output)
	}
}

func TestRandomSourcePolicyCheckRejectsMissingContractTest(t *testing.T) {
	fixture := randomSourcePolicyFixture(t)
	context := randomSourcePolicyContext()
	policies := context["random_source_policy"].(map[string]any)["policies"].([]any)
	policies[0].(map[string]any)["contract_tests"] = []string{"internal/rand/random_bytes_test.go:TestMissingFixture"}
	fixture.WriteJSON("ai-context.json", context)

	output, err := fixture.RunRandomSourcePolicyCheck()
	if err == nil {
		t.Fatalf("randomsourcepolicycheck unexpectedly passed:\n%s", output)
	}
	if !strings.Contains(output, "RANDOM_SOURCE_POLICY_CONTRACT_TEST_MISSING") {
		t.Fatalf("random source policy output missing missing-test rule id:\n%s", output)
	}
}

func TestRandomSourcePolicyCheckRejectsPackageCoverageDrift(t *testing.T) {
	fixture := randomSourcePolicyFixture(t)
	context := randomSourcePolicyContext()
	policy := context["random_source_policy"].(map[string]any)
	policy["packages"] = []string{"vrand", "vid", "vcrypto"}
	fixture.WriteJSON("ai-context.json", context)

	output, err := fixture.RunRandomSourcePolicyCheck()
	if err == nil {
		t.Fatalf("randomsourcepolicycheck unexpectedly passed:\n%s", output)
	}
	if !strings.Contains(output, "RANDOM_SOURCE_POLICY_PACKAGE_COVERAGE") {
		t.Fatalf("random source policy output missing package coverage rule id:\n%s", output)
	}
}

func TestThreatModelCheckAcceptsValidBoundaryFixture(t *testing.T) {
	fixture := threatModelFixture(t)
	output, err := fixture.RunThreatModelCheck()
	if err != nil {
		t.Fatalf("threatmodelcheck failed: %v\n%s", err, output)
	}
	if !strings.Contains(output, "threat model boundary contracts are valid") {
		t.Fatalf("threat model output missing success:\n%s", output)
	}
}

func TestThreatModelCheckRejectsUnknownBoundaryName(t *testing.T) {
	fixture := threatModelFixture(t)
	context := threatModelContext()
	boundaries := context["threat_model"].(map[string]any)["boundary_contracts"].([]any)
	boundaries[0].(map[string]any)["name"] = "default_timeot"
	fixture.WriteJSON("ai-context.json", context)

	output, err := fixture.RunThreatModelCheckJSON()
	if err == nil {
		t.Fatalf("threatmodelcheck -json unexpectedly passed:\n%s", output)
	}
	var result struct {
		Status   string `json:"status"`
		Findings []struct {
			RuleID string `json:"rule_id"`
		} `json:"findings"`
	}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("threatmodelcheck -json output is not valid JSON: %v\n%s", err, output)
	}
	if result.Status != "failed" {
		t.Fatalf("JSON status = %q, want failed\n%s", result.Status, output)
	}
	var sawUnknown bool
	for _, finding := range result.Findings {
		if finding.RuleID == "THREAT_MODEL_BOUNDARY_UNKNOWN_NAME" {
			sawUnknown = true
		}
	}
	if !sawUnknown {
		t.Fatalf("threat model JSON findings missing unknown boundary rule id:\n%s", output)
	}
}

func TestThreatModelCheckRejectsMissingContractTest(t *testing.T) {
	fixture := threatModelFixture(t)
	context := threatModelContext()
	boundaries := context["threat_model"].(map[string]any)["boundary_contracts"].([]any)
	boundaries[0].(map[string]any)["contract_tests"] = []string{"internal/httpx/contract_request_test.go:TestMissingFixture", "internal/httpx/http/request_timeout_redirect_test.go:TestRequestTimeout"}
	fixture.WriteJSON("ai-context.json", context)

	output, err := fixture.RunThreatModelCheck()
	if err == nil {
		t.Fatalf("threatmodelcheck unexpectedly passed:\n%s", output)
	}
	if !strings.Contains(output, "THREAT_MODEL_BOUNDARY_CONTRACT_TEST_MISSING") {
		t.Fatalf("threat model output missing missing-test rule id:\n%s", output)
	}
}

func TestThreatModelCheckRejectsUnknownPackage(t *testing.T) {
	fixture := threatModelFixture(t)
	context := threatModelContext()
	boundaries := context["threat_model"].(map[string]any)["boundary_contracts"].([]any)
	boundaries[0].(map[string]any)["packages"] = []string{"vhttp", "vmissing"}
	fixture.WriteJSON("ai-context.json", context)

	output, err := fixture.RunThreatModelCheck()
	if err == nil {
		t.Fatalf("threatmodelcheck unexpectedly passed:\n%s", output)
	}
	if !strings.Contains(output, "THREAT_MODEL_BOUNDARY_UNKNOWN_PACKAGE") {
		t.Fatalf("threat model output missing unknown-package rule id:\n%s", output)
	}
}

func TestDynamicContractsCheckAcceptsValidFixture(t *testing.T) {
	fixture := dynamicContractsFixture(t)
	output, err := fixture.RunDynamicContractsCheck()
	if err != nil {
		t.Fatalf("dynamiccontractscheck failed: %v\n%s", err, output)
	}
	if !strings.Contains(output, "dynamic semantic contracts are valid") {
		t.Fatalf("dynamic contracts output missing success:\n%s", output)
	}
}

func TestDynamicContractsCheckRejectsUnknownDomain(t *testing.T) {
	fixture := dynamicContractsFixture(t)
	context := dynamicContractsContext()
	contracts := context["dynamic_semantic_contracts"].(map[string]any)
	domains := contracts["domains"].(map[string]any)
	domains["vjson_dyanmic"] = domains["vjson_dynamic"]
	fixture.WriteJSON("ai-context.json", context)

	output, err := fixture.RunDynamicContractsCheckJSON()
	if err == nil {
		t.Fatalf("dynamiccontractscheck -json unexpectedly passed:\n%s", output)
	}
	var result struct {
		Status   string `json:"status"`
		Findings []struct {
			RuleID string `json:"rule_id"`
		} `json:"findings"`
	}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("dynamiccontractscheck -json output is not valid JSON: %v\n%s", err, output)
	}
	if result.Status != "failed" {
		t.Fatalf("JSON status = %q, want failed\n%s", result.Status, output)
	}
	var sawUnknown bool
	for _, finding := range result.Findings {
		if finding.RuleID == "DYNAMIC_CONTRACTS_DOMAIN_UNKNOWN" {
			sawUnknown = true
		}
	}
	if !sawUnknown {
		t.Fatalf("dynamic contracts JSON findings missing unknown-domain rule id:\n%s", output)
	}
}

func TestDynamicContractsCheckRejectsMissingReference(t *testing.T) {
	fixture := dynamicContractsFixture(t)
	context := dynamicContractsContext()
	domains := context["dynamic_semantic_contracts"].(map[string]any)["domains"].(map[string]any)
	domains["vjson_dynamic"].(map[string]any)["contract_tests"] = []string{"vjson/json_conversion_test.go:TestMissingFixture"}
	fixture.WriteJSON("ai-context.json", context)

	output, err := fixture.RunDynamicContractsCheck()
	if err == nil {
		t.Fatalf("dynamiccontractscheck unexpectedly passed:\n%s", output)
	}
	if !strings.Contains(output, "DYNAMIC_CONTRACTS_REFERENCE_MISSING") {
		t.Fatalf("dynamic contracts output missing missing-reference rule id:\n%s", output)
	}
}

func TestDynamicContractsCheckRejectsPackageCoverageDrift(t *testing.T) {
	fixture := dynamicContractsFixture(t)
	context := dynamicContractsContext()
	domains := context["dynamic_semantic_contracts"].(map[string]any)["domains"].(map[string]any)
	domains["vref_reflection_boundaries"].(map[string]any)["packages"] = []string{"internal/ref"}
	fixture.WriteJSON("ai-context.json", context)

	output, err := fixture.RunDynamicContractsCheck()
	if err == nil {
		t.Fatalf("dynamiccontractscheck unexpectedly passed:\n%s", output)
	}
	if !strings.Contains(output, "DYNAMIC_CONTRACTS_PACKAGE_COVERAGE") {
		t.Fatalf("dynamic contracts output missing package-coverage rule id:\n%s", output)
	}
}

func TestErrorModelCheckAcceptsValidFixture(t *testing.T) {
	fixture := errorModelFixture(t)
	output, err := fixture.RunErrorModelCheck()
	if err != nil {
		t.Fatalf("errormodelcheck failed: %v\n%s", err, output)
	}
	if !strings.Contains(output, "error model governance is valid") {
		t.Fatalf("error model output missing success:\n%s", output)
	}
}

func TestErrorModelCheckRejectsMissingCode(t *testing.T) {
	fixture := errorModelFixture(t)
	context := errorModelContext()
	taxonomy := context["error_model"].(map[string]any)["taxonomy"].([]any)
	taxonomy = taxonomy[:len(taxonomy)-1]
	context["error_model"].(map[string]any)["taxonomy"] = taxonomy
	fixture.WriteJSON("ai-context.json", context)

	output, err := fixture.RunErrorModelCheckJSON()
	if err == nil {
		t.Fatalf("errormodelcheck -json unexpectedly passed:\n%s", output)
	}
	var result struct {
		Status   string `json:"status"`
		Findings []struct {
			RuleID string `json:"rule_id"`
		} `json:"findings"`
	}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("errormodelcheck -json output is not valid JSON: %v\n%s", err, output)
	}
	if result.Status != "failed" {
		t.Fatalf("JSON status = %q, want failed\n%s", result.Status, output)
	}
	var sawCoverage bool
	for _, finding := range result.Findings {
		if finding.RuleID == "ERROR_MODEL_TAXONOMY_COVERAGE" {
			sawCoverage = true
		}
	}
	if !sawCoverage {
		t.Fatalf("error model JSON findings missing taxonomy coverage rule id:\n%s", output)
	}
}

func TestErrorModelCheckRejectsMissingConstant(t *testing.T) {
	fixture := errorModelFixture(t)
	fixture.WriteFile("errors.go", `package knifer

const ErrCodeInvalidInput = "GK_INVALID_INPUT"
`)

	output, err := fixture.RunErrorModelCheck()
	if err == nil {
		t.Fatalf("errormodelcheck unexpectedly passed:\n%s", output)
	}
	if !strings.Contains(output, "ERROR_MODEL_ERROR_CONSTANT_MISSING") {
		t.Fatalf("error model output missing missing-constant rule id:\n%s", output)
	}
}

func TestErrorModelCheckRejectsMissingContractTest(t *testing.T) {
	fixture := errorModelFixture(t)
	context := errorModelContext()
	context["error_model"].(map[string]any)["contract_tests"] = []string{"errors_test.go:TestMissingFixture"}
	fixture.WriteJSON("ai-context.json", context)

	output, err := fixture.RunErrorModelCheck()
	if err == nil {
		t.Fatalf("errormodelcheck unexpectedly passed:\n%s", output)
	}
	if !strings.Contains(output, "ERROR_MODEL_CONTRACT_TEST_MISSING") {
		t.Fatalf("error model output missing missing-contract-test rule id:\n%s", output)
	}
}

func TestBenchmarkRegressionCheckAcceptsValidFixture(t *testing.T) {
	fixture := benchmarkRegressionFixture(t)
	output, err := fixture.RunBenchmarkRegressionCheck()
	if err != nil {
		t.Fatalf("benchmarkregressioncheck failed: %v\n%s", err, output)
	}
	if !strings.Contains(output, "benchmark regression metadata is valid") {
		t.Fatalf("benchmark regression output missing success:\n%s", output)
	}
}

func TestBenchmarkRegressionCheckRejectsHotPathMissingBenchTarget(t *testing.T) {
	fixture := benchmarkRegressionFixture(t)
	fixture.WriteFile("Makefile", benchmarkRegressionMakefileWithout("./vjson"))

	output, err := fixture.RunBenchmarkRegressionCheckJSON()
	if err == nil {
		t.Fatalf("benchmarkregressioncheck -json unexpectedly passed:\n%s", output)
	}
	var result struct {
		Status   string `json:"status"`
		Findings []struct {
			RuleID string `json:"rule_id"`
		} `json:"findings"`
	}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("benchmarkregressioncheck -json output is not valid JSON: %v\n%s", err, output)
	}
	if result.Status != "failed" {
		t.Fatalf("JSON status = %q, want failed\n%s", result.Status, output)
	}
	var sawBenchTarget bool
	for _, finding := range result.Findings {
		if finding.RuleID == "BENCHMARK_REGRESSION_HOT_PATH_BENCH_TARGET" {
			sawBenchTarget = true
		}
	}
	if !sawBenchTarget {
		t.Fatalf("benchmark regression JSON findings missing bench-target rule id:\n%s", output)
	}
}

func TestBenchmarkRegressionCheckRejectsMissingBenchmarkFunction(t *testing.T) {
	fixture := benchmarkRegressionFixture(t)
	context := benchmarkRegressionContext()
	hotPaths := context["benchmark_regression"].(map[string]any)["hot_path_packages"].([]any)
	hotPaths[0].(map[string]any)["benchmarks"] = []string{"vjson/json_benchmark_test.go:BenchmarkMissingFixture"}
	fixture.WriteJSON("ai-context.json", context)

	output, err := fixture.RunBenchmarkRegressionCheck()
	if err == nil {
		t.Fatalf("benchmarkregressioncheck unexpectedly passed:\n%s", output)
	}
	if !strings.Contains(output, "BENCHMARK_REGRESSION_BENCHMARK_MISSING") {
		t.Fatalf("benchmark regression output missing missing-benchmark rule id:\n%s", output)
	}
}

func TestBenchmarkRegressionCheckRejectsMinimumCountTooLow(t *testing.T) {
	fixture := benchmarkRegressionFixture(t)
	context := benchmarkRegressionContext()
	thresholds := context["benchmark_regression"].(map[string]any)["thresholds"].(map[string]any)
	thresholds["minimum_count"] = 1
	fixture.WriteJSON("ai-context.json", context)

	output, err := fixture.RunBenchmarkRegressionCheck()
	if err == nil {
		t.Fatalf("benchmarkregressioncheck unexpectedly passed:\n%s", output)
	}
	if !strings.Contains(output, "BENCHMARK_REGRESSION_THRESHOLD_INVALID") {
		t.Fatalf("benchmark regression output missing threshold rule id:\n%s", output)
	}
}

func TestBenchmarkRegressionCheckRejectsMissingMakeTarget(t *testing.T) {
	fixture := benchmarkRegressionFixture(t)
	fixture.WriteFile("Makefile", benchmarkRegressionMakefileWithoutTarget("benchstat"))

	output, err := fixture.RunBenchmarkRegressionCheck()
	if err == nil {
		t.Fatalf("benchmarkregressioncheck unexpectedly passed:\n%s", output)
	}
	if !strings.Contains(output, "BENCHMARK_REGRESSION_MAKE_TARGET_MISSING") {
		t.Fatalf("benchmark regression output missing make-target rule id:\n%s", output)
	}
}

func TestAPIConvergenceCheckAcceptsValidFixture(t *testing.T) {
	fixture := apiConvergenceFixture(t)
	output, err := fixture.RunAPIConvergenceCheck()
	if err != nil {
		t.Fatalf("apiconvergencecheck failed: %v\n%s", err, output)
	}
	if !strings.Contains(output, "api convergence metadata is valid") {
		t.Fatalf("api convergence output missing success:\n%s", output)
	}
}

func TestAPIConvergenceCheckRejectsGoldenPathDrift(t *testing.T) {
	fixture := apiConvergenceFixture(t)
	context := apiConvergenceContext()
	context["api_convergence"].(map[string]any)["facades"].(map[string]any)["vjson"].(map[string]any)["primary"] = []string{"Format", "Parse"}
	fixture.WriteJSON("ai-context.json", context)

	output, err := fixture.RunAPIConvergenceCheckJSON()
	if err == nil {
		t.Fatalf("apiconvergencecheck -json unexpectedly passed:\n%s", output)
	}
	var result struct {
		Status   string `json:"status"`
		Findings []struct {
			RuleID string `json:"rule_id"`
		} `json:"findings"`
	}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("apiconvergencecheck -json output is not valid JSON: %v\n%s", err, output)
	}
	if result.Status != "failed" {
		t.Fatalf("JSON status = %q, want failed\n%s", result.Status, output)
	}
	var sawDrift bool
	for _, finding := range result.Findings {
		if finding.RuleID == "API_CONVERGENCE_PRIMARY_DRIFT" {
			sawDrift = true
		}
	}
	if !sawDrift {
		t.Fatalf("api convergence JSON findings missing primary drift rule id:\n%s", output)
	}
}

func TestAPIConvergenceCheckRejectsUnknownAPI(t *testing.T) {
	fixture := apiConvergenceFixture(t)
	context := apiConvergenceContext()
	context["api_convergence"].(map[string]any)["facades"].(map[string]any)["vjson"].(map[string]any)["advanced"] = []string{"MissingAPI"}
	fixture.WriteJSON("ai-context.json", context)

	output, err := fixture.RunAPIConvergenceCheck()
	if err == nil {
		t.Fatalf("apiconvergencecheck unexpectedly passed:\n%s", output)
	}
	if !strings.Contains(output, "API_CONVERGENCE_UNKNOWN_API") {
		t.Fatalf("api convergence output missing unknown API rule id:\n%s", output)
	}
}

func TestAPIConvergenceCheckRejectsCompatibilityStatusDrift(t *testing.T) {
	fixture := apiConvergenceFixture(t)
	context := apiConvergenceContext()
	context["api_convergence"].(map[string]any)["facades"].(map[string]any)["vjson"].(map[string]any)["compatibility"] = []string{"Format"}
	fixture.WriteJSON("ai-context.json", context)

	output, err := fixture.RunAPIConvergenceCheck()
	if err == nil {
		t.Fatalf("apiconvergencecheck unexpectedly passed:\n%s", output)
	}
	if !strings.Contains(output, "API_CONVERGENCE_COMPATIBILITY_STATUS") {
		t.Fatalf("api convergence output missing compatibility status rule id:\n%s", output)
	}
}

func TestLifecycleCheckAcceptsValidFixture(t *testing.T) {
	fixture := lifecycleFixture(t)
	output, err := fixture.RunLifecycleCheck()
	if err != nil {
		t.Fatalf("lifecyclecheck failed: %v\n%s", err, output)
	}
	if !strings.Contains(output, "package lifecycle metadata is valid") {
		t.Fatalf("lifecycle output missing success:\n%s", output)
	}
}

func TestLifecycleCheckRejectsHeavyGradeMismatch(t *testing.T) {
	fixture := lifecycleFixture(t)
	context := lifecycleContext()
	context["package_lifecycle"].(map[string]any)["packages"].(map[string]any)["vimg"].(map[string]any)["grade"] = "core"
	fixture.WriteJSON("ai-context.json", context)

	output, err := fixture.RunLifecycleCheckJSON()
	if err == nil {
		t.Fatalf("lifecyclecheck -json unexpectedly passed:\n%s", output)
	}
	var result struct {
		Status   string `json:"status"`
		Findings []struct {
			RuleID string `json:"rule_id"`
		} `json:"findings"`
	}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("lifecyclecheck -json output is not valid JSON: %v\n%s", err, output)
	}
	if result.Status != "failed" {
		t.Fatalf("JSON status = %q, want failed\n%s", result.Status, output)
	}
	var sawMismatch bool
	for _, finding := range result.Findings {
		if finding.RuleID == "LIFECYCLE_HEAVY_GRADE_MISMATCH" {
			sawMismatch = true
		}
	}
	if !sawMismatch {
		t.Fatalf("lifecycle JSON findings missing heavy mismatch rule id:\n%s", output)
	}
}

func TestLifecycleCheckRejectsTierOverlap(t *testing.T) {
	fixture := lifecycleFixture(t)
	context := lifecycleContext()
	context["dependency_tiers"].(map[string]any)["provider_contract_facades"] = []string{"vai", "vimg"}
	fixture.WriteJSON("ai-context.json", context)

	output, err := fixture.RunLifecycleCheck()
	if err == nil {
		t.Fatalf("lifecyclecheck unexpectedly passed:\n%s", output)
	}
	if !strings.Contains(output, "LIFECYCLE_DEPENDENCY_TIER_OVERLAP") {
		t.Fatalf("lifecycle output missing tier overlap rule id:\n%s", output)
	}
}

func TestLifecycleCheckRejectsMissingRationale(t *testing.T) {
	fixture := lifecycleFixture(t)
	context := lifecycleContext()
	context["package_lifecycle"].(map[string]any)["packages"].(map[string]any)["vjson"].(map[string]any)["rationale"] = ""
	fixture.WriteJSON("ai-context.json", context)

	output, err := fixture.RunLifecycleCheck()
	if err == nil {
		t.Fatalf("lifecyclecheck unexpectedly passed:\n%s", output)
	}
	if !strings.Contains(output, "LIFECYCLE_RATIONALE_MISSING") {
		t.Fatalf("lifecycle output missing rationale rule id:\n%s", output)
	}
}

func TestDependencyTiersCheckAcceptsValidFixture(t *testing.T) {
	fixture := dependencyTiersFixture(t)
	output, err := fixture.RunDependencyTiersCheck()
	if err != nil {
		t.Fatalf("dependencytierscheck failed: %v\n%s", err, output)
	}
	if !strings.Contains(output, "dependency tiers metadata is valid") {
		t.Fatalf("dependency tiers output missing success:\n%s", output)
	}
}

func TestDependencyTiersCheckRejectsUnknownFacade(t *testing.T) {
	fixture := dependencyTiersFixture(t)
	context := dependencyTiersContext()
	context["dependency_tiers"].(map[string]any)["core_facades"] = []string{"vjson", "vmissing"}
	fixture.WriteJSON("ai-context.json", context)

	output, err := fixture.RunDependencyTiersCheckJSON()
	if err == nil {
		t.Fatalf("dependencytierscheck -json unexpectedly passed:\n%s", output)
	}
	var result struct {
		Status   string `json:"status"`
		Findings []struct {
			RuleID string `json:"rule_id"`
		} `json:"findings"`
	}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("dependencytierscheck -json output is not valid JSON: %v\n%s", err, output)
	}
	if result.Status != "failed" {
		t.Fatalf("JSON status = %q, want failed\n%s", result.Status, output)
	}
	var sawUnknown bool
	for _, finding := range result.Findings {
		if finding.RuleID == "DEPENDENCY_TIERS_UNKNOWN_FACADE" {
			sawUnknown = true
		}
	}
	if !sawUnknown {
		t.Fatalf("dependency tiers JSON findings missing unknown facade rule id:\n%s", output)
	}
}

func TestDependencyTiersCheckRejectsTierOverlap(t *testing.T) {
	fixture := dependencyTiersFixture(t)
	context := dependencyTiersContext()
	context["dependency_tiers"].(map[string]any)["provider_contract_facades"] = []string{"vai", "vimg"}
	fixture.WriteJSON("ai-context.json", context)

	output, err := fixture.RunDependencyTiersCheck()
	if err == nil {
		t.Fatalf("dependencytierscheck unexpectedly passed:\n%s", output)
	}
	if !strings.Contains(output, "DEPENDENCY_TIERS_OVERLAP") {
		t.Fatalf("dependency tiers output missing overlap rule id:\n%s", output)
	}
}

func TestDependencyTiersCheckRejectsUnknownAllowlistPrefix(t *testing.T) {
	fixture := dependencyTiersFixture(t)
	context := dependencyTiersContext()
	context["dependency_tiers"].(map[string]any)["heavy_dependency_allowlist"].(map[string]any)["example.com/heavy"] = []string{"internal/missing"}
	fixture.WriteJSON("ai-context.json", context)

	output, err := fixture.RunDependencyTiersCheck()
	if err == nil {
		t.Fatalf("dependencytierscheck unexpectedly passed:\n%s", output)
	}
	if !strings.Contains(output, "DEPENDENCY_TIERS_ALLOWLIST_PREFIX_UNKNOWN") {
		t.Fatalf("dependency tiers output missing allowlist prefix rule id:\n%s", output)
	}
}

func TestDependencyTiersCheckRejectsAllowlistSchema(t *testing.T) {
	fixture := dependencyTiersFixture(t)
	context := dependencyTiersContext()
	context["dependency_tiers"].(map[string]any)["heavy_dependency_allowlist"].(map[string]any)["example.com/heavy"] = "vimg"
	fixture.WriteJSON("ai-context.json", context)

	output, err := fixture.RunDependencyTiersCheck()
	if err == nil {
		t.Fatalf("dependencytierscheck unexpectedly passed:\n%s", output)
	}
	if !strings.Contains(output, "DEPENDENCY_TIERS_ALLOWLIST_SCHEMA") {
		t.Fatalf("dependency tiers output missing allowlist schema rule id:\n%s", output)
	}
}

func TestGovernanceMigrationCheckAcceptsValidFixture(t *testing.T) {
	fixture := governanceMigrationFixture(t)
	output, err := fixture.RunGovernanceMigrationCheck()
	if err != nil {
		t.Fatalf("governancemigrationcheck failed: %v\n%s", err, output)
	}
	if !strings.Contains(output, "governance migration status is valid") {
		t.Fatalf("governance migration output missing success:\n%s", output)
	}
}

func TestGovernanceMigrationCheckRejectsPythonRuleRegression(t *testing.T) {
	fixture := governanceMigrationFixture(t)
	fixture.WriteFile("bin/check_governance_maturity.sh", "def validate_error_model() -> None:\n\tpass\n")

	output, err := fixture.RunGovernanceMigrationCheckJSON()
	if err == nil {
		t.Fatalf("governancemigrationcheck -json unexpectedly passed:\n%s", output)
	}
	var result struct {
		Status   string `json:"status"`
		Findings []struct {
			RuleID string `json:"rule_id"`
		} `json:"findings"`
	}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("governancemigrationcheck -json output is not valid JSON: %v\n%s", err, output)
	}
	if result.Status != "failed" {
		t.Fatalf("JSON status = %q, want failed\n%s", result.Status, output)
	}
	var sawRegression bool
	for _, finding := range result.Findings {
		if finding.RuleID == "GOVERNANCE_MIGRATION_PYTHON_RULE_REGRESSION" {
			sawRegression = true
		}
	}
	if !sawRegression {
		t.Fatalf("governance migration JSON findings missing regression rule id:\n%s", output)
	}
}

func TestGovernanceMigrationCheckRejectsMissingMakeTarget(t *testing.T) {
	fixture := governanceMigrationFixture(t)
	fixture.WriteFile("Makefile", `governance-maturity-check:
	bash bin/check_governance_maturity.sh
	$(MAKE) random-source-policy-check

bench-regression-check:
	go run ./bin/benchmarkregressioncheck -root .
`)

	output, err := fixture.RunGovernanceMigrationCheck()
	if err == nil {
		t.Fatalf("governancemigrationcheck unexpectedly passed:\n%s", output)
	}
	if !strings.Contains(output, "GOVERNANCE_MIGRATION_MAKE_TARGET_MISSING") {
		t.Fatalf("governance migration output missing make target rule id:\n%s", output)
	}
}

func TestProviderContractCheckRejectsConcreteNetworkAndMissingProvider(t *testing.T) {
	fixture := providerContractFixture(t)
	fixture.WriteFile("vbad/bad.go", `package vbad

import "github.com/imajinyun/knifer-go/internal/bad"

type Client = bad.Client
`)
	fixture.WriteFile("internal/bad/client.go", `package bad

import (
	"net/http"
	"os"
)

type Client struct{}

func New() *Client {
	_ = os.Getenv("TOKEN")
	_ = http.Client{}
	return &Client{}
}
`)

	output, err := fixture.RunProviderContractCheck()
	if err == nil {
		t.Fatalf("check_provider_contracts.sh unexpectedly passed:\n%s", output)
	}
	for _, want := range []string{
		"PROVIDER_CONTRACT_MISSING_PROVIDER_INTERFACE",
		"PROVIDER_CONTRACT_FORBIDDEN_IMPORT",
		"PROVIDER_CONTRACT_FORBIDDEN_SIDE_EFFECT",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("provider contract output missing rule id %q:\n%s", want, output)
		}
	}
	for _, want := range []string{
		"must define a Provider interface contract",
		`concrete provider/network SDK dependency "net/http"`,
		"os.Getenv",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("provider contract output missing %q:\n%s", want, output)
		}
	}
}

func TestProviderContractCheckEmitsJSONFindings(t *testing.T) {
	fixture := providerContractFixture(t)
	fixture.WriteFile("vbad/bad.go", `package vbad

import "github.com/imajinyun/knifer-go/internal/bad"

type Client = bad.Client
`)
	fixture.WriteFile("internal/bad/client.go", `package bad

import "os"

type Client struct{}

func New() *Client {
	_ = os.Getenv("TOKEN")
	return &Client{}
}
`)

	output, err := fixture.RunProviderContractCheckJSON()
	if err == nil {
		t.Fatalf("providercontractcheck -json unexpectedly passed:\n%s", output)
	}
	var result struct {
		Status   string `json:"status"`
		Findings []struct {
			RuleID string `json:"rule_id"`
		} `json:"findings"`
	}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("providercontractcheck -json output is not valid JSON: %v\n%s", err, output)
	}
	if result.Status != "failed" {
		t.Fatalf("JSON status = %q, want failed\n%s", result.Status, output)
	}
	var sawMissingProvider bool
	var sawSideEffect bool
	for _, finding := range result.Findings {
		if finding.RuleID == "PROVIDER_CONTRACT_MISSING_PROVIDER_INTERFACE" {
			sawMissingProvider = true
		}
		if finding.RuleID == "PROVIDER_CONTRACT_FORBIDDEN_SIDE_EFFECT" {
			sawSideEffect = true
		}
	}
	if !sawMissingProvider || !sawSideEffect {
		t.Fatalf("provider JSON findings missing expected rule ids:\n%s", output)
	}
}

func TestProviderContractCheckAcceptsProviderOnlyContract(t *testing.T) {
	fixture := providerContractFixture(t)
	fixture.WriteFile("vbad/bad.go", `package vbad

import "github.com/imajinyun/knifer-go/internal/bad"

type Provider = bad.Provider
type Client = bad.Client
type Option = bad.Option

func WithProvider(provider Provider) Option { return bad.WithProvider(provider) }
func New(opts ...Option) *Client { return bad.New(opts...) }
`)
	fixture.WriteFile("internal/bad/client.go", `package bad

import "context"

type Provider interface {
	Run(context.Context, string) (string, error)
}

type Client struct{ provider Provider }
type Option func(*Client)

func WithProvider(provider Provider) Option {
	return func(c *Client) {
		if provider != nil {
			c.provider = provider
		}
	}
}

func New(opts ...Option) *Client {
	c := &Client{}
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}
	return c
}
`)

	output, err := fixture.RunProviderContractCheck()
	if err != nil {
		t.Fatalf("check_provider_contracts.sh failed: %v\n%s", err, output)
	}
	if !strings.Contains(output, "provider contract governance is valid") {
		t.Fatalf("provider contract output missing success:\n%s", output)
	}
}

func TestCIWorkflowCheckRejectsUnknownMakeTarget(t *testing.T) {
	fixture := ciWorkflowFixture(t, "make missing-target")
	output, err := fixture.RunCIWorkflowCheck()
	if err == nil {
		t.Fatalf("check_ci_workflows.sh unexpectedly passed:\n%s", output)
	}
	if !strings.Contains(output, "CI_WORKFLOW_UNKNOWN_MAKE_TARGET") {
		t.Fatalf("CI workflow output missing CI_WORKFLOW_UNKNOWN_MAKE_TARGET rule id:\n%s", output)
	}
	if !strings.Contains(output, "references unknown Makefile target") || !strings.Contains(output, "missing-target") {
		t.Fatalf("CI workflow output missing unknown target error:\n%s", output)
	}
}

func TestCIWorkflowCheckEmitsJSONFindings(t *testing.T) {
	fixture := ciWorkflowFixture(t, "make missing-target")
	output, err := fixture.RunCIWorkflowCheckJSON()
	if err == nil {
		t.Fatalf("ciworkflowcheck -json unexpectedly passed:\n%s", output)
	}

	var result struct {
		Status   string `json:"status"`
		Findings []struct {
			RuleID   string `json:"rule_id"`
			Path     string `json:"path"`
			Message  string `json:"message"`
			Severity string `json:"severity"`
		} `json:"findings"`
	}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("ciworkflowcheck -json output is not valid JSON: %v\n%s", err, output)
	}
	if result.Status != "failed" {
		t.Fatalf("JSON status = %q, want failed\n%s", result.Status, output)
	}
	if len(result.Findings) == 0 {
		t.Fatalf("JSON findings unexpectedly empty:\n%s", output)
	}
	finding := result.Findings[0]
	if finding.RuleID != "CI_WORKFLOW_UNKNOWN_MAKE_TARGET" {
		t.Fatalf("JSON rule_id = %q, want CI_WORKFLOW_UNKNOWN_MAKE_TARGET\n%s", finding.RuleID, output)
	}
	if finding.Path != ".github/workflows/go.yml" {
		t.Fatalf("JSON path = %q, want .github/workflows/go.yml\n%s", finding.Path, output)
	}
	if finding.Severity != "error" {
		t.Fatalf("JSON severity = %q, want error\n%s", finding.Severity, output)
	}
	if !strings.Contains(finding.Message, "missing-target") {
		t.Fatalf("JSON message missing target name:\n%s", output)
	}
}

func TestCIWorkflowCheckAcceptsDeclaredMakeTarget(t *testing.T) {
	fixture := ciWorkflowFixture(t, "make ci-agent-governance")
	output, err := fixture.RunCIWorkflowCheck()
	if err != nil {
		t.Fatalf("check_ci_workflows.sh failed: %v\n%s", err, output)
	}
	if !strings.Contains(output, "CI workflow governance is valid") {
		t.Fatalf("CI workflow output missing success:\n%s", output)
	}
}

func TestCIWorkflowCheckEmitsPassingJSON(t *testing.T) {
	fixture := ciWorkflowFixture(t, "make ci-agent-governance")
	output, err := fixture.RunCIWorkflowCheckJSON()
	if err != nil {
		t.Fatalf("ciworkflowcheck -json failed: %v\n%s", err, output)
	}

	var result struct {
		Status   string            `json:"status"`
		Findings []json.RawMessage `json:"findings"`
	}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("ciworkflowcheck -json output is not valid JSON: %v\n%s", err, output)
	}
	if result.Status != "passed" {
		t.Fatalf("JSON status = %q, want passed\n%s", result.Status, output)
	}
	if len(result.Findings) != 0 {
		t.Fatalf("JSON findings length = %d, want 0\n%s", len(result.Findings), output)
	}
}

func TestDocsQuickstartCheckRejectsMissingQualitySections(t *testing.T) {
	fixture := docsQuickstartFixture(t,
		"# vbad Quickstart\n\n"+
			"`vbad` has an intentionally incomplete quickstart.\n\n"+
			"## Which helper should I use?\n\n"+
			"Use `vbad` when testing fixture failures.\n",
	)
	output, err := fixture.RunDocsQuickstartCheck()
	if err == nil {
		t.Fatalf("check_docs_quickstart.sh unexpectedly passed:\n%s", output)
	}
	if !strings.Contains(output, "DOCS_QUICKSTART_SECTION_MISSING") {
		t.Fatalf("docs quickstart output missing DOCS_QUICKSTART_SECTION_MISSING rule id:\n%s", output)
	}
	for _, want := range []string{
		"## Golden path APIs",
		"## Benchmarks and trade-offs",
		"checklist",
		"error behavior",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("docs quickstart output missing %q:\n%s", want, output)
		}
	}
}

func TestDocsQuickstartCheckAcceptsCompleteFixture(t *testing.T) {
	fixture := docsQuickstartFixture(t,
		"# vbad Quickstart\n\n"+
			"`vbad` has a complete fixture quickstart.\n\n"+
			"## Golden path APIs\n\n"+
			"- `Run`\n\n"+
			"## Which helper should I use?\n\n"+
			"Use `Run` for fixture examples.\n\n"+
			"## Fixture correctness checklist\n\n"+
			"- Handle errors returned by `Run`.\n\n"+
			"## When not to use vbad\n\n"+
			"- Use direct code outside fixture tests.\n\n"+
			"## Related packages\n\n"+
			"- Use `vjson` when fixture data needs JSON.\n\n"+
			"## Benchmarks and trade-offs\n\n"+
			"Benchmark only if fixture code becomes hot.\n\n"+
			"## FAQ\n\n"+
			"### How do I handle errors?\n\n"+
			"Check `err != nil` and return the error to the caller.\n\n"+
			"```go\n"+
			"package main\n\n"+
			"import (\n"+
			"\t\"fmt\"\n\n"+
			"\t\"github.com/imajinyun/knifer-go/vbad\"\n"+
			")\n\n"+
			"func main() {\n"+
			"\tfmt.Println(vbad.Run())\n"+
			"}\n"+
			"```\n",
	)
	output, err := fixture.RunDocsQuickstartCheck()
	if err != nil {
		t.Fatalf("check_docs_quickstart.sh failed: %v\n%s", err, output)
	}
	if !strings.Contains(output, "quickstart docs are valid") {
		t.Fatalf("docs quickstart output missing success:\n%s", output)
	}
}

func TestDocsQuickstartCheckEmitsJSONReport(t *testing.T) {
	fixture := docsQuickstartFixture(t,
		"# vbad Quickstart\n\n"+
			"`vbad` has a complete fixture quickstart.\n\n"+
			"## Golden path APIs\n\n"+
			"- `Run`\n\n"+
			"## Which helper should I use?\n\n"+
			"Use `Run` for fixture examples.\n\n"+
			"## Fixture correctness checklist\n\n"+
			"- Handle errors returned by `Run`.\n\n"+
			"## When not to use vbad\n\n"+
			"- Use direct code outside fixture tests.\n\n"+
			"## Related packages\n\n"+
			"- Use `vjson` when fixture data needs JSON.\n\n"+
			"## Benchmarks and trade-offs\n\n"+
			"Benchmark only if fixture code becomes hot.\n\n"+
			"## FAQ\n\n"+
			"### How do I handle errors?\n\n"+
			"Check `err != nil` and return the error to the caller.\n\n"+
			"```go\n"+
			"package main\n\n"+
			"import (\n"+
			"\t\"fmt\"\n\n"+
			"\t\"github.com/imajinyun/knifer-go/vbad\"\n"+
			")\n\n"+
			"func main() {\n"+
			"\tfmt.Println(vbad.Run())\n"+
			"}\n"+
			"```\n",
	)
	output, err := fixture.RunDocsQuickstartCheckJSON()
	if err != nil {
		t.Fatalf("docsquickstartcheck -json failed: %v\n%s", err, output)
	}
	var result struct {
		Status            string            `json:"status"`
		Findings          []json.RawMessage `json:"findings"`
		PublicFacadeCount int               `json:"public_facade_count"`
	}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("docsquickstartcheck -json output is not valid JSON: %v\n%s", err, output)
	}
	if result.Status != "passed" || len(result.Findings) != 0 || result.PublicFacadeCount != 1 {
		t.Fatalf("unexpected docs quickstart JSON report:\n%s", output)
	}
}

func TestDocsQuickstartCheckAcceptsNoErrorProfile(t *testing.T) {
	fixture := docsQuickstartFixtureWithProfiles(t,
		[]string{"no_error_returning"},
		"# vbad Quickstart\n\n"+
			"`vbad` has deterministic value helpers.\n\n"+
			"## Golden path APIs\n\n"+
			"- `Run`\n\n"+
			"## Which helper should I use?\n\n"+
			"Use `Run` for fixture examples.\n\n"+
			"## Fixture correctness checklist\n\n"+
			"- Keep calls side-effect free.\n\n"+
			"## When not to use vbad\n\n"+
			"- Use direct code outside fixture tests.\n\n"+
			"## Related packages\n\n"+
			"- Use `vjson` when fixture data needs JSON.\n\n"+
			"## Benchmarks and trade-offs\n\n"+
			"Benchmark only if fixture code becomes hot.\n\n"+
			"## FAQ\n\n"+
			"### Do vbad helpers return errors?\n\n"+
			"No. Fixture helpers do not return errors.\n\n"+
			"```go\n"+
			"package main\n\n"+
			"import (\n"+
			"\t\"fmt\"\n\n"+
			"\t\"github.com/imajinyun/knifer-go/vbad\"\n"+
			")\n\n"+
			"func main() {\n"+
			"\tfmt.Println(vbad.Run())\n"+
			"}\n"+
			"```\n",
	)
	output, err := fixture.RunDocsQuickstartCheck()
	if err != nil {
		t.Fatalf("check_docs_quickstart.sh failed: %v\n%s", err, output)
	}
	if !strings.Contains(output, "quickstart docs are valid") {
		t.Fatalf("docs quickstart output missing success:\n%s", output)
	}
}

func TestDocsQuickstartCheckRejectsProviderProfileWithoutBoundary(t *testing.T) {
	fixture := docsQuickstartFixtureWithProfiles(t,
		[]string{"error_returning", "provider_contract"},
		"# vbad Quickstart\n\n"+
			"`vbad` has an incomplete provider fixture.\n\n"+
			"## Golden path APIs\n\n"+
			"- `Run`\n\n"+
			"## Which helper should I use?\n\n"+
			"Use `Run` for fixture examples.\n\n"+
			"## Fixture correctness checklist\n\n"+
			"- Handle errors returned by `Run`.\n\n"+
			"## When not to use vbad\n\n"+
			"- Use direct code outside fixture tests.\n\n"+
			"## Related packages\n\n"+
			"- Use `vjson` when fixture data needs JSON.\n\n"+
			"## Benchmarks and trade-offs\n\n"+
			"Benchmark only if fixture code becomes hot.\n\n"+
			"## FAQ\n\n"+
			"### How do I handle errors?\n\n"+
			"Check `err != nil` and return the error to the caller.\n\n"+
			"```go\n"+
			"package main\n\n"+
			"import (\n"+
			"\t\"fmt\"\n\n"+
			"\t\"github.com/imajinyun/knifer-go/vbad\"\n"+
			")\n\n"+
			"func main() {\n"+
			"\tfmt.Println(vbad.Run())\n"+
			"}\n"+
			"```\n",
	)
	output, err := fixture.RunDocsQuickstartCheck()
	if err == nil {
		t.Fatalf("check_docs_quickstart.sh unexpectedly passed:\n%s", output)
	}
	if !strings.Contains(output, "DOCS_QUICKSTART_PROFILE_GUIDANCE_MISSING") {
		t.Fatalf("docs quickstart output missing DOCS_QUICKSTART_PROFILE_GUIDANCE_MISSING rule id:\n%s", output)
	}
	if !strings.Contains(output, "profile provider_contract") {
		t.Fatalf("docs quickstart output missing provider profile error:\n%s", output)
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
	structuredChecks := map[string]any{
		"change_policy_check": map[string]any{
			"cmd":       "go run ./bin/changepolicycheck -root /fixture -json",
			"exit_code": 0,
			"json": map[string]any{
				"detected_policies": []string{"security_sensitive"},
				"findings":          []any{},
				"policy_paths": map[string]any{
					"security_sensitive": []string{"internal/db/scan.go"},
				},
				"required_commands": []string{"change_policy_check", "security_sensitive_diff", "agent_full_check", "agent_security_check", "agent_evidence", "agent_evidence_check"},
				"rule_ids":          []string{"CHANGE_SECURITY_SENSITIVE"},
				"semantic_rule_ids": []string{},
				"status":            "passed",
			},
			"status": "passed",
			"stderr": "",
			"stdout": `{
  "status": "passed",
  "findings": [],
  "detected_policies": ["security_sensitive"],
  "rule_ids": ["CHANGE_SECURITY_SENSITIVE"],
  "semantic_rule_ids": [],
  "required_commands": ["change_policy_check", "security_sensitive_diff", "agent_full_check", "agent_security_check", "agent_evidence", "agent_evidence_check"],
  "policy_paths": {"security_sensitive": ["internal/db/scan.go"]}
}`,
		},
		"ci_workflow_check": map[string]any{
			"cmd":       "go run ./bin/ciworkflowcheck -root /fixture -json",
			"exit_code": 0,
			"json": map[string]any{
				"findings": []any{},
				"status":   "passed",
			},
			"status": "passed",
			"stderr": "",
			"stdout": `{
  "status": "passed",
  "findings": []
}`,
		},
		"random_source_policy_check": structuredPassedCheck("go run ./bin/randomsourcepolicycheck -root /fixture -json"),
		"threat_model_check":         structuredPassedCheck("go run ./bin/threatmodelcheck -root /fixture -json"),
		"dynamic_contracts_check":    structuredPassedCheck("go run ./bin/dynamiccontractscheck -root /fixture -json"),
		"error_model_check":          structuredPassedCheck("go run ./bin/errormodelcheck -root /fixture -json"),
		"api_convergence_check":      structuredPassedCheck("go run ./bin/apiconvergencecheck -root /fixture -json"),
		"lifecycle_check":            structuredPassedCheck("go run ./bin/lifecyclecheck -root /fixture -json"),
		"dependency_tiers_check":     structuredPassedCheck("go run ./bin/dependencytierscheck -root /fixture -json"),
		"benchmark_regression_check": structuredPassedCheck("go run ./bin/benchmarkregressioncheck -root /fixture -json"),
		"provider_contract_check":    structuredPassedCheck("go run ./bin/providercontractcheck -root /fixture -json"),
		"arch_imports_check":         structuredPassedCheck("go run ./bin/archimportscheck -root /fixture -json"),
		"panic_policy_check":         structuredPassedCheck("go run ./bin/panicpolicycheck -root /fixture -json"),
		"facade_boundary_check":      structuredPassedCheck("go run ./bin/facadeboundarycheck -root /fixture -json"),
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
		"structured_checks":        structuredChecks,
		"worktree_status":          " M internal/db/scan.go",
	}
}

func structuredPassedCheck(cmd string) map[string]any {
	return map[string]any{
		"cmd":       cmd,
		"exit_code": 0,
		"json": map[string]any{
			"findings": []any{},
			"status":   "passed",
		},
		"status": "passed",
		"stderr": "",
		"stdout": `{
  "status": "passed",
  "findings": []
}`,
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

func providerContractFixture(t *testing.T) *governanceFixture {
	t.Helper()
	fixture := newGovernanceFixture(t)
	fixture.WriteJSON("ai-context.json", map[string]any{
		"public_facades": []any{
			map[string]any{"package": "vbad", "internal": "internal/bad"},
		},
		"dependency_tiers": map[string]any{
			"provider_contract_facades": []string{"vbad"},
		},
	})
	return fixture
}

func randomSourcePolicyFixture(t *testing.T) *governanceFixture {
	t.Helper()
	fixture := newGovernanceFixture(t)
	fixture.WriteJSON("ai-context.json", randomSourcePolicyContext())
	fixture.WriteFile("internal/rand/random_bytes_test.go", `package rand

func TestSecureRandomBytesFailClosed() {}
func TestRandomBytesWithOptionsReaderAndStrictMode() {}
func TestFillRandomBytesFallbackKeepsLength() {}
`)
	fixture.WriteFile("internal/crypto/random_test.go", `package crypto

func TestRandomProviderFallbacksAndErrors() {}
`)
	fixture.WriteFile("vrand/rand_test.go", `package vrand

func TestRandFacadeBytesFailureBoundaries() {}
`)
	fixture.WriteFile("internal/id/uuid_test.go", `package id

func TestDefaultFallbackRandomSourceProviderCanBeConfiguredAndReset() {}
`)
	fixture.WriteFile("vid/random_source_test.go", `package vid

func TestIDFacadeFallbackRandomSourceProvider() {}
func TestIDFacadeRandomFallbackBoundaries() {}
`)
	fixture.WriteFile("internal/jwt/jwt_key_set_test.go", `package jwt

func TestSetKeyRejectsNone() {}
`)
	fixture.WriteFile("internal/jwt/signer_test.go", `package jwt

func TestHMACSignerStrictRejectsWeakKeys() {}
`)
	fixture.WriteFile("vjwt/signer_hmac_test.go", `package vjwt

func TestStrictHMACSignerRejectsWeakKey() {}
`)
	fixture.WriteFile("internal/crypto/rsa_test.go", `package crypto

func TestRSANilKeyAndProviderErrorContracts() {}
`)
	return fixture
}

func randomSourcePolicyContext() map[string]any {
	return map[string]any{
		"random_source_policy": map[string]any{
			"packages": []string{"vrand", "vid", "vcrypto", "vjwt"},
			"policies": []any{
				map[string]any{
					"name":            "secure_bytes_fail_closed",
					"packages":        []string{"vrand", "vcrypto"},
					"behavior":        "Security-sensitive byte helpers fail closed.",
					"allowed_sources": []string{"crypto/rand.Reader"},
					"forbidden_uses":  []string{"tokens from math/rand"},
					"contract_tests":  []string{"internal/rand/random_bytes_test.go:TestSecureRandomBytesFailClosed", "internal/crypto/random_test.go:TestRandomProviderFallbacksAndErrors"},
				},
				map[string]any{
					"name":            "compatibility_byte_fallback",
					"packages":        []string{"vrand"},
					"behavior":        "Compatibility byte helpers may fall back outside strict mode.",
					"allowed_sources": []string{"WithRandomReader"},
					"forbidden_uses":  []string{"secrets"},
					"contract_tests":  []string{"internal/rand/random_bytes_test.go:TestRandomBytesWithOptionsReaderAndStrictMode", "internal/rand/random_bytes_test.go:TestFillRandomBytesFallbackKeepsLength", "vrand/rand_test.go:TestRandFacadeBytesFailureBoundaries"},
				},
				map[string]any{
					"name":            "identifier_fallback_compatibility",
					"packages":        []string{"vid"},
					"behavior":        "Identifier helpers may use compatibility fallback.",
					"allowed_sources": []string{"WithFallbackRandomSource"},
					"forbidden_uses":  []string{"security-sensitive bearer tokens"},
					"contract_tests":  []string{"internal/id/uuid_test.go:TestDefaultFallbackRandomSourceProviderCanBeConfiguredAndReset", "vid/random_source_test.go:TestIDFacadeFallbackRandomSourceProvider", "vid/random_source_test.go:TestIDFacadeRandomFallbackBoundaries"},
				},
				map[string]any{
					"name":            "jwt_key_and_signer_policy",
					"packages":        []string{"vjwt", "vcrypto"},
					"behavior":        "JWT and crypto helpers reject unsafe algorithms.",
					"allowed_sources": []string{"WithSignerRandomReader"},
					"forbidden_uses":  []string{"alg=none signer fallback"},
					"contract_tests":  []string{"internal/jwt/jwt_key_set_test.go:TestSetKeyRejectsNone", "internal/jwt/signer_test.go:TestHMACSignerStrictRejectsWeakKeys", "vjwt/signer_hmac_test.go:TestStrictHMACSignerRejectsWeakKey", "internal/crypto/rsa_test.go:TestRSANilKeyAndProviderErrorContracts"},
				},
			},
		},
	}
}

func threatModelFixture(t *testing.T) *governanceFixture {
	t.Helper()
	fixture := newGovernanceFixture(t)
	fixture.WriteJSON("ai-context.json", threatModelContext())
	for path, functions := range threatModelTestFunctions() {
		var body strings.Builder
		body.WriteString("package fixture\n\n")
		for _, fn := range functions {
			body.WriteString("func ")
			body.WriteString(fn)
			body.WriteString("() {}\n")
		}
		fixture.WriteFile(path, body.String())
	}
	return fixture
}

func threatModelContext() map[string]any {
	publicFacades := []string{"vhttp", "vresty", "vurl", "vconf"}
	var publicFacadeEntries []any
	for _, pkg := range publicFacades {
		publicFacadeEntries = append(publicFacadeEntries, map[string]any{"package": pkg})
	}
	return map[string]any{
		"public_facades": publicFacadeEntries,
		"threat_model": map[string]any{
			"boundary_contracts": []any{
				threatBoundary("default_timeout", []string{"vhttp", "vresty"}, []string{"positive default timeout", "timeout errors classified as GK_TIMEOUT"}, []string{"internal/httpx/contract_request_test.go:TestHTTPContractDefaultTimeout", "internal/httpx/http/request_timeout_redirect_test.go:TestRequestTimeout"}),
				threatBoundary("redirect_revalidation", []string{"vhttp", "vresty", "vurl", "vconf"}, []string{"redirect target URL policy", "round-trip host revalidation"}, []string{"internal/httpx/http/safe_request_test.go:TestSafeRequestRejectsPrivateAndUnsafeRedirects", "internal/conf/load_remote_safe_hosts_test.go:TestLoadRemoteSafeRejectsUnsafeRedirectTarget"}),
				threatBoundary("private_host_rejection", []string{"vhttp", "vresty", "vurl", "vconf"}, []string{"loopback rejection", "multicast rejection"}, []string{"internal/httpx/http/safe_request_test.go:TestSafeRequestAllowedHostsDoesNotBypassPrivateRejection", "internal/url/safe_resource_allowlist_test.go:TestOpenSafeAllowedHostsDoesNotBypassPrivateRejection"}),
				threatBoundary("bounded_response_reads", []string{"vhttp", "vresty", "vurl", "vconf"}, []string{"max response bytes", "ContentLength precheck"}, []string{"internal/httpx/contract_response_limits_test.go:TestHTTPContractMaxResponseBytes", "internal/url/safe_resource_limit_test.go:TestOpenSafeMaxBytes"}),
				threatBoundary("safe_download_paths", []string{"vhttp", "vresty", "vurl"}, []string{"safe filename normalization", "download path scoping"}, []string{"internal/httpx/internal/shared/util_test.go:TestSafeDownloadedFilenameAndJoin", "internal/httpx/http/save_as_content_disposition_test.go:TestSaveAsRejectsUnsafeContentDispositionFilename"}),
				threatBoundary("remote_config_boundary", []string{"vconf"}, []string{"safe remote loader", "allowed host policy"}, []string{"vconf/load_remote_safe_test.go:TestFacadeRemoteSafeWrappers", "internal/conf/load_remote_safe_roundtrip_test.go:TestLoadRemoteSafeRevalidatesHostAtRoundTrip"}),
			},
		},
	}
}

func threatBoundary(name string, packages, controls, tests []string) map[string]any {
	return map[string]any{
		"name":              name,
		"packages":          packages,
		"required_controls": controls,
		"contract_tests":    tests,
	}
}

func threatModelTestFunctions() map[string][]string {
	return map[string][]string{
		"internal/httpx/contract_request_test.go":                 {"TestHTTPContractDefaultTimeout"},
		"internal/httpx/http/request_timeout_redirect_test.go":    {"TestRequestTimeout"},
		"internal/httpx/http/safe_request_test.go":                {"TestSafeRequestRejectsPrivateAndUnsafeRedirects", "TestSafeRequestAllowedHostsDoesNotBypassPrivateRejection"},
		"internal/conf/load_remote_safe_hosts_test.go":            {"TestLoadRemoteSafeRejectsUnsafeRedirectTarget"},
		"internal/url/safe_resource_allowlist_test.go":            {"TestOpenSafeAllowedHostsDoesNotBypassPrivateRejection"},
		"internal/httpx/contract_response_limits_test.go":         {"TestHTTPContractMaxResponseBytes"},
		"internal/url/safe_resource_limit_test.go":                {"TestOpenSafeMaxBytes"},
		"internal/httpx/internal/shared/util_test.go":             {"TestSafeDownloadedFilenameAndJoin"},
		"internal/httpx/http/save_as_content_disposition_test.go": {"TestSaveAsRejectsUnsafeContentDispositionFilename"},
		"vconf/load_remote_safe_test.go":                          {"TestFacadeRemoteSafeWrappers"},
		"internal/conf/load_remote_safe_roundtrip_test.go":        {"TestLoadRemoteSafeRevalidatesHostAtRoundTrip"},
	}
}

func dynamicContractsFixture(t *testing.T) *governanceFixture {
	t.Helper()
	fixture := newGovernanceFixture(t)
	fixture.WriteJSON("ai-context.json", dynamicContractsContext())
	for _, dir := range []string{"internal/bean", "internal/conv", "internal/ref", "vjson", "vobj", "vconf", "vconv", "vref"} {
		fixture.WriteFile(dir+"/doc.go", "package fixture\n")
	}
	for path, functions := range dynamicContractTestFunctions() {
		var body strings.Builder
		body.WriteString("package fixture\n\n")
		for _, fn := range functions {
			body.WriteString("func ")
			body.WriteString(fn)
			body.WriteString("() {}\n")
		}
		fixture.WriteFile(path, body.String())
	}
	return fixture
}

func dynamicContractsContext() map[string]any {
	return map[string]any{
		"dynamic_semantic_contracts": map[string]any{
			"required_domains": []string{"vbean_decode_copy_merge", "vjson_dynamic", "vobj_dynamic", "vconf_dynamic", "vconv_conversion_matrix", "vref_reflection_boundaries"},
			"domains": map[string]any{
				"vbean_decode_copy_merge":    dynamicDomain([]string{"internal/bean"}, []string{"embedded pointer nil handling", "unused field reporting", "map replace semantics"}, []string{"internal/bean/copy_options_test.go:TestDecodeContractEmbeddedPointerNilAndUnused"}, nil),
				"vjson_dynamic":              dynamicDomain([]string{"vjson"}, []string{"scalar conversion", "path lookup", "invalid input errors"}, []string{"vjson/json_conversion_test.go:TestDynamicJSONContractMatrix"}, []string{"vjson/json_conversion_test.go:FuzzDynamicJSONStringContract"}),
				"vobj_dynamic":               dynamicDomain([]string{"vobj"}, []string{"map serialization", "clone semantics", "nil predicates"}, []string{"vobj/serialization_test.go:TestDynamicObjectContractMatrix"}, []string{"vobj/serialization_test.go:FuzzDynamicObjectStringContract"}),
				"vconf_dynamic":              dynamicDomain([]string{"vconf"}, []string{"profile lookup", "scalar conversion", "missing defaults"}, []string{"vconf/config_bind_options_test.go:TestDynamicConfigContractMatrix"}, []string{"vconf/config_bind_options_test.go:FuzzDynamicConfigScalarContract"}),
				"vconv_conversion_matrix":    dynamicDomain([]string{"internal/conv", "vconv"}, []string{"nil conversion", "scalar parsing", "overflow rejection"}, []string{"internal/conv/conversion_matrix_test.go:TestConversionMatrixPropertyContract", "vconv/conv_test.go:TestConvFacadeConversionMatrix"}, []string{"internal/conv/conversion_matrix_test.go:FuzzConversionMatrixStringScalars"}),
				"vref_reflection_boundaries": dynamicDomain([]string{"internal/ref", "vref"}, []string{"unsafe opt-in", "numeric boundaries", "invalid reflection input"}, []string{"internal/ref/checked_convert_test.go:TestCheckedConvertNumericBoundaries", "vref/ref_field_test.go:TestFacadeReflectionHelpers"}, nil),
			},
		},
	}
}

func dynamicDomain(packages, guarantees, contractTests, fuzzTests []string) map[string]any {
	return map[string]any{
		"packages":       packages,
		"guarantees":     guarantees,
		"contract_tests": contractTests,
		"fuzz_tests":     fuzzTests,
	}
}

func dynamicContractTestFunctions() map[string][]string {
	return map[string][]string{
		"internal/bean/copy_options_test.go":      {"TestDecodeContractEmbeddedPointerNilAndUnused"},
		"vjson/json_conversion_test.go":           {"TestDynamicJSONContractMatrix", "FuzzDynamicJSONStringContract"},
		"vobj/serialization_test.go":              {"TestDynamicObjectContractMatrix", "FuzzDynamicObjectStringContract"},
		"vconf/config_bind_options_test.go":       {"TestDynamicConfigContractMatrix", "FuzzDynamicConfigScalarContract"},
		"internal/conv/conversion_matrix_test.go": {"TestConversionMatrixPropertyContract", "FuzzConversionMatrixStringScalars"},
		"vconv/conv_test.go":                      {"TestConvFacadeConversionMatrix"},
		"internal/ref/checked_convert_test.go":    {"TestCheckedConvertNumericBoundaries"},
		"vref/ref_field_test.go":                  {"TestFacadeReflectionHelpers"},
	}
}

func errorModelFixture(t *testing.T) *governanceFixture {
	t.Helper()
	fixture := newGovernanceFixture(t)
	fixture.WriteJSON("ai-context.json", errorModelContext())
	fixture.WriteFile("errors.go", `package knifer

const (
	ErrCodeInvalidInput     = "GK_INVALID_INPUT"
	ErrCodeNotFound         = "GK_NOT_FOUND"
	ErrCodeUnsupported      = "GK_UNSUPPORTED"
	ErrCodeUnsafeResource   = "GK_UNSAFE_RESOURCE"
	ErrCodeTimeout          = "GK_TIMEOUT"
	ErrCodeProviderFailure  = "GK_PROVIDER_FAILURE"
	ErrCodeInternal         = "GK_INTERNAL"
)
`)
	fixture.WriteFile("errors_test.go", `package knifer

func TestUnifiedErrorTaxonomyCodes() {}
`)
	fixture.WriteFile("internal/bean/error_contract_test.go", `package bean

func TestBeanErrorContract() {}
`)
	fixture.WriteFile("internal/db/error_contract_test.go", `package db

func TestDBErrorContract() {}
`)
	fixture.WriteFile("vdb/error_contract_test.go", `package vdb

func TestVDBErrorContract() {}
`)
	return fixture
}

func errorModelContext() map[string]any {
	return map[string]any{
		"error_model": map[string]any{
			"taxonomy": []any{
				errorTaxonomy("invalid input", "GK_INVALID_INPUT", "invalid caller input"),
				errorTaxonomy("not found", "GK_NOT_FOUND", "missing resource"),
				errorTaxonomy("unsupported type", "GK_UNSUPPORTED", "unsupported operation"),
				errorTaxonomy("unsafe resource", "GK_UNSAFE_RESOURCE", "unsafe resource"),
				errorTaxonomy("timeout", "GK_TIMEOUT", "deadline exceeded"),
				errorTaxonomy("provider failure", "GK_PROVIDER_FAILURE", "provider failed"),
				errorTaxonomy("internal", "GK_INTERNAL", "internal failure"),
			},
			"contract_tests": []string{
				"errors_test.go:TestUnifiedErrorTaxonomyCodes",
				"internal/bean/error_contract_test.go:TestBeanErrorContract",
				"internal/db/error_contract_test.go:TestDBErrorContract",
				"vdb/error_contract_test.go:TestVDBErrorContract",
			},
		},
	}
}

func errorTaxonomy(category, code, useWhen string) map[string]any {
	return map[string]any{
		"category": category,
		"code":     code,
		"use_when": useWhen,
	}
}

func benchmarkRegressionFixture(t *testing.T) *governanceFixture {
	t.Helper()
	fixture := newGovernanceFixture(t)
	fixture.WriteJSON("ai-context.json", benchmarkRegressionContext())
	fixture.WriteFile("Makefile", benchmarkRegressionMakefileWithout(""))
	for _, dir := range []string{
		"internal/slice", "internal/maps", "internal/str", "internal/num", "internal/bean", "internal/json", "internal/db", "internal/httpx/http", "internal/codec",
		"vbean", "vjson", "vmap", "vslice", "vstr", "vdb", "vhttp", "vcodec",
	} {
		fixture.WriteFile(dir+"/doc.go", "package fixture\n")
	}
	for path, functions := range benchmarkRegressionFunctions() {
		var body strings.Builder
		body.WriteString("package fixture\n\nimport \"testing\"\n\n")
		for _, fn := range functions {
			body.WriteString("func ")
			body.WriteString(fn)
			body.WriteString("(b *testing.B) {}\n")
		}
		fixture.WriteFile(path, body.String())
	}
	return fixture
}

func benchmarkRegressionContext() map[string]any {
	hotPaths := []any{
		hotPath("./vjson", "data_transform", []string{"vjson/json_benchmark_test.go:BenchmarkParseObj"}),
		hotPath("./vstr", "text_parsing", []string{"vstr/str_benchmark_test.go:BenchmarkReverse"}),
		hotPath("./vslice", "collections", []string{"vslice/slice_benchmark_test.go:BenchmarkFilter"}),
		hotPath("./vmap", "collections", []string{"vmap/maps_benchmark_test.go:BenchmarkFilter"}),
		hotPath("./vdb", "data_and_cli_boundaries", []string{"vdb/builder_test.go:BenchmarkFacadePageOrders"}),
		hotPath("./vhttp", "network_clients", []string{"vhttp/http_benchmark_test.go:BenchmarkGetStringE"}),
		hotPath("./vcodec", "data_transform", []string{"vcodec/codec_benchmark_test.go:BenchmarkBase64Encode"}),
	}
	return map[string]any{
		"benchmark_regression": map[string]any{
			"baseline_command":   "make bench-baseline BENCHCOUNT=10 BENCHTIME=3s",
			"compare_command":    "make bench-compare BENCHCOUNT=10 BENCHTIME=3s",
			"benchstat_required": true,
			"thresholds": map[string]any{
				"ns_per_op_regression_percent":     10,
				"bytes_per_op_regression_percent":  10,
				"allocs_per_op_regression_percent": 5,
				"minimum_count":                    10,
			},
			"tracked_packages":  []string{"./internal/slice", "./internal/maps", "./internal/str", "./internal/num", "./internal/bean", "./internal/json", "./internal/db", "./internal/httpx/http", "./internal/codec", "./vbean", "./vjson", "./vmap", "./vslice", "./vstr", "./vdb", "./vhttp", "./vcodec"},
			"hot_path_packages": hotPaths,
		},
	}
}

func hotPath(pkg, owner string, benchmarks []string) map[string]any {
	return map[string]any{
		"package":           pkg,
		"owner":             owner,
		"benchmarks":        benchmarks,
		"threshold_profile": "default",
	}
}

func benchmarkRegressionMakefileWithout(excluded string) string {
	benchPkgs := []string{"./internal/slice", "./internal/maps", "./internal/str", "./internal/num", "./internal/bean", "./internal/json", "./internal/db", "./internal/httpx/http", "./internal/codec"}
	benchFacadePkgs := []string{"./vbean", "./vjson", "./vmap", "./vslice", "./vstr", "./vdb", "./vhttp", "./vcodec"}
	benchCodecPkgs := []string{}
	remove := func(values []string) []string {
		var out []string
		for _, value := range values {
			if value != excluded {
				out = append(out, value)
			}
		}
		return out
	}
	return "BENCH_PKGS ?= " + strings.Join(remove(benchPkgs), " ") + "\n" +
		"BENCH_FACADE_PKGS ?= " + strings.Join(remove(benchFacadePkgs), " ") + "\n" +
		"BENCH_CODEC_PKGS ?= " + strings.Join(remove(benchCodecPkgs), " ") + "\n\n" +
		"bench-baseline:\n\t@true\n\n" +
		"bench-compare:\n\t@true\n\n" +
		"bench-regression-check:\n\t@true\n\n" +
		"benchstat:\n\t@true\n"
}

func benchmarkRegressionMakefileWithoutTarget(excludedTarget string) string {
	var body strings.Builder
	body.WriteString("BENCH_PKGS ?= ./internal/slice ./internal/maps ./internal/str ./internal/num ./internal/bean ./internal/json ./internal/db ./internal/httpx/http ./internal/codec\n")
	body.WriteString("BENCH_FACADE_PKGS ?= ./vbean ./vjson ./vmap ./vslice ./vstr ./vdb ./vhttp ./vcodec\n")
	body.WriteString("BENCH_CODEC_PKGS ?=\n\n")
	for _, target := range []string{"bench-baseline", "bench-compare", "bench-regression-check", "benchstat"} {
		if target == excludedTarget {
			continue
		}
		body.WriteString(target)
		body.WriteString(":\n\t@true\n\n")
	}
	return body.String()
}

func benchmarkRegressionFunctions() map[string][]string {
	return map[string][]string{
		"vjson/json_benchmark_test.go":   {"BenchmarkParseObj"},
		"vstr/str_benchmark_test.go":     {"BenchmarkReverse"},
		"vslice/slice_benchmark_test.go": {"BenchmarkFilter"},
		"vmap/maps_benchmark_test.go":    {"BenchmarkFilter"},
		"vdb/builder_test.go":            {"BenchmarkFacadePageOrders"},
		"vhttp/http_benchmark_test.go":   {"BenchmarkGetStringE"},
		"vcodec/codec_benchmark_test.go": {"BenchmarkBase64Encode"},
	}
}

func apiConvergenceFixture(t *testing.T) *governanceFixture {
	t.Helper()
	fixture := newGovernanceFixture(t)
	fixture.WriteJSON("ai-context.json", apiConvergenceContext())
	fixture.WriteJSON("docs/api/tools.json", apiConvergenceTools())
	return fixture
}

func apiConvergenceContext() map[string]any {
	return map[string]any{
		"public_facades": []any{
			map[string]any{"package": "vjson"},
			map[string]any{"package": "vstr"},
		},
		"api_convergence": map[string]any{
			"max_golden_path_per_facade": 5,
			"required_classifications":   []string{"primary", "advanced", "compatibility", "avoid"},
			"facades": map[string]any{
				"vjson": map[string]any{
					"primary":       []string{"Parse", "Format"},
					"advanced":      []string{},
					"compatibility": []string{"Legacy"},
					"avoid":         []string{},
					"decision":      "Use primary JSON helpers.",
				},
				"vstr": map[string]any{
					"primary":       []string{"Reverse"},
					"advanced":      []string{},
					"compatibility": []string{},
					"avoid":         []string{},
					"decision":      "Use primary string helpers.",
				},
			},
		},
	}
}

func apiConvergenceTools() map[string]any {
	return map[string]any{
		"packages": []any{
			map[string]any{
				"name": "vjson",
				"golden_path": []any{
					map[string]any{"name": "Parse"},
					map[string]any{"name": "Format"},
				},
				"functions": []any{
					map[string]any{"name": "Parse", "status": "recommended"},
					map[string]any{"name": "Format", "status": "recommended"},
					map[string]any{"name": "Legacy", "status": "compatibility"},
				},
			},
			map[string]any{
				"name": "vstr",
				"golden_path": []any{
					map[string]any{"name": "Reverse"},
				},
				"functions": []any{
					map[string]any{"name": "Reverse", "status": "recommended"},
				},
			},
		},
	}
}

func lifecycleFixture(t *testing.T) *governanceFixture {
	t.Helper()
	fixture := newGovernanceFixture(t)
	fixture.WriteJSON("ai-context.json", lifecycleContext())
	return fixture
}

func lifecycleContext() map[string]any {
	publicFacades := []string{"vai", "vjson", "vimg"}
	var publicFacadeEntries []any
	for _, pkg := range publicFacades {
		publicFacadeEntries = append(publicFacadeEntries, map[string]any{"package": pkg})
	}
	return map[string]any{
		"public_facades": publicFacadeEntries,
		"dependency_tiers": map[string]any{
			"heavy_extension_facades":   []string{"vimg"},
			"provider_contract_facades": []string{"vai"},
			"core_facades":              []string{"vjson"},
		},
		"package_lifecycle": map[string]any{
			"allowed_grades": []string{"core", "stable", "maintenance", "adapter", "heavy", "candidate-for-split", "candidate-for-deprecation"},
			"packages": map[string]any{
				"vai":   map[string]any{"grade": "adapter", "rationale": "Provider contract facade."},
				"vjson": map[string]any{"grade": "core", "rationale": "Core JSON facade."},
				"vimg":  map[string]any{"grade": "heavy", "rationale": "Heavy image facade."},
			},
		},
	}
}

func dependencyTiersFixture(t *testing.T) *governanceFixture {
	t.Helper()
	fixture := newGovernanceFixture(t)
	fixture.WriteJSON("ai-context.json", dependencyTiersContext())
	for _, dir := range []string{"vjson", "vai", "vimg", "internal/json", "internal/ai", "internal/imgx"} {
		fixture.WriteFile(dir+"/doc.go", "package fixture\n")
	}
	return fixture
}

func dependencyTiersContext() map[string]any {
	return map[string]any{
		"public_facades": []any{
			map[string]any{"package": "vjson", "internal": "internal/json"},
			map[string]any{"package": "vai", "internal": "internal/ai"},
			map[string]any{"package": "vimg", "internal": "internal/imgx"},
		},
		"dependency_tiers": map[string]any{
			"core_facades":              []string{"vjson"},
			"provider_contract_facades": []string{"vai"},
			"heavy_extension_facades":   []string{"vimg"},
			"heavy_dependency_allowlist": map[string]any{
				"example.com/heavy": []string{"internal/imgx", "vimg"},
			},
		},
	}
}

func governanceMigrationFixture(t *testing.T) *governanceFixture {
	t.Helper()
	fixture := newGovernanceFixture(t)
	fixture.WriteFile("bin/check_governance_maturity.sh", "# migrated rules stay out of Python\n")
	fixture.WriteFile("Makefile", `governance-maturity-check:
	bash bin/check_governance_maturity.sh
	$(MAKE) random-source-policy-check
	$(MAKE) threat-model-check
	$(MAKE) dynamic-contracts-check
	$(MAKE) error-model-check
	$(MAKE) api-convergence-check
	$(MAKE) lifecycle-check
	$(MAKE) dependency-tiers-check

bench-regression-check:
	go run ./bin/benchmarkregressioncheck -root .
`)
	return fixture
}

func ciWorkflowFixture(t *testing.T, workflowCommand string) *governanceFixture {
	t.Helper()
	fixture := newGovernanceFixture(t)
	fixture.WriteFile("Makefile", "ci-agent-governance:\n\t@true\n")
	fixture.WriteJSON("ai-context.json", map[string]any{
		"commands": map[string]any{
			"ci_agent_governance": map[string]any{"cmd": "make ci-agent-governance"},
		},
		"ci_workflows": map[string]any{
			"tool_versions": map[string]any{
				"go_1_25_patch": "1.25.11",
				"golangci_lint": "v2.12.2",
			},
			"github_actions": map[string]any{
				"go": map[string]any{
					"path":          ".github/workflows/go.yml",
					"required_jobs": []string{"agent-governance"},
					"agent_governance": map[string]any{
						"required_commands":  []string{"make ci-agent-governance"},
						"required_env":       []string{"AGENT_CHANGE_BASE_REF", "AGENT_EVIDENCE_FILE"},
						"required_artifacts": []string{"agent-validation-evidence"},
					},
				},
			},
		},
	})
	fixture.WriteFile(".github/workflows/go.yml", `name: fixture

env:
  GOLANGCI_LINT_VERSION: v2.12.2
  GO_1_25_PATCH_VERSION: "1.25.11"

jobs:
  agent-governance:
    runs-on: ubuntu-latest
    steps:
      - run: echo AGENT_CHANGE_BASE_REF AGENT_EVIDENCE_FILE agent-validation-evidence
      - run: make ci-agent-governance
      - run: `+workflowCommand+`
  test:
    strategy:
      matrix:
        go-version: ["1.25.11", "1.26"]
`)
	return fixture
}

func writeSemanticChangePolicyFixture(fixture *governanceFixture) {
	fixture.t.Helper()
	fixture.WriteFile("Makefile", "quick-check:\n\t@true\n\nfull-check:\n\t@true\n\nrelease-check:\n\t$(MAKE) full-check\n")
	fixture.WriteFile("ai-context.json", `{
  "change_type_policies": {
    "bug_fix": {"required_commands": ["agent_check"]},
    "ci_governance": {"required_commands": ["agent_check"]},
    "dependency_change": {"required_commands": ["agent_check"]},
    "documentation": {"required_commands": ["agent_check"]},
    "internal_refactor": {"required_commands": ["agent_check"]},
    "public_api": {"required_commands": ["agent_check"]},
    "security_sensitive": {"required_commands": ["agent_security_check"]}
  },
  "commands": {
    "agent_check": {"cmd": "make agent-check"}
  },
  "api_freeze": {
    "v1_candidate": true
  },
  "coverage_gates": {
    "repository_threshold": 75.2
  },
  "security_sensitive_packages": ["vcrypto"],
  "random_source_policy": {
    "policies": []
  },
  "threat_model": {
    "boundary_contracts": []
  }
}
`)
}

func changePolicySemanticDiffFixture() string {
	return `diff --git a/ai-context.json b/ai-context.json
--- a/ai-context.json
+++ b/ai-context.json
@@ -15,1 +15,1 @@
-    "v1_candidate": true
+    "v1_candidate": false
@@ -18,1 +18,1 @@
-    "repository_threshold": 75.2
+    "repository_threshold": 76.0
@@ -20,1 +20,1 @@
-  "security_sensitive_packages": ["vcrypto"],
+  "security_sensitive_packages": ["vcrypto", "vjwt"],
@@ -22,1 +22,1 @@
-    "policies": []
+    "policies": [{"name": "crypto_reader"}]
@@ -25,1 +25,1 @@
-    "boundary_contracts": []
+    "boundary_contracts": [{"name": "timeout"}]
diff --git a/Makefile b/Makefile
--- a/Makefile
+++ b/Makefile
@@ -8,1 +8,1 @@
-	$(MAKE) full-check
+	COVERAGE_CHECK_ALL_PACKAGES=1 $(MAKE) full-check
`
}

func stringSliceContains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func docsQuickstartFixture(t *testing.T, quickstart string) *governanceFixture {
	t.Helper()
	return docsQuickstartFixtureWithProfiles(t, []string{"error_returning"}, quickstart)
}

func docsQuickstartFixtureWithProfiles(t *testing.T, profiles []string, quickstart string) *governanceFixture {
	t.Helper()
	fixture := newGovernanceFixture(t)
	fixture.WriteJSON("ai-context.json", map[string]any{
		"public_facades": []any{
			map[string]any{"package": "vbad"},
		},
		"docs_quality_profiles": map[string]any{
			"allowed_profiles": []string{"error_returning", "no_error_returning", "security_sensitive", "provider_contract", "heavy_extension"},
			"packages": map[string]any{
				"vbad": profiles,
			},
		},
	})
	fixture.WriteFile("docs/doc/README.md", "- [vbad](01-vbad.md)\n")
	fixture.WriteFile("docs/doc/01-vbad.md", quickstart)
	return fixture
}
