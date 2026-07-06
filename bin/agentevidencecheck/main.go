package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const diffFilter = "ACDMRTUXB"

type checker struct {
	root     string
	context  map[string]any
	evidence map[string]any
	errors   []ruleError
}

type ruleError struct {
	id      string
	message string
}

func main() {
	rootFlag := flag.String("root", "", "repository root")
	contextFlag := flag.String("ai-context", "", "ai-context.json path")
	evidenceFlag := flag.String("evidence", "", "agent evidence JSON path")
	flag.Parse()

	root := strings.TrimSpace(*rootFlag)
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "agent evidence check error: cannot resolve working directory: %v\n", err)
			os.Exit(1)
		}
	}
	contextPath := strings.TrimSpace(*contextFlag)
	if contextPath == "" {
		contextPath = filepath.Join(root, "ai-context.json")
	}
	evidencePath := strings.TrimSpace(*evidenceFlag)
	if evidencePath == "" {
		evidencePath = "/tmp/knifer-go-agent-validation.json"
	}

	context, err := loadJSON(contextPath, "ai-context.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	evidence, err := loadJSON(evidencePath, "Agent evidence")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	c := &checker{root: root, context: context, evidence: evidence}
	c.run()
	if len(c.errors) > 0 {
		for _, err := range c.errors {
			fmt.Fprintf(os.Stderr, "agent evidence check error: [%s] %s\n", err.id, err.message)
		}
		os.Exit(1)
	}

	displayPath := evidencePath
	if strings.HasPrefix(evidencePath, root+string(os.PathSeparator)) {
		if rel, err := filepath.Rel(root, evidencePath); err == nil {
			displayPath = rel
		}
	}
	detectedPolicies := stringList(c.evidence["detected_change_policies"])
	requiredCommands := stringList(c.evidence["required_commands"])
	fmt.Printf(
		"agent evidence is valid (%s; %d policies, %d required commands, merge_ready=%s)\n",
		displayPath,
		len(detectedPolicies),
		len(requiredCommands),
		strings.ToLower(fmt.Sprint(c.evidence["merge_ready"])),
	)
}

