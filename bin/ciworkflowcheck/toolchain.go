package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/imajinyun/knifer-go/bin/internal/toolchainpolicy"
	"gopkg.in/yaml.v3"
)

type workflowStep struct {
	Uses            string            `yaml:"uses"`
	Run             string            `yaml:"run"`
	If              string            `yaml:"if"`
	Env             map[string]string `yaml:"env"`
	With            map[string]string `yaml:"with"`
	ContinueOnError bool              `yaml:"continue-on-error"`
}

type workflowJob struct {
	Env      map[string]string `yaml:"env"`
	Steps    []workflowStep    `yaml:"steps"`
	Strategy struct {
		Matrix map[string][]string `yaml:"matrix"`
	} `yaml:"strategy"`
}

func (c *checker) checkToolchainWorkflow(text, path string, p toolchainpolicy.Policy) {
	var workflow struct {
		Env  map[string]string      `yaml:"env"`
		Jobs map[string]workflowJob `yaml:"jobs"`
	}
	if err := yaml.Unmarshal([]byte(text), &workflow); err != nil {
		c.addError("CI_WORKFLOW_SCHEMA_INVALID", path, "cannot parse workflow: "+err.Error())
		return
	}
	if workflow.Env["GO_RELEASE_VERSION"] != p.Release || workflow.Env["GOLANGCI_LINT_VERSION"] != p.Lint {
		c.addError("CI_WORKFLOW_VERSION_DRIFT", path, "workflow tool pins must match ci_workflows.tool_versions")
	}
	if workflow.Env["GOTOOLCHAIN"] != "local" {
		c.addError("CI_WORKFLOW_AUTO_TOOLCHAIN", path, "workflow must set GOTOOLCHAIN=local")
	}
	for jobName, job := range workflow.Jobs {
		if jobName == "scorecard" {
			continue
		}
		if mode, ok := job.Env["GOTOOLCHAIN"]; ok && mode != "local" {
			c.addError("CI_WORKFLOW_AUTO_TOOLCHAIN", path, jobName+" overrides GOTOOLCHAIN=local")
		}
		matrixJob := jobName == "test"
		if matrixJob && !slices.Equal(job.Strategy.Matrix["go-version"], p.Test) {
			c.addError("CI_WORKFLOW_VERSION_DRIFT", path, fmt.Sprintf("test matrix must equal %v", p.Test))
		}
		setup, recorded := false, false
		for _, step := range job.Steps {
			if mode, ok := step.Env["GOTOOLCHAIN"]; ok && mode != "local" {
				c.addError("CI_WORKFLOW_AUTO_TOOLCHAIN", path, jobName+" step overrides GOTOOLCHAIN=local")
			}
			if strings.HasPrefix(step.Uses, "actions/setup-go@") {
				want := "${{ env.GO_RELEASE_VERSION }}"
				if matrixJob {
					want = "${{ matrix.go-version }}"
				}
				if step.With["go-version"] != want || step.If != "" {
					c.addError("CI_WORKFLOW_VERSION_DRIFT", path, jobName+" must unconditionally setup "+want)
				}
				setup = true
				continue
			}
			if setup && strings.Contains(step.Run, "go version") && strings.Contains(step.Run, "go env GOTOOLCHAIN") {
				recorded = true
			}
			goWork := len(workflowMakeTargets(step.Run)) > 0 || strings.Contains(step.Run, "go install ") || strings.HasPrefix(step.Uses, "github/codeql-action/autobuild@")
			if goWork && (!setup || !recorded) {
				c.addError("CI_WORKFLOW_GO_SETUP_MISSING", path, jobName+" must setup and record Go before validation")
			}
		}
		if !setup || !recorded {
			c.addError("CI_WORKFLOW_GO_SETUP_MISSING", path, jobName+" must setup and record its fixed Go toolchain")
		}
	}
}
