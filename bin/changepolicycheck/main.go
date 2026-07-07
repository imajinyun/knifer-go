package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/imajinyun/knifer-go/bin/internal/govreport"
)

const diffFilter = "ACDMRTUXB"

var policyRuleIDs = map[string]string{
	"bug_fix":            "CHANGE_BUG_FIX",
	"ci_governance":      "CHANGE_CI_GOVERNANCE",
	"dependency_change":  "CHANGE_DEPENDENCY",
	"documentation":      "CHANGE_DOCS",
	"internal_refactor":  "CHANGE_INTERNAL_REFACTOR",
	"public_api":         "CHANGE_PUBLIC_API",
	"security_sensitive": "CHANGE_SECURITY_SENSITIVE",
}

var aiContextSemanticFields = map[string]string{
	"api_freeze":                  "SEMANTIC_AI_CONTEXT_API_FREEZE_CHANGE",
	"coverage_gates":              "SEMANTIC_AI_CONTEXT_COVERAGE_GATES_CHANGE",
	"random_source_policy":        "SEMANTIC_AI_CONTEXT_RANDOM_SOURCE_POLICY_CHANGE",
	"security_sensitive_packages": "SEMANTIC_AI_CONTEXT_SECURITY_SENSITIVE_PACKAGES_CHANGE",
	"threat_model":                "SEMANTIC_AI_CONTEXT_THREAT_MODEL_CHANGE",
}

type checker struct {
	root string
	data map[string]any
}

type changePolicyReport struct {
	Status           string              `json:"status"`
	Findings         []govreport.Finding `json:"findings"`
	DetectedPolicies []string            `json:"detected_policies"`
	RuleIDs          []string            `json:"rule_ids"`
	SemanticRuleIDs  []string            `json:"semantic_rule_ids"`
	RequiredCommands []string            `json:"required_commands"`
	PolicyPaths      map[string][]string `json:"policy_paths"`
}

func main() {
	rootFlag := flag.String("root", "", "repository root")
	changedFilesFlag := flag.String("changed-files", "", "newline-separated changed files override")
	jsonFlag := flag.Bool("json", false, "emit machine-readable JSON output")
	flag.Parse()

	root := strings.TrimSpace(*rootFlag)
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			report := failedReport("CHANGE_POLICY_INPUT_ERROR", "", fmt.Sprintf("cannot resolve working directory: %v", err))
			writeReport(*jsonFlag, report)
			os.Exit(1)
		}
	}
	data, err := loadJSON(filepath.Join(root, "ai-context.json"))
	if err != nil {
		report := failedReport("CHANGE_POLICY_INPUT_ERROR", "ai-context.json", err.Error())
		writeReport(*jsonFlag, report)
		os.Exit(1)
	}
	changedFiles := strings.TrimSpace(*changedFilesFlag)
	if changedFiles == "" {
		changedFiles = os.Getenv("CHANGE_POLICY_CHANGED_FILES")
	}
	c := checker{root: root, data: data}
	report := c.run(changedFiles)
	writeReport(*jsonFlag, report)
	if report.Status == govreport.StatusFailed {
		os.Exit(1)
	}
}