func (c *checker) run() {
	project := c.requireMapping(c.context["project"], "ai-context.json.project")
	commands := c.requireMapping(c.context["commands"], "ai-context.json.commands")
	policies := c.requireMapping(c.context["change_type_policies"], "ai-context.json.change_type_policies")

	if schemaVersion := c.requireString(c.evidence["schema_version"], "schema_version"); schemaVersion != "" && schemaVersion != "1.0" {
		c.addError("AGENT_EVIDENCE_SCHEMA_VERSION", "schema_version must be 1.0")
	}
	if c.requireString(c.evidence["repository"], "repository") != stringValue(project["name"]) {
		c.addError("AGENT_EVIDENCE_PROJECT_MISMATCH", "repository must match ai-context.json.project.name")
	}
	if c.requireString(c.evidence["module"], "module") != stringValue(project["module"]) {
		c.addError("AGENT_EVIDENCE_PROJECT_MISMATCH", "module must match ai-context.json.project.module")
	}
	for _, key := range []string{"generated_at", "branch", "commit"} {
		c.requireString(c.evidence[key], key)
	}
	c.requireOptionalString(c.evidence["change_base_ref"], "change_base_ref")
	if c.requireString(c.evidence["diff_filter"], "diff_filter") != diffFilter {
		c.addError("AGENT_EVIDENCE_DIFF_FILTER_INVALID", "diff_filter must be "+diffFilter)
	}

	changedFiles := c.requireStringList(c.evidence["changed_files"], "changed_files")
	detectedPolicies := c.requireStringList(c.evidence["detected_change_policies"], "detected_change_policies")
	requiredCommands := c.requireStringList(c.evidence["required_commands"], "required_commands")
	securitySensitivePaths := c.requireStringList(c.evidence["security_sensitive_paths"], "security_sensitive_paths")

	if unknown := differenceStrings(detectedPolicies, mapKeys(policies)); len(unknown) > 0 {
		c.addError("AGENT_EVIDENCE_POLICY_UNKNOWN", "detected_change_policies contains unknown policies: "+strings.Join(unknown, ", "))
	}
	if unknown := differenceStrings(requiredCommands, mapKeys(commands)); len(unknown) > 0 {
		c.addError("AGENT_EVIDENCE_COMMAND_UNKNOWN", "required_commands contains unknown commands: "+strings.Join(unknown, ", "))
	}

	expectedRequiredCommands := []string{}
	for _, policy := range sortedStrings(detectedPolicies) {
		policySpec := c.requireMapping(policies[policy], "ai-context.json.change_type_policies."+policy)
		for _, command := range stringList(policySpec["required_commands"]) {
			if !contains(expectedRequiredCommands, command) {
				expectedRequiredCommands = append(expectedRequiredCommands, command)
			}
		}
	}
	if !equalStrings(requiredCommands, expectedRequiredCommands) {
		c.addError("AGENT_EVIDENCE_REQUIRED_COMMANDS_MISMATCH", fmt.Sprintf("required_commands must match detected policies; got %v, want %v", requiredCommands, expectedRequiredCommands))
	}

	riskRank := map[string]int{"low": 1, "medium": 2, "high": 3, "forbidden_for_agent": 4}
	highestRisk := "low"
	for _, command := range requiredCommands {
		commandSpec := c.requireMapping(commands[command], "ai-context.json.commands."+command)
		risk := stringValue(commandSpec["risk_level"])
		rank, ok := riskRank[risk]
		if !ok {
			c.addError("AGENT_EVIDENCE_COMMAND_RISK_INVALID", fmt.Sprintf("ai-context.json.commands.%s.risk_level is invalid", command))
			continue
		}
		if rank > riskRank[highestRisk] {
			highestRisk = risk
		}
	}
	if c.requireString(c.evidence["highest_required_command_risk"], "highest_required_command_risk") != highestRisk {
		c.addError("AGENT_EVIDENCE_RISK_MISMATCH", "highest_required_command_risk must be "+highestRisk)
	}

	commandAttestations := c.requireMapping(c.evidence["command_attestations"], "command_attestations")
	c.validateCommandAttestations(requiredCommands, commandAttestations)
	c.validateEmbeddedChecks(commandAttestations)

	if len(securitySensitivePaths) > 0 && !contains(detectedPolicies, "security_sensitive") {
		c.addError("AGENT_EVIDENCE_SECURITY_PATHS_POLICY_MISSING", "security_sensitive_paths requires detected security_sensitive policy")
	}
	expectedSecuritySensitivePaths := c.expectedSecuritySensitivePaths(changedFiles)
	if !equalStrings(sortedStrings(securitySensitivePaths), expectedSecuritySensitivePaths) {
		c.addError("AGENT_EVIDENCE_SECURITY_PATHS_MISMATCH", fmt.Sprintf("security_sensitive_paths must match changed security-sensitive paths; got %v, want %v", sortedStrings(securitySensitivePaths), expectedSecuritySensitivePaths))
	}

	c.validateSecurityReview(detectedPolicies, requiredCommands, commandAttestations, expectedSecuritySensitivePaths, policies)
	c.validateSecuritySensitiveDiff(commandAttestations, expectedSecuritySensitivePaths)
	c.validateMergeReady(requiredCommands, detectedPolicies, commandAttestations)

	if _, ok := c.evidence["worktree_status"].(string); !ok {
		c.addError("AGENT_EVIDENCE_WORKTREE_STATUS_INVALID", "worktree_status must be a string")
	}
}

