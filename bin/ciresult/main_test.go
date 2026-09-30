package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/imajinyun/knifer-go/bin/internal/cievidence"
	"github.com/imajinyun/knifer-go/bin/internal/toolchainpolicy"
)

func cleanFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for name, content := range map[string]string{".gitignore": ".aiflow/\n", "go.mod": "module example.com/fixture\ngo 1.26.0\n", "go.sum": "", "input.txt": "initial"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{{"init", "-q"}, {"add", "."}, {"-c", "user.name=Fixture", "-c", "user.email=fixture@example.com", "-c", "commit.gpgsign=false", "-c", "core.hooksPath=/dev/null", "commit", "-qm", "fixture"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git: %v %s", err, out)
		}
	}
	t.Setenv("GITHUB_SHA", "")
	t.Setenv("GITHUB_RUN_ID", "123")
	t.Setenv("GITHUB_RUN_ATTEMPT", "1")
	t.Setenv("AGENT_CHANGE_BASE_REF", "")
	t.Setenv("GOTOOLCHAIN", "local")
	t.Setenv("GOWORK", "off")
	return root
}

func TestRun(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses POSIX shell fixtures")
	}
	for _, tt := range []struct {
		name, script, status string
		code                 int
		cancel               bool
	}{
		{name: "success", script: "exit 0", status: "passed"},
		{name: "failure", script: "exit 17", status: "failed", code: 17},
		{name: "changed_tree", script: "echo changed > input.txt", status: "failed", code: 1},
		{name: "cancelled", script: "exit 0", status: "cancelled", code: 130, cancel: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			root := cleanFixture(t)
			out := filepath.Join(root, ".aiflow/result.json")
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if tt.cancel {
				cancel()
			}
			if code := run(ctx, root, "probe", out, []string{"sh", "-c", tt.script}); code != tt.code {
				t.Fatalf("exit=%d want %d", code, tt.code)
			}
			data, err := os.ReadFile(out)
			if err != nil {
				t.Fatal(err)
			}
			var r cievidence.Record
			if err := json.Unmarshal(data, &r); err != nil {
				t.Fatal(err)
			}
			if r.Status != tt.status || r.ExitCode == nil || *r.ExitCode != tt.code || r.FinishedAt.Before(r.StartedAt) {
				t.Fatalf("invalid record: %+v", r)
			}
			if tt.code == 0 && (r.Before != r.After || r.GoVersion == "" || r.Before.ModulesSHA256 == "") {
				t.Fatalf("unbound record: %+v", r)
			}
		})
	}
}

func TestCollect(t *testing.T) {
	root := cleanFixture(t)
	p := toolchainpolicy.Policy{Minimum: "1.26.0", Release: "1.27.1", Test: []string{"1.26.8", "1.27.1"}, Lint: "v2.14.0"}
	config := map[string]any{"project": map[string]string{"go_version": ">=1.26"}, "ci_workflows": map[string]any{"tool_versions": p}}
	if err := cievidence.WriteJSON(filepath.Join(root, "ai-context.json"), config); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "add", "ai-context.json")
	cmd.Dir = root
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	cmd = exec.Command("git", "-c", "user.name=Fixture", "-c", "user.email=fixture@example.com", "-c", "commit.gpgsign=false", "-c", "core.hooksPath=/dev/null", "commit", "-qm", "config")
	cmd.Dir = root
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	c, err := cievidence.Capture(root)
	if err != nil {
		t.Fatal(err)
	}
	jobs := map[string]cievidence.JobResult{}
	for _, job := range cievidence.RequiredJobs() {
		jobs[job] = cievidence.JobResult{Result: "success"}
	}
	needs, err := json.Marshal(jobs)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(root, ".aiflow/results")
	for _, req := range cievidence.Requirements(p) {
		r := cievidence.Record{Schema: 1, ID: req.ID, Before: c, After: c, Command: req.Command, GoVersion: "go" + req.GoVersion, GoToolchain: "local", Status: "passed"}
		zero := 0
		r.ExitCode = &zero
		r.StartedAt = time.Unix(1, 0).UTC()
		r.FinishedAt = r.StartedAt
		if err := cievidence.WriteJSON(filepath.Join(dir, req.ID+".json"), r); err != nil {
			t.Fatal(err)
		}
	}
	m, err := collect(root, dir, string(needs))
	if err != nil || m.Status != "passed" {
		t.Fatalf("collect: %v %+v", err, m)
	}
	if err := os.Remove(filepath.Join(dir, "govulncheck.json")); err != nil {
		t.Fatal(err)
	}
	m, err = collect(root, dir, string(needs))
	if err != nil || m.Status != "failed" {
		t.Fatalf("missing scan was accepted: %v %+v", err, m)
	}
	if _, err := collect(root, dir, "invalid JSON"); err == nil {
		t.Fatal("invalid needs accepted")
	}
}
