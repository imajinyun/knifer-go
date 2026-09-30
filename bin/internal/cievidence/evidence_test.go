package cievidence

import (
	"testing"
	"time"

	"github.com/imajinyun/knifer-go/bin/internal/toolchainpolicy"
)

func fixture() (Candidate, toolchainpolicy.Policy, []Record, map[string]JobResult) {
	c := Candidate{Commit: "commit", Tree: "tree", ModulesSHA256: "modules", RunID: "42", Attempt: "1", Base: "base"}
	p := toolchainpolicy.Policy{Minimum: "1.26.0", Release: "1.27.1", Test: []string{"1.26.8", "1.27.1"}, Lint: "v2.14.0"}
	records := make([]Record, 0, len(Requirements(p)))
	for _, req := range Requirements(p) {
		zero := 0
		records = append(records, Record{Schema: 1, ID: req.ID, Before: c, After: c, Command: req.Command, GoVersion: "go" + req.GoVersion, GoToolchain: "local", GoWork: "off", Status: "passed", ExitCode: &zero, StartedAt: time.Now().UTC(), FinishedAt: time.Now().UTC()})
	}
	jobs := map[string]JobResult{}
	for _, job := range RequiredJobs() {
		jobs[job] = JobResult{Result: "success"}
	}
	return c, p, records, jobs
}

func TestValidate(t *testing.T) {
	for _, name := range []string{"valid", "missing_test", "failed_scan", "cancelled_job", "skipped_job", "missing_job", "wrong_commit", "wrong_tree", "wrong_modules", "old_run", "old_attempt", "changed_during_run", "wrong_toolchain", "auto_toolchain", "wrong_command", "missing_exit", "duplicate_record", "native_failure", "empty_scope"} {
		t.Run(name, func(t *testing.T) {
			c, p, records, jobs := fixture()
			switch name {
			case "missing_test":
				records = records[1:]
			case "failed_scan":
				for i := range records {
					if records[i].ID == "govulncheck" {
						code := 1
						records[i].ExitCode = &code
						records[i].Status = "failed"
					}
				}
			case "cancelled_job":
				jobs["test"] = JobResult{Result: "cancelled"}
			case "skipped_job":
				jobs["test"] = JobResult{Result: "skipped"}
			case "missing_job":
				delete(jobs, "test")
			case "wrong_commit":
				records[0].Before.Commit = "different"
			case "wrong_tree":
				records[0].Before.Tree = "different"
			case "wrong_modules":
				records[0].Before.ModulesSHA256 = "different"
			case "old_run":
				records[0].Before.RunID = "41"
			case "old_attempt":
				records[0].Before.Attempt = "2"
			case "changed_during_run":
				records[0].After.Tree = "different"
			case "wrong_toolchain":
				records[0].GoVersion = "go1.25.13"
			case "auto_toolchain":
				records[0].GoToolchain = "auto"
			case "wrong_command":
				records[0].Command = []string{"true"}
			case "missing_exit":
				records[0].ExitCode = nil
			case "duplicate_record":
				records = append(records, records[0])
			case "native_failure":
				jobs["codeql"] = JobResult{Result: "failure"}
			case "empty_scope":
				c.RunID = ""
			}
			m := Validate(c, p, records, jobs)
			if (m.Status == "passed") != (name == "valid") {
				t.Fatalf("status=%s failures=%v", m.Status, m.Failures)
			}
			if name == "valid" {
				if m.Attestations["agent_full_check"].Status != "covered_by_ci" || m.Attestations["agent_security_check"].CIJob == "" {
					t.Fatalf("missing composite attestations: %+v", m.Attestations)
				}
			} else if len(m.Attestations) != 0 {
				t.Fatal("failed inputs must not produce passing attestations")
			}
		})
	}
}
