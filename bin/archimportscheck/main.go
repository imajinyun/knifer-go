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

	"github.com/imajinyun/knifer-go/bin/internal/govreport"
)

type checker struct {
	root      string
	module    string
	allowlist map[string][]string
	findings  []govreport.Finding
	quiet     bool
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
			writeReport(*jsonFlag, govreport.Failed([]govreport.Finding{govreport.Error("ARCH_IMPORT_INPUT_ERROR", "", fmt.Sprintf("cannot resolve working directory: %v", err))}))
			os.Exit(1)
		}
	}

	c := &checker{root: root, quiet: *jsonFlag}
	if err := c.run(); err != nil {
		c.addViolation("ARCH_IMPORT_INPUT_ERROR", err.Error())
	}
	if len(c.findings) > 0 {
		writeReport(*jsonFlag, govreport.Failed(c.findings))
		os.Exit(1)
	}
	writeReport(*jsonFlag, govreport.Passed())
}

func (c *checker) run() error {
	c.logStage("arch imports: resolving module")
	module, err := c.goOutput("list", "-m")
	if err != nil {
		return fmt.Errorf("cannot resolve module path via 'go list -m'")
	}
	for _, line := range strings.Split(module, "\n") {
		if strings.Contains(line, "knifer-go") {
			c.module = strings.TrimSpace(line)
			break
		}
	}
	if c.module == "" {
		return fmt.Errorf("cannot resolve module path via 'go list -m'")
	}
	allowlist, err := loadHeavyAllowlist(filepath.Join(c.root, "ai-context.json"))
	if err != nil {
		return err
	}
	c.allowlist = allowlist
	c.scanPublicFacades()
	c.scanInternalPackages()
	c.scanHeavyDependencyIsolation()
	return nil
}

func (c *checker) scanPublicFacades() {
	c.logStage("arch imports: scanning public facades")
	dirs, _ := filepath.Glob(filepath.Join(c.root, "v*"))
	sort.Strings(dirs)
	for _, dir := range dirs {
		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			continue
		}
		pkg := filepath.Base(dir)
		goFiles, _ := filepath.Glob(filepath.Join(dir, "*.go"))
		if len(goFiles) == 0 {
			continue
		}
		if _, err := os.Stat(filepath.Join(dir, "doc.go")); err != nil {
			c.addViolation("ARCH_IMPORT_MISSING_DOC", fmt.Sprintf("%s: missing doc.go", pkg))
		}
		imports := c.importsForPattern("./" + pkg)
		for _, imp := range imports {
			switch {
			case strings.HasPrefix(imp, c.module+"/v"):
				c.addViolation("ARCH_IMPORT_FACADE_TO_FACADE", fmt.Sprintf("%s: imports another public package %s (v* packages must not depend on each other)", pkg, imp))
			case strings.HasPrefix(imp, c.module+"/internal/"):
				rel := strings.TrimPrefix(imp, c.module+"/")
				if stat, err := os.Stat(filepath.Join(c.root, filepath.FromSlash(rel))); err != nil || !stat.IsDir() {
					c.addViolation("ARCH_IMPORT_MISSING_INTERNAL", fmt.Sprintf("%s: imports non-existent internal path %s", pkg, imp))
				}
			case strings.HasPrefix(imp, c.module):
			default:
				if isExternalImport(imp) && !c.allowedHeavyExternalImport(pkg, imp) {
					c.addViolation("ARCH_IMPORT_FACADE_EXTERNAL_DEP", fmt.Sprintf("%s: imports third-party dependency %s (facade dependency surface must be allowlisted)", pkg, imp))
				}
			}
		}
		for _, file := range goFiles {
			base := filepath.Base(file)
			if base == "doc.go" || strings.HasSuffix(base, "_test.go") {
				continue
			}
			fileImports := c.importsForPattern(filepath.ToSlash(strings.TrimPrefix(file, c.root+string(os.PathSeparator))))
			internalCount := 0
			for _, imp := range fileImports {
				if strings.HasPrefix(imp, c.module+"/internal/") {
					internalCount++
				}
			}
			if internalCount == 0 {
				rel := filepath.ToSlash(strings.TrimPrefix(file, c.root+string(os.PathSeparator)))
				c.addViolation("ARCH_IMPORT_FACADE_NO_INTERNAL_DELEGATION", fmt.Sprintf("%s: does not import any internal/ implementation (each facade source file must delegate to internal)", rel))
			}
		}
	}
}

