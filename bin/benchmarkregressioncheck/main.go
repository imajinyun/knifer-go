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

var expectedHotPackages = setOf("./vjson", "./vstr", "./vslice", "./vmap", "./vdb", "./vhttp", "./vcodec")

type checker struct {
	root     string
	makefile string
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
				govreport.Error("BENCHMARK_REGRESSION_INPUT_ERROR", "", fmt.Sprintf("cannot resolve working directory: %v", err)),
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
			govreport.Error("BENCHMARK_REGRESSION_INPUT_ERROR", contextPath, err.Error()),
		}))
		os.Exit(1)
	}
	makefileBytes, err := os.ReadFile(filepath.Join(root, "Makefile"))
	if err != nil {
		writeReport(*jsonFlag, govreport.Failed([]govreport.Finding{
			govreport.Error("BENCHMARK_REGRESSION_INPUT_ERROR", "Makefile", fmt.Sprintf("cannot read Makefile: %v", err)),
		}))
		os.Exit(1)
	}

	c := &checker{root: root, makefile: string(makefileBytes)}
	c.run(data)
	if len(c.findings) > 0 {
		writeReport(*jsonFlag, govreport.Failed(c.findings))
		os.Exit(1)
	}
	writeReport(*jsonFlag, govreport.Passed())
}

func (c *checker) run(data map[string]any) {
	bench := mapValue(data["benchmark_regression"])
	if bench == nil {
		c.addError("BENCHMARK_REGRESSION_SCHEMA_INVALID", "benchmark_regression must be an object")
		return
	}
	if value, ok := bench["benchstat_required"].(bool); !ok || !value {
		c.addError("BENCHMARK_REGRESSION_BENCHSTAT_REQUIRED", "benchmark_regression.benchstat_required must be true")
	}
	for _, field := range []string{"baseline_command", "compare_command"} {
		value := stringValue(bench[field])
		if !strings.HasPrefix(value, "make bench-") {
			c.addError("BENCHMARK_REGRESSION_COMMAND_INVALID", fmt.Sprintf("benchmark_regression.%s must start with a make bench-* target", field))
		}
	}
	c.validateThresholds(mapValue(bench["thresholds"]))
	tracked := stringList(bench["tracked_packages"])
	trackedSet := setOf(tracked...)
	c.validateTrackedPackages(tracked)
	benchPackageSet := c.benchPackageSet()
	c.validateHotPaths(bench["hot_path_packages"], trackedSet, benchPackageSet)
	for _, target := range []string{"bench-baseline", "bench-compare", "bench-regression-check", "benchstat"} {
		if !regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(target) + `:(?:\s|$)`).MatchString(c.makefile) {
			c.addError("BENCHMARK_REGRESSION_MAKE_TARGET_MISSING", "Makefile must define benchmark target "+target)
		}
	}
}

func (c *checker) validateThresholds(thresholds map[string]any) {
	if thresholds == nil {
		c.addError("BENCHMARK_REGRESSION_SCHEMA_INVALID", "benchmark_regression.thresholds must be an object")
		return
	}
	for _, key := range []string{"ns_per_op_regression_percent", "bytes_per_op_regression_percent", "allocs_per_op_regression_percent"} {
		if value := numberValue(thresholds[key]); value <= 0 {
			c.addError("BENCHMARK_REGRESSION_THRESHOLD_INVALID", fmt.Sprintf("benchmark_regression.thresholds.%s must be a positive number", key))
		}
	}
	if value := numberValue(thresholds["minimum_count"]); value < 10 {
		c.addError("BENCHMARK_REGRESSION_THRESHOLD_INVALID", "benchmark_regression.thresholds.minimum_count must be at least 10")
	}
}

func (c *checker) validateTrackedPackages(tracked []string) {
	if len(tracked) < 5 {
		c.addError("BENCHMARK_REGRESSION_TRACKED_PACKAGES_INCOMPLETE", "benchmark_regression.tracked_packages must include representative core and facade packages")
	}
	hasInternal := false
	hasFacade := false
	for _, pkg := range tracked {
		if strings.HasPrefix(pkg, "./internal/") {
			hasInternal = true
		}
		if strings.HasPrefix(pkg, "./v") {
			hasFacade = true
		}
		if strings.HasPrefix(pkg, "./") && !isDir(filepath.Join(c.root, filepath.FromSlash(strings.TrimPrefix(pkg, "./")))) {
			c.addError("BENCHMARK_REGRESSION_PACKAGE_MISSING", "benchmark_regression.tracked_packages references missing package directory "+pkg)
		}
	}
	if !hasInternal {
		c.addError("BENCHMARK_REGRESSION_TRACKED_PACKAGES_INCOMPLETE", "benchmark_regression.tracked_packages must include at least one internal package")
	}
	if !hasFacade {
		c.addError("BENCHMARK_REGRESSION_TRACKED_PACKAGES_INCOMPLETE", "benchmark_regression.tracked_packages must include at least one public facade package")
	}
}

