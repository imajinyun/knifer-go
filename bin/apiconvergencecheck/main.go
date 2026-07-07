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

type checker struct {
	findings []govreport.Finding
}

func main() {
	rootFlag := flag.String("root", "", "repository root")
	contextFlag := flag.String("ai-context", "", "ai-context.json path")
	toolsFlag := flag.String("tools", "", "docs/api/tools.json path")
	jsonFlag := flag.Bool("json", false, "emit machine-readable JSON output")
	flag.Parse()

	root := strings.TrimSpace(*rootFlag)
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			writeReport(*jsonFlag, govreport.Failed([]govreport.Finding{
				govreport.Error("API_CONVERGENCE_INPUT_ERROR", "", fmt.Sprintf("cannot resolve working directory: %v", err)),
			}))
			os.Exit(1)
		}
	}
	contextPath := strings.TrimSpace(*contextFlag)
	if contextPath == "" {
		contextPath = filepath.Join(root, "ai-context.json")
	}
	toolsPath := strings.TrimSpace(*toolsFlag)
	if toolsPath == "" {
		toolsPath = filepath.Join(root, "docs", "api", "tools.json")
	}
	context, err := loadJSON(contextPath)
	if err != nil {
		writeReport(*jsonFlag, govreport.Failed([]govreport.Finding{
			govreport.Error("API_CONVERGENCE_INPUT_ERROR", contextPath, err.Error()),
		}))
		os.Exit(1)
	}
	tools, err := loadJSON(toolsPath)
	if err != nil {
		writeReport(*jsonFlag, govreport.Failed([]govreport.Finding{
			govreport.Error("API_CONVERGENCE_INPUT_ERROR", toolsPath, err.Error()),
		}))
		os.Exit(1)
	}

	c := &checker{}
	c.run(context, tools)
	if len(c.findings) > 0 {
		writeReport(*jsonFlag, govreport.Failed(c.findings))
		os.Exit(1)
	}
	writeReport(*jsonFlag, govreport.Passed())
}

func (c *checker) run(context, tools map[string]any) {
	apiConvergence := mapValue(context["api_convergence"])
	if apiConvergence == nil {
		c.addError("API_CONVERGENCE_SCHEMA_INVALID", "api_convergence must be an object")
		return
	}
	maxGolden := intValue(apiConvergence["max_golden_path_per_facade"])
	if maxGolden != 5 {
		c.addError("API_CONVERGENCE_MAX_GOLDEN_INVALID", "api_convergence.max_golden_path_per_facade must be 5")
		maxGolden = 5
	}
	required := setOf(stringList(apiConvergence["required_classifications"])...)
	for _, name := range []string{"primary", "advanced", "compatibility", "avoid"} {
		if !required[name] {
			c.addError("API_CONVERGENCE_CLASSIFICATION_MISSING", "api_convergence.required_classifications must include "+name)
		}
	}
	publicFacades := map[string]bool{}
	for _, item := range list(context["public_facades"]) {
		entry := mapValue(item)
		if pkg := stringValue(entry["package"]); pkg != "" {
			publicFacades[pkg] = true
		}
	}
	facades := mapValue(apiConvergence["facades"])
	if facades == nil {
		c.addError("API_CONVERGENCE_SCHEMA_INVALID", "api_convergence.facades must be an object")
		return
	}
	if missing := differenceSet(publicFacades, boolKeys(facades)); len(missing) > 0 {
		c.addError("API_CONVERGENCE_FACADE_COVERAGE", "api_convergence.facades missing public facade(s): "+strings.Join(missing, ", "))
	}
	if extra := differenceSet(boolKeys(facades), publicFacades); len(extra) > 0 {
		c.addError("API_CONVERGENCE_FACADE_COVERAGE", "api_convergence.facades includes non-public facade(s): "+strings.Join(extra, ", "))
	}

	toolPackages := toolsPackages(tools)
	for _, packageName := range sortedBoolKeys(publicFacades) {
		entry := mapValue(facades[packageName])
		if entry == nil {
			continue
		}
		toolPackage := mapValue(toolPackages[packageName])
		if toolPackage == nil {
			c.addError("API_CONVERGENCE_TOOLS_PACKAGE_MISSING", "docs/api/tools.json missing package "+packageName)
			continue
		}
		functionNames := map[string]bool{}
		compatibilityFunctions := map[string]bool{}
		for _, item := range list(toolPackage["functions"]) {
			fn := mapValue(item)
			name := stringValue(fn["name"])
			if name == "" {
				continue
			}
			functionNames[name] = true
			if stringValue(fn["status"]) == "compatibility" {
				compatibilityFunctions[name] = true
			}
		}
		golden := goldenPathNames(toolPackage)
		if len(golden) == 0 {
			c.addError("API_CONVERGENCE_GOLDEN_PATH_MISSING", packageName+" must expose at least one golden_path entry")
		}
		if len(golden) > maxGolden {
			c.addError("API_CONVERGENCE_GOLDEN_PATH_TOO_LARGE", fmt.Sprintf("%s golden_path has %d entries; max is %d", packageName, len(golden), maxGolden))
		}
		primary := stringList(entry["primary"])
		if len(primary) == 0 || len(primary) > maxGolden {
			c.addError("API_CONVERGENCE_PRIMARY_INVALID", fmt.Sprintf("api_convergence.facades.%s.primary must contain 1-%d APIs", packageName, maxGolden))
		}
		if !equalStrings(primary, golden) {
			c.addError("API_CONVERGENCE_PRIMARY_DRIFT", fmt.Sprintf("api_convergence.facades.%s.primary must match docs/api/tools.json golden_path order", packageName))
		}
		buckets := map[string][]string{}
		for _, bucket := range []string{"primary", "advanced", "compatibility", "avoid"} {
			values := stringList(entry[bucket])
			buckets[bucket] = values
			if hasDuplicates(values) {
				c.addError("API_CONVERGENCE_BUCKET_DUPLICATE", fmt.Sprintf("api_convergence.facades.%s.%s must not contain duplicates", packageName, bucket))
			}
			for _, fnName := range values {
				if !functionNames[fnName] {
					c.addError("API_CONVERGENCE_UNKNOWN_API", fmt.Sprintf("api_convergence.facades.%s.%s references unknown API %s", packageName, bucket, fnName))
				}
			}
		}
		for _, pair := range [][2]string{{"primary", "advanced"}, {"primary", "avoid"}, {"advanced", "compatibility"}, {"advanced", "avoid"}, {"compatibility", "avoid"}} {
			if overlap := overlapStrings(buckets[pair[0]], buckets[pair[1]]); len(overlap) > 0 {
				c.addError("API_CONVERGENCE_BUCKET_OVERLAP", fmt.Sprintf("api_convergence.facades.%s.%s and %s overlap: %s", packageName, pair[0], pair[1], strings.Join(overlap, ", ")))
			}
		}
		for _, fnName := range buckets["compatibility"] {
			if !compatibilityFunctions[fnName] {
				c.addError("API_CONVERGENCE_COMPATIBILITY_STATUS", fmt.Sprintf("api_convergence.facades.%s.compatibility includes non-compatibility API %s", packageName, fnName))
			}
		}
		if strings.TrimSpace(stringValue(entry["decision"])) == "" {
			c.addError("API_CONVERGENCE_DECISION_MISSING", fmt.Sprintf("api_convergence.facades.%s.decision must be non-empty", packageName))
		}
	}
}