func (c *checker) validateCommandAttestations(requiredCommands []string, commandAttestations map[string]any) {
	allowedStatuses := setOf("passed", "failed", "pending", "not_recorded", "skipped", "covered_by_ci")
	allowedSources := setOf("embedded_check", "current_process", "post_generation", "required_by_policy", "agent_run", "ci_job", "manual_review")
	for _, command := range requiredCommands {
		attestation := c.requireMapping(commandAttestations[command], "command_attestations."+command)
		status := c.requireString(attestation["status"], "command_attestations."+command+".status")
		if status != "" && !allowedStatuses[status] {
			c.addError("AGENT_EVIDENCE_ATTESTATION_STATUS", "command_attestations."+command+".status must be one of: "+strings.Join(sortedSet(allowedStatuses), ", "))
		}
		source := c.requireString(attestation["source"], "command_attestations."+command+".source")
		if source != "" && !allowedSources[source] {
			c.addError("AGENT_EVIDENCE_ATTESTATION_SOURCE", "command_attestations."+command+".source must be one of: "+strings.Join(sortedSet(allowedSources), ", "))
		}
		if status == "pending" || status == "not_recorded" || status == "skipped" {
			c.requireString(attestation["reason"], "command_attestations."+command+".reason")
		}
		if status == "passed" || status == "failed" {
			c.requireString(attestation["cmd"], "command_attestations."+command+".cmd")
			exitCode, ok := intValue(attestation["exit_code"])
			if !ok {
				c.addError("AGENT_EVIDENCE_ATTESTATION_EXIT_CODE", "command_attestations."+command+".exit_code must be an integer")
			} else if status == "passed" && exitCode != 0 {
				c.addError("AGENT_EVIDENCE_ATTESTATION_EXIT_CODE", "command_attestations."+command+".exit_code must be 0 when status is passed")
			} else if status == "failed" && exitCode == 0 {
				c.addError("AGENT_EVIDENCE_ATTESTATION_EXIT_CODE", "command_attestations."+command+".exit_code must be non-zero when status is failed")
			}
		}
		if status == "covered_by_ci" {
			c.requireString(attestation["ci_job"], "command_attestations."+command+".ci_job")
		}
	}

	agentEvidence := c.requireMapping(commandAttestations["agent_evidence"], "command_attestations.agent_evidence")
	if c.requireString(agentEvidence["status"], "command_attestations.agent_evidence.status") != "passed" {
		c.addError("AGENT_EVIDENCE_ATTESTATION_STATUS", "command_attestations.agent_evidence.status must be passed")
	}
	if c.requireString(agentEvidence["source"], "command_attestations.agent_evidence.source") != "current_process" {
		c.addError("AGENT_EVIDENCE_ATTESTATION_SOURCE", "command_attestations.agent_evidence.source must be current_process")
	}
	agentEvidenceCheck := c.requireMapping(commandAttestations["agent_evidence_check"], "command_attestations.agent_evidence_check")
	if c.requireString(agentEvidenceCheck["status"], "command_attestations.agent_evidence_check.status") != "pending" {
		c.addError("AGENT_EVIDENCE_ATTESTATION_STATUS", "command_attestations.agent_evidence_check.status must be pending")
	}
	if c.requireString(agentEvidenceCheck["source"], "command_attestations.agent_evidence_check.source") != "post_generation" {
		c.addError("AGENT_EVIDENCE_ATTESTATION_SOURCE", "command_attestations.agent_evidence_check.source must be post_generation")
	}
	c.requireString(agentEvidenceCheck["reason"], "command_attestations.agent_evidence_check.reason")
}

