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

	"github.com/imajinyun/knifer-go/bin/internal/govreport"
)

var requiredGrades = setOf("core", "stable", "maintenance", "adapter", "heavy", "candidate-for-split", "candidate-for-deprecation")
var coreCompatibleGrades = setOf("core", "stable", "maintenance", "candidate-for-split", "candidate-for-deprecation")

type checker struct {
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
				govreport.Error("LIFECYCLE_INPUT_ERROR", "", fmt.Sprintf("cannot resolve working directory: %v", err)),
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
			govreport.Error("LIFECYCLE_INPUT_ERROR", contextPath, err.Error()),
		}))
		os.Exit(1)
	}

	c := &checker{}
	c.run(data)
	if len(c.findings) > 0 {
		writeReport(*jsonFlag, govreport.Failed(c.findings))
		os.Exit(1)
	}
	writeReport(*jsonFlag, govreport.Passed())
}

func (c *checker) run(data map[string]any) {
	publicFacades := map[string]bool{}
	for _, item := range list(data["public_facades"]) {
		entry := mapValue(item)
		if pkg := stringValue(entry["package"]); pkg != "" {
			publicFacades[pkg] = true
		}
	}
	lifecycle := mapValue(data["package_lifecycle"])
	if lifecycle == nil {
		c.addError("LIFECYCLE_SCHEMA_INVALID", "package_lifecycle must be an object")
		return
	}
	allowed := setOf(stringList(lifecycle["allowed_grades"])...)
	for _, grade := range sortedBoolKeys(requiredGrades) {
		if !allowed[grade] {
			c.addError("LIFECYCLE_ALLOWED_GRADES_INCOMPLETE", "package_lifecycle.allowed_grades must include "+grade)
		}
	}
	packages := mapValue(lifecycle["packages"])
	if packages == nil {
		c.addError("LIFECYCLE_SCHEMA_INVALID", "package_lifecycle.packages must be an object")
		return
	}
	if missing := differenceSet(publicFacades, boolKeys(packages)); len(missing) > 0 {
		c.addError("LIFECYCLE_PACKAGE_COVERAGE", "package_lifecycle.packages missing public facade(s): "+strings.Join(missing, ", "))
	}
	if extra := differenceSet(boolKeys(packages), publicFacades); len(extra) > 0 {
		c.addError("LIFECYCLE_PACKAGE_COVERAGE", "package_lifecycle.packages includes non-public facade(s): "+strings.Join(extra, ", "))
	}

	tiers := mapValue(data["dependency_tiers"])
	heavy := setOf(stringList(tiers["heavy_extension_facades"])...)
	adapters := setOf(stringList(tiers["provider_contract_facades"])...)
	core := setOf(stringList(tiers["core_facades"])...)
	for _, tier := range []struct {
		name   string
		values map[string]bool
	}{
		{"heavy_extension_facades", heavy},
		{"provider_contract_facades", adapters},
		{"core_facades", core},
	} {
		if unknown := differenceSet(tier.values, publicFacades); len(unknown) > 0 {
			c.addError("LIFECYCLE_DEPENDENCY_TIER_UNKNOWN_PACKAGE", fmt.Sprintf("dependency_tiers.%s includes non-public facade(s): %s", tier.name, strings.Join(unknown, ", ")))
		}
	}
	if len(intersection(heavy, adapters)) > 0 || len(intersection(heavy, core)) > 0 || len(intersection(adapters, core)) > 0 {
		c.addError("LIFECYCLE_DEPENDENCY_TIER_OVERLAP", "dependency_tiers facade sets must be mutually exclusive")
	}
	for _, packageName := range sortedAnyKeys(packages) {
		entry := mapValue(packages[packageName])
		if entry == nil {
			c.addError("LIFECYCLE_SCHEMA_INVALID", fmt.Sprintf("package_lifecycle.packages.%s must be an object", packageName))
			continue
		}
		grade := stringValue(entry["grade"])
		if !allowed[grade] {
			c.addError("LIFECYCLE_GRADE_INVALID", fmt.Sprintf("package_lifecycle.packages.%s.grade must be an allowed lifecycle grade", packageName))
		}
		if strings.TrimSpace(stringValue(entry["rationale"])) == "" {
			c.addError("LIFECYCLE_RATIONALE_MISSING", fmt.Sprintf("package_lifecycle.packages.%s.rationale must be non-empty", packageName))
		}
		if heavy[packageName] && grade != "heavy" {
			c.addError("LIFECYCLE_HEAVY_GRADE_MISMATCH", fmt.Sprintf("package_lifecycle.packages.%s.grade must be heavy", packageName))
		}
		if adapters[packageName] && grade != "adapter" {
			c.addError("LIFECYCLE_ADAPTER_GRADE_MISMATCH", fmt.Sprintf("package_lifecycle.packages.%s.grade must be adapter", packageName))
		}
		if core[packageName] && !coreCompatibleGrades[grade] {
			c.addError("LIFECYCLE_CORE_GRADE_MISMATCH", fmt.Sprintf("package_lifecycle.packages.%s.grade must remain core-compatible", packageName))
		}
	}
}

func (c *checker) addError(ruleID, message string) {
	c.findings = append(c.findings, govreport.Error(ruleID, "ai-context.json", message))
}

func writeReport(jsonOutput bool, report govreport.Envelope) {
	if jsonOutput {
		if err := govreport.WriteJSON(os.Stdout, report); err != nil {
			fmt.Fprintf(os.Stderr, "lifecycle check error: [LIFECYCLE_INPUT_ERROR] cannot encode JSON output: %v\n", err)
		}
		return
	}
	if report.Status == govreport.StatusFailed {
		for _, finding := range report.Findings {
			fmt.Fprintf(os.Stderr, "lifecycle check error: [%s] %s\n", finding.RuleID, finding.Message)
		}
		return
	}
	fmt.Println("package lifecycle metadata is valid")
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
	return mapping
}

func list(value any) []any {
	values, _ := value.([]any)
	return values
}

func stringValue(value any) string {
	text, _ := value.(string)
	return strings.TrimSpace(text)
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

func boolKeys(values map[string]any) map[string]bool {
	out := map[string]bool{}
	for key := range values {
		out[key] = true
	}
	return out
}

func sortedAnyKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedBoolKeys(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
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

func intersection(left, right map[string]bool) []string {
	var out []string
	for key := range left {
		if right[key] {
			out = append(out, key)
		}
	}
	sort.Strings(out)
	return out
}