func (c checker) run(changedFilesOverride string) changePolicyReport {
	policies := mapValue(c.data["change_type_policies"])
	facades := map[string]string{}
	for _, entry := range list(c.data["public_facades"]) {
		mapping := mapValue(entry)
		pkg := stringValue(mapping["package"])
		internal := strings.TrimRight(stringValue(mapping["internal"]), "/")
		if pkg != "" {
			facades[pkg] = internal
		}
	}
	securityPrefixes := map[string]struct{}{}
	for _, pkg := range stringList(c.data["security_sensitive_packages"]) {
		securityPrefixes[strings.TrimRight(pkg, "/")+"/"] = struct{}{}
		if internal := facades[pkg]; internal != "" {
			securityPrefixes[strings.TrimRight(internal, "/")+"/"] = struct{}{}
		}
	}

	changedFiles := parseChangedFiles(changedFilesOverride)
	if len(changedFiles) == 0 && strings.TrimSpace(changedFilesOverride) == "" {
		changedFiles = c.changedFilesFromGit()
	}
	detected := map[string]struct{}{}
	matched := map[string]map[string]struct{}{}
	for policy := range policies {
		matched[policy] = map[string]struct{}{}
	}

	for _, path := range changedFiles {
		if path == "go.mod" || path == "go.sum" {
			addPolicy(detected, matched, "dependency_change", path)
		}
		if path == "ai-context.json" || path == "Makefile" || strings.HasPrefix(path, ".github/") || strings.HasPrefix(path, "bin/check_") || strings.HasPrefix(path, "bin/agent_") {
			addPolicy(detected, matched, "ci_governance", path)
		}
		if path == "docs/api/exports.txt" {
			addPolicy(detected, matched, "public_api", path)
		}
		if strings.HasSuffix(path, ".md") || path == "CLAUDE.md" || path == "llms.txt" || strings.HasPrefix(path, "docs/") {
			addPolicy(detected, matched, "documentation", path)
		}
		if hasPrefix(path, securityPrefixes) {
			addPolicy(detected, matched, "security_sensitive", path)
		}
		facadePath := ""
		for pkg := range facades {
			if strings.HasPrefix(path, pkg+"/") {
				facadePath = pkg
				break
			}
		}
		if facadePath != "" && strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			addPolicy(detected, matched, "public_api", path)
		} else if facadePath != "" && strings.HasSuffix(path, "_test.go") {
			addPolicy(detected, matched, "bug_fix", path)
		}
		if strings.HasPrefix(path, "internal/") && !hasPrefix(path, securityPrefixes) {
			if strings.HasSuffix(path, "_test.go") {
				addPolicy(detected, matched, "bug_fix", path)
			} else {
				addPolicy(detected, matched, "internal_refactor", path)
			}
		}
	}

	if len(changedFiles) == 0 {
		return changePolicyReport{
			Status:           govreport.StatusPassed,
			Findings:         []govreport.Finding{},
			DetectedPolicies: []string{},
			RuleIDs:          []string{},
			SemanticRuleIDs:  []string{},
			RequiredCommands: []string{},
			PolicyPaths:      map[string][]string{},
		}
	}
	if len(detected) == 0 {
		for _, path := range changedFiles {
			addPolicy(detected, matched, "bug_fix", path)
		}
	}

	var unknown []string
	for policy := range detected {
		if _, ok := policies[policy]; !ok {
			unknown = append(unknown, policy)
		}
	}
	sort.Strings(unknown)
	if len(unknown) > 0 {
		return failedReport("CHANGE_POLICY_UNKNOWN_POLICY", "", "detected unknown policies: "+strings.Join(unknown, ", "))
	}

	detectedPolicies := sortedSet(detected)
	requiredCommands := []string{}
	for _, policy := range detectedPolicies {
		for _, command := range stringList(mapValue(policies[policy])["required_commands"]) {
			if !contains(requiredCommands, command) {
				requiredCommands = append(requiredCommands, command)
			}
		}
	}

	policyPaths := map[string][]string{}
	for _, policy := range detectedPolicies {
		paths := sortedSet(matched[policy])
		if len(paths) == 0 {
			continue
		}
		policyPaths[policy] = paths
	}
	return changePolicyReport{
		Status:           govreport.StatusPassed,
		Findings:         []govreport.Finding{},
		DetectedPolicies: detectedPolicies,
		RuleIDs:          ruleIDsForPolicies(detectedPolicies),
		SemanticRuleIDs:  c.semanticRuleIDs(changedFiles),
		RequiredCommands: requiredCommands,
		PolicyPaths:      policyPaths,
	}
}

func (c checker) changedFilesFromGit() []string {
	files := map[string]struct{}{}
	baseRef := os.Getenv("AGENT_CHANGE_BASE_REF")
	if baseRef == "" && os.Getenv("GITHUB_BASE_REF") != "" {
		baseRef = "origin/" + os.Getenv("GITHUB_BASE_REF")
	}
	if baseRef != "" && gitOK(c.root, "rev-parse", "--verify", "--quiet", baseRef+"^{commit}") {
		for _, file := range gitLines(c.root, "diff", "--name-only", "--diff-filter="+diffFilter, baseRef+"...HEAD", "--") {
			files[strings.Trim(file, "/")] = struct{}{}
		}
	}
	for _, args := range [][]string{
		{"diff", "--name-only", "--diff-filter=" + diffFilter, "HEAD", "--"},
		{"diff", "--name-only", "--cached", "--diff-filter=" + diffFilter, "--"},
		{"ls-files", "--others", "--exclude-standard", "--"},
	} {
		for _, file := range gitLines(c.root, args...) {
			files[strings.Trim(file, "/")] = struct{}{}
		}
	}
	return sortedSet(files)
}