func (c *checker) validateEmbeddedChecks(commandAttestations map[string]any) {
	checks := c.requireMapping(c.evidence["checks"], "checks")
	for _, checkName := range []string{"ai_context_check", "change_policy_check"} {
		check := c.requireMapping(checks[checkName], "checks."+checkName)
		if c.requireString(check["status"], "checks."+checkName+".status") != "passed" {
			c.addError("AGENT_EVIDENCE_EMBEDDED_CHECK_STATUS", "checks."+checkName+".status must be passed")
		}
		exitCode, ok := intValue(check["exit_code"])
		if !ok {
			c.addError("AGENT_EVIDENCE_EMBEDDED_CHECK_EXIT_CODE", "checks."+checkName+".exit_code must be an integer")
		} else if exitCode != 0 {
			c.addError("AGENT_EVIDENCE_EMBEDDED_CHECK_EXIT_CODE", "checks."+checkName+".exit_code must be 0")
		}
		c.requireString(check["cmd"], "checks."+checkName+".cmd")
		attestation := c.requireMapping(commandAttestations[checkName], "command_attestations."+checkName)
		if attestation["status"] != check["status"] {
			c.addError("AGENT_EVIDENCE_EMBEDDED_CHECK_ATTESTATION", "command_attestations."+checkName+".status must match checks."+checkName+".status")
		}
		if !sameInt(attestation["exit_code"], check["exit_code"]) {
			c.addError("AGENT_EVIDENCE_EMBEDDED_CHECK_ATTESTATION", "command_attestations."+checkName+".exit_code must match checks."+checkName+".exit_code")
		}
		if attestation["cmd"] != check["cmd"] {
			c.addError("AGENT_EVIDENCE_EMBEDDED_CHECK_ATTESTATION", "command_attestations."+checkName+".cmd must match checks."+checkName+".cmd")
		}
		if attestation["source"] != "embedded_check" {
			c.addError("AGENT_EVIDENCE_EMBEDDED_CHECK_ATTESTATION", "command_attestations."+checkName+".source must be embedded_check")
		}
	}
}

func (c *checker) expectedSecuritySensitivePaths(changedFiles []string) []string {
	facades := map[string]string{}
	for _, entry := range list(c.context["public_facades"]) {
		mapping, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		pkg := stringValue(mapping["package"])
		internal := strings.TrimRight(stringValue(mapping["internal"]), "/")
		if pkg != "" && internal != "" {
			facades[pkg] = internal
		}
	}
	prefixes := []string{}
	for _, pkg := range stringList(c.context["security_sensitive_packages"]) {
		prefixes = append(prefixes, strings.TrimRight(pkg, "/")+"/")
		if internal := facades[pkg]; internal != "" {
			prefixes = append(prefixes, strings.TrimRight(internal, "/")+"/")
		}
	}
	var paths []string
	for _, path := range changedFiles {
		for _, prefix := range prefixes {
			if strings.HasPrefix(path, prefix) {
				paths = append(paths, path)
				break
			}
		}
	}
	return sortedStrings(paths)
}

