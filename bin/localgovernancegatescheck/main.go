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

var requiredTargets = []string{"quick-check", "full-check", "ci-workflow-check", "release-check"}

type checker struct {
	makefile string
	findings []govreport.Finding
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
				govreport.Error("LOCAL_GOVERNANCE_INPUT_ERROR", "", fmt.Sprintf("cannot resolve working directory: %v", err)),
			}))
			os.Exit(1)
		}
	}
	makefilePath := filepath.Join(root, "Makefile")
	makefileBytes, err := os.ReadFile(makefilePath)
	if err != nil {
		writeReport(*jsonFlag, govreport.Failed([]govreport.Finding{
			govreport.Error("LOCAL_GOVERNANCE_INPUT_ERROR", "Makefile", fmt.Sprintf("cannot read Makefile: %v", err)),
		}))
		os.Exit(1)
	}

	c := &checker{makefile: string(makefileBytes)}
	c.run()
	if len(c.findings) > 0 {
		writeReport(*jsonFlag, govreport.Failed(c.findings))
		os.Exit(1)
	}
	writeReport(*jsonFlag, govreport.Passed())
}

func (c *checker) run() {
	for _, target := range requiredTargets {
		if !c.targetExists(target) {
			c.addError("LOCAL_GOVERNANCE_MAKE_TARGET_MISSING", "Makefile must define target "+target)
			continue
		}
		if !c.targetDependsOn(target, "bench-regression-check", map[string]bool{}) {
			c.addError("LOCAL_GOVERNANCE_BENCH_REGRESSION_DEP_MISSING", "Makefile target "+target+" must depend on bench-regression-check")
		}
	}
}

func (c *checker) targetExists(target string) bool {
	return targetHeaderPattern(target).MatchString(c.makefile)
}

func (c *checker) targetDependsOn(target, dependency string, seen map[string]bool) bool {
	if target == dependency {
		return true
	}
	if seen[target] {
		return false
	}
	seen[target] = true
	for _, dep := range c.targetDependencies(target) {
		if dep == dependency {
			return true
		}
		if validTargetName(dep) && c.targetDependsOn(dep, dependency, seen) {
			return true
		}
	}
	for _, called := range c.targetRecipeCalls(target) {
		if called == dependency {
			return true
		}
		if c.targetDependsOn(called, dependency, seen) {
			return true
		}
	}
	return false
}

func (c *checker) targetDependencies(target string) []string {
	match := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(target) + `:\s*(.*)$`).FindStringSubmatch(c.makefile)
	if len(match) != 2 {
		return nil
	}
	var out []string
	for _, dep := range strings.Fields(match[1]) {
		if dep != "" && !strings.HasPrefix(dep, "$") {
			out = append(out, dep)
		}
	}
	return out
}

func (c *checker) targetRecipeCalls(target string) []string {
	body := c.targetRecipe(target)
	matches := regexp.MustCompile(`(?:\$\(MAKE\)|make)\s+([A-Za-z0-9_.-]+)`).FindAllStringSubmatch(body, -1)
	var out []string
	for _, match := range matches {
		out = append(out, match[1])
	}
	return out
}

func (c *checker) targetRecipe(target string) string {
	lines := strings.Split(c.makefile, "\n")
	targetPattern := targetHeaderPattern(target)
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

func targetHeaderPattern(target string) *regexp.Regexp {
	return regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(target) + `:(?:\s|$)`)
}

func validTargetName(value string) bool {
	return regexp.MustCompile(`^[A-Za-z0-9_.-]+$`).MatchString(value)
}

func (c *checker) addError(ruleID, message string) {
	c.findings = append(c.findings, govreport.Error(ruleID, "Makefile", message))
}

func writeReport(jsonOutput bool, report govreport.Envelope) {
	if jsonOutput {
		if err := govreport.WriteJSON(os.Stdout, report); err != nil {
			fmt.Fprintf(os.Stderr, "local governance gates check error: [LOCAL_GOVERNANCE_INPUT_ERROR] cannot encode JSON output: %v\n", err)
		}
		return
	}
	if report.Status == govreport.StatusFailed {
		for _, finding := range report.Findings {
			fmt.Fprintf(os.Stderr, "local governance gates check error: [%s] %s\n", finding.RuleID, finding.Message)
		}
		return
	}
	fmt.Println("local governance gates are valid")
}
