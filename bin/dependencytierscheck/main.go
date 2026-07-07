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
				govreport.Error("DEPENDENCY_TIERS_INPUT_ERROR", "", fmt.Sprintf("cannot resolve working directory: %v", err)),
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
			govreport.Error("DEPENDENCY_TIERS_INPUT_ERROR", contextPath, err.Error()),
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
	publicFacades := map[string]bool{}
	knownPrefixes := map[string]bool{}
	for _, item := range list(data["public_facades"]) {
		entry := mapValue(item)
		pkg := stringValue(entry["package"])
		if pkg == "" {
			continue
		}
		publicFacades[pkg] = true
		knownPrefixes[pkg] = true
		if internal := strings.TrimRight(stringValue(entry["internal"]), "/"); internal != "" {
			knownPrefixes[internal] = true
		}
	}
	tiers := mapValue(data["dependency_tiers"])
	if tiers == nil {
		c.addError("DEPENDENCY_TIERS_SCHEMA_INVALID", "dependency_tiers must be an object")
		return
	}

	core := c.requireTier(tiers, "core_facades", publicFacades)
	heavy := c.requireTier(tiers, "heavy_extension_facades", publicFacades)
	providers := c.requireTier(tiers, "provider_contract_facades", publicFacades)
	if len(intersection(core, heavy)) > 0 || len(intersection(core, providers)) > 0 || len(intersection(heavy, providers)) > 0 {
		c.addError("DEPENDENCY_TIERS_OVERLAP", "dependency_tiers core/provider/heavy facade sets must be mutually exclusive")
	}
	combined := union(core, heavy, providers)
	if missing := differenceSet(publicFacades, combined); len(missing) > 0 {
		c.addError("DEPENDENCY_TIERS_FACADE_COVERAGE", "dependency_tiers must classify public facade(s): "+strings.Join(missing, ", "))
	}

	allowlist := mapValue(tiers["heavy_dependency_allowlist"])
	if allowlist == nil {
		c.addError("DEPENDENCY_TIERS_ALLOWLIST_SCHEMA", "dependency_tiers.heavy_dependency_allowlist must be an object")
		return
	}
	for importPattern, raw := range allowlist {
		if strings.TrimSpace(importPattern) == "" {
			c.addError("DEPENDENCY_TIERS_ALLOWLIST_SCHEMA", "dependency_tiers.heavy_dependency_allowlist import pattern must be non-empty")
			continue
		}
		prefixes, ok := raw.([]any)
		if !ok || len(prefixes) == 0 {
			c.addError("DEPENDENCY_TIERS_ALLOWLIST_SCHEMA", fmt.Sprintf("dependency_tiers.heavy_dependency_allowlist.%s must be a non-empty list", importPattern))
			continue
		}
		for index, item := range prefixes {
			prefix, ok := item.(string)
			prefix = strings.TrimRight(strings.TrimSpace(prefix), "/")
			if !ok || prefix == "" {
				c.addError("DEPENDENCY_TIERS_ALLOWLIST_SCHEMA", fmt.Sprintf("dependency_tiers.heavy_dependency_allowlist.%s[%d] must be a non-empty string", importPattern, index))
				continue
			}
			if !c.prefixExists(prefix, knownPrefixes) {
				c.addError("DEPENDENCY_TIERS_ALLOWLIST_PREFIX_UNKNOWN", fmt.Sprintf("dependency_tiers.heavy_dependency_allowlist.%s references unknown package prefix %s", importPattern, prefix))
			}
		}
	}
}

func (c *checker) requireTier(tiers map[string]any, name string, publicFacades map[string]bool) map[string]bool {
	values := stringList(tiers[name])
	if len(values) == 0 {
		c.addError("DEPENDENCY_TIERS_SCHEMA_INVALID", "dependency_tiers."+name+" must be a non-empty list")
	}
	out := map[string]bool{}
	for _, value := range values {
		out[value] = true
		if !publicFacades[value] {
			c.addError("DEPENDENCY_TIERS_UNKNOWN_FACADE", fmt.Sprintf("dependency_tiers.%s includes non-public facade %s", name, value))
		}
	}
	return out
}

func (c *checker) prefixExists(prefix string, knownPrefixes map[string]bool) bool {
	if knownPrefixes[prefix] {
		return true
	}
	if stat, err := os.Stat(filepath.Join(c.root, filepath.FromSlash(prefix))); err == nil && stat.IsDir() {
		return true
	}
	for known := range knownPrefixes {
		if strings.HasPrefix(known, prefix+"/") || strings.HasPrefix(prefix, known+"/") {
			return true
		}
	}
	return false
}

func (c *checker) addError(ruleID, message string) {
	c.findings = append(c.findings, govreport.Error(ruleID, "ai-context.json", message))
}

func writeReport(jsonOutput bool, report govreport.Envelope) {
	if jsonOutput {
		if err := govreport.WriteJSON(os.Stdout, report); err != nil {
			fmt.Fprintf(os.Stderr, "dependency tiers check error: [DEPENDENCY_TIERS_INPUT_ERROR] cannot encode JSON output: %v\n", err)
		}
		return
	}
	if report.Status == govreport.StatusFailed {
		for _, finding := range report.Findings {
			fmt.Fprintf(os.Stderr, "dependency tiers check error: [%s] %s\n", finding.RuleID, finding.Message)
		}
		return
	}
	fmt.Println("dependency tiers metadata is valid")
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

func union(sets ...map[string]bool) map[string]bool {
	out := map[string]bool{}
	for _, set := range sets {
		for key := range set {
			out[key] = true
		}
	}
	return out
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