func (c *checker) validateSecurityReview(detectedPolicies, requiredCommands []string, commandAttestations map[string]any, expectedSecuritySensitivePaths []string, policies map[string]any) {
	securityReview := c.requireMapping(c.evidence["security_review"], "security_review")
	reviewRequired, reviewRequiredOK := boolValue(securityReview["required"])
	if !reviewRequiredOK {
		c.addError("AGENT_EVIDENCE_SECURITY_REVIEW_SCHEMA", "security_review.required must be a boolean")
	}
	securityReviewRequired, securityReviewRequiredOK := boolValue(securityReview["security_review_required"])
	if !securityReviewRequiredOK {
		c.addError("AGENT_EVIDENCE_SECURITY_REVIEW_SCHEMA", "security_review.security_review_required must be a boolean")
	}
	expectedReviewRequired := contains(detectedPolicies, "security_sensitive")
	if reviewRequiredOK && reviewRequired != expectedReviewRequired {
		c.addError("AGENT_EVIDENCE_SECURITY_REVIEW_REQUIRED", fmt.Sprintf("security_review.required must be %s", strings.ToLower(fmt.Sprint(expectedReviewRequired))))
	}
	if securityReviewRequiredOK && securityReviewRequired != expectedReviewRequired {
		c.addError("AGENT_EVIDENCE_SECURITY_REVIEW_REQUIRED", fmt.Sprintf("security_review.security_review_required must be %s", strings.ToLower(fmt.Sprint(expectedReviewRequired))))
	}
	if expectedReviewRequired {
		securityPolicy := mapValue(policies["security_sensitive"])
		if required, ok := boolValue(securityPolicy["security_review_required"]); !ok || !required {
			c.addError("AGENT_EVIDENCE_SECURITY_REVIEW_POLICY", "security_sensitive policy must require security review when security_review is required")
		}
	}
	reviewPaths := c.requireStringList(securityReview["paths"], "security_review.paths")
	if !equalStrings(sortedStrings(reviewPaths), expectedSecuritySensitivePaths) {
		c.addError("AGENT_EVIDENCE_SECURITY_REVIEW_PATHS", fmt.Sprintf("security_review.paths must match changed security-sensitive paths; got %v, want %v", sortedStrings(reviewPaths), expectedSecuritySensitivePaths))
	}
	expectedReviewCommands := []string{}
	if expectedReviewRequired {
		for _, command := range []string{"agent_full_check", "agent_security_check"} {
			if contains(requiredCommands, command) {
				expectedReviewCommands = append(expectedReviewCommands, command)
			}
		}
	}
	reviewCommands := c.requireStringList(securityReview["required_commands"], "security_review.required_commands")
	if !equalStrings(reviewCommands, expectedReviewCommands) {
		c.addError("AGENT_EVIDENCE_SECURITY_REVIEW_COMMANDS", fmt.Sprintf("security_review.required_commands must match security validation commands; got %v, want %v", reviewCommands, expectedReviewCommands))
	}
	reviewAttestations := c.requireMapping(securityReview["command_attestations"], "security_review.command_attestations")
	for _, command := range reviewCommands {
		reviewAttestation := c.requireMapping(reviewAttestations[command], "security_review.command_attestations."+command)
		topAttestation := c.requireMapping(commandAttestations[command], "command_attestations."+command)
		for _, key := range []string{"status", "source", "cmd"} {
			if reviewAttestation[key] != topAttestation[key] {
				c.addError("AGENT_EVIDENCE_SECURITY_REVIEW_ATTESTATION", fmt.Sprintf("security_review.command_attestations.%s.%s must match command_attestations.%s.%s", command, key, command, key))
			}
		}
		if _, ok := topAttestation["exit_code"]; ok && !sameInt(reviewAttestation["exit_code"], topAttestation["exit_code"]) {
			c.addError("AGENT_EVIDENCE_SECURITY_REVIEW_ATTESTATION", "security_review.command_attestations."+command+".exit_code must match command_attestations."+command+".exit_code")
		}
		status := stringValue(topAttestation["status"])
		if status == "skipped" || status == "not_recorded" {
			c.requireString(reviewAttestation["reason"], "security_review.command_attestations."+command+".reason")
		}
		if status == "covered_by_ci" {
			c.requireString(reviewAttestation["ci_job"], "security_review.command_attestations."+command+".ci_job")
		}
	}
	reviewStatus := c.requireString(securityReview["status"], "security_review.status")
	if reviewStatus != "" && !setOf("not_required", "blocked", "ready")[reviewStatus] {
		c.addError("AGENT_EVIDENCE_SECURITY_REVIEW_STATUS", "security_review.status must be one of: blocked, not_required, ready")
	}
	reviewReady := expectedReviewRequired && len(expectedSecuritySensitivePaths) > 0
	for _, command := range expectedReviewCommands {
		status := stringValue(mapValue(commandAttestations[command])["status"])
		if status != "passed" && status != "covered_by_ci" {
			reviewReady = false
		}
	}
	expectedReviewStatus := "not_required"
	if expectedReviewRequired {
		if reviewReady {
			expectedReviewStatus = "ready"
		} else {
			expectedReviewStatus = "blocked"
		}
	}
	if reviewStatus != "" && reviewStatus != expectedReviewStatus {
		c.addError("AGENT_EVIDENCE_SECURITY_REVIEW_STATUS", "security_review.status must be "+expectedReviewStatus)
	}
	auditConclusion := c.requireString(securityReview["audit_conclusion"], "security_review.audit_conclusion")
	if auditConclusion != "" {
		if expectedReviewStatus == "ready" && !strings.Contains(auditConclusion, "validation attestations") {
			c.addError("AGENT_EVIDENCE_SECURITY_REVIEW_AUDIT", "security_review.audit_conclusion must describe validation attestations when ready")
		}
		if expectedReviewStatus == "blocked" && !strings.Contains(strings.ToLower(auditConclusion), "blocked") {
			c.addError("AGENT_EVIDENCE_SECURITY_REVIEW_AUDIT", "security_review.audit_conclusion must explain the blocked security review")
		}
	}
}