func (c *checker) validateHotPaths(value any, trackedSet, benchPackageSet map[string]bool) {
	hotPaths, ok := value.([]any)
	if !ok || len(hotPaths) < 7 {
		c.addError("BENCHMARK_REGRESSION_HOT_PATHS_INCOMPLETE", "benchmark_regression.hot_path_packages must include at least 7 hot-path package entries")
		hotPaths = nil
	}
	seen := map[string]bool{}
	for index, raw := range hotPaths {
		entry := mapValue(raw)
		if entry == nil {
			c.addError("BENCHMARK_REGRESSION_SCHEMA_INVALID", fmt.Sprintf("benchmark_regression.hot_path_packages[%d] must be an object", index))
			continue
		}
		pkg := stringValue(entry["package"])
		if pkg == "" {
			c.addError("BENCHMARK_REGRESSION_SCHEMA_INVALID", fmt.Sprintf("benchmark_regression.hot_path_packages[%d].package must be non-empty", index))
			continue
		}
		seen[pkg] = true
		if !trackedSet[pkg] {
			c.addError("BENCHMARK_REGRESSION_HOT_PATH_TRACKING", fmt.Sprintf("benchmark_regression.hot_path_packages.%s must also be listed in tracked_packages", pkg))
		}
		if !benchPackageSet[pkg] {
			c.addError("BENCHMARK_REGRESSION_HOT_PATH_BENCH_TARGET", fmt.Sprintf("benchmark_regression.hot_path_packages.%s must be covered by BENCH_PKGS, BENCH_FACADE_PKGS, or BENCH_CODEC_PKGS", pkg))
		}
		if strings.HasPrefix(pkg, "./") && !isDir(filepath.Join(c.root, filepath.FromSlash(strings.TrimPrefix(pkg, "./")))) {
			c.addError("BENCHMARK_REGRESSION_PACKAGE_MISSING", fmt.Sprintf("benchmark_regression.hot_path_packages.%s references missing package directory", pkg))
		}
		if strings.TrimSpace(stringValue(entry["owner"])) == "" {
			c.addError("BENCHMARK_REGRESSION_HOT_PATH_OWNER", fmt.Sprintf("benchmark_regression.hot_path_packages.%s.owner must be non-empty", pkg))
		}
		if stringValue(entry["threshold_profile"]) != "default" {
			c.addError("BENCHMARK_REGRESSION_THRESHOLD_PROFILE", fmt.Sprintf("benchmark_regression.hot_path_packages.%s.threshold_profile must be default", pkg))
		}
		benchmarks := stringList(entry["benchmarks"])
		if len(benchmarks) == 0 {
			c.addError("BENCHMARK_REGRESSION_BENCHMARKS_MISSING", fmt.Sprintf("benchmark_regression.hot_path_packages.%s.benchmarks must not be empty", pkg))
		}
		for _, reference := range benchmarks {
			if !c.benchmarkExists(reference) {
				c.addError("BENCHMARK_REGRESSION_BENCHMARK_MISSING", fmt.Sprintf("benchmark_regression.hot_path_packages.%s.benchmarks references missing benchmark %s", pkg, reference))
			}
		}
	}
	if missing := differenceSet(expectedHotPackages, seen); len(missing) > 0 {
		c.addError("BENCHMARK_REGRESSION_HOT_PATH_MISSING", "benchmark_regression.hot_path_packages missing package(s): "+strings.Join(missing, ", "))
	}
}

func (c *checker) benchPackageSet() map[string]bool {
	out := map[string]bool{}
	for _, variable := range []string{"BENCH_PKGS", "BENCH_FACADE_PKGS", "BENCH_CODEC_PKGS"} {
		for _, pkg := range makeVariablePackages(c.makefile, variable) {
			out[pkg] = true
		}
	}
	return out
}

func (c *checker) benchmarkExists(reference string) bool {
	match := regexp.MustCompile(`^(.+\.go):(Benchmark[A-Za-z_]\w*)$`).FindStringSubmatch(reference)
	if len(match) != 3 {
		return false
	}
	data, err := os.ReadFile(filepath.Join(c.root, filepath.FromSlash(match[1])))
	if err != nil {
		return false
	}
	return regexp.MustCompile(`(?m)^func\s+` + regexp.QuoteMeta(match[2]) + `\s*\(\s*b\s+\*testing\.B\s*\)`).Match(data)
}

func makeVariablePackages(makefile, variable string) []string {
	pattern := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(variable) + `\s*\??=\s*(.+)$`)
	match := pattern.FindStringSubmatch(makefile)
	if len(match) != 2 {
		return nil
	}
	return strings.Fields(match[1])
}

func (c *checker) addError(ruleID, message string) {
	c.findings = append(c.findings, govreport.Error(ruleID, "ai-context.json", message))
}

func writeReport(jsonOutput bool, report govreport.Envelope) {
	if jsonOutput {
		if err := govreport.WriteJSON(os.Stdout, report); err != nil {
			fmt.Fprintf(os.Stderr, "benchmark regression check error: [BENCHMARK_REGRESSION_INPUT_ERROR] cannot encode JSON output: %v\n", err)
		}
		return
	}
	if report.Status == govreport.StatusFailed {
		for _, finding := range report.Findings {
			fmt.Fprintf(os.Stderr, "benchmark regression check error: [%s] %s\n", finding.RuleID, finding.Message)
		}
		return
	}
	fmt.Println("benchmark regression metadata is valid")
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

func numberValue(value any) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case int:
		return float64(v)
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

func isDir(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && stat.IsDir()
}
