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

	"github.com/imajinyun/knifer-go/bin/internal/govreport"
)

var expectedPackages = setOf("vrand", "vid", "vcrypto", "vjwt")
var expectedPolicyNames = setOf(
	"secure_bytes_fail_closed",
	"compatibility_byte_fallback",
	"identifier_fallback_compatibility",
	"jwt_key_and_signer_policy",
)

type checker struct {
	root     string
	findings []govreport.Finding
}

func main() {
	rootFlag := flag.String("root", "", "repository root")
	contextFlag := flag.String("ai-context", "", "ai-context.json path")
	jsonFlag := flag.Bool("json", false, "emit machine-readable JSON output")
	flag.Parse()

	root := strings.TrimSpace(*rootFlag)
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			writeReport(*jsonFlag, govreport.Failed([]govreport.Finding{
				govreport.Error("RANDOM_SOURCE_POLICY_INPUT_ERROR", "", fmt.Sprintf("cannot resolve working directory: %v", err)),
			}))
			os.Exit(1)
		}
	}
	contextPath := strings.TrimSpace(*contextFlag)
	if contextPath == "" {
		contextPath = filepath.Join(root, "ai-context.json")
	}
	data, err := loadJSON(contextPath)
	if err != nil {
		writeReport(*jsonFlag, govreport.Failed([]govreport.Finding{
			govreport.Error("RANDOM_SOURCE_POLICY_INPUT_ERROR", contextPath, err.Error()),
		}))
		os.Exit(1)
	}

	c := &checker{root: root}
	c.run(data)
	if len(c.findings) > 0 {
		writeReport(*jsonFlag, govreport.Failed(c.findings))
		os.Exit(1)
	}
	writeReport(*jsonFlag, govreport.Passed())
}

func (c *checker) run(data map[string]any) {
	policy := mapValue(data["random_source_policy"])
	if policy == nil {
		c.addError("RANDOM_SOURCE_POLICY_SCHEMA_INVALID", "random_source_policy must be an object")
		return
	}
	packages := setOf(stringList(policy["packages"])...)
	if !sameSet(packages, expectedPackages) {
		c.addError("RANDOM_SOURCE_POLICY_PACKAGE_COVERAGE", "random_source_policy.packages must cover exactly: "+strings.Join(sortedBoolKeys(expectedPackages), ", "))
	}

	policiesRaw, ok := policy["policies"].([]any)
	if !ok || len(policiesRaw) == 0 {
		c.addError("RANDOM_SOURCE_POLICY_SCHEMA_INVALID", "random_source_policy.policies must be a non-empty list")
		return
	}

	seenNames := map[string]bool{}
	coveredPackages := map[string]bool{}
	for index, raw := range policiesRaw {
		entry := mapValue(raw)
		if entry == nil {
			c.addError("RANDOM_SOURCE_POLICY_SCHEMA_INVALID", fmt.Sprintf("random_source_policy.policies[%d] must be an object", index))
			continue
		}
		name := strings.TrimSpace(stringValue(entry["name"]))
		if name == "" {
			c.addError("RANDOM_SOURCE_POLICY_SCHEMA_INVALID", fmt.Sprintf("random_source_policy.policies[%d].name must be non-empty", index))
			continue
		}
		seenNames[name] = true
		if !expectedPolicyNames[name] {
			c.addError("RANDOM_SOURCE_POLICY_UNKNOWN_NAME", "random_source_policy.policies includes unknown policy: "+name)
		}
		entryPackages := stringList(entry["packages"])
		for _, pkg := range entryPackages {
			coveredPackages[pkg] = true
		}
		if unknown := differenceSet(setOf(entryPackages...), expectedPackages); len(unknown) > 0 {
			c.addError("RANDOM_SOURCE_POLICY_UNKNOWN_PACKAGE", fmt.Sprintf("random_source_policy.policies.%s.packages contains unknown package(s): %s", name, strings.Join(unknown, ", ")))
		}
		if strings.TrimSpace(stringValue(entry["behavior"])) == "" {
			c.addError("RANDOM_SOURCE_POLICY_SCHEMA_INVALID", fmt.Sprintf("random_source_policy.policies.%s.behavior must be non-empty", name))
		}
		for _, field := range []string{"allowed_sources", "forbidden_uses", "contract_tests"} {
			items := stringList(entry[field])
			if len(items) == 0 {
				c.addError("RANDOM_SOURCE_POLICY_SCHEMA_INVALID", fmt.Sprintf("random_source_policy.policies.%s.%s must be non-empty", name, field))
			}
		}
		for _, reference := range stringList(entry["contract_tests"]) {
			if !referencesFunction(reference) {
				c.addError("RANDOM_SOURCE_POLICY_CONTRACT_TEST_INVALID", fmt.Sprintf("random_source_policy.policies.%s.contract_tests must reference explicit test functions, got %s", name, reference))
				continue
			}
			if !c.referenceExists(reference) {
				c.addError("RANDOM_SOURCE_POLICY_CONTRACT_TEST_MISSING", fmt.Sprintf("random_source_policy.policies.%s.contract_tests references missing file or function %s", name, reference))
			}
		}
	}
	if missing := differenceSet(expectedPolicyNames, seenNames); len(missing) > 0 {
		c.addError("RANDOM_SOURCE_POLICY_MISSING_NAME", "random_source_policy.policies missing policy/policies: "+strings.Join(missing, ", "))
	}
	if extra := differenceSet(seenNames, expectedPolicyNames); len(extra) > 0 {
		c.addError("RANDOM_SOURCE_POLICY_UNKNOWN_NAME", "random_source_policy.policies includes unknown policy/policies: "+strings.Join(extra, ", "))
	}
	if !sameSet(coveredPackages, expectedPackages) {
		c.addError("RANDOM_SOURCE_POLICY_PACKAGE_COVERAGE", "random_source_policy policies must cover package(s): "+strings.Join(sortedBoolKeys(expectedPackages), ", "))
	}
}

