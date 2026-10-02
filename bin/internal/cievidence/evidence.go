// Package cievidence records and validates command results for one CI candidate.
package cievidence

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/imajinyun/knifer-go/bin/internal/toolchainpolicy"
)

// Candidate identifies the immutable checkout and workflow execution.
type Candidate struct {
	Commit        string `json:"commit"`
	Tree          string `json:"tree"`
	ModulesSHA256 string `json:"modules_sha256"`
	RunID         string `json:"run_id"`
	Attempt       string `json:"run_attempt"`
	Base          string `json:"base_ref"`
}

// Record is emitted by the command runner, including unsuccessful executions.
type Record struct {
	Schema      int       `json:"schema_version"`
	ID          string    `json:"id"`
	Before      Candidate `json:"before"`
	After       Candidate `json:"after"`
	Command     []string  `json:"command"`
	GoVersion   string    `json:"go_version"`
	GoToolchain string    `json:"go_toolchain"`
	GoWork      string    `json:"go_work"`
	Status      string    `json:"status"`
	ExitCode    *int      `json:"exit_code"`
	StartedAt   time.Time `json:"started_at"`
	FinishedAt  time.Time `json:"finished_at"`
	Error       string    `json:"error,omitempty"`
}

// JobResult is the authoritative result supplied by GitHub's needs context.
type JobResult struct {
	Result string `json:"result"`
}

// Requirement fixes the expected command and toolchain for one result artifact.
type Requirement struct {
	ID, Job, GoVersion string
	Command            []string
}

// Attestation maps successful real CI jobs to an existing validation contract.
type Attestation struct {
	Status string `json:"status"`
	Source string `json:"source"`
	CIJob  string `json:"ci_job"`
	Reason string `json:"reason"`
}

// Manifest contains validated inputs; passing it alone does not grant admission.
type Manifest struct {
	Schema       int                    `json:"schema_version"`
	Status       string                 `json:"status"`
	Candidate    Candidate              `json:"candidate"`
	Records      []Record               `json:"records"`
	Jobs         map[string]JobResult   `json:"jobs"`
	Failures     []string               `json:"failures"`
	Attestations map[string]Attestation `json:"attestations"`
}

// RequiredJobs includes command jobs and native security actions.
func RequiredJobs() []string {
	return []string{"agent-governance", "test", "lint", "govulncheck", "benchmark-smoke", "fuzz-smoke", "codeql", "scorecard", "database"}
}

// Requirements derives every matrix artifact from the canonical toolchain pins.
func Requirements(p toolchainpolicy.Policy) []Requirement {
	out := make([]Requirement, 0, len(p.Test)+6)
	for _, v := range p.Test {
		out = append(out, Requirement{ID: "test-" + v, Job: "test", GoVersion: v, Command: []string{"make", "ci-test", "COVERAGE_FILE=.aiflow/ci/coverage.out"}})
	}
	for _, pair := range [][2]string{{"agent-governance", "ci-governance-check"}, {"lint", "lint"}, {"govulncheck", "govulncheck"}, {"benchmark-smoke", "bench-smoke"}, {"fuzz-smoke", "fuzz-smoke"}, {"database", "db-integration-check"}} {
		out = append(out, Requirement{ID: pair[0], Job: pair[0], GoVersion: p.Release, Command: []string{"make", pair[1]}})
	}
	return out
}

// Validate fails closed for incomplete, stale or unsuccessful CI evidence.
func Validate(c Candidate, p toolchainpolicy.Policy, records []Record, jobs map[string]JobResult) Manifest {
	records = slices.Clone(records)
	slices.SortFunc(records, func(a, b Record) int { return cmp.Compare(a.ID, b.ID) })
	m := Manifest{Schema: 1, Status: "failed", Candidate: c, Records: records, Jobs: jobs, Failures: []string{}, Attestations: map[string]Attestation{}}
	fail := func(message string) { m.Failures = append(m.Failures, message) }
	if c.Commit == "" || c.Tree == "" || c.ModulesSHA256 == "" || c.RunID == "" || c.Attempt == "" {
		fail("candidate identity and workflow scope must be complete")
	}
	for _, job := range RequiredJobs() {
		if jobs[job].Result != "success" {
			fail(fmt.Sprintf("job %s result is %q, require success", job, jobs[job].Result))
		}
	}
	byID := map[string]Record{}
	for _, r := range records {
		if _, exists := byID[r.ID]; exists {
			fail("duplicate record " + r.ID)
		}
		byID[r.ID] = r
	}
	for _, req := range Requirements(p) {
		r, ok := byID[req.ID]
		if !ok {
			fail("missing record " + req.ID)
			continue
		}
		delete(byID, req.ID)
		if r.Schema != 1 {
			fail(req.ID + ": unsupported record schema")
		}
		if r.Before != c || r.After != c {
			fail(req.ID + ": candidate commit/tree/dependencies/run changed or stale")
		}
		if !slices.Equal(r.Command, req.Command) {
			fail(req.ID + ": command does not match required validation")
		}
		if r.GoVersion != "go"+req.GoVersion || r.GoToolchain != "local" || (r.GoWork != "" && r.GoWork != "off") {
			fail(req.ID + ": wrong toolchain or workspace")
		}
		if r.Status != "passed" || r.ExitCode == nil || *r.ExitCode != 0 || r.Error != "" {
			fail(req.ID + ": command did not pass")
		}
		if r.StartedAt.IsZero() || r.FinishedAt.Before(r.StartedAt) {
			fail(req.ID + ": missing or invalid execution timestamps")
		}
	}
	for id := range byID {
		fail("unexpected record " + id)
	}
	if len(m.Failures) > 0 {
		return m
	}
	m.Status = "passed"
	groups := map[string][]string{
		"agent_check":          {"agent-governance", "test"},
		"agent_full_check":     {"agent-governance", "test", "lint", "govulncheck"},
		"agent_security_check": {"lint", "govulncheck"},
		"coverage_profile":     {"test"}, "coverage_check": {"test"},
		"api_check": {"agent-governance"}, "tools_check": {"agent-governance"}, "docs_check": {"agent-governance"},
		"aiflow_layout_check": {"agent-governance"}, "governance_maturity_check": {"agent-governance"},
		"bench_regression_check": {"agent-governance"}, "ci_workflow_check": {"agent-governance"},
	}
	for command, covered := range groups {
		m.Attestations[command] = Attestation{Status: "covered_by_ci", Source: "ci_job", CIJob: strings.Join(covered, ","), Reason: "Validated command records for this candidate and workflow run cover this contract; security-sensitive diff remains an embedded classification signal."}
	}
	return m
}