func (c *checker) validateSecuritySensitiveDiff(commandAttestations map[string]any, expectedSecuritySensitivePaths []string) {
	checks := c.requireMapping(c.evidence["checks"], "checks")
	check := c.requireMapping(checks["security_sensitive_diff"], "checks.security_sensitive_diff")
	attestation := c.requireMapping(commandAttestations["security_sensitive_diff"], "command_attestations.security_sensitive_diff")
	status := c.requireString(check["status"], "checks.security_sensitive_diff.status")
	exitCode, exitCodeOK := intValue(check["exit_code"])
	if !exitCodeOK {
		c.addError("AGENT_EVIDENCE_SECURITY_DIFF_EXIT_CODE", "checks.security_sensitive_diff.exit_code must be an integer")
	}
	c.requireString(check["cmd"], "checks.security_sensitive_diff.cmd")
	if attestation["status"] != status {
		c.addError("AGENT_EVIDENCE_SECURITY_DIFF_ATTESTATION", "command_attestations.security_sensitive_diff.status must match checks.security_sensitive_diff.status")
	}
	if !sameInt(attestation["exit_code"], check["exit_code"]) {
		c.addError("AGENT_EVIDENCE_SECURITY_DIFF_ATTESTATION", "command_attestations.security_sensitive_diff.exit_code must match checks.security_sensitive_diff.exit_code")
	}
	if attestation["cmd"] != check["cmd"] {
		c.addError("AGENT_EVIDENCE_SECURITY_DIFF_ATTESTATION", "command_attestations.security_sensitive_diff.cmd must match checks.security_sensitive_diff.cmd")
	}
	if attestation["source"] != "embedded_check" {
		c.addError("AGENT_EVIDENCE_SECURITY_DIFF_ATTESTATION", "command_attestations.security_sensitive_diff.source must be embedded_check")
	}
	stdout := stringValue(check["stdout"])
	stderr := stringValue(check["stderr"])
	if len(expectedSecuritySensitivePaths) == 0 {
		if status != "passed" {
			c.addError("AGENT_EVIDENCE_SECURITY_DIFF_STATUS", "checks.security_sensitive_diff.status must be passed when no security-sensitive paths changed")
		}
		if exitCodeOK && exitCode != 0 {
			c.addError("AGENT_EVIDENCE_SECURITY_DIFF_EXIT_CODE", "checks.security_sensitive_diff.exit_code must be 0 when no security-sensitive paths changed")
		}
		return
	}
	combined := strings.TrimSpace(stdout + "\n" + stderr)
	if strings.Contains(combined, "no changed files") {
		c.addError("AGENT_EVIDENCE_SECURITY_DIFF_OUTPUT", "checks.security_sensitive_diff output conflicts with security_sensitive_paths")
	}
	for _, path := range expectedSecuritySensitivePaths {
		if !strings.Contains(combined, path) {
			c.addError("AGENT_EVIDENCE_SECURITY_DIFF_OUTPUT", fmt.Sprintf("checks.security_sensitive_diff output must mention changed security-sensitive path %q", path))
		}
	}
	documentationOnly := true
	for _, path := range expectedSecuritySensitivePaths {
		if !strings.HasSuffix(path, "/example_test.go") && !c.isDocGoCommentOnly(path) {
			documentationOnly = false
			break
		}
	}
	if documentationOnly {
		if status != "passed" {
			c.addError("AGENT_EVIDENCE_SECURITY_DIFF_STATUS", "checks.security_sensitive_diff.status may be passed for security-sensitive example/doc-only diffs")
		}
		if exitCodeOK && exitCode != 0 {
			c.addError("AGENT_EVIDENCE_SECURITY_DIFF_EXIT_CODE", "checks.security_sensitive_diff.exit_code must be 0 for security-sensitive example/doc-only diffs")
		}
		if !strings.Contains(combined, "example/doc-only diff") {
			c.addError("AGENT_EVIDENCE_SECURITY_DIFF_OUTPUT", "checks.security_sensitive_diff output must explain security-sensitive example/doc-only diff")
		}
	} else {
		if status != "failed" {
			c.addError("AGENT_EVIDENCE_SECURITY_DIFF_STATUS", "checks.security_sensitive_diff.status must be failed when security-sensitive non-example paths changed")
		}
		if exitCodeOK && exitCode == 0 {
			c.addError("AGENT_EVIDENCE_SECURITY_DIFF_EXIT_CODE", "checks.security_sensitive_diff.exit_code must be non-zero when security-sensitive non-example paths changed")
		}
	}
}

