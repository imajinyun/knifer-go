package main

import (
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
		"must define a Provider interface contract",
		`concrete provider/network SDK dependency "net/http"`,
		"os.Getenv",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("provider contract output missing %q:\n%s", want, output)
		}
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
