package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
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

type checker struct {
	root string
	data map[string]any
}

func main() {
	rootFlag := flag.String("root", "", "repository root")
	changedFilesFlag := flag.String("changed-files", "", "newline-separated changed files override")
	flag.Parse()

	root := strings.TrimSpace(*rootFlag)
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "CHANGE POLICY CHECK ERROR: cannot resolve working directory: %v\n", err)
			os.Exit(1)
		}
	}
	data, err := loadJSON(filepath.Join(root, "ai-context.json"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "CHANGE POLICY CHECK ERROR: %v\n", err)
		os.Exit(1)
	}
	changedFiles := strings.TrimSpace(*changedFilesFlag)
	if changedFiles == "" {
		changedFiles = os.Getenv("CHANGE_POLICY_CHANGED_FILES")
	}
	c := checker{root: root, data: data}
	c.run(changedFiles)
}

func (c checker) run(changedFilesOverride string) {
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
		fmt.Println("change policy check passed: no changed files")
		return
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
		fmt.Fprintln(os.Stderr, "CHANGE POLICY CHECK ERROR: detected unknown policies: "+strings.Join(unknown, ", "))
		os.Exit(1)
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

	fmt.Println("change policy check passed")
	fmt.Println("detected policies: " + strings.Join(detectedPolicies, ", "))
	fmt.Println("rule ids: " + strings.Join(ruleIDsForPolicies(detectedPolicies), ", "))
	fmt.Println("required commands: " + strings.Join(requiredCommands, ", "))
	for _, policy := range detectedPolicies {
		paths := sortedSet(matched[policy])
		if len(paths) == 0 {
			continue
		}
		fmt.Printf("%s paths:\n", policy)
		for _, path := range paths {
			fmt.Printf("  - %s\n", path)
		}
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

func gitOK(root string, args ...string) bool {
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	return cmd.Run() == nil
}

func gitLines(root string, args ...string) []string {
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	var lines []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
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