func (c checker) changedDiffFromGit() string {
	var chunks []string
	baseRef := os.Getenv("AGENT_CHANGE_BASE_REF")
	if baseRef == "" && os.Getenv("GITHUB_BASE_REF") != "" {
		baseRef = "origin/" + os.Getenv("GITHUB_BASE_REF")
	}
	if baseRef != "" && gitOK(c.root, "rev-parse", "--verify", "--quiet", baseRef+"^{commit}") {
		chunks = append(chunks, gitOutput(c.root, "diff", "--unified=0", "--diff-filter="+diffFilter, baseRef+"...HEAD", "--", "ai-context.json", "Makefile"))
	}
	for _, args := range [][]string{
		{"diff", "--unified=0", "--diff-filter=" + diffFilter, "HEAD", "--", "ai-context.json", "Makefile"},
		{"diff", "--cached", "--unified=0", "--diff-filter=" + diffFilter, "--", "ai-context.json", "Makefile"},
	} {
		chunks = append(chunks, gitOutput(c.root, args...))
	}
	return strings.TrimSpace(strings.Join(chunks, "\n"))
}

func addPolicy(detected map[string]struct{}, matched map[string]map[string]struct{}, policy, path string) {
	detected[policy] = struct{}{}
	if matched[policy] == nil {
		matched[policy] = map[string]struct{}{}
	}
	matched[policy][path] = struct{}{}
}

