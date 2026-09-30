package main

import (
	"slices"
	"strings"

	"github.com/imajinyun/knifer-go/bin/internal/cievidence"
	"gopkg.in/yaml.v3"
)

func (c *checker) checkAdmissionWorkflow(text, path string) {
	var workflow struct {
		Env  map[string]string `yaml:"env"`
		Jobs map[string]struct {
			If    string            `yaml:"if"`
			Needs []string          `yaml:"needs"`
			Env   map[string]string `yaml:"env"`
			Steps []workflowStep    `yaml:"steps"`
		} `yaml:"jobs"`
	}
	if err := yaml.Unmarshal([]byte(text), &workflow); err != nil {
		c.addError("CI_ADMISSION_SCHEMA", path, err.Error())
		return
	}
	problem := func(message string) { c.addError("CI_ADMISSION_WIRING", path, message) }
	if workflow.Env["AGENT_CHANGE_BASE_REF"] != "${{ github.event.pull_request.base.sha || github.event.before }}" || workflow.Env["GOWORK"] != "off" {
		problem("all jobs must share the event base and disable ambient workspaces")
	}
	final, ok := workflow.Jobs["admission"]
	if !ok {
		problem("missing admission job")
		return
	}
	if final.If != "always()" {
		problem("admission must run with if: always()")
	}
	needs := slices.Clone(final.Needs)
	slices.Sort(needs)
	want := cievidence.RequiredJobs()
	slices.Sort(want)
	if !slices.Equal(needs, want) {
		problem("admission needs must cover exactly the required validation jobs")
	}
	if final.Env["CI_NEEDS_JSON"] != "${{ toJSON(needs) }}" {
		problem("admission must use the current needs context")
	}
	download, strict, upload := false, false, false
	for _, step := range final.Steps {
		if strings.HasPrefix(step.Uses, "actions/download-artifact@") && step.With["pattern"] == "ci-result-${{ github.run_attempt }}-*" && step.With["path"] == ".aiflow/ci/results" && step.With["merge-multiple"] == "true" && step.With["run-id"] == "" {
			download = true
		}
		if strings.TrimSpace(step.Run) == "make ci-admission-check" && step.If == "" && !step.ContinueOnError {
			strict = download
		}
		if strings.HasPrefix(step.Uses, "actions/upload-artifact@") && step.If == "always()" && step.With["name"] == "agent-validation-evidence-${{ github.run_attempt }}" && step.With["include-hidden-files"] == "true" {
			upload = true
		}
	}
	if !strict || !upload {
		problem("admission must collect current artifacts, run the strict gate, and always upload its report")
	}
	for _, pair := range [][3]string{
		{"agent-governance", "agent-governance", "make ci-governance-check"},
		{"test", "test-${{ matrix.go-version }}", "make ci-test COVERAGE_FILE=.aiflow/ci/coverage.out"},
		{"lint", "lint", "make lint"},
		{"govulncheck", "govulncheck", "make govulncheck"},
		{"benchmark-smoke", "benchmark-smoke", "make bench-smoke"},
		{"fuzz-smoke", "fuzz-smoke", "make fuzz-smoke"},
	} {
		job, id, command := pair[0], pair[1], pair[2]
		built, recorded, saved := false, false, false
		for _, step := range workflow.Jobs[job].Steps {
			if strings.TrimSpace(step.Run) == "go build -o .aiflow/ci/ciresult ./bin/ciresult" && step.If == "" && !step.ContinueOnError {
				built = true
			}
			if strings.TrimSpace(step.Run) == ".aiflow/ci/ciresult run -id "+id+" -out .aiflow/ci/records -- "+command && step.If == "" && !step.ContinueOnError {
				recorded = built
			}
			if strings.HasPrefix(step.Uses, "actions/upload-artifact@") && step.If == "always()" && step.With["name"] == "ci-result-${{ github.run_attempt }}-"+id && step.With["path"] == ".aiflow/ci/records/"+id+".json" && step.With["include-hidden-files"] == "true" && step.With["if-no-files-found"] == "error" {
				saved = true
			}
		}
		if !recorded || !saved {
			problem(job + " must execute through the recorder and always upload its unique result")
		}
	}
}
