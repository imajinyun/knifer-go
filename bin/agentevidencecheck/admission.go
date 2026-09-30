package main

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/imajinyun/knifer-go/bin/internal/cievidence"
	"github.com/imajinyun/knifer-go/bin/internal/toolchainpolicy"
)

func (c *checker) requireAdmission(results, needs string) {
	if ready, ok := boolValue(c.evidence["merge_ready"]); !ok || !ready {
		c.addError("AGENT_ADMISSION_NOT_READY", "merge_ready must be true for admission")
	}
	candidate, err := cievidence.Capture(c.root)
	if err != nil {
		c.addError("AGENT_ADMISSION_CANDIDATE", err.Error())
		return
	}
	if stringValue(c.evidence["commit"]) != candidate.Commit {
		c.addError("AGENT_ADMISSION_COMMIT", "report commit does not match checkout")
	}
	p, err := toolchainpolicy.Load(c.root)
	if err != nil {
		c.addError("AGENT_ADMISSION_POLICY", err.Error())
		return
	}
	records, err := cievidence.ReadRecords(results)
	if err != nil {
		c.addError("AGENT_ADMISSION_RESULTS", err.Error())
		return
	}
	var jobs map[string]cievidence.JobResult
	if err := json.Unmarshal([]byte(needs), &jobs); err != nil {
		c.addError("AGENT_ADMISSION_JOBS", err.Error())
		return
	}
	expected := cievidence.Validate(candidate, p, records, jobs)
	if expected.Status != "passed" {
		for _, message := range expected.Failures {
			c.addError("AGENT_ADMISSION_RESULTS", message)
		}
		return
	}
	data, err := json.Marshal(c.evidence["ci_evidence"])
	if err != nil {
		c.addError("AGENT_ADMISSION_BINDING", err.Error())
		return
	}
	var supplied cievidence.Manifest
	if err := json.Unmarshal(data, &supplied); err != nil {
		c.addError("AGENT_ADMISSION_BINDING", err.Error())
		return
	}
	if !reflect.DeepEqual(supplied, expected) {
		c.addError("AGENT_ADMISSION_BINDING", "report CI evidence does not match validated original records")
	}
	attestations := mapValue(c.evidence["command_attestations"])
	for _, command := range stringList(c.evidence["required_commands"]) {
		switch command {
		case "agent_evidence", "agent_evidence_check", "ai_context_check", "change_policy_check", "security_sensitive_diff":
			continue
		}
		want, ok := expected.Attestations[command]
		if !ok {
			c.addError("AGENT_ADMISSION_UNCOVERED_COMMAND", "no validated CI coverage for "+command)
			continue
		}
		actual := mapValue(attestations[command])
		if stringValue(actual["status"]) != want.Status || stringValue(actual["source"]) != want.Source || stringValue(actual["ci_job"]) != want.CIJob {
			c.addError("AGENT_ADMISSION_ATTESTATION", fmt.Sprintf("%s must use its validated CI-derived attestation", command))
		}
	}
}
