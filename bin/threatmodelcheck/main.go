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

var expectedBoundaryNames = setOf(
	"default_timeout",
	"redirect_revalidation",
	"private_host_rejection",
	"bounded_response_reads",
	"safe_download_paths",
	"remote_config_boundary",
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
				govreport.Error("THREAT_MODEL_INPUT_ERROR", "", fmt.Sprintf("cannot resolve working directory: %v", err)),
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
			govreport.Error("THREAT_MODEL_INPUT_ERROR", contextPath, err.Error()),
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
	for _, item := range list(data["public_facades"]) {
		entry := mapValue(item)
		if pkg := stringValue(entry["package"]); pkg != "" {
			publicFacades[pkg] = true
		}
	}

	threatModel := mapValue(data["threat_model"])
	if threatModel == nil {
		c.addError("THREAT_MODEL_SCHEMA_INVALID", "threat_model must be an object")
		return
	}
	c.validateMethodology(threatModel)
	c.validateTopLevelMisuseTests(threatModel)
	c.validateDomains(threatModel, data, publicFacades)
	c.validateBoundaryContracts(threatModel, publicFacades)
}

func (c *checker) validateMethodology(threatModel map[string]any) {
	methodology := stringValue(threatModel["methodology"])
	if !strings.Contains(methodology, "STRIDE") || !strings.Contains(methodology, "DREAD") {
		c.addError("THREAT_MODEL_METHODOLOGY_INCOMPLETE", "threat_model.methodology must mention STRIDE and DREAD")
	}
}

func (c *checker) validateTopLevelMisuseTests(threatModel map[string]any) {
	for _, reference := range stringList(threatModel["misuse_tests"]) {
		if !referencesFunction(reference) {
			c.addError("THREAT_MODEL_MISUSE_TEST_INVALID", "threat_model.misuse_tests must reference explicit test functions, got "+reference)
			continue
		}
		if !c.referenceExists(reference) {
			c.addError("THREAT_MODEL_MISUSE_TEST_MISSING", "threat_model.misuse_tests references missing file or function "+reference)
		}
	}
}

func (c *checker) validateDomains(threatModel map[string]any, data map[string]any, publicFacades map[string]bool) {
	domains := mapValue(threatModel["domains"])
	if domains == nil {
		c.addError("THREAT_MODEL_DOMAINS_SCHEMA_INVALID", "threat_model.domains must be an object")
		return
	}
	coveredPackages := map[string]bool{}
	for _, domainName := range sortedAnyKeys(domains) {
		domain := mapValue(domains[domainName])
		if domain == nil {
			c.addError("THREAT_MODEL_DOMAIN_SCHEMA_INVALID", fmt.Sprintf("threat_model.domains.%s must be an object", domainName))
			continue
		}
		c.validateDomain(domainName, domain, coveredPackages)
	}

	securitySensitive := setOf(stringList(data["security_sensitive_packages"])...)
	if unknown := differenceSet(securitySensitive, publicFacades); len(unknown) > 0 {
		c.addError("THREAT_MODEL_SECURITY_SENSITIVE_UNKNOWN", "security_sensitive_packages includes non-public facade(s): "+strings.Join(unknown, ", "))
	}
	if missing := differenceSet(securitySensitive, coveredPackages); len(missing) > 0 {
		c.addError("THREAT_MODEL_SECURITY_SENSITIVE_COVERAGE", "threat_model.domains do not cover security-sensitive package(s): "+strings.Join(missing, ", "))
	}
	if unexpected := differenceSet(coveredPackages, securitySensitive); len(unexpected) > 0 {
		c.addError("THREAT_MODEL_DOMAIN_UNEXPECTED_PACKAGE", "threat_model.domains cover non-security-sensitive package(s): "+strings.Join(unexpected, ", "))
	}
}