func (c *checker) validateMergeReady(requiredCommands, detectedPolicies []string, commandAttestations map[string]any) {
	expectedMergeBlockers := []string{}
	for _, command := range requiredCommands {
		if !attestationReady(command, mapValue(commandAttestations[command])) {
			expectedMergeBlockers = append(expectedMergeBlockers, command)
		}
	}
	checks := c.requireMapping(c.evidence["checks"], "checks")
	if mapValue(checks["ai_context_check"])["status"] != "passed" {
		expectedMergeBlockers = append(expectedMergeBlockers, "ai_context_check")
	}
	if mapValue(checks["change_policy_check"])["status"] != "passed" {
		expectedMergeBlockers = append(expectedMergeBlockers, "change_policy_check")
	}
	if contains(detectedPolicies, "security_sensitive") {
		if status := stringValue(mapValue(commandAttestations["agent_security_check"])["status"]); status != "passed" && status != "covered_by_ci" {
			expectedMergeBlockers = append(expectedMergeBlockers, "agent_security_check")
		}
		if status := stringValue(mapValue(commandAttestations["agent_full_check"])["status"]); status != "passed" && status != "covered_by_ci" {
			expectedMergeBlockers = append(expectedMergeBlockers, "agent_full_check")
		}
	}
	expectedMergeBlockers = uniqueSorted(expectedMergeBlockers)
	mergeReady, ok := boolValue(c.evidence["merge_ready"])
	if !ok {
		c.addError("AGENT_EVIDENCE_MERGE_READY_SCHEMA", "merge_ready must be a boolean")
	} else if mergeReady != (len(expectedMergeBlockers) == 0) {
		c.addError("AGENT_EVIDENCE_MERGE_READY_MISMATCH", fmt.Sprintf("merge_ready must be %s", strings.ToLower(fmt.Sprint(len(expectedMergeBlockers) == 0))))
	}
	mergeBlockers := c.requireStringList(c.evidence["merge_blockers"], "merge_blockers")
	if !equalStrings(sortedStrings(mergeBlockers), expectedMergeBlockers) {
		c.addError("AGENT_EVIDENCE_MERGE_BLOCKERS_MISMATCH", fmt.Sprintf("merge_blockers must be %v, got %v", expectedMergeBlockers, sortedStrings(mergeBlockers)))
	}
}

func attestationReady(command string, attestation map[string]any) bool {
	if command == "security_sensitive_diff" {
		return true
	}
	status := stringValue(attestation["status"])
	if command == "agent_evidence_check" {
		return status == "pending"
	}
	return status == "passed" || status == "covered_by_ci"
}

