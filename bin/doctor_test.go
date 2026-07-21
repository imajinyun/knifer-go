package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestAgentAndGovernanceTargetsDefaultToIsolatedGoCache(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(repoRoot(t), "Makefile"))
	if err != nil {
		t.Fatalf("ReadFile(Makefile) error = %v", err)
	}
	makefile := string(data)
	for _, want := range []string{
		"ISOLATED_GOCACHE ?= /tmp/knifer-go-gocache",
		"EFFECTIVE_ISOLATED_GOCACHE := $(if $(strip $(GOCACHE)),$(GOCACHE),$(ISOLATED_GOCACHE))",
		"$(AGENT_GOVERNANCE_TARGETS): export GOCACHE := $(EFFECTIVE_ISOLATED_GOCACHE)",
	} {
		if !strings.Contains(makefile, want) {
			t.Fatalf("Makefile missing isolated cache contract %q", want)
		}
	}
	match := regexp.MustCompile(`(?m)^AGENT_GOVERNANCE_TARGETS := (.+)$`).FindStringSubmatch(makefile)
	if len(match) != 2 {
		t.Fatal("Makefile must define AGENT_GOVERNANCE_TARGETS")
	}
	targets := map[string]bool{}
	for _, target := range strings.Fields(match[1]) {
		targets[target] = true
	}
	for _, target := range []string{
		"doctor",
		"agent-check",
		"agent-full-check",
		"agent-security-check",
		"ci-agent-governance",
		"governance-maturity-check",
		"api-check",
		"api-freeze-check",
		"tools-check",
		"tools-gen",
		"tools-report",
		"docs-check",
		"docs-gen",
		"ai-context-check",
		"arch",
		"lint",
		"govulncheck",
		"coverage-check",
		"release-notes-check",
	} {
		if !targets[target] {
			t.Fatalf("AGENT_GOVERNANCE_TARGETS missing %s", target)
		}
	}
}

func TestDoctorRejectsStderrEvenWhenGoListExitsZero(t *testing.T) {
	fixture := newDoctorFixture(t, `echo "go: writing stat cache: operation not permitted" >&2
exit 0`)

	output, err := fixture.RunScript(
		"bin/doctor.sh",
		"DOCTOR_ROOT="+fixture.Root(),
		"PATH="+filepath.Join(fixture.Root(), "bin")+string(os.PathListSeparator)+os.Getenv("PATH"),
		"GO=go",
	)
	if err == nil {
		t.Fatalf("doctor unexpectedly succeeded:\n%s", output)
	}
	if !strings.Contains(output, "go: writing stat cache: operation not permitted") {
		t.Fatalf("doctor did not preserve stderr:\n%s", output)
	}
	if !strings.Contains(output, "DOCTOR ERROR: go list ./... wrote to stderr") {
		t.Fatalf("doctor did not classify stderr as a failure:\n%s", output)
	}
	if strings.Contains(output, "go list ./... OK") {
		t.Fatalf("doctor reported a false success:\n%s", output)
	}
}

func TestDoctorPreservesOutputAndRejectsNonZeroStatus(t *testing.T) {
	fixture := newDoctorFixture(t, `echo "partial package output"
echo "package load failed" >&2
exit 7`)

	output, err := fixture.RunScript(
		"bin/doctor.sh",
		"DOCTOR_ROOT="+fixture.Root(),
		"PATH="+filepath.Join(fixture.Root(), "bin")+string(os.PathListSeparator)+os.Getenv("PATH"),
		"GO=go",
	)
	if err == nil {
		t.Fatalf("doctor unexpectedly succeeded:\n%s", output)
	}
	for _, want := range []string{
		"partial package output",
		"package load failed",
		"DOCTOR ERROR: go list ./... exited with status 7",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("doctor output missing %q:\n%s", want, output)
		}
	}
	if strings.Contains(output, "go list ./... OK") {
		t.Fatalf("doctor reported a false success:\n%s", output)
	}
}

func newDoctorFixture(t *testing.T, packageListBehavior string) *governanceFixture {
	t.Helper()
	fixture := newGovernanceFixture(t)
	fakeBin := filepath.Join(fixture.Root(), "bin")
	if err := os.MkdirAll(fakeBin, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", fakeBin, err)
	}
	writeExecutableTestFile(t, fakeBin, "go", `#!/usr/bin/env bash
if [ "${1:-}" = "-C" ]; then
	shift 2
fi
case "${1:-}" in
	version)
		echo "go version go1.25.0 test/arch"
		;;
	env)
		echo "go1.25.0"
		;;
	tool)
		exit 1
		;;
	list)
		if [ "${2:-}" = "-m" ]; then
			echo "github.com/imajinyun/knifer-go"
			exit 0
		fi
		`+packageListBehavior+`
		;;
	*)
		exit 0
		;;
esac
`)
	writeExecutableTestFile(t, fakeBin, "git", "#!/usr/bin/env bash\necho '## main'\n")
	fixture.WriteFile("bin/check_go_module_cache.sh", "#!/usr/bin/env bash\nexit 0\n")
	if err := os.Chmod(filepath.Join(fixture.Root(), "bin", "check_go_module_cache.sh"), 0o755); err != nil {
		t.Fatalf("Chmod(check_go_module_cache.sh) error = %v", err)
	}
	return fixture
}

func writeExecutableTestFile(t *testing.T, root, name, content string) {
	t.Helper()
	writeTestFile(t, root, name, content)
	if err := os.Chmod(filepath.Join(root, name), 0o755); err != nil {
		t.Fatalf("Chmod(%q) error = %v", filepath.Join(root, name), err)
	}
}
