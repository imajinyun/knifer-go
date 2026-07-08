package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/imajinyun/knifer-go/bin/internal/govreport"
)

type checker struct {
	root     string
	findings []govreport.Finding
}

var forbiddenPythonTokens = []string{
	"def validate_benchmark_regression",
	"validate_benchmark_regression()",
	"--bench-only",
	"BENCH_ONLY",
	"def validate_random_source_policy",
	"validate_random_source_policy()",
	"def validate_error_model",
	"validate_error_model()",
	"def validate_dynamic_semantic_contracts",
	"validate_dynamic_semantic_contracts()",
	"def validate_lifecycle",
	"validate_lifecycle()",
	"def validate_api_convergence",
	"validate_api_convergence()",
	"def validate_dependency_isolation",
	"validate_dependency_isolation()",
	"def validate_capability_domains",
	"validate_capability_domains()",
	"threat_model.boundary_contracts",
}

var requiredMaturityTargets = []string{
	"random-source-policy-check",
	"threat-model-check",
	"dynamic-contracts-check",
	"error-model-check",
	"api-convergence-check",
	"lifecycle-check",
	"dependency-tiers-check",
	"capability-domains-check",
}

func main() {
	rootFlag := flag.String("root", "", "repository root")
	jsonFlag := flag.Bool("json", false, "emit machine-readable JSON output")
	flag.Parse()

	root := strings.TrimSpace(*rootFlag)
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			writeReport(*jsonFlag, govreport.Failed([]govreport.Finding{
				govreport.Error("GOVERNANCE_MIGRATION_INPUT_ERROR", "", fmt.Sprintf("cannot resolve working directory: %v", err)),
			}))
			os.Exit(1)
		}
	}
	c := &checker{root: root}
	c.run()
	if len(c.findings) > 0 {
		writeReport(*jsonFlag, govreport.Failed(c.findings))
		os.Exit(1)
	}
	writeReport(*jsonFlag, govreport.Passed())
}

func (c *checker) run() {
	maturityPath := filepath.Join(c.root, "bin", "check_governance_maturity.sh")
	maturityBytes, err := os.ReadFile(maturityPath)
	if err != nil {
		c.addError("GOVERNANCE_MIGRATION_INPUT_ERROR", "bin/check_governance_maturity.sh", fmt.Sprintf("cannot read maturity script: %v", err))
		return
	}
	maturityText := string(maturityBytes)
	for _, token := range forbiddenPythonTokens {
		if strings.Contains(maturityText, token) {
			c.addError("GOVERNANCE_MIGRATION_PYTHON_RULE_REGRESSION", "bin/check_governance_maturity.sh", "migrated Python governance rule reintroduced: "+token)
		}
	}

	makefilePath := filepath.Join(c.root, "Makefile")
	makefileBytes, err := os.ReadFile(makefilePath)
	if err != nil {
		c.addError("GOVERNANCE_MIGRATION_INPUT_ERROR", "Makefile", fmt.Sprintf("cannot read Makefile: %v", err))
		return
	}
	makefileText := string(makefileBytes)
	maturityRecipe := makeTargetRecipe(makefileText, "governance-maturity-check")
	for _, target := range requiredMaturityTargets {
		if !strings.Contains(maturityRecipe, "$(MAKE) "+target) && !strings.Contains(maturityRecipe, "make "+target) {
			c.addError("GOVERNANCE_MIGRATION_MAKE_TARGET_MISSING", "Makefile", "governance-maturity-check must run "+target)
		}
	}
	benchRecipe := makeTargetRecipe(makefileText, "bench-regression-check")
	if !strings.Contains(benchRecipe, "benchmarkregressioncheck") {
		c.addError("GOVERNANCE_MIGRATION_MAKE_TARGET_MISSING", "Makefile", "bench-regression-check must run benchmarkregressioncheck")
	}
	if makeTargetDependencies(makefileText, "bench-regression-check")["governance-maturity-check"] {
		c.addError("GOVERNANCE_MIGRATION_MAKE_TARGET_REGRESSION", "Makefile", "bench-regression-check must not depend on governance-maturity-check")
	}
}

func makeTargetRecipe(makefileText, target string) string {
	lines := strings.Split(makefileText, "\n")
	targetPattern := regexp.MustCompile(`^` + regexp.QuoteMeta(target) + `:(?:\s|$)`)
	otherTargetPattern := regexp.MustCompile(`^[A-Za-z0-9_.-]+:(?:\s|$)`)
	var out []string
	inTarget := false
	for _, line := range lines {
		if targetPattern.MatchString(line) {
			inTarget = true
			continue
		}
		if inTarget && otherTargetPattern.MatchString(line) {
			break
		}
		if inTarget {
			out = append(out, line)
		}
	}
	return strings.Join(out, "\n")
}

func makeTargetDependencies(makefileText, target string) map[string]bool {
	pattern := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(target) + `:\s*(.*)$`)
	match := pattern.FindStringSubmatch(makefileText)
	if len(match) != 2 {
		return nil
	}
	out := map[string]bool{}
	for _, dep := range strings.Fields(match[1]) {
		out[dep] = true
	}
	return out
}

func (c *checker) addError(ruleID, path, message string) {
	c.findings = append(c.findings, govreport.Error(ruleID, path, message))
}

func writeReport(jsonOutput bool, report govreport.Envelope) {
	if jsonOutput {
		if err := govreport.WriteJSON(os.Stdout, report); err != nil {
			fmt.Fprintf(os.Stderr, "governance migration check error: [GOVERNANCE_MIGRATION_INPUT_ERROR] cannot encode JSON output: %v\n", err)
		}
		return
	}
	if report.Status == govreport.StatusFailed {
		for _, finding := range report.Findings {
			fmt.Fprintf(os.Stderr, "governance migration check error: [%s] %s\n", finding.RuleID, finding.Message)
		}
		return
	}
	fmt.Println("governance migration status is valid")
}