func toolsPackages(tools map[string]any) map[string]any {
	out := map[string]any{}
	switch packages := tools["packages"].(type) {
	case []any:
		for _, item := range packages {
			pkg := mapValue(item)
			if name := stringValue(pkg["name"]); name != "" {
				out[name] = pkg
			}
		}
	case map[string]any:
		return packages
	}
	return out
}

func goldenPathNames(toolPackage map[string]any) []string {
	var out []string
	for _, item := range list(toolPackage["golden_path"]) {
		entry := mapValue(item)
		if name := stringValue(entry["name"]); name != "" {
			out = append(out, name)
		}
	}
	return out
}

func (c *checker) addError(ruleID, message string) {
	c.findings = append(c.findings, govreport.Error(ruleID, "ai-context.json", message))
}

func writeReport(jsonOutput bool, report govreport.Envelope) {
	if jsonOutput {
		if err := govreport.WriteJSON(os.Stdout, report); err != nil {
			fmt.Fprintf(os.Stderr, "api convergence check error: [API_CONVERGENCE_INPUT_ERROR] cannot encode JSON output: %v\n", err)
		}
		return
	}
	if report.Status == govreport.StatusFailed {
		for _, finding := range report.Findings {
			fmt.Fprintf(os.Stderr, "api convergence check error: [%s] %s\n", finding.RuleID, finding.Message)
		}
		return
	}
	fmt.Println("api convergence metadata is valid")
}

func loadJSON(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("missing JSON file: %s", path)
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

func intValue(value any) int {
	switch v := value.(type) {
	case float64:
		return int(v)
	case int:
		return v
	default:
		return 0
	}
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

func hasDuplicates(values []string) bool {
	seen := map[string]bool{}
	for _, value := range values {
		if seen[value] {
			return true
		}
		seen[value] = true
	}
	return false
}

func overlapStrings(left, right []string) []string {
	rightSet := setOf(right...)
	var out []string
	for _, value := range left {
		if rightSet[value] {
			out = append(out, value)
		}
	}
	sort.Strings(out)
	return out
}
