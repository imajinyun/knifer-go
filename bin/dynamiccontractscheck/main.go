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

var expectedDomains = setOf(
	"vbean_decode_copy_merge",
	"vjson_dynamic",
	"vobj_dynamic",
	"vconf_dynamic",
	"vconv_conversion_matrix",
	"vref_reflection_boundaries",
)

var expectedPackages = setOf(
	"internal/bean",
	"internal/conv",
	"internal/ref",
	"vjson",
	"vobj",
	"vconf",
	"vconv",
	"vref",
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
				govreport.Error("DYNAMIC_CONTRACTS_INPUT_ERROR", "", fmt.Sprintf("cannot resolve working directory: %v", err)),
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
			govreport.Error("DYNAMIC_CONTRACTS_INPUT_ERROR", contextPath, err.Error()),
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
	contracts := mapValue(data["dynamic_semantic_contracts"])
	if contracts == nil {
		c.addError("DYNAMIC_CONTRACTS_SCHEMA_INVALID", "dynamic_semantic_contracts must be an object")
		return
	}
	requiredDomains := setOf(stringList(contracts["required_domains"])...)
	if !sameSet(requiredDomains, expectedDomains) {
		c.addError("DYNAMIC_CONTRACTS_DOMAIN_COVERAGE", "dynamic_semantic_contracts.required_domains must cover exactly: "+strings.Join(sortedBoolKeys(expectedDomains), ", "))
	}
	domains := mapValue(contracts["domains"])
	if domains == nil {
		c.addError("DYNAMIC_CONTRACTS_SCHEMA_INVALID", "dynamic_semantic_contracts.domains must be an object")
		return
	}
	if missing := differenceSet(requiredDomains, boolKeys(domains)); len(missing) > 0 {
		c.addError("DYNAMIC_CONTRACTS_DOMAIN_MISSING", "dynamic_semantic_contracts.domains missing required domain(s): "+strings.Join(missing, ", "))
	}
	if extra := differenceSet(boolKeys(domains), requiredDomains); len(extra) > 0 {
		c.addError("DYNAMIC_CONTRACTS_DOMAIN_UNKNOWN", "dynamic_semantic_contracts.domains includes unknown domain(s): "+strings.Join(extra, ", "))
	}

	coveredPackages := map[string]bool{}
	for _, domainName := range sortedAnyKeys(domains) {
		domain := mapValue(domains[domainName])
		if domain == nil {
			c.addError("DYNAMIC_CONTRACTS_SCHEMA_INVALID", fmt.Sprintf("dynamic_semantic_contracts.domains.%s must be an object", domainName))
			continue
		}
		packages := stringList(domain["packages"])
		for _, pkg := range packages {
			coveredPackages[pkg] = true
			if stat, err := os.Stat(filepath.Join(c.root, filepath.FromSlash(pkg))); err != nil || !stat.IsDir() {
				c.addError("DYNAMIC_CONTRACTS_PACKAGE_MISSING", fmt.Sprintf("dynamic_semantic_contracts.domains.%s.packages references missing directory %s", domainName, pkg))
			}
		}
		if len(stringList(domain["guarantees"])) < 3 {
			c.addError("DYNAMIC_CONTRACTS_GUARANTEES_INCOMPLETE", fmt.Sprintf("dynamic_semantic_contracts.domains.%s.guarantees must contain at least 3 semantic guarantees", domainName))
		}
		contractTests := stringList(domain["contract_tests"])
		if len(contractTests) == 0 {
			c.addError("DYNAMIC_CONTRACTS_CONTRACT_TESTS_EMPTY", fmt.Sprintf("dynamic_semantic_contracts.domains.%s.contract_tests must be non-empty", domainName))
		}
		c.validateReferences(domainName, "contract_tests", contractTests)
		c.validateReferences(domainName, "fuzz_tests", stringList(domain["fuzz_tests"]))
	}
	if !sameSet(coveredPackages, expectedPackages) {
		c.addError("DYNAMIC_CONTRACTS_PACKAGE_COVERAGE", "dynamic_semantic_contracts must cover exactly package directories: "+strings.Join(sortedBoolKeys(expectedPackages), ", "))
	}
}

func (c *checker) validateReferences(domainName, field string, references []string) {
	for _, reference := range references {
		if !referencesFunction(reference) {
			c.addError("DYNAMIC_CONTRACTS_REFERENCE_INVALID", fmt.Sprintf("dynamic_semantic_contracts.domains.%s.%s must reference explicit test functions, got %s", domainName, field, reference))
			continue
		}
		if !c.referenceExists(reference) {
			c.addError("DYNAMIC_CONTRACTS_REFERENCE_MISSING", fmt.Sprintf("dynamic_semantic_contracts.domains.%s.%s references missing file or function %s", domainName, field, reference))
		}
	}
}

func (c *checker) addError(ruleID, message string) {
	c.findings = append(c.findings, govreport.Error(ruleID, "ai-context.json", message))
}

func writeReport(jsonOutput bool, report govreport.Envelope) {
	if jsonOutput {
		if err := govreport.WriteJSON(os.Stdout, report); err != nil {
			fmt.Fprintf(os.Stderr, "dynamic contracts check error: [DYNAMIC_CONTRACTS_INPUT_ERROR] cannot encode JSON output: %v\n", err)
		}
		return
	}
	if report.Status == govreport.StatusFailed {
		for _, finding := range report.Findings {
			fmt.Fprintf(os.Stderr, "dynamic contracts check error: [%s] %s\n", finding.RuleID, finding.Message)
		}
		return
	}
	fmt.Println("dynamic semantic contracts are valid")
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

var referencePattern = regexp.MustCompile(`^(.+\.go):([A-Za-z_]\w*)$`)

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
	return regexp.MustCompile(`(?m)^func\s+` + regexp.QuoteMeta(match[2]) + `\s*\(`).Match(data)
}

func mapValue(value any) map[string]any {
	mapping, _ := value.(map[string]any)
	return mapping
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

func boolKeys(values map[string]any) map[string]bool {
	out := map[string]bool{}
	for key := range values {
		out[key] = true
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