func (c *checker) isDocGoCommentOnly(path string) bool {
	if !strings.HasSuffix(path, "/doc.go") {
		return false
	}
	data, err := os.ReadFile(filepath.Join(c.root, path))
	if err != nil {
		return false
	}
	inBlockComment := false
	seenPackage := false
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if inBlockComment {
			if strings.Contains(line, "*/") {
				inBlockComment = false
				line = strings.TrimSpace(strings.SplitN(line, "*/", 2)[1])
				if line == "" {
					continue
				}
			} else {
				continue
			}
		}
		if strings.HasPrefix(line, "//") {
			continue
		}
		if strings.HasPrefix(line, "/*") {
			if !strings.Contains(line, "*/") {
				inBlockComment = true
				continue
			}
			line = strings.TrimSpace(strings.SplitN(line, "*/", 2)[1])
			if line == "" {
				continue
			}
		}
		if strings.HasPrefix(line, "package ") {
			seenPackage = true
			continue
		}
		return false
	}
	return seenPackage
}

func loadJSON(path, label string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("missing %s: %s", label, path)
		}
		return nil, fmt.Errorf("cannot read %s: %w", label, err)
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("invalid %s: %w", label, err)
	}
	return out, nil
}

func (c *checker) addError(ruleID, message string) {
	c.errors = append(c.errors, ruleError{id: ruleID, message: message})
}

func (c *checker) requireMapping(value any, path string) map[string]any {
	mapping, ok := value.(map[string]any)
	if !ok {
		c.addError("AGENT_EVIDENCE_SCHEMA_INVALID", path+" must be an object")
		return map[string]any{}
	}
	return mapping
}

func (c *checker) requireString(value any, path string) string {
	text, ok := value.(string)
	if !ok || strings.TrimSpace(text) == "" {
		c.addError("AGENT_EVIDENCE_SCHEMA_INVALID", path+" must be a non-empty string")
		return ""
	}
	return text
}

func (c *checker) requireOptionalString(value any, path string) string {
	if value == nil {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		c.addError("AGENT_EVIDENCE_SCHEMA_INVALID", path+" must be a string")
		return ""
	}
	return text
}

func (c *checker) requireStringList(value any, path string) []string {
	values, ok := value.([]any)
	if !ok {
		c.addError("AGENT_EVIDENCE_SCHEMA_INVALID", path+" must be a list")
		return nil
	}
	out := []string{}
	for i, item := range values {
		text, ok := item.(string)
		if !ok || strings.TrimSpace(text) == "" {
			c.addError("AGENT_EVIDENCE_SCHEMA_INVALID", fmt.Sprintf("%s[%d] must be a non-empty string", path, i))
			continue
		}
		out = append(out, text)
	}
	return out
}

func stringValue(value any) string {
	text, _ := value.(string)
	return text
}

func list(value any) []any {
	values, _ := value.([]any)
	return values
}

func mapValue(value any) map[string]any {
	mapping, _ := value.(map[string]any)
	if mapping == nil {
		return map[string]any{}
	}
	return mapping
}

func stringList(value any) []string {
	values, ok := value.([]any)
	if !ok {
		return nil
	}
	out := []string{}
	for _, item := range values {
		if text, ok := item.(string); ok {
			out = append(out, text)
		}
	}
	return out
}

func intValue(value any) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		if v == float64(int(v)) {
			return int(v), true
		}
		return 0, false
	default:
		return 0, false
	}
}

func boolValue(value any) (bool, bool) {
	v, ok := value.(bool)
	return v, ok
}

func sameInt(left, right any) bool {
	l, lok := intValue(left)
	r, rok := intValue(right)
	return lok && rok && l == r
}

func mapKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func sortedStrings(values []string) []string {
	out := append([]string(nil), values...)
	sort.Strings(out)
	return out
}

func uniqueSorted(values []string) []string {
	seen := map[string]struct{}{}
	for _, value := range values {
		seen[value] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for value := range seen {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func differenceStrings(values []string, allowed []string) []string {
	allowedSet := map[string]struct{}{}
	for _, value := range allowed {
		allowedSet[value] = struct{}{}
	}
	var out []string
	for _, value := range values {
		if _, ok := allowedSet[value]; !ok {
			out = append(out, value)
		}
	}
	return uniqueSorted(out)
}

func setOf(values ...string) map[string]bool {
	out := map[string]bool{}
	for _, value := range values {
		out[value] = true
	}
	return out
}

func sortedSet(values map[string]bool) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}
