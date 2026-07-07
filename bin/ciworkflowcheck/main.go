package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type checker struct {
	root              string
	workflowCount     int
	workflowFileCount int
	findings          []finding
}

type finding struct {
	RuleID   string `json:"rule_id"`
	Path     string `json:"path"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

type checkResult struct {
	Status   string    `json:"status"`
	Findings []finding `json:"findings"`
}

func main() {
	rootFlag := flag.String("root", "", "repository root to validate")
	jsonFlag := flag.Bool("json", false, "emit machine-readable JSON output")
	flag.Parse()

	root := strings.TrimSpace(*rootFlag)
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			findings := []finding{{
				RuleID:   "CI_WORKFLOW_CHECK_INPUT_ERROR",
				Path:     "",
				Message:  fmt.Sprintf("cannot resolve working directory: %v", err),
				Severity: "error",
			}}
			writeOutput(*jsonFlag, checkResult{Status: "failed", Findings: findings}, 0, 0)
			os.Exit(1)
		}
	}

	c := &checker{root: root}
	if err := c.run(); err != nil {
		c.addError("CI_WORKFLOW_CHECK_INPUT_ERROR", "", err.Error())
	}
	if len(c.findings) > 0 {
		writeOutput(*jsonFlag, checkResult{Status: "failed", Findings: c.findings}, 0, 0)
		os.Exit(1)
	}
	writeOutput(*jsonFlag, checkResult{Status: "passed", Findings: []finding{}}, c.workflowCount, c.workflowFileCount)
}

func (c *checker) run() error {
	data, err := readJSON(filepath.Join(c.root, "ai-context.json"))
	if err != nil {
		return err
	}

	ciWorkflows := c.requireMapping(data["ci_workflows"], "ci_workflows")
	commands := c.requireMapping(data["commands"], "commands")
	commandMakeTargets := map[string]map[string]struct{}{}
	for name, value := range commands {
		spec, ok := value.(map[string]any)
		if !ok {
			continue
		}
		target := makeTargetFromCommand(stringValue(spec["cmd"]))
		if target == "" {
			continue
		}
		if commandMakeTargets[target] == nil {
			commandMakeTargets[target] = map[string]struct{}{}
		}
		commandMakeTargets[target][name] = struct{}{}
	}

	toolVersions := c.requireMapping(ciWorkflows["tool_versions"], "ci_workflows.tool_versions")
	go125Patch := c.requireString(toolVersions["go_1_25_patch"], "ci_workflows.tool_versions.go_1_25_patch")
	golangciLintVersion := c.requireString(toolVersions["golangci_lint"], "ci_workflows.tool_versions.golangci_lint")
	githubActions := c.requireMapping(ciWorkflows["github_actions"], "ci_workflows.github_actions")

	makefileText, err := os.ReadFile(filepath.Join(c.root, "Makefile"))
	if err != nil {
		return fmt.Errorf("missing Makefile: %w", err)
	}
	definedMakeTargets := makeTargets(string(makefileText))
	actualPaths := workflowFiles(filepath.Join(c.root, ".github", "workflows"))
	c.workflowCount = len(githubActions)
	c.workflowFileCount = len(actualPaths)
	declaredPaths := map[string]struct{}{}

	names := sortedKeys(githubActions)
	for _, name := range names {
		workflow := c.requireMapping(githubActions[name], "ci_workflows.github_actions."+name)
		workflowPath := c.requireString(workflow["path"], "ci_workflows.github_actions."+name+".path")
		requiredJobs := c.requireStringList(workflow["required_jobs"], "ci_workflows.github_actions."+name+".required_jobs")
		agentGovernance := c.requireMapping(workflow["agent_governance"], "ci_workflows.github_actions."+name+".agent_governance")
		requiredCommands := c.requireStringList(agentGovernance["required_commands"], "ci_workflows.github_actions."+name+".agent_governance.required_commands")
		requiredEnv := c.requireStringList(agentGovernance["required_env"], "ci_workflows.github_actions."+name+".agent_governance.required_env")
		requiredArtifacts := c.requireStringList(agentGovernance["required_artifacts"], "ci_workflows.github_actions."+name+".agent_governance.required_artifacts")
		if workflowPath == "" {
			continue
		}
		declaredPaths[workflowPath] = struct{}{}
		workflowBytes, err := os.ReadFile(filepath.Join(c.root, workflowPath))
		if err != nil {
			c.addError("CI_WORKFLOW_MISSING_FILE", workflowPath, fmt.Sprintf("ci_workflows.github_actions.%s.path references missing workflow %q", name, workflowPath))
			continue
		}
		workflowText := string(workflowBytes)

		for _, target := range sortedSet(workflowMakeTargets(workflowText)) {
			if _, ok := definedMakeTargets[target]; !ok {
				c.addError("CI_WORKFLOW_UNKNOWN_MAKE_TARGET", workflowPath, fmt.Sprintf("%s references unknown Makefile target %q", workflowPath, target))
			}
		}

		jobs := workflowJobNames(workflowText)
		for _, job := range requiredJobs {
			if _, ok := jobs[job]; !ok {
				c.addError("CI_WORKFLOW_MISSING_REQUIRED_JOB", workflowPath, fmt.Sprintf("%s is missing required job %q", workflowPath, job))
			}
		}

		for _, requiredText := range requiredCommands {
			if !strings.Contains(workflowText, requiredText) {
				c.addError("CI_WORKFLOW_REQUIRED_COMMAND_MISSING", workflowPath, fmt.Sprintf("%s must contain required command %q", workflowPath, requiredText))
			}
			target := makeTargetFromCommand(requiredText)
			if target == "" {
				continue
			}
			if _, ok := definedMakeTargets[target]; !ok {
				c.addError("CI_WORKFLOW_UNKNOWN_MAKE_TARGET", "ai-context.json", fmt.Sprintf("ci_workflows.github_actions.%s required command %q references unknown Makefile target", name, requiredText))
			}
			if _, ok := commandMakeTargets[target]; !ok {
				c.addError("CI_WORKFLOW_COMMAND_NOT_IN_AI_CONTEXT", "ai-context.json", fmt.Sprintf("ci_workflows.github_actions.%s required command %q has no ai-context.commands entry", name, requiredText))
			}
		}
		for _, envName := range requiredEnv {
			if !strings.Contains(workflowText, envName) {
				c.addError("CI_WORKFLOW_REQUIRED_ENV_MISSING", workflowPath, fmt.Sprintf("%s must contain required env %q", workflowPath, envName))
			}
		}
		for _, artifact := range requiredArtifacts {
			if !strings.Contains(workflowText, artifact) {
				c.addError("CI_WORKFLOW_REQUIRED_ARTIFACT_MISSING", workflowPath, fmt.Sprintf("%s must upload required artifact %q", workflowPath, artifact))
			}
		}

		if name == "go" || name == "release" {
			if !strings.Contains(workflowText, "GO_1_25_PATCH_VERSION") || !strings.Contains(workflowText, go125Patch) {
				c.addError("CI_WORKFLOW_VERSION_DRIFT", workflowPath, fmt.Sprintf("%s must use declared Go patch version %q", workflowPath, go125Patch))
			}
		}
		if name == "go" {
			if !strings.Contains(workflowText, "GOLANGCI_LINT_VERSION") || !strings.Contains(workflowText, golangciLintVersion) {
				c.addError("CI_WORKFLOW_VERSION_DRIFT", workflowPath, fmt.Sprintf("%s must use declared golangci-lint version %q", workflowPath, golangciLintVersion))
			}
			matrix := fmt.Sprintf(`go-version: ["%s", "1.26"]`, go125Patch)
			if go125Patch != "" && !strings.Contains(workflowText, matrix) {
				c.addError("CI_WORKFLOW_VERSION_DRIFT", workflowPath, fmt.Sprintf("%s test matrix must include %q and '1.26'", workflowPath, go125Patch))
			}
			for _, duplicateStep := range []string{"make race-test", "make shuffle-test", "make mod-check"} {
				if strings.Contains(workflowText, duplicateStep) {
					c.addError("CI_WORKFLOW_DUPLICATE_SUBSTEP", workflowPath, fmt.Sprintf("%s should not duplicate ci-test sub-step %q", workflowPath, duplicateStep))
				}
			}
		}
	}

	undeclared := difference(actualPaths, declaredPaths)
	missingFiles := difference(declaredPaths, actualPaths)
	if len(undeclared) > 0 {
		c.addError("CI_WORKFLOW_UNDECLARED_FILE", ".github/workflows", "undeclared GitHub workflow file(s): "+strings.Join(undeclared, ", "))
	}
	if len(missingFiles) > 0 {
		c.addError("CI_WORKFLOW_MISSING_FILE", ".github/workflows", "declared GitHub workflow path(s) not present on disk: "+strings.Join(missingFiles, ", "))
	}
	return nil
}

func readJSON(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("missing ai-context.json")
		}
		return nil, fmt.Errorf("cannot read ai-context.json: %w", err)
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("invalid ai-context.json: %w", err)
	}
	return out, nil
}

func (c *checker) addError(ruleID, path, message string) {
	c.findings = append(c.findings, finding{
		RuleID:   ruleID,
		Path:     path,
		Message:  message,
		Severity: "error",
	})
}

func (c *checker) requireMapping(value any, path string) map[string]any {
	mapping, ok := value.(map[string]any)
	if !ok {
		c.addError("CI_WORKFLOW_SCHEMA_INVALID", pathToFile(path), path+" must be an object")
		return map[string]any{}
	}
	return mapping
}

func (c *checker) requireString(value any, path string) string {
	text, ok := value.(string)
	if !ok || strings.TrimSpace(text) == "" {
		c.addError("CI_WORKFLOW_SCHEMA_INVALID", pathToFile(path), path+" must be a non-empty string")
		return ""
	}
	return text
}

func (c *checker) requireStringList(value any, path string) []string {
	values, ok := value.([]any)
	if !ok {
		c.addError("CI_WORKFLOW_SCHEMA_INVALID", pathToFile(path), path+" must be a list")
		return nil
	}
	var out []string
	for i, item := range values {
		text, ok := item.(string)
		if !ok || strings.TrimSpace(text) == "" {
			c.addError("CI_WORKFLOW_SCHEMA_INVALID", pathToFile(path), fmt.Sprintf("%s[%d] must be a non-empty string", path, i))
			continue
		}
		out = append(out, text)
	}
	return out
}

func writeOutput(jsonOutput bool, result checkResult, workflowCount, fileCount int) {
	if jsonOutput {
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "CI WORKFLOW CHECK FAILED:\n- [CI_WORKFLOW_CHECK_INPUT_ERROR] cannot encode JSON output: %v\n", err)
			return
		}
		fmt.Println(string(data))
		return
	}
	if result.Status == "passed" {
		fmt.Printf("CI workflow governance is valid (%d workflows, %d files)\n", workflowCount, fileCount)
		return
	}
	fmt.Fprintln(os.Stderr, "CI WORKFLOW CHECK FAILED:")
	for _, finding := range result.Findings {
		fmt.Fprintf(os.Stderr, "- [%s] %s\n", finding.RuleID, finding.Message)
	}
}

func pathToFile(path string) string {
	if strings.HasPrefix(path, "ci_workflows") || strings.HasPrefix(path, "commands") {
		return "ai-context.json"
	}
	return ""
}

func stringValue(value any) string {
	text, _ := value.(string)
	return text
}

var (
	makeTargetPattern         = regexp.MustCompile(`(?m)^([A-Za-z0-9_.-]+):(?:\s|$)`)
	makeCommandPattern        = regexp.MustCompile(`^\s*make\s+([A-Za-z0-9_.-]+)`)
	workflowMakeTargetPattern = regexp.MustCompile(`(?:^|[\s;&|])make\s+([A-Za-z0-9_.-]+)`)
	jobNamePattern            = regexp.MustCompile(`^  ([A-Za-z0-9_-]+):\s*$`)
)

func makeTargets(makefileText string) map[string]struct{} {
	targets := map[string]struct{}{}
	for _, match := range makeTargetPattern.FindAllStringSubmatch(makefileText, -1) {
		targets[match[1]] = struct{}{}
	}
	return targets
}

func makeTargetFromCommand(command string) string {
	match := makeCommandPattern.FindStringSubmatch(command)
	if len(match) == 0 {
		return ""
	}
	return match[1]
}

func workflowMakeTargets(workflowText string) map[string]struct{} {
	targets := map[string]struct{}{}
	for _, match := range workflowMakeTargetPattern.FindAllStringSubmatch(workflowText, -1) {
		targets[match[1]] = struct{}{}
	}
	return targets
}

func workflowJobNames(workflowText string) map[string]struct{} {
	jobs := map[string]struct{}{}
	inJobs := false
	for _, line := range strings.Split(workflowText, "\n") {
		if regexp.MustCompile(`^jobs:\s*$`).MatchString(line) {
			inJobs = true
			continue
		}
		if inJobs && line != "" && !strings.HasPrefix(line, " ") {
			break
		}
		if !inJobs {
			continue
		}
		match := jobNamePattern.FindStringSubmatch(line)
		if len(match) > 0 {
			jobs[match[1]] = struct{}{}
		}
	}
	return jobs
}

func workflowFiles(workflowDir string) map[string]struct{} {
	entries, err := os.ReadDir(workflowDir)
	if err != nil {
		return map[string]struct{}{}
	}
	paths := map[string]struct{}{}
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasSuffix(name, ".yml") || strings.HasSuffix(name, ".yaml") {
			paths[filepath.ToSlash(filepath.Join(".github", "workflows", name))] = struct{}{}
		}
	}
	return paths
}

func sortedKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedSet(values map[string]struct{}) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func difference(left, right map[string]struct{}) []string {
	var out []string
	for value := range left {
		if _, ok := right[value]; !ok {
			out = append(out, value)
		}
	}
	sort.Strings(out)
	return out
}
