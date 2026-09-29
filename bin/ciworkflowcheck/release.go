package main

import (
	"strings"

	"gopkg.in/yaml.v3"
)

// checkReleaseLint requires the release job to supply its own pinned linter;
// tools installed by another job are not available in the release environment.
func (c *checker) checkReleaseLint(text, path, version string) {
	var workflow struct {
		Env  map[string]string `yaml:"env"`
		Jobs map[string]struct {
			Steps []struct {
				Run string `yaml:"run"`
			} `yaml:"steps"`
		} `yaml:"jobs"`
	}
	if err := yaml.Unmarshal([]byte(text), &workflow); err != nil {
		c.addError("CI_WORKFLOW_SCHEMA_INVALID", path, "cannot parse release workflow: "+err.Error())
		return
	}
	if workflow.Env["GOLANGCI_LINT_VERSION"] != version {
		c.addError("CI_WORKFLOW_VERSION_DRIFT", path, "release workflow must use declared golangci-lint version "+version)
	}
	const install = "go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${{ env.GOLANGCI_LINT_VERSION }}"
	installed := false
	for _, step := range workflow.Jobs["release"].Steps {
		for _, line := range strings.Split(step.Run, "\n") {
			line = strings.TrimSpace(line)
			if line == install {
				installed = true
			}
			if makeTargetFromCommand(line) == "release-check" && !installed {
				c.addError("CI_WORKFLOW_RELEASE_LINT_INSTALL_MISSING", path,
					"release job must install the pinned golangci-lint before make release-check")
			}
		}
	}
}
