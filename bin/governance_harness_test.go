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

func (f *governanceFixture) WriteTempFile(name, content string) string {
	f.t.Helper()
	f.WriteFile(name, content)
	return filepath.Join(f.root, name)
}

func (f *governanceFixture) WriteGoMod() {
	f.t.Helper()
	f.WriteFile("go.mod", "module github.com/imajinyun/knifer-go\n\ngo 1.25.0\n")
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
	return f.RunScriptArgs(script, nil, env...)
}

func (f *governanceFixture) RunScriptArgs(script string, args []string, env ...string) (string, error) {
	f.t.Helper()
	root := repoRoot(f.t)
	cmdArgs := append([]string{filepath.Join(root, script)}, args...)
	cmd := exec.Command("bash", cmdArgs...)
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

func (f *governanceFixture) RunCIWorkflowCheckJSON() (string, error) {
	f.t.Helper()
	return f.RunGoTool("ciworkflowcheck", "-root", f.root, "-json")
}

func (f *governanceFixture) RunDocsQuickstartCheck() (string, error) {
	f.t.Helper()
	return f.RunScript("bin/check_docs_quickstart.sh", "DOCS_QUICKSTART_ROOT="+f.root)
}

func (f *governanceFixture) RunArchImportsCheck() (string, error) {
	f.t.Helper()
	return f.RunScript("bin/check_arch_imports.sh", "ARCH_CHECK_ROOT="+f.root)
}

func (f *governanceFixture) RunPanicPolicyCheck() (string, error) {
	f.t.Helper()
	return f.RunScript("bin/check_panic_policy.sh", "ARCH_CHECK_ROOT="+f.root)
}

func (f *governanceFixture) RunFacadeBoundaryCheck() (string, error) {
	f.t.Helper()
	return f.RunScript("bin/check_facade_boundary.sh", "ARCH_CHECK_ROOT="+f.root)
}

func (f *governanceFixture) RunReleaseNotesCheck(changelog, template, version string) (string, error) {
	f.t.Helper()
	var args []string
	if version != "" {
		args = append(args, version)
	}
	return f.RunScriptArgs(
		"bin/check_release_notes.sh",
		args,
		"CHANGELOG_FILE="+changelog,
		"GOVERNANCE_RELEASE_TEMPLATE_FILE="+template,
	)
}

func (f *governanceFixture) RunChangePolicyCheck(changedFiles string) (string, error) {
	f.t.Helper()
	return f.RunScript("bin/check_change_policy.sh", "CHANGE_POLICY_CHANGED_FILES="+changedFiles)
}

func (f *governanceFixture) RunChangePolicyCheckWithDiff(changedFiles, diffText string) (string, error) {
	f.t.Helper()
	return f.RunGoToolEnv(
		"changepolicycheck",
		[]string{"CHANGE_POLICY_DIFF=" + diffText},
		"-root", f.root,
		"-changed-files", changedFiles,
	)
}

func (f *governanceFixture) RunChangePolicyCheckJSON(changedFiles, diffText string) (string, error) {
	f.t.Helper()
	env := []string{}
	if diffText != "" {
		env = append(env, "CHANGE_POLICY_DIFF="+diffText)
	}
	return f.RunGoToolEnv(
		"changepolicycheck",
		env,
		"-root", f.root,
		"-changed-files", changedFiles,
		"-json",
	)
}

func (f *governanceFixture) RunAgentEvidenceCheck(evidence map[string]any) (string, error) {
	f.t.Helper()
	path := filepath.Join(f.root, "agent-evidence.json")
	f.WriteJSON("agent-evidence.json", evidence)
	return f.RunScript("bin/check_agent_evidence.sh", "AGENT_EVIDENCE_FILE="+path)
}

func (f *governanceFixture) RunAgentEvidenceCheckJSON(evidence map[string]any) (string, error) {
	f.t.Helper()
	path := filepath.Join(f.root, "agent-evidence.json")
	f.WriteJSON("agent-evidence.json", evidence)
	return f.RunGoTool(
		"agentevidencecheck",
		"-root", repoRoot(f.t),
		"-evidence", path,
		"-json",
	)
}

func (f *governanceFixture) RunCoverageCheck(coverageFile string, env ...string) (string, error) {
	f.t.Helper()
	return f.RunScriptArgs("bin/check_coverage.sh", []string{coverageFile}, env...)
}

func (f *governanceFixture) RunAPIFreezeCheck(contextPath, toolsPath string) (string, error) {
	f.t.Helper()
	return f.RunScript(
		"bin/check_api_freeze.sh",
		"AI_CONTEXT_FILE="+contextPath,
		"TOOLS_JSON_FILE="+toolsPath,
	)
}

func (f *governanceFixture) RunAPIFreezeCheckJSON(contextPath, toolsPath string) (string, error) {
	f.t.Helper()
	return f.RunGoTool(
		"apifreezecheck",
		"-ai-context", contextPath,
		"-tools", toolsPath,
		"-json",
	)
}

func (f *governanceFixture) RunGoTool(tool string, args ...string) (string, error) {
	f.t.Helper()
	return f.RunGoToolEnv(tool, nil, args...)
}

func (f *governanceFixture) RunGoToolEnv(tool string, env []string, args ...string) (string, error) {
	f.t.Helper()
	root := repoRoot(f.t)
	binary := filepath.Join(f.t.TempDir(), tool)
	build := exec.Command("go", "build", "-o", binary, "./bin/"+tool)
	build.Dir = root
	build.Env = os.Environ()
	if output, err := build.CombinedOutput(); err != nil {
		f.t.Fatalf("go build ./bin/%s failed: %v\n%s", tool, err, string(output))
	}

	cmd := exec.Command(binary, args...)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), env...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}
