package main

import (
	"strings"
	"testing"
)

func TestAgentEvidenceAdmissionRequiresReady(t *testing.T) {
	fixture := newGovernanceFixture(t)
	path := fixture.WriteTempFile("evidence.json", "")
	fixture.WriteJSON("evidence.json", baseAgentEvidence())
	output, err := fixture.RunGoToolEnv("agentevidencecheck", []string{"CHANGE_POLICY_DIFF=" + agentEvidenceFixtureDiff}, "-root", repoRoot(t), "-evidence", path, "-require-ready")
	if err == nil || !strings.Contains(output, "AGENT_ADMISSION_NOT_READY") {
		t.Fatalf("strict mode accepted not-ready report: %v\n%s", err, output)
	}
}