func (c *checker) validateDomain(domainName string, domain map[string]any, coveredPackages map[string]bool) {
	packages := stringList(domain["packages"])
	if len(packages) == 0 {
		c.addError("THREAT_MODEL_DOMAIN_PACKAGES_MISSING", fmt.Sprintf("threat_model.domains.%s.packages must be non-empty", domainName))
	}
	for _, pkg := range packages {
		coveredPackages[pkg] = true
	}

	threats := stringList(domain["threats"])
	if len(threats) == 0 {
		c.addError("THREAT_MODEL_DOMAIN_THREATS_MISSING", fmt.Sprintf("threat_model.domains.%s.threats must be non-empty", domainName))
	}
	threatSet := setOf(threats...)
	contractTests := mapValue(domain["contract_tests"])
	if contractTests == nil {
		c.addError("THREAT_MODEL_DOMAIN_CONTRACT_TESTS_SCHEMA_INVALID", fmt.Sprintf("threat_model.domains.%s.contract_tests must be an object", domainName))
		contractTests = map[string]any{}
	}
	if missing := differenceSet(threatSet, anyKeys(contractTests)); len(missing) > 0 {
		c.addError("THREAT_MODEL_DOMAIN_CONTRACT_TEST_MISSING_THREAT", fmt.Sprintf("threat_model.domains.%s.contract_tests missing threat(s): %s", domainName, strings.Join(missing, ", ")))
	}
	for _, threat := range sortedAnyKeys(contractTests) {
		if !threatSet[threat] {
			c.addError("THREAT_MODEL_DOMAIN_CONTRACT_TEST_UNKNOWN_THREAT", fmt.Sprintf("threat_model.domains.%s.contract_tests declares unknown threat %s", domainName, threat))
		}
		c.validateThreatReferences(domainName, threat, stringList(contractTests[threat]))
	}
	for _, reference := range stringList(domain["misuse_tests"]) {
		if !c.referencePathExists(reference) {
			c.addError("THREAT_MODEL_DOMAIN_MISUSE_TEST_MISSING", fmt.Sprintf("threat_model.domains.%s.misuse_tests references missing file %s", domainName, reference))
		}
	}
}

func (c *checker) validateThreatReferences(domainName, threat string, references []string) {
	minimumReferences := 2
	if domainName == "network_clients" || domainName == "crypto_identity_randomness" || domainName == "data_and_cli_boundaries" {
		minimumReferences = 3
	}
	if len(references) < minimumReferences {
		c.addError("THREAT_MODEL_DOMAIN_CONTRACT_TEST_INCOMPLETE", fmt.Sprintf("threat_model.domains.%s.contract_tests.%s must reference at least %d tests", domainName, threat, minimumReferences))
	}
	hasFunctionReference := false
	for _, reference := range references {
		if referencesFunction(reference) {
			hasFunctionReference = true
		}
		if !c.referenceExists(reference) {
			c.addError("THREAT_MODEL_DOMAIN_CONTRACT_TEST_MISSING", fmt.Sprintf("threat_model.domains.%s.contract_tests.%s references missing file or function %s", domainName, threat, reference))
		}
	}
	if !hasFunctionReference {
		c.addError("THREAT_MODEL_DOMAIN_CONTRACT_TEST_INVALID", fmt.Sprintf("threat_model.domains.%s.contract_tests.%s must reference at least one explicit test function", domainName, threat))
	}
}

