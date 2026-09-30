package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/imajinyun/knifer-go/bin/internal/cievidence"
	"github.com/imajinyun/knifer-go/bin/internal/toolchainpolicy"
)

func admissionFixture(t *testing.T) (*checker, string, string) {
	t.Helper()
	root := t.TempDir()
	p := toolchainpolicy.Policy{Minimum: "1.26.0", Release: "1.27.1", Test: []string{"1.26.8", "1.27.1"}, Lint: "v2.14.0"}
	for name, content := range map[string]string{".gitignore": ".aiflow/\n", "go.mod": "module example.com/fixture\ngo 1.26.0\n", "go.sum": ""} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	config := map[string]any{"project": map[string]string{"go_version": ">=1.26"}, "ci_workflows": map[string]any{"tool_versions": p}}
	if err := cievidence.WriteJSON(filepath.Join(root, "ai-context.json"), config); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"init", "-q"}, {"add", "."}, {"-c", "user.name=Fixture", "-c", "user.email=fixture@example.com", "-c", "commit.gpgsign=false", "-c", "core.hooksPath=/dev/null", "commit", "-qm", "fixture"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git: %v %s", err, out)
		}
	}
	t.Setenv("GITHUB_SHA", "")
	t.Setenv("GITHUB_RUN_ID", "321")
	t.Setenv("GITHUB_RUN_ATTEMPT", "1")
	t.Setenv("AGENT_CHANGE_BASE_REF", "")
	candidate, err := cievidence.Capture(root)
	if err != nil {
		t.Fatal(err)
	}
	jobs := map[string]cievidence.JobResult{}
	for _, name := range cievidence.RequiredJobs() {
		jobs[name] = cievidence.JobResult{Result: "success"}
	}
	records := make([]cievidence.Record, 0, len(cievidence.Requirements(p)))
	dir := filepath.Join(root, ".aiflow/results")
	for _, req := range cievidence.Requirements(p) {
		code := 0
		r := cievidence.Record{Schema: 1, ID: req.ID, Before: candidate, After: candidate, Command: req.Command, GoVersion: "go" + req.GoVersion, GoToolchain: "local", Status: "passed", ExitCode: &code, StartedAt: time.Unix(1, 0).UTC(), FinishedAt: time.Unix(2, 0).UTC()}
		records = append(records, r)
		if err := cievidence.WriteJSON(filepath.Join(dir, req.ID+".json"), r); err != nil {
			t.Fatal(err)
		}
	}
	m := cievidence.Validate(candidate, p, records, jobs)
	data, err := json.Marshal(map[string]any{"commit": candidate.Commit, "merge_ready": true, "required_commands": []string{"agent_full_check", "agent_security_check"}, "command_attestations": m.Attestations, "ci_evidence": m})
	if err != nil {
		t.Fatal(err)
	}
	var evidence map[string]any
	if err := json.Unmarshal(data, &evidence); err != nil {
		t.Fatal(err)
	}
	needs, err := json.Marshal(jobs)
	if err != nil {
		t.Fatal(err)
	}
	return &checker{root: root, evidence: evidence}, dir, string(needs)
}

func TestRequireAdmission(t *testing.T) {
	for _, name := range []string{"valid", "not_ready", "wrong_report_commit", "missing_test", "failed_job", "forged_attestation", "changed_modules", "old_attempt", "missing_manifest"} {
		t.Run(name, func(t *testing.T) {
			c, dir, needs := admissionFixture(t)
			switch name {
			case "not_ready":
				c.evidence["merge_ready"] = false
			case "wrong_report_commit":
				c.evidence["commit"] = "other"
			case "missing_test":
				if err := os.Remove(filepath.Join(dir, "test-1.26.8.json")); err != nil {
					t.Fatal(err)
				}
			case "failed_job":
				needs = `{"test":{"result":"failure"}}`
			case "forged_attestation":
				mapValue(mapValue(c.evidence["command_attestations"])["agent_full_check"])["source"] = "agent_run"
			case "changed_modules":
				if err := os.WriteFile(filepath.Join(c.root, "go.sum"), []byte("changed"), 0o600); err != nil {
					t.Fatal(err)
				}
			case "old_attempt":
				t.Setenv("GITHUB_RUN_ATTEMPT", "2")
			case "missing_manifest":
				delete(c.evidence, "ci_evidence")
			}
			c.requireAdmission(dir, needs)
			if (len(c.findings) == 0) != (name == "valid") {
				t.Fatalf("findings=%+v", c.findings)
			}
		})
	}
}
