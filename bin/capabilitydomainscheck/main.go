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

var expectedDomains = setOf(
	"data_transform",
	"collections",
	"text_parsing",
	"trust_boundary",
	"security_primitives",
	"runtime_adapters",
	"domain_helpers",
)

var allowedTestTypes = setOf("benchmark", "contract", "error_contract", "example", "fuzz", "misuse", "provider_contract", "security")

var requiredTestMatrix = map[string]map[string]bool{
	"data_transform":      setOf("contract", "fuzz", "error_contract"),
	"collections":         setOf("contract", "benchmark"),
	"text_parsing":        setOf("contract", "fuzz", "provider_contract"),
	"trust_boundary":      setOf("contract", "security", "misuse", "fuzz", "error_contract"),
	"security_primitives": setOf("contract", "security", "misuse", "error_contract"),
	"runtime_adapters":    setOf("contract", "provider_contract"),
	"domain_helpers":      setOf("contract", "example"),
}

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
				govreport.Error("CAPABILITY_DOMAINS_INPUT_ERROR", "", fmt.Sprintf("cannot resolve working directory: %v", err)),
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
			govreport.Error("CAPABILITY_DOMAINS_INPUT_ERROR", contextPath, err.Error()),
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
	publicFacades := publicFacadeSet(data)
	domains := mapValue(data["capability_domains"])
	if domains == nil {
		c.addError("CAPABILITY_DOMAINS_SCHEMA_INVALID", "capability_domains must be an object")
		return
	}
	if missing := differenceSet(expectedDomains, anyKeys(domains)); len(missing) > 0 {
		c.addError("CAPABILITY_DOMAINS_REQUIRED_DOMAIN_MISSING", "capability_domains missing required domain(s): "+strings.Join(missing, ", "))
	}
	if extra := differenceSet(anyKeys(domains), expectedDomains); len(extra) > 0 {
		c.addError("CAPABILITY_DOMAINS_UNKNOWN_DOMAIN", "capability_domains includes unknown domain(s): "+strings.Join(extra, ", "))
	}

	coveredPackages := map[string]bool{}
	for _, domainName := range sortedAnyKeys(domains) {
		domain := mapValue(domains[domainName])
		if domain == nil {
			c.addError("CAPABILITY_DOMAINS_SCHEMA_INVALID", fmt.Sprintf("capability_domains.%s must be an object", domainName))
			continue
		}
		c.validateDomain(domainName, domain, publicFacades, coveredPackages)
	}
	if missing := differenceSet(publicFacades, coveredPackages); len(missing) > 0 {
		c.addError("CAPABILITY_DOMAINS_FACADE_COVERAGE", "capability_domains do not cover public facade(s): "+strings.Join(missing, ", "))
	}

	c.validateSecuritySensitive(data, domains, publicFacades)
}

func (c *checker) validateDomain(domainName string, domain map[string]any, publicFacades, coveredPackages map[string]bool) {
	if strings.TrimSpace(stringValue(domain["purpose"])) == "" {
		c.addError("CAPABILITY_DOMAINS_PURPOSE_MISSING", fmt.Sprintf("capability_domains.%s.purpose must be non-empty", domainName))
	}
	packages := stringList(domain["packages"])
	if len(packages) < 2 {
		c.addError("CAPABILITY_DOMAINS_PACKAGES_INCOMPLETE", fmt.Sprintf("capability_domains.%s.packages must include at least 2 facades", domainName))
	}
	packageSet := setOf(packages...)
	if unknown := differenceSet(packageSet, publicFacades); len(unknown) > 0 {
		c.addError("CAPABILITY_DOMAINS_UNKNOWN_FACADE", fmt.Sprintf("capability_domains.%s.packages includes non-public facade(s): %s", domainName, strings.Join(unknown, ", ")))
	}
	for _, pkg := range packages {
		coveredPackages[pkg] = true
	}

	if len(stringList(domain["required_focus"])) < 2 {
		c.addError("CAPABILITY_DOMAINS_REQUIRED_FOCUS_INCOMPLETE", fmt.Sprintf("capability_domains.%s.required_focus must include at least 2 focus areas", domainName))
	}
	requiredTests := setOf(stringList(domain["required_tests"])...)
	if len(requiredTests) == 0 {
		c.addError("CAPABILITY_DOMAINS_REQUIRED_TESTS_MISSING", fmt.Sprintf("capability_domains.%s.required_tests must be non-empty", domainName))
	}
	if unknown := differenceSet(requiredTests, allowedTestTypes); len(unknown) > 0 {
		c.addError("CAPABILITY_DOMAINS_REQUIRED_TEST_UNKNOWN", fmt.Sprintf("capability_domains.%s.required_tests includes unknown test type(s): %s", domainName, strings.Join(unknown, ", ")))
	}
	if required, ok := requiredTestMatrix[domainName]; ok {
		if missing := differenceSet(required, requiredTests); len(missing) > 0 {
			c.addError("CAPABILITY_DOMAINS_REQUIRED_TEST_MISSING", fmt.Sprintf("capability_domains.%s.required_tests missing required test type(s): %s", domainName, strings.Join(missing, ", ")))
		}
	}
}

func (c *checker) validateSecuritySensitive(data map[string]any, domains map[string]any, publicFacades map[string]bool) {
	securitySensitive := setOf(stringList(data["security_sensitive_packages"])...)
	if unknown := differenceSet(securitySensitive, publicFacades); len(unknown) > 0 {
		c.addError("CAPABILITY_DOMAINS_SECURITY_SENSITIVE_UNKNOWN", "security_sensitive_packages includes non-public facade(s): "+strings.Join(unknown, ", "))
	}
	coveredSensitive := map[string]bool{}
	for _, domainName := range []string{"trust_boundary", "security_primitives", "runtime_adapters"} {
		domain := mapValue(domains[domainName])
		for _, pkg := range stringList(domain["packages"]) {
			coveredSensitive[pkg] = true
		}
	}
	if missing := differenceSet(securitySensitive, coveredSensitive); len(missing) > 0 {
		c.addError("CAPABILITY_DOMAINS_SECURITY_SENSITIVE_COVERAGE", "security-sensitive facades must be represented by trust_boundary, security_primitives, or runtime_adapters: "+strings.Join(missing, ", "))
	}
}

func publicFacadeSet(data map[string]any) map[string]bool {
	out := map[string]bool{}
	for _, item := range list(data["public_facades"]) {
		entry := mapValue(item)
		if pkg := stringValue(entry["package"]); pkg != "" {
			out[pkg] = true
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
			fmt.Fprintf(os.Stderr, "capability domains check error: [CAPABILITY_DOMAINS_INPUT_ERROR] cannot encode JSON output: %v\n", err)
		}
		return
	}
	if report.Status == govreport.StatusFailed {
		for _, finding := range report.Findings {
			fmt.Fprintf(os.Stderr, "capability domains check error: [%s] %s\n", finding.RuleID, finding.Message)
		}
		return
	}
	fmt.Println("capability domains metadata is valid")
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

func anyKeys(values map[string]any) map[string]bool {
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
