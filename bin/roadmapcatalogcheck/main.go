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
	"strconv"
	"strings"

	"github.com/imajinyun/knifer-go/bin/internal/govreport"
)

var starDomains = map[string][]string{
	"Safe HTTP (`vhttp`, `vresty`, `vurl`)":    {"vhttp", "vresty", "vurl"},
	"Safe crypto (`vcrypto`, `vrand`, `vjwt`)": {"vcrypto", "vrand", "vjwt"},
	"Daily JSON/file (`vjson`, `vfile`)":       {"vjson", "vfile"},
}

type checker struct {
	root     string
	tools    map[string]any
	roadmap  string
	findings []govreport.Finding
}

func main() {
	rootFlag := flag.String("root", "", "repository root")
	contextFlag := flag.String("ai-context", "", "ai-context.json path")
	toolsFlag := flag.String("tools", "", "docs/api/tools.json path")
	roadmapFlag := flag.String("roadmap", "", "roadmap markdown path")
	jsonFlag := flag.Bool("json", false, "emit machine-readable JSON output")
	flag.Parse()

	root := strings.TrimSpace(*rootFlag)
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			writeReport(*jsonFlag, govreport.Failed([]govreport.Finding{
				govreport.Error("ROADMAP_CATALOG_INPUT_ERROR", "", fmt.Sprintf("cannot resolve working directory: %v", err)),
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
	roadmapPath := strings.TrimSpace(*roadmapFlag)
	if roadmapPath == "" {
		roadmapPath = filepath.Join(root, "docs", "superpowers", "plans", "49-roadmap.md")
	}
	if _, err := loadJSON(contextPath); err != nil {
		writeReport(*jsonFlag, govreport.Failed([]govreport.Finding{
			govreport.Error("ROADMAP_CATALOG_INPUT_ERROR", contextPath, err.Error()),
		}))
		os.Exit(1)
	}
	tools, err := loadJSON(toolsPath)
	if err != nil {
		writeReport(*jsonFlag, govreport.Failed([]govreport.Finding{
			govreport.Error("ROADMAP_CATALOG_INPUT_ERROR", toolsPath, err.Error()),
		}))
		os.Exit(1)
	}
	roadmap, err := os.ReadFile(roadmapPath)
	if err != nil {
		writeReport(*jsonFlag, govreport.Failed([]govreport.Finding{
			govreport.Error("ROADMAP_CATALOG_INPUT_ERROR", roadmapPath, fmt.Sprintf("cannot read roadmap: %v", err)),
		}))
		os.Exit(1)
	}

	c := &checker{root: root, tools: tools, roadmap: string(roadmap)}
	c.run()
	if len(c.findings) > 0 {
		writeReport(*jsonFlag, govreport.Failed(c.findings))
		os.Exit(1)
	}
	writeReport(*jsonFlag, govreport.Passed())
}

func (c *checker) run() {
	c.validateBaseline()
	c.validateStarDomainScorecard()
}

func (c *checker) validateBaseline() {
	summary := mapValue(c.tools["summary"])
	statusCounts := mapValue(summary["status_counts"])
	synopsisSources := mapValue(summary["synopsis_sources"])
	expected := map[string]int{
		"Public facade packages":             intValue(summary["package_count"], "docs/api/tools.json.summary.package_count", c),
		"Public functions":                   intValue(summary["function_count"], "docs/api/tools.json.summary.function_count", c),
		"Functions with executable examples": intValue(summary["functions_with_examples"], "docs/api/tools.json.summary.functions_with_examples", c),
		"Context-aware functions":            intValue(summary["context_aware_functions"], "docs/api/tools.json.summary.context_aware_functions", c),
		"Functions returning errors":         intValue(summary["returns_error_functions"], "docs/api/tools.json.summary.returns_error_functions", c),
		"Recommended public functions":       intValue(statusCounts["recommended"], "docs/api/tools.json.summary.status_counts.recommended", c),
		"Compatibility public functions":     intValue(statusCounts["compatibility"], "docs/api/tools.json.summary.status_counts.compatibility", c),
		"Empty function synopses":            intValue(synopsisSources["empty"], "docs/api/tools.json.summary.synopsis_sources.empty", c),
		"Facade-sourced function synopses":   intValue(synopsisSources["facade"], "docs/api/tools.json.summary.synopsis_sources.facade", c),
		"Internal-sourced function synopses": intValue(synopsisSources["internal"], "docs/api/tools.json.summary.synopsis_sources.internal", c),
	}
	actual := c.extractMetricTable("Baseline")
	for metric, expectedValue := range expected {
		actualValue, ok := actual[metric]
		if !ok {
			c.addError("ROADMAP_CATALOG_BASELINE_METRIC_MISSING", "docs/superpowers/plans/49-roadmap.md Baseline missing metric "+metric)
			continue
		}
		if actualValue != expectedValue {
			c.addError("ROADMAP_CATALOG_BASELINE_DRIFT", fmt.Sprintf("docs/superpowers/plans/49-roadmap.md Baseline %s=%d must match tools catalog value %d", metric, actualValue, expectedValue))
		}
	}
	if extra := differenceStringSet(keys(actual), keys(expected)); len(extra) > 0 {
		c.addError("ROADMAP_CATALOG_BASELINE_EXTRA_METRIC", "docs/superpowers/plans/49-roadmap.md Baseline includes unmanaged metric(s): "+strings.Join(extra, ", "))
	}
}

func (c *checker) validateStarDomainScorecard() {
	rows := map[string]map[string]string{}
	for _, row := range c.extractMarkdownRows("90-Day Star Domain Scorecard") {
		rows[row["Domain"]] = row
	}
	if missing := differenceStringSet(keys(starDomains), keys(rows)); len(missing) > 0 {
		c.addError("ROADMAP_CATALOG_SCORECARD_DOMAIN_MISSING", "docs/superpowers/plans/49-roadmap.md scorecard missing domain row(s): "+strings.Join(missing, ", "))
	}
	if extra := differenceStringSet(keys(rows), keys(starDomains)); len(extra) > 0 {
		c.addError("ROADMAP_CATALOG_SCORECARD_DOMAIN_EXTRA", "docs/superpowers/plans/49-roadmap.md scorecard includes unmanaged domain row(s): "+strings.Join(extra, ", "))
	}
	for domain, packages := range starDomains {
		row := rows[domain]
		if row == nil {
			continue
		}
		functionCount := 0
		exampleCount := 0
		for _, pkg := range packages {
			functionCount += c.packageSummaryInt(pkg, "function_count")
			exampleCount += c.packageSummaryInt(pkg, "functions_with_examples")
		}
		actualFunctions, ok := parseIntCell(row["Public functions"])
		if !ok {
			c.addError("ROADMAP_CATALOG_SCORECARD_CELL_INVALID", domain+" Public functions must be an integer")
		} else if actualFunctions != functionCount {
			c.addError("ROADMAP_CATALOG_SCORECARD_DRIFT", fmt.Sprintf("%s Public functions=%d must match tools catalog value %d", domain, actualFunctions, functionCount))
		}
		actualExamples, ok := parseIntCell(row["Examples"])
		if !ok {
			c.addError("ROADMAP_CATALOG_SCORECARD_CELL_INVALID", domain+" Examples must be an integer")
		} else if actualExamples != exampleCount {
			c.addError("ROADMAP_CATALOG_SCORECARD_DRIFT", fmt.Sprintf("%s Examples=%d must match tools catalog value %d", domain, actualExamples, exampleCount))
		}
		expectedRatio := "0.0%"
		if functionCount > 0 {
			expectedRatio = fmt.Sprintf("%.1f%%", float64(exampleCount)/float64(functionCount)*100)
		}
		if row["Example ratio"] != expectedRatio {
			c.addError("ROADMAP_CATALOG_SCORECARD_RATIO_DRIFT", fmt.Sprintf("%s Example ratio=%q must match tools catalog value %q", domain, row["Example ratio"], expectedRatio))
		}
	}
}

func (c *checker) packageSummaryInt(packageName, field string) int {
	for _, raw := range list(c.tools["packages"]) {
		pkg := mapValue(raw)
		if stringValue(pkg["name"]) != packageName {
			continue
		}
		summary := mapValue(pkg["summary"])
		return intValue(summary[field], fmt.Sprintf("docs/api/tools.json.packages.%s.summary.%s", packageName, field), c)
	}
	c.addError("ROADMAP_CATALOG_TOOLS_PACKAGE_MISSING", "docs/api/tools.json missing package "+packageName)
	return 0
}

func (c *checker) extractMetricTable(heading string) map[string]int {
	out := map[string]int{}
	for _, row := range c.extractMarkdownRows(heading) {
		metric := row["Metric"]
		value := row["Value"]
		parsed, ok := parseIntCell(value)
		if !ok {
			c.addError("ROADMAP_CATALOG_BASELINE_CELL_INVALID", fmt.Sprintf("docs/superpowers/plans/49-roadmap.md %s metric %q must be an integer, got %q", heading, metric, value))
			continue
		}
		out[metric] = parsed
	}
	return out
}

func (c *checker) extractMarkdownRows(heading string) []map[string]string {
	body, ok := markdownSection(c.roadmap, heading)
	if !ok {
		c.addError("ROADMAP_CATALOG_SECTION_MISSING", "docs/superpowers/plans/49-roadmap.md must contain ## "+heading)
		return nil
	}
	var tableLines []string
	for _, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(line, "|") {
			tableLines = append(tableLines, line)
		}
	}
	if len(tableLines) < 2 {
		c.addError("ROADMAP_CATALOG_TABLE_MISSING", fmt.Sprintf("docs/superpowers/plans/49-roadmap.md %s must contain a markdown table", heading))
		return nil
	}
	headers := splitMarkdownRow(tableLines[0])
	var rows []map[string]string
	for _, line := range tableLines[2:] {
		columns := splitMarkdownRow(line)
		if len(columns) != len(headers) {
			c.addError("ROADMAP_CATALOG_TABLE_INVALID", fmt.Sprintf("docs/superpowers/plans/49-roadmap.md %s row has %d columns, want %d", heading, len(columns), len(headers)))
			continue
		}
		row := map[string]string{}
		for index, header := range headers {
			row[header] = columns[index]
		}
		rows = append(rows, row)
	}
	return rows
}

func markdownSection(markdown, heading string) (string, bool) {
	pattern := regexp.MustCompile(`(?ms)^## ` + regexp.QuoteMeta(heading) + `\n(?P<body>.*?)(?:^## |\z)`)
	match := pattern.FindStringSubmatch(markdown)
	if len(match) < 2 {
		return "", false
	}
	return match[1], true
}

func splitMarkdownRow(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")
	parts := strings.Split(line, "|")
	for index := range parts {
		parts[index] = strings.TrimSpace(parts[index])
	}
	return parts
}

func parseIntCell(value string) (int, bool) {
	value = strings.ReplaceAll(strings.TrimSpace(value), ",", "")
	if !regexp.MustCompile(`^\d+$`).MatchString(value) {
		return 0, false
	}
	parsed, err := strconv.Atoi(value)
	return parsed, err == nil
}

func (c *checker) addError(ruleID, message string) {
	c.findings = append(c.findings, govreport.Error(ruleID, "docs/superpowers/plans/49-roadmap.md", message))
}

func writeReport(jsonOutput bool, report govreport.Envelope) {
	if jsonOutput {
		if err := govreport.WriteJSON(os.Stdout, report); err != nil {
			fmt.Fprintf(os.Stderr, "roadmap catalog check error: [ROADMAP_CATALOG_INPUT_ERROR] cannot encode JSON output: %v\n", err)
		}
		return
	}
	if report.Status == govreport.StatusFailed {
		for _, finding := range report.Findings {
			fmt.Fprintf(os.Stderr, "roadmap catalog check error: [%s] %s\n", finding.RuleID, finding.Message)
		}
		return
	}
	fmt.Println("roadmap catalog governance is valid")
}

func loadJSON(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("missing JSON file")
		}
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func intValue(value any, path string, c *checker) int {
	switch typed := value.(type) {
	case float64:
		if typed == float64(int(typed)) {
			return int(typed)
		}
	case int:
		return typed
	}
	c.addError("ROADMAP_CATALOG_TOOLS_VALUE_INVALID", path+" must be an integer")
	return 0
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

func keys[T any](values map[string]T) map[string]bool {
	out := map[string]bool{}
	for key := range values {
		out[key] = true
	}
	return out
}

func differenceStringSet(left, right map[string]bool) []string {
	var out []string
	for key := range left {
		if !right[key] {
			out = append(out, key)
		}
	}
	sort.Strings(out)
	return out
}