func (c *checker) addError(ruleID, message string) {
	c.findings = append(c.findings, govreport.Error(ruleID, "ai-context.json", message))
}

func writeReport(jsonOutput bool, report govreport.Envelope) {
	if jsonOutput {
		if err := govreport.WriteJSON(os.Stdout, report); err != nil {
			fmt.Fprintf(os.Stderr, "random source policy check error: [RANDOM_SOURCE_POLICY_INPUT_ERROR] cannot encode JSON output: %v\n", err)
		}
		return
	}
	if report.Status == govreport.StatusFailed {
		for _, finding := range report.Findings {
			fmt.Fprintf(os.Stderr, "random source policy check error: [%s] %s\n", finding.RuleID, finding.Message)
		}
		return
	}
	fmt.Println("random source policy governance is valid")
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

var referencePattern = regexp.MustCompile(`^(.+\.go):(Test|Fuzz|Example)[A-Za-z0-9_]*$`)

func referencesFunction(reference string) bool {
	return referencePattern.MatchString(reference)
}

func (c *checker) referenceExists(reference string) bool {
	match := referencePattern.FindStringSubmatch(reference)
	if len(match) != 3 {
		return false
	}
	path := filepath.Join(c.root, filepath.FromSlash(match[1]))
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return regexp.MustCompile(`func\s+` + regexp.QuoteMeta(functionName(reference)) + `\s*\(`).Match(data)
}

func functionName(reference string) string {
	if idx := strings.LastIndex(reference, ":"); idx >= 0 {
		return reference[idx+1:]
	}
	return ""
}

func mapValue(value any) map[string]any {
	mapping, _ := value.(map[string]any)
	return mapping
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
		if text, ok := value.(string); ok && strings.TrimSpace(text) != "" {
			out = append(out, strings.TrimSpace(text))
		}
	}
	return out
}

func setOf(values ...string) map[string]bool {
	out := map[string]bool{}
	for _, value := range values {
		out[value] = true
	}
	return out
}

func sameSet(left, right map[string]bool) bool {
	if len(left) != len(right) {
		return false
	}
	for key := range left {
		if !right[key] {
			return false
		}
	}
	return true
}

func differenceSet(left, right map[string]bool) []string {
	var out []string
	for key := range left {
		if !right[key] {
			out = append(out, key)
		}
	}
	sort.Strings(out)
	return out
}

func sortedBoolKeys(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