func hasPrefix(path string, prefixes map[string]struct{}) bool {
	for prefix := range prefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func parseChangedFiles(value string) []string {
	seen := map[string]struct{}{}
	for _, line := range strings.Split(value, "\n") {
		line = strings.Trim(strings.TrimSpace(line), "/")
		if line != "" {
			seen[line] = struct{}{}
		}
	}
	return sortedSet(seen)
}

func ruleIDsForPolicies(policies []string) []string {
	var ids []string
	for _, policy := range policies {
		if id := policyRuleIDs[policy]; id != "" {
			ids = append(ids, id)
		}
	}
	return ids
}

func (c checker) semanticRuleIDs(paths []string) []string {
	diffText := strings.TrimSpace(os.Getenv("CHANGE_POLICY_DIFF"))
	if diffText == "" {
		diffText = c.changedDiffFromGit()
	}
	if diffText != "" {
		if ids := c.semanticRuleIDsFromDiff(diffText, paths); len(ids) > 0 {
			return ids
		}
		if diffMentionsAnyFile(diffText, "ai-context.json", "Makefile") {
			return semanticRuleIDsFromPaths(withoutPaths(paths, "ai-context.json", "Makefile"))
		}
	}
	return semanticRuleIDsFromPaths(paths)
}

func semanticRuleIDsFromPaths(paths []string) []string {
	ids := map[string]struct{}{}
	for _, path := range paths {
		switch {
		case path == "ai-context.schema.json":
			ids["SEMANTIC_SCHEMA_CONTRACT_CHANGE"] = struct{}{}
		case path == "ai-context.json":
			ids["SEMANTIC_AI_CONTEXT_CHANGE"] = struct{}{}
		case path == "Makefile":
			ids["SEMANTIC_MAKEFILE_CHANGE"] = struct{}{}
		}
		if path == "Makefile" || strings.HasPrefix(path, ".github/workflows/") {
			ids["SEMANTIC_RELEASE_GATE_CHANGE"] = struct{}{}
		}
		if path == "ai-context.json" || path == "ai-context.schema.json" || strings.Contains(path, "coverage") {
			ids["SEMANTIC_COVERAGE_POLICY_CHANGE"] = struct{}{}
		}
		if path == "ai-context.json" || path == "ai-context.schema.json" || strings.Contains(path, "api_freeze") || strings.Contains(path, "api-freeze") {
			ids["SEMANTIC_API_FREEZE_POLICY_CHANGE"] = struct{}{}
		}
		if path == "ai-context.json" || path == "ai-context.schema.json" || strings.Contains(path, "security") || strings.Contains(path, "crypto") || strings.Contains(path, "jwt") || strings.Contains(path, "rand") {
			ids["SEMANTIC_SECURITY_POLICY_CHANGE"] = struct{}{}
		}
	}
	return sortedSet(ids)
}

func (c checker) semanticRuleIDsFromDiff(diffText string, paths []string) []string {
	ids := map[string]struct{}{}
	for _, path := range paths {
		if path == "ai-context.schema.json" {
			ids["SEMANTIC_SCHEMA_CONTRACT_CHANGE"] = struct{}{}
		}
	}

	changes := changedLinesByFile(diffText)
	if change := changes["ai-context.json"]; len(change.lines) > 0 || len(change.texts) > 0 {
		ranges := topLevelJSONRanges(filepath.Join(c.root, "ai-context.json"))
		for _, line := range change.lines {
			if id := semanticIDForRangeLine(ranges, line, aiContextSemanticFields); id != "" {
				ids[id] = struct{}{}
			}
		}
		for _, text := range change.texts {
			if id := semanticIDForJSONLine(text); id != "" {
				ids[id] = struct{}{}
			}
		}
	}
	if change := changes["Makefile"]; len(change.lines) > 0 || len(change.texts) > 0 {
		ranges := makefileTargetRanges(filepath.Join(c.root, "Makefile"))
		for _, line := range change.lines {
			if rangeNameForLine(ranges, line) == "release-check" {
				ids["SEMANTIC_MAKEFILE_RELEASE_CHECK_CHANGE"] = struct{}{}
			}
		}
		for _, text := range change.texts {
			if strings.HasPrefix(strings.TrimSpace(text), "release-check:") {
				ids["SEMANTIC_MAKEFILE_RELEASE_CHECK_CHANGE"] = struct{}{}
			}
		}
	}
	return sortedSet(ids)
}

type fileChange struct {
	lines []int
	texts []string
}

func changedLinesByFile(diffText string) map[string]fileChange {
	changes := map[string]fileChange{}
	currentPath := ""
	oldLine := 0
	newLine := 0
	inHunk := false
	for _, line := range strings.Split(diffText, "\n") {
		switch {
		case strings.HasPrefix(line, "+++ "):
			currentPath = parseDiffPath(strings.TrimSpace(strings.TrimPrefix(line, "+++ ")))
			inHunk = false
			continue
		case strings.HasPrefix(line, "@@ "):
			oldStart, newStart, ok := parseHunkStart(line)
			if !ok {
				inHunk = false
				continue
			}
			oldLine = oldStart
			newLine = newStart
			inHunk = true
			continue
		}
		if currentPath == "" || !inHunk || line == `\ No newline at end of file` {
			continue
		}
		change := changes[currentPath]
		switch {
		case strings.HasPrefix(line, "+"):
			change.lines = append(change.lines, newLine)
			change.texts = append(change.texts, strings.TrimSpace(strings.TrimPrefix(line, "+")))
			newLine++
		case strings.HasPrefix(line, "-"):
			change.lines = append(change.lines, max(newLine, 1))
			change.texts = append(change.texts, strings.TrimSpace(strings.TrimPrefix(line, "-")))
			oldLine++
		default:
			oldLine++
			newLine++
		}
		changes[currentPath] = change
	}
	return changes
}

func parseDiffPath(path string) string {
	if path == "/dev/null" {
		return ""
	}
	path = strings.TrimPrefix(path, "b/")
	return strings.Trim(path, "/")
}

func diffMentionsAnyFile(diffText string, paths ...string) bool {
	wanted := map[string]struct{}{}
	for _, path := range paths {
		wanted[path] = struct{}{}
	}
	for _, line := range strings.Split(diffText, "\n") {
		if !strings.HasPrefix(line, "+++ ") && !strings.HasPrefix(line, "--- ") {
			continue
		}
		path := parseDiffPath(strings.TrimSpace(line[4:]))
		if _, ok := wanted[path]; ok {
			return true
		}
	}
	return false
}

func withoutPaths(paths []string, excluded ...string) []string {
	excludedSet := map[string]struct{}{}
	for _, path := range excluded {
		excludedSet[path] = struct{}{}
	}
	var out []string
	for _, path := range paths {
		if _, ok := excludedSet[path]; ok {
			continue
		}
		out = append(out, path)
	}
	return out
}

var hunkStartPattern = regexp.MustCompile(`^@@ -(\d+)(?:,\d+)? \+(\d+)(?:,\d+)? @@`)

func parseHunkStart(line string) (int, int, bool) {
	match := hunkStartPattern.FindStringSubmatch(line)
	if len(match) != 3 {
		return 0, 0, false
	}
	oldStart := atoi(match[1])
	newStart := atoi(match[2])
	return oldStart, newStart, true
}

type namedRange struct {
	name  string
	start int
	end   int
}

func topLevelJSONRanges(path string) []namedRange {
	lines := readLines(path)
	var ranges []namedRange
	keyPattern := regexp.MustCompile(`^  "([^"]+)":`)
	for index, line := range lines {
		match := keyPattern.FindStringSubmatch(line)
		if len(match) != 2 {
			continue
		}
		if len(ranges) > 0 {
			ranges[len(ranges)-1].end = index
		}
		ranges = append(ranges, namedRange{name: match[1], start: index + 1, end: len(lines)})
	}
	return ranges
}

func makefileTargetRanges(path string) []namedRange {
	lines := readLines(path)
	targetPattern := regexp.MustCompile(`^([A-Za-z0-9_.-]+):(?:\s|$)`)
	var ranges []namedRange
	for index, line := range lines {
		match := targetPattern.FindStringSubmatch(line)
		if len(match) != 2 {
			continue
		}
		if len(ranges) > 0 {
			ranges[len(ranges)-1].end = index
		}
		ranges = append(ranges, namedRange{name: match[1], start: index + 1, end: len(lines)})
	}
	return ranges
}

func semanticIDForRangeLine(ranges []namedRange, line int, idsByName map[string]string) string {
	name := rangeNameForLine(ranges, line)
	return idsByName[name]
}

func rangeNameForLine(ranges []namedRange, line int) string {
	for _, candidate := range ranges {
		if line >= candidate.start && line <= candidate.end {
			return candidate.name
		}
	}
	return ""
}

func semanticIDForJSONLine(line string) string {
	keyPattern := regexp.MustCompile(`"([^"]+)":`)
	match := keyPattern.FindStringSubmatch(line)
	if len(match) != 2 {
		return ""
	}
	return aiContextSemanticFields[match[1]]
}

func readLines(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	return strings.Split(string(data), "\n")
}

func atoi(value string) int {
	var out int
	for _, r := range value {
		if r < '0' || r > '9' {
			return out
		}
		out = out*10 + int(r-'0')
	}
	return out
}

func gitOK(root string, args ...string) bool {
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	return cmd.Run() == nil
}

func gitLines(root string, args ...string) []string {
	out := gitOutput(root, args...)
	if out == "" {
		return nil
	}
	var lines []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func gitOutput(root string, args ...string) string {
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(out)
}

func failedReport(ruleID, path, message string) changePolicyReport {
	return changePolicyReport{
		Status:           govreport.StatusFailed,
		Findings:         []govreport.Finding{govreport.Error(ruleID, path, message)},
		DetectedPolicies: []string{},
		RuleIDs:          []string{},
		SemanticRuleIDs:  []string{},
		RequiredCommands: []string{},
		PolicyPaths:      map[string][]string{},
	}
}

func writeReport(jsonOutput bool, report changePolicyReport) {
	if jsonOutput {
		if err := govreport.WriteJSON(os.Stdout, report); err != nil {
			fmt.Fprintf(os.Stderr, "CHANGE POLICY CHECK ERROR: cannot encode JSON output: %v\n", err)
		}
		return
	}
	if report.Status == govreport.StatusFailed {
		for _, finding := range report.Findings {
			fmt.Fprintf(os.Stderr, "CHANGE POLICY CHECK ERROR: %s\n", finding.Message)
		}
		return
	}
	if len(report.DetectedPolicies) == 0 {
		fmt.Println("change policy check passed: no changed files")
		return
	}
	fmt.Println("change policy check passed")
	fmt.Println("detected policies: " + strings.Join(report.DetectedPolicies, ", "))
	fmt.Println("rule ids: " + strings.Join(report.RuleIDs, ", "))
	if len(report.SemanticRuleIDs) > 0 {
		fmt.Println("semantic rule ids: " + strings.Join(report.SemanticRuleIDs, ", "))
	}
	fmt.Println("required commands: " + strings.Join(report.RequiredCommands, ", "))
	for _, policy := range report.DetectedPolicies {
		paths := report.PolicyPaths[policy]
		if len(paths) == 0 {
			continue
		}
		fmt.Printf("%s paths:\n", policy)
		for _, path := range paths {
			fmt.Printf("  - %s\n", path)
		}
	}
}

func loadJSON(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("missing ai-context.json")
		}
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func mapValue(value any) map[string]any {
	mapping, _ := value.(map[string]any)
	if mapping == nil {
		return map[string]any{}
	}
	return mapping
}

func list(value any) []any {
	values, _ := value.([]any)
	return values
}

func stringValue(value any) string {
	text, _ := value.(string)
	return text
}

func stringList(value any) []string {
	values, ok := value.([]any)
	if !ok {
		return nil
	}
	var out []string
	for _, value := range values {
		if text, ok := value.(string); ok {
			out = append(out, text)
		}
	}
	return out
}

func sortedSet(values map[string]struct{}) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		if value != "" {
			out = append(out, value)
		}
	}
	sort.Strings(out)
	return out
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
