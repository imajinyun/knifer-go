package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const diffFilter = "ACDMRTUXB"

type config struct {
	root                          string
	coverageFile                  string
	module                        string
	repositoryThreshold           float64
	packageThresholds             map[string]float64
	changedPackageThresholds      map[string]float64
	securitySensitivePaths        []string
	securitySensitiveMinThreshold float64
	changedSecuritySensitivePaths []string
	coverageCheckAllPackages      bool
}

type profileLine struct {
	file       string
	statements int
	count      int
}

func main() {
	rootFlag := flag.String("root", "", "repository root")
	flag.Parse()

	root := strings.TrimSpace(*rootFlag)
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "COVERAGE CHECK ERROR: cannot resolve working directory: %v\n", err)
			os.Exit(2)
		}
	}
	coverageFile := "coverage.out"
	if flag.NArg() > 0 {
		coverageFile = flag.Arg(0)
	}

	cfg, err := loadConfig(root, coverageFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "COVERAGE CHECK ERROR: %v\n", err)
		os.Exit(2)
	}
	lines, err := parseCoverageProfile(cfg.coverageFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "COVERAGE CHECK ERROR: %v\n", err)
		os.Exit(2)
	}

	total, ok := totalCoverage(lines)
	if !ok {
		fmt.Fprintln(os.Stderr, "COVERAGE CHECK ERROR: cannot read total coverage from "+cfg.coverageFile)
		os.Exit(2)
	}
	if total < cfg.repositoryThreshold {
		fmt.Fprintf(os.Stderr, "coverage %.1f%% is below required %.1f%%\n", total, cfg.repositoryThreshold)
		os.Exit(1)
	}
	fmt.Printf("coverage %.1f%% meets required %.1f%%\n", total, cfg.repositoryThreshold)

	if !cfg.coverageCheckAllPackages {
		fmt.Println("package coverage thresholds skipped for unchanged packages; set COVERAGE_CHECK_ALL_PACKAGES=1 to enforce all package thresholds")
	} else {
		if code := checkPackageThresholds(lines, cfg.packageThresholds, "", false); code != 0 {
			os.Exit(code)
		}
	}

	if code := checkPackageThresholds(lines, cfg.changedPackageThresholds, "changed package ", true); code != 0 {
		os.Exit(code)
	}

	if code := checkSecuritySensitive(lines, cfg.securitySensitivePaths, cfg.securitySensitiveMinThreshold, false); code != 0 {
		os.Exit(code)
	}
	if len(cfg.changedSecuritySensitivePaths) == 0 {
		return
	}
	if code := checkSecuritySensitive(lines, cfg.changedSecuritySensitivePaths, cfg.securitySensitiveMinThreshold, true); code != 0 {
		os.Exit(code)
	}
}

func loadConfig(root, coverageFile string) (config, error) {
	data, err := readJSON(filepath.Join(root, "ai-context.json"))
	if err != nil {
		return config{}, err
	}
	module := stringValue(mapValue(data["project"])["module"])
	coverageGates := mapValue(data["coverage_gates"])
	repositoryThreshold := numberValue(coverageGates["repository_threshold"])
	securitySensitiveMinThreshold := numberValue(coverageGates["security_sensitive_min_threshold"])
	packageThresholds := numberMap(coverageGates["package_thresholds"])

	changedFiles := changedFiles(root)
	facadeToInternal := map[string]string{}
	for _, entry := range list(data["public_facades"]) {
		mapping := mapValue(entry)
		pkg := stringValue(mapping["package"])
		internal := strings.TrimRight(stringValue(mapping["internal"]), "/")
		if pkg != "" && internal != "" {
			facadeToInternal[pkg] = internal
		}
	}

	securitySensitiveSet := map[string]struct{}{}
	changedSecuritySensitiveSet := map[string]struct{}{}
	securityPrefixToPackageDir := map[string]string{}
	for _, pkg := range stringList(data["security_sensitive_packages"]) {
		packageDir := strings.TrimRight(pkg, "/")
		securityPrefixToPackageDir[packageDir+"/"] = packageDir
		if hasStatementSource(root, packageDir) {
			securitySensitiveSet[module+"/"+packageDir] = struct{}{}
		}
		if internal := facadeToInternal[pkg]; internal != "" {
			internal = strings.TrimRight(internal, "/")
			securityPrefixToPackageDir[internal+"/"] = internal
			if hasStatementSource(root, internal) {
				securitySensitiveSet[module+"/"+internal] = struct{}{}
			}
		}
	}

	changedPackageThresholds := map[string]float64{}
	for _, path := range changedFiles {
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "/doc.go") {
			continue
		}
		packagePath := module + "/" + filepath.ToSlash(filepath.Dir(path))
		if threshold, ok := packageThresholds[packagePath]; ok {
			changedPackageThresholds[packagePath] = threshold
		}
		for prefix, packageDir := range securityPrefixToPackageDir {
			if strings.HasPrefix(path, prefix) && hasStatementSource(root, packageDir) {
				changedSecuritySensitiveSet[module+"/"+packageDir] = struct{}{}
			}
		}
	}

	cfg := config{
		root:                          root,
		coverageFile:                  coverageFile,
		module:                        module,
		repositoryThreshold:           repositoryThreshold,
		packageThresholds:             packageThresholds,
		changedPackageThresholds:      changedPackageThresholds,
		securitySensitivePaths:        sortedSet(securitySensitiveSet),
		securitySensitiveMinThreshold: securitySensitiveMinThreshold,
		changedSecuritySensitivePaths: sortedSet(changedSecuritySensitiveSet),
		coverageCheckAllPackages:      os.Getenv("COVERAGE_CHECK_ALL_PACKAGES") == "1",
	}
	cfg.applyEnvOverrides()
	return cfg, nil
}

