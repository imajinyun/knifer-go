package main

import (
	"strings"

	"gopkg.in/yaml.v3"
)

func (c *checker) checkFuzzArtifacts(text, path, name string) {
	jobName := ""
	switch name {
	case "go":
		jobName = "fuzz-smoke"
	case "release":
		jobName = "release"
	case "fuzz":
		jobName = "fuzz-extended"
	default:
		return
	}
	var workflow struct {
		On   map[string]any `yaml:"on"`
		Jobs map[string]struct {
			Timeout int            `yaml:"timeout-minutes"`
			Steps   []workflowStep `yaml:"steps"`
		} `yaml:"jobs"`
	}
	if err := yaml.Unmarshal([]byte(text), &workflow); err != nil {
		c.addError("CI_FUZZ_SCHEMA", path, err.Error())
		return
	}
	job := workflow.Jobs[jobName]
	uploaded := false
	extended := false
	for _, step := range job.Steps {
		if strings.HasPrefix(step.Uses, "actions/upload-artifact@") && step.If == "always()" && step.With["path"] == ".aiflow/fuzz/" && step.With["include-hidden-files"] == "true" {
			uploaded = true
		}
		if strings.TrimSpace(step.Run) == "make fuzz-extended" && !step.ContinueOnError {
			extended = true
		}
	}
	if !uploaded {
		c.addError("CI_FUZZ_ARTIFACTS_MISSING", path, jobName+" must always upload isolated fuzz artifacts")
	}
	if name == "fuzz" {
		_, manual := workflow.On["workflow_dispatch"]
		schedules, ok := workflow.On["schedule"].([]any)
		if !manual || !ok || len(schedules) == 0 || !extended || job.Timeout <= 0 || job.Timeout > 20 {
			c.addError("CI_FUZZ_EXTENDED_POLICY", path, "extended fuzz must run on schedule/manual dispatch with a bounded job timeout")
		}
	}
}