func (c *checker) scanInternalPackages() {
	c.logStage("arch imports: scanning internal packages")
	packages := c.goListPackages("./internal/...")
	for _, pkg := range packages {
		imports := c.importsForPattern(pkg)
		for _, imp := range imports {
			if strings.HasPrefix(imp, c.module+"/v") {
				rel := strings.TrimPrefix(pkg, c.module+"/")
				c.addViolation("ARCH_IMPORT_INTERNAL_TO_FACADE", fmt.Sprintf("%s: imports public facade %s (internal packages must not depend on v* packages)", rel, imp))
			}
		}
	}
}

func (c *checker) scanHeavyDependencyIsolation() {
	c.logStage("arch imports: scanning heavy dependency isolation")
	packages := append(c.goListPackages("./internal/..."), c.goListPackages("./v...")...)
	sort.Strings(packages)
	for _, pkg := range packages {
		rel := strings.TrimPrefix(pkg, c.module+"/")
		imports := c.importsForPattern(pkg)
		for _, imp := range imports {
			if c.isHeavyExternalImport(imp) && !c.allowedHeavyExternalImport(rel, imp) {
				c.addViolation("ARCH_IMPORT_HEAVY_DEPENDENCY_LEAK", fmt.Sprintf("%s: imports heavy optional dependency %s outside its isolated package family", rel, imp))
			}
		}
	}
}

func (c *checker) importsForPattern(pattern string) []string {
	out, err := c.goOutput("list", "-f", "{{range .Imports}}{{println .}}{{end}}", pattern)
	if err != nil {
		return nil
	}
	return nonEmptyLines(out)
}

func (c *checker) goListPackages(pattern string) []string {
	out, err := c.goOutput("list", pattern)
	if err != nil {
		return nil
	}
	return nonEmptyLines(out)
}

func (c *checker) goOutput(args ...string) (string, error) {
	cmd := exec.Command("go", args...)
	cmd.Dir = c.root
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func loadHeavyAllowlist(path string) (map[string][]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("missing ai-context.json")
		}
		return nil, err
	}
	var context map[string]any
	if err := json.Unmarshal(data, &context); err != nil {
		return nil, err
	}
	allowlist := map[string][]string{}
	tiers, _ := context["dependency_tiers"].(map[string]any)
	raw, _ := tiers["heavy_dependency_allowlist"].(map[string]any)
	for importPath, value := range raw {
		for _, prefix := range list(value) {
			if text, ok := prefix.(string); ok {
				allowlist[importPath] = append(allowlist[importPath], text)
			}
		}
	}
	return allowlist, nil
}

func (c *checker) isHeavyExternalImport(importPath string) bool {
	for pattern := range c.allowlist {
		if importPatternMatches(pattern, importPath) {
			return true
		}
	}
	return false
}

func (c *checker) allowedHeavyExternalImport(rel, importPath string) bool {
	for pattern, prefixes := range c.allowlist {
		if !importPatternMatches(pattern, importPath) {
			continue
		}
		for _, prefix := range prefixes {
			if rel == prefix || strings.HasPrefix(rel, prefix+"/") {
				return true
			}
		}
	}
	return false
}

func importPatternMatches(pattern, importPath string) bool {
	if strings.Contains(pattern, "*") {
		match, err := filepath.Match(pattern, importPath)
		return err == nil && match
	}
	return pattern == importPath
}

func isExternalImport(importPath string) bool {
	first := strings.Split(importPath, "/")[0]
	return strings.Contains(first, ".")
}

func list(value any) []any {
	values, _ := value.([]any)
	return values
}

func nonEmptyLines(value string) []string {
	var out []string
	for _, line := range strings.Split(value, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

func (c *checker) addViolation(ruleID, message string) {
	c.findings = append(c.findings, govreport.Error(ruleID, "", message))
}

func (c *checker) logStage(message string) {
	if !c.quiet {
		fmt.Println(message)
	}
}

func writeReport(jsonOutput bool, report govreport.Envelope) {
	if jsonOutput {
		if err := govreport.WriteJSON(os.Stdout, report); err != nil {
			fmt.Fprintf(os.Stderr, "ARCH IMPORT VIOLATION: [ARCH_IMPORT_INPUT_ERROR] cannot encode JSON output: %v\n", err)
		}
		return
	}
	if report.Status == govreport.StatusFailed {
		for _, finding := range report.Findings {
			fmt.Fprintf(os.Stderr, "ARCH IMPORT VIOLATION: [%s] %s\n", finding.RuleID, finding.Message)
		}
		fmt.Fprintln(os.Stderr, "Architecture import check failed. See violations above.")
		return
	}
	fmt.Println("architecture import governance passed")
}
