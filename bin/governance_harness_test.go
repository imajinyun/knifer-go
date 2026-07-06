package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

type governanceFixture struct {
	t    *testing.T
	root string
}

func newGovernanceFixture(t *testing.T) *governanceFixture {
	t.Helper()
	return &governanceFixture{t: t, root: t.TempDir()}
}

func (f *governanceFixture) Root() string {
	return f.root
}

func (f *governanceFixture) WriteFile(name, content string) {
	f.t.Helper()
	writeTestFile(f.t, f.root, name, content)
}

func (f *governanceFixture) WriteJSON(name string, value any) {
	f.t.Helper()
	path := filepath.Join(f.root, name)
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		f.t.Fatalf("MarshalIndent() error = %v", err)
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		f.t.Fatalf("MkdirAll(%q) error = %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		f.t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

func (f *governanceFixture) RunScript(script string, env ...string) (string, error) {
	f.t.Helper()
	root := repoRoot(f.t)
	cmd := exec.Command("bash", filepath.Join(root, script))
	cmd.Dir = root
	cmd.Env = append(os.Environ(), env...)
	combined, err := cmd.CombinedOutput()
	return string(combined), err
}

func (f *governanceFixture) RunProviderContractCheck() (string, error) {
	f.t.Helper()
	return f.RunScript("bin/check_provider_contracts.sh", "PROVIDER_CONTRACT_ROOT="+f.root)
}

func (f *governanceFixture) RunCIWorkflowCheck() (string, error) {
	f.t.Helper()
	return f.RunScript("bin/check_ci_workflows.sh", "CI_WORKFLOW_ROOT="+f.root)
}

func (f *governanceFixture) RunDocsQuickstartCheck() (string, error) {
	f.t.Helper()
	return f.RunScript("bin/check_docs_quickstart.sh", "DOCS_QUICKSTART_ROOT="+f.root)
}