func (c *config) applyEnvOverrides() {
	if value := strings.TrimSpace(os.Getenv("COVERAGE_THRESHOLD")); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			c.repositoryThreshold = parsed
		}
	}
	if value, ok := os.LookupEnv("PACKAGE_COVERAGE_THRESHOLDS"); ok {
		c.packageThresholds = parseThresholdList(value)
	}
	if value, ok := os.LookupEnv("CHANGED_PACKAGE_COVERAGE_THRESHOLDS"); ok {
		c.changedPackageThresholds = parseThresholdList(value)
	}
	if value, ok := os.LookupEnv("SECURITY_SENSITIVE_COVERAGE_PATHS"); ok {
		c.securitySensitivePaths = fields(value)
	}
	if value := strings.TrimSpace(os.Getenv("SECURITY_SENSITIVE_MIN_COVERAGE_THRESHOLD")); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			c.securitySensitiveMinThreshold = parsed
		}
	}
	if value, ok := os.LookupEnv("CHANGED_SECURITY_SENSITIVE_COVERAGE_PATHS"); ok {
		c.changedSecuritySensitivePaths = fields(value)
	}
}

func parseCoverageProfile(path string) ([]profileLine, error) {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("%s does not exist", path)
		}
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return nil, fmt.Errorf("cannot read total coverage from %s", path)
	}
	var lines []profileLine
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 3 {
			continue
		}
		filePath := fields[0]
		if idx := strings.Index(filePath, ":"); idx >= 0 {
			filePath = filePath[:idx]
		}
		statements, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		count, err := strconv.Atoi(fields[2])
		if err != nil {
			continue
		}
		lines = append(lines, profileLine{file: filepath.ToSlash(filePath), statements: statements, count: count})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func packageCoverage(lines []profileLine, pkg string) (float64, bool) {
	var statements, covered int
	prefix := strings.TrimRight(pkg, "/") + "/"
	for _, line := range lines {
		if !strings.HasPrefix(line.file, prefix) {
			continue
		}
		rel := strings.TrimPrefix(line.file, prefix)
		if strings.Contains(rel, "/") || !strings.HasSuffix(rel, ".go") {
			continue
		}
		statements += line.statements
		if line.count > 0 {
			covered += line.statements
		}
	}
	if statements == 0 {
		return 0, false
	}
	return float64(covered) * 100 / float64(statements), true
}

func totalCoverage(lines []profileLine) (float64, bool) {
	var statements, covered int
	for _, line := range lines {
		statements += line.statements
		if line.count > 0 {
			covered += line.statements
		}
	}
	if statements == 0 {
		return 0, false
	}
	return float64(covered) * 100 / float64(statements), true
}

func checkPackageThresholds(lines []profileLine, thresholds map[string]float64, prefix string, changed bool) int {
	for _, packagePath := range sortedKeys(thresholds) {
		threshold := thresholds[packagePath]
		total, ok := packageCoverage(lines, packagePath)
		if !ok {
			if changed {
				fmt.Fprintf(os.Stderr, "COVERAGE CHECK ERROR: changed package %s has no coverage data\n", packagePath)
			} else {
				fmt.Fprintf(os.Stderr, "COVERAGE CHECK ERROR: package %s has no coverage data\n", packagePath)
			}
			return 2
		}
		if total < threshold {
			fmt.Fprintf(os.Stderr, "%s%s coverage %.1f%% is below required %.1f%%\n", prefix, packagePath, total, threshold)
			return 1
		}
		fmt.Printf("%s%s coverage %.1f%% meets required %.1f%%\n", prefix, packagePath, total, threshold)
	}
	return 0
}

