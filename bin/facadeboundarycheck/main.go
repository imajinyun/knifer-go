package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/imajinyun/knifer-go/bin/internal/govreport"
)

type checker struct {
	root     string
	findings []govreport.Finding
}

var (
	packageCommentPattern = regexp.MustCompile(`(?m)^//\s+Package\s+\w+`)
	unsafeRefAccess       = regexp.MustCompile(`fieldAccessConfig\{\s*unsafeAccess:\s*true\s*\}`)
	facadeLogicPattern    = regexp.MustCompile(`^(?:if|for|switch|select|defer|go)\b|:=`)
)

var allowedFacadeLogicPaths = map[string]struct{}{
	"vcache/cache.go": {},
	"vjob/job.go":     {},
	"vnum/arith.go":   {},
	"vrand/rand.go":   {},
	"vset/set.go":     {},
	"vskt/socket.go":  {},
	"vxml/element.go": {},
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
			writeReport(*jsonFlag, govreport.Failed([]govreport.Finding{govreport.Error("FACADE_BOUNDARY_INPUT_ERROR", "", fmt.Sprintf("cannot resolve working directory: %v", err))}))
			os.Exit(1)
		}
	}

	c := &checker{root: root}
	if err := c.run(); err != nil {
		c.addViolation("FACADE_BOUNDARY_INPUT_ERROR", err.Error())
	}
	if len(c.findings) > 0 {
		writeReport(*jsonFlag, govreport.Failed(c.findings))
		os.Exit(1)
	}
	writeReport(*jsonFlag, govreport.Passed())
}

func (c *checker) run() error {
	c.checkPackageDocs()
	c.checkUnsafeRefOptIn()
	return c.checkFacadeLogic()
}

func (c *checker) checkPackageDocs() {
	paths := []string{filepath.Join(c.root, "doc.go")}
	matches, _ := filepath.Glob(filepath.Join(c.root, "v*", "doc.go"))
	sort.Strings(matches)
	paths = append(paths, matches...)
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if !packageCommentPattern.Match(data) {
			rel := c.rel(path)
			c.addViolation("FACADE_BOUNDARY_DOC_COMMENT_MISSING", fmt.Sprintf("%s: doc.go must contain a package comment starting with 'Package <name>'", rel))
		}
	}
}

func (c *checker) checkUnsafeRefOptIn() {
	path := filepath.Join(c.root, "internal", "ref", "ref.go")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if unsafeRefAccess.MatchString(line) && !strings.Contains(enclosingFunc(lines, i), "call(") {
			c.addViolation("FACADE_BOUNDARY_UNSAFE_OPT_IN", fmt.Sprintf("internal/ref/ref.go:%d: unsafe field access must require explicit FieldAccessOption opt-in", i+1))
		}
	}
}

func (c *checker) checkFacadeLogic() error {
	matches, err := filepath.Glob(filepath.Join(c.root, "v*", "*.go"))
	if err != nil {
		return err
	}
	sort.Strings(matches)
	for _, path := range matches {
		base := filepath.Base(path)
		if base == "doc.go" || strings.HasSuffix(base, "_test.go") {
			continue
		}
		rel := c.rel(path)
		if _, ok := allowedFacadeLogicPaths[rel]; ok {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			c.addViolation("FACADE_BOUNDARY_READ_ERROR", fmt.Sprintf("%s: cannot read facade source", rel))
			continue
		}
		for i, raw := range strings.Split(string(data), "\n") {
			line := strings.TrimSpace(raw)
			if line == "" || strings.HasPrefix(line, "//") {
				continue
			}
			if facadeLogicPattern.MatchString(line) {
				c.addViolation(
					"FACADE_BOUNDARY_THIN_FACADE_VIOLATION",
					fmt.Sprintf("%s:%d: facade packages should not contain implementation control flow or local state; move logic to internal/*", rel, i+1),
				)
			}
		}
	}
	return nil
}

func enclosingFunc(lines []string, idx int) string {
	for i := idx; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "func ") {
			return line
		}
	}
	return ""
}

func (c *checker) rel(path string) string {
	rel, err := filepath.Rel(c.root, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(rel)
}

func (c *checker) addViolation(ruleID, message string) {
	c.findings = append(c.findings, govreport.Error(ruleID, "", message))
}

func writeReport(jsonOutput bool, report govreport.Envelope) {
	if jsonOutput {
		if err := govreport.WriteJSON(os.Stdout, report); err != nil {
			fmt.Fprintf(os.Stderr, "FACADE BOUNDARY VIOLATION: [FACADE_BOUNDARY_INPUT_ERROR] cannot encode JSON output: %v\n", err)
		}
		return
	}
	if report.Status == govreport.StatusFailed {
		for _, finding := range report.Findings {
			fmt.Fprintf(os.Stderr, "FACADE BOUNDARY VIOLATION: [%s] %s\n", finding.RuleID, finding.Message)
		}
		return
	}
	fmt.Println("facade boundary governance is valid")
}