func (c *checker) validateBoundaryContracts(threatModel map[string]any, publicFacades map[string]bool) {
	boundaryContracts, ok := threatModel["boundary_contracts"].([]any)
	if !ok || len(boundaryContracts) == 0 {
		c.addError("THREAT_MODEL_BOUNDARY_CONTRACTS_MISSING", "threat_model.boundary_contracts must be a non-empty list")
		return
	}

	seenNames := map[string]bool{}
	for index, raw := range boundaryContracts {
		contract := mapValue(raw)
		if contract == nil {
			c.addError("THREAT_MODEL_BOUNDARY_SCHEMA_INVALID", fmt.Sprintf("threat_model.boundary_contracts[%d] must be an object", index))
			continue
		}
		name := strings.TrimSpace(stringValue(contract["name"]))
		if name == "" {
			c.addError("THREAT_MODEL_BOUNDARY_SCHEMA_INVALID", fmt.Sprintf("threat_model.boundary_contracts[%d].name must be non-empty", index))
			continue
		}
		seenNames[name] = true
		if !expectedBoundaryNames[name] {
			c.addError("THREAT_MODEL_BOUNDARY_UNKNOWN_NAME", "threat_model.boundary_contracts includes unknown boundary: "+name)
		}
		for _, pkg := range stringList(contract["packages"]) {
			if !publicFacades[pkg] {
				c.addError("THREAT_MODEL_BOUNDARY_UNKNOWN_PACKAGE", fmt.Sprintf("threat_model.boundary_contracts.%s.packages contains non-public facade: %s", name, pkg))
			}
		}
		if len(stringList(contract["required_controls"])) < 2 {
			c.addError("THREAT_MODEL_BOUNDARY_CONTROL_INCOMPLETE", fmt.Sprintf("threat_model.boundary_contracts.%s.required_controls must contain at least two controls", name))
		}
		references := stringList(contract["contract_tests"])
		if len(references) < 2 {
			c.addError("THREAT_MODEL_BOUNDARY_CONTRACT_TEST_INCOMPLETE", fmt.Sprintf("threat_model.boundary_contracts.%s.contract_tests must reference at least two tests", name))
		}
		for _, reference := range references {
			if !referencesFunction(reference) {
				c.addError("THREAT_MODEL_BOUNDARY_CONTRACT_TEST_INVALID", fmt.Sprintf("threat_model.boundary_contracts.%s.contract_tests must reference explicit test functions, got %s", name, reference))
				continue
			}
			if !c.referenceExists(reference) {
				c.addError("THREAT_MODEL_BOUNDARY_CONTRACT_TEST_MISSING", fmt.Sprintf("threat_model.boundary_contracts.%s.contract_tests references missing file or function %s", name, reference))
			}
		}
	}
	if missing := differenceSet(expectedBoundaryNames, seenNames); len(missing) > 0 {
		c.addError("THREAT_MODEL_BOUNDARY_MISSING_NAME", "threat_model.boundary_contracts missing boundary/boundaries: "+strings.Join(missing, ", "))
	}
	if extra := differenceSet(seenNames, expectedBoundaryNames); len(extra) > 0 {
		c.addError("THREAT_MODEL_BOUNDARY_UNKNOWN_NAME", "threat_model.boundary_contracts includes unknown boundary/boundaries: "+strings.Join(extra, ", "))
	}
}

func (c *checker) addError(ruleID, message string) {
	c.findings = append(c.findings, govreport.Error(ruleID, "ai-context.json", message))
}

func writeReport(jsonOutput bool, report govreport.Envelope) {
	if jsonOutput {
		if err := govreport.WriteJSON(os.Stdout, report); err != nil {
			fmt.Fprintf(os.Stderr, "threat model check error: [THREAT_MODEL_INPUT_ERROR] cannot encode JSON output: %v\n", err)
		}
		return
	}
	if report.Status == govreport.StatusFailed {
		for _, finding := range report.Findings {
			fmt.Fprintf(os.Stderr, "threat model check error: [%s] %s\n", finding.RuleID, finding.Message)
		}
		return
	}
	fmt.Println("threat model governance is valid")
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

func (c *checker) referencePathExists(reference string) bool {
	path := reference
	if match := referencePattern.FindStringSubmatch(reference); len(match) == 3 {
		path = match[1]
	}
	stat, err := os.Stat(filepath.Join(c.root, filepath.FromSlash(path)))
	return err == nil && !stat.IsDir()
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