func checkSecuritySensitive(lines []profileLine, packagePaths []string, threshold float64, changed bool) int {
	if len(packagePaths) == 0 {
		return 0
	}
	var missing []string
	var below []string
	count := 0
	for _, packagePath := range packagePaths {
		total, ok := packageCoverage(lines, packagePath)
		if !ok {
			missing = append(missing, packagePath)
			continue
		}
		count++
		if threshold > 0 && total < threshold {
			below = append(below, fmt.Sprintf("%s coverage %.1f%% is below required %.1f%%", packagePath, total, threshold))
		}
	}
	if len(missing) > 0 {
		if changed {
			fmt.Fprintln(os.Stderr, "COVERAGE CHECK ERROR: changed security-sensitive package(s) have no coverage data:")
		} else {
			fmt.Fprintln(os.Stderr, "COVERAGE CHECK ERROR: security-sensitive package(s) have no coverage data:")
		}
		for _, packagePath := range missing {
			fmt.Fprintf(os.Stderr, "  - %s\n", packagePath)
		}
		return 2
	}
	if len(below) > 0 {
		if changed {
			fmt.Fprintln(os.Stderr, "COVERAGE CHECK ERROR: changed security-sensitive package(s) are below coverage threshold:")
		} else {
			fmt.Fprintln(os.Stderr, "COVERAGE CHECK ERROR: security-sensitive package(s) are below coverage threshold:")
		}
		for _, message := range below {
			fmt.Fprintf(os.Stderr, "  - %s\n", message)
		}
		return 1
	}
	label := "security-sensitive"
	if changed {
		label = "changed security-sensitive"
	}
	if threshold > 0 {
		fmt.Printf("%s coverage data present for %d package path(s), all at or above %.1f%%\n", label, count, threshold)
	} else {
		fmt.Printf("%s coverage data present for %d package path(s)\n", label, count)
	}
	return 0
}

func changedFiles(root string) []string {
	files := map[string]struct{}{}
	baseRef := os.Getenv("AGENT_CHANGE_BASE_REF")
	if baseRef == "" && os.Getenv("GITHUB_BASE_REF") != "" {
		baseRef = "origin/" + os.Getenv("GITHUB_BASE_REF")
	}
	if baseRef != "" && gitOK(root, "rev-parse", "--verify", "--quiet", baseRef+"^{commit}") {
		for _, file := range gitLines(root, "diff", "--name-only", "--diff-filter="+diffFilter, baseRef+"...HEAD", "--") {
			files[strings.Trim(file, "/")] = struct{}{}
		}
	}
	for _, args := range [][]string{
		{"diff", "--name-only", "--diff-filter=" + diffFilter, "HEAD", "--"},
		{"diff", "--name-only", "--cached", "--diff-filter=" + diffFilter, "--"},
		{"ls-files", "--others", "--exclude-standard", "--"},
	} {
		for _, file := range gitLines(root, args...) {
			files[strings.Trim(file, "/")] = struct{}{}
		}
	}
	return sortedSet(files)
}

func gitOK(root string, args ...string) bool {
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	return cmd.Run() == nil
}

func gitLines(root string, args ...string) []string {
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	var lines []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func hasStatementSource(root, packageDir string) bool {
	entries, err := os.ReadDir(filepath.Join(root, packageDir))
	if err != nil {
		return false
	}
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go") && name != "doc.go" {
			return true
		}
	}
	return false
}

func readJSON(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
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
	if mapping == nil {
		return map[string]any{}
	}
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
		if text, ok := value.(string); ok {
			out = append(out, text)
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

func numberMap(value any) map[string]float64 {
	mapping := mapValue(value)
	out := map[string]float64{}
	for key, value := range mapping {
		out[key] = numberValue(value)
	}
	return out
}

func parseThresholdList(value string) map[string]float64 {
	out := map[string]float64{}
	for _, gate := range fields(value) {
		parts := strings.SplitN(gate, "=", 2)
		if len(parts) != 2 {
			continue
		}
		threshold, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			continue
		}
		out[parts[0]] = threshold
	}
	return out
}

func fields(value string) []string {
	return strings.Fields(strings.TrimSpace(value))
}

func sortedSet(values map[string]struct{}) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		if value != "" {
			out = append(out, value)
		}
	}
	sort.Strings(out)
	return out
}

func sortedKeys(values map[string]float64) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
