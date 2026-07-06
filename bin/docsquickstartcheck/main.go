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
)

type checker struct {
	root   string
	errors []ruleError
}

type ruleError struct {
	id      string
	message string
}

const panicErrToken = "panic" + "(err)"

func main() {
	rootFlag := flag.String("root", "", "repository root to validate")
	flag.Parse()

	root := strings.TrimSpace(*rootFlag)
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "docs quickstart check error: cannot resolve working directory: %v\n", err)
			os.Exit(1)
		}
	}

	c := &checker{root: root}
	if err := c.run(); err != nil {
		c.addError("DOCS_QUICKSTART_INPUT_ERROR", err.Error())
	}
	if len(c.errors) > 0 {
		for _, err := range c.errors {
			fmt.Fprintf(os.Stderr, "docs quickstart check error: [%s] %s\n", err.id, err.message)
		}
		os.Exit(1)
	}
}

func (c *checker) run() error {
	data, err := readJSON(filepath.Join(c.root, "ai-context.json"))
	if err != nil {
		return err
	}

	publicFacades, ok := data["public_facades"].([]any)
	if !ok {
		c.addError("DOCS_QUICKSTART_SCHEMA_INVALID", "ai-context.json public_facades must be a list")
		publicFacades = nil
	}
	docsQualityProfiles, _ := data["docs_quality_profiles"].(map[string]any)
	profilePackages := map[string]any{}
	allowedProfiles := map[string]struct{}{}
	if docsQualityProfiles != nil {
		if packages, ok := docsQualityProfiles["packages"].(map[string]any); ok {
			profilePackages = packages
		}
		if allowed, ok := docsQualityProfiles["allowed_profiles"].([]any); ok {
			for _, value := range allowed {
				if text, ok := value.(string); ok {
					allowedProfiles[text] = struct{}{}
				}
			}
		}
	}

	docDir := filepath.Join(c.root, "docs", "doc")
	indexText := ""
	indexBytes, err := os.ReadFile(filepath.Join(docDir, "README.md"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			c.addError("DOCS_QUICKSTART_INDEX_MISSING", "docs/doc/README.md is missing")
		} else {
			c.addError("DOCS_QUICKSTART_READ_ERROR", fmt.Sprintf("cannot read docs/doc/README.md: %v", err))
		}
	} else {
		indexText = string(indexBytes)
	}

	docsByPackage := c.quickstartDocs(docDir)
	knownPackages := map[string]struct{}{}
	for _, entry := range publicFacades {
		mapping, ok := entry.(map[string]any)
		if !ok {
			c.addError("DOCS_QUICKSTART_SCHEMA_INVALID", fmt.Sprintf("invalid public_facades entry: %v", entry))
			continue
		}
		pkg, ok := mapping["package"].(string)
		if !ok || pkg == "" {
			c.addError("DOCS_QUICKSTART_SCHEMA_INVALID", fmt.Sprintf("invalid public_facades entry: %v", entry))
			continue
		}
		knownPackages[pkg] = struct{}{}
	}

	if docsQualityProfiles == nil {
		c.addError("DOCS_QUICKSTART_PROFILE_SCHEMA", "ai-context.json docs_quality_profiles must be an object")
	}
	if len(allowedProfiles) == 0 {
		c.addError("DOCS_QUICKSTART_PROFILE_SCHEMA", "ai-context.json docs_quality_profiles.allowed_profiles must be non-empty")
	}
	if _, ok := docsQualityProfiles["packages"].(map[string]any); !ok {
		c.addError("DOCS_QUICKSTART_PROFILE_SCHEMA", "ai-context.json docs_quality_profiles.packages must be an object")
	}

	profilePackageNames := map[string]struct{}{}
	for pkg := range profilePackages {
		profilePackageNames[pkg] = struct{}{}
	}
	if missing := difference(knownPackages, profilePackageNames); len(missing) > 0 {
		c.addError("DOCS_QUICKSTART_PROFILE_COVERAGE", "docs_quality_profiles.packages missing facade package(s): "+strings.Join(missing, ", "))
	}
	if extra := difference(profilePackageNames, knownPackages); len(extra) > 0 {
		c.addError("DOCS_QUICKSTART_PROFILE_COVERAGE", "docs_quality_profiles.packages contains unknown facade package(s): "+strings.Join(extra, ", "))
	}

	for _, entry := range publicFacades {
		mapping, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		pkg, ok := mapping["package"].(string)
		if !ok || pkg == "" {
			continue
		}
		profiles := c.packageProfiles(pkg, profilePackages, allowedProfiles)
		matches := docsByPackage[pkg]
		if len(matches) == 0 {
			c.addError("DOCS_QUICKSTART_MISSING_DOC", "missing quickstart doc for "+pkg)
			continue
		}
		if len(matches) > 1 {
			c.addError("DOCS_QUICKSTART_DUPLICATE_DOC", fmt.Sprintf("multiple quickstart docs for %s: %s", pkg, strings.Join(matches, ", ")))
			continue
		}
		filename := matches[0]
		textBytes, err := os.ReadFile(filepath.Join(docDir, filename))
		if err != nil {
			c.addError("DOCS_QUICKSTART_READ_ERROR", fmt.Sprintf("cannot read %s: %v", filename, err))
			continue
		}
		c.validateDoc(pkg, filename, string(textBytes), profiles, indexText)
	}

	if extraDocs := difference(keys(docsByPackage), knownPackages); len(extraDocs) > 0 {
		c.addError("DOCS_QUICKSTART_UNKNOWN_DOC", "quickstart docs exist for unknown facade package(s): "+strings.Join(extraDocs, ", "))
	}

	if len(c.errors) == 0 {
		fmt.Printf("quickstart docs are valid (%d public facades)\n", len(publicFacades))
	}
	return nil
}

func readJSON(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("missing ai-context.json")
		}
		return nil, fmt.Errorf("cannot read ai-context.json: %w", err)
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("invalid ai-context.json: %w", err)
	}
	return out, nil
}

var docFilePattern = regexp.MustCompile(`^\d{2}-v[a-z0-9]+\.md$`)

func (c *checker) quickstartDocs(docDir string) map[string][]string {
	entries, err := os.ReadDir(docDir)
	if err != nil {
		c.addError("DOCS_QUICKSTART_READ_ERROR", fmt.Sprintf("cannot read docs/doc: %v", err))
		return map[string][]string{}
	}
	docs := map[string][]string{}
	for _, entry := range entries {
		name := entry.Name()
		if !docFilePattern.MatchString(name) {
			continue
		}
		pkg := strings.TrimSuffix(regexp.MustCompile(`^\d{2}-`).ReplaceAllString(name, ""), ".md")
		docs[pkg] = append(docs[pkg], name)
	}
	for pkg := range docs {
		sort.Strings(docs[pkg])
	}
	return docs
}

func (c *checker) packageProfiles(pkg string, profilePackages map[string]any, allowedProfiles map[string]struct{}) map[string]struct{} {
	raw, ok := profilePackages[pkg].([]any)
	if !ok || len(raw) == 0 {
		c.addError("DOCS_QUICKSTART_PROFILE_COVERAGE", fmt.Sprintf("docs_quality_profiles.packages.%s must contain at least one profile", pkg))
		return map[string]struct{}{"error_returning": {}}
	}
	profiles := map[string]struct{}{}
	for _, value := range raw {
		text, ok := value.(string)
		if !ok {
			continue
		}
		profiles[text] = struct{}{}
	}
	if unknown := difference(profiles, allowedProfiles); len(unknown) > 0 {
		c.addError("DOCS_QUICKSTART_PROFILE_SCHEMA", fmt.Sprintf("docs_quality_profiles.packages.%s contains unknown profile(s): %s", pkg, strings.Join(unknown, ", ")))
	}
	if _, errorReturning := profiles["error_returning"]; errorReturning {
		if _, noErrorReturning := profiles["no_error_returning"]; noErrorReturning {
			c.addError("DOCS_QUICKSTART_PROFILE_SCHEMA", fmt.Sprintf("docs_quality_profiles.packages.%s cannot combine error_returning and no_error_returning", pkg))
		}
	}
	return profiles
}

var (
	checklistPattern     = regexp.MustCompile(`(?m)^## .*checklist$|^## Safety notes$`)
	goFencePattern       = regexp.MustCompile("(?s)```go\n(.*?)\n```")
	relatedPackagePat    = regexp.MustCompile(`(?m)^- Use ` + "`" + `v[a-z0-9]+` + "`" + ` `)
	packageMainPattern   = regexp.MustCompile(`(?m)^package\s+main\b`)
	importBlockPattern   = regexp.MustCompile(`(?s)import\s*\((.*?)\)`)
	singleImportPattern  = regexp.MustCompile(`import\s+"([^"]+)"`)
	importLineInBlockPat = regexp.MustCompile(`"([^"]+)"`)
)

func (c *checker) validateDoc(pkg, filename, text string, profiles map[string]struct{}, indexText string) {
	titleExact := regexp.MustCompile(`(?m)^# ` + regexp.QuoteMeta(pkg) + ` Quickstart\s*$`)
	titleAdapter := regexp.MustCompile(`(?m)^# ` + regexp.QuoteMeta(pkg) + `: .+\s*$`)
	if !titleExact.MatchString(text) && !titleAdapter.MatchString(text) {
		c.addError("DOCS_QUICKSTART_TITLE_INVALID", fmt.Sprintf("%s must start with '# %s Quickstart' or an approved adapter title", filename, pkg))
	}
	for _, section := range []string{
		"## Golden path APIs",
		"## Which helper should I use?",
		"## Related packages",
		"## Benchmarks and trade-offs",
		"## FAQ",
	} {
		if !strings.Contains(text, section) {
			c.addError("DOCS_QUICKSTART_SECTION_MISSING", fmt.Sprintf("%s is missing required section %q", filename, section))
		}
	}
	if !relatedPackagePat.MatchString(text) {
		c.addError("DOCS_QUICKSTART_RELATED_MISSING", fmt.Sprintf("%s must include at least one related-package bullet using 'Use `v...`'", filename))
	}
	if !strings.Contains(text, "Prefer") && !strings.Contains(text, "Use") {
		c.addError("DOCS_QUICKSTART_HELPER_GUIDANCE_MISSING", fmt.Sprintf("%s helper guidance must include explicit use/prefer wording", filename))
	}
	lowerText := strings.ToLower(text)
	if hasProfile(profiles, "error_returning") && !containsAny(lowerText, []string{"error", "errors.is", panicErrToken, "err != nil"}) {
		c.addError("DOCS_QUICKSTART_ERROR_BEHAVIOR_MISSING", fmt.Sprintf("%s must document error behavior or explicitly show error handling", filename))
	}
	if hasProfile(profiles, "no_error_returning") && !containsAny(lowerText, []string{"no error", "do not return errors", "does not return errors", "return errors?"}) {
		c.addError("DOCS_QUICKSTART_NO_ERROR_CONTRACT_MISSING", fmt.Sprintf("%s profile no_error_returning must explicitly state that facade helpers do not return errors", filename))
	}
	if hasProfile(profiles, "security_sensitive") {
		c.requirePhrase(filename, lowerText, []string{"untrusted", "trust boundary", "safety checklist", "safe ", "security", "credential", "secret"}, "security_sensitive", "trust-boundary or safe-input guidance")
	}
	if hasProfile(profiles, "provider_contract") {
		c.requirePhrase(filename, lowerText, []string{"provider injection", "injected provider", "no built-in", "does not read", "does not open", "credential"}, "provider_contract", "provider injection and no-default-provider boundaries")
	}
	if hasProfile(profiles, "heavy_extension") {
		c.requirePhrase(filename, lowerText, []string{"dependency", "trade-off", "optional", "directly", "benchmark"}, "heavy_extension", "dependency or trade-off guidance")
	}
	if !regexp.MustCompile(`(?m)^## When not to use`).MatchString(text) {
		c.addError("DOCS_QUICKSTART_WHEN_NOT_MISSING", fmt.Sprintf("%s is missing required section '## When not to use ...'", filename))
	}
	if !checklistPattern.MatchString(text) {
		c.addError("DOCS_QUICKSTART_CHECKLIST_MISSING", fmt.Sprintf("%s is missing a checklist section", filename))
	}
	if strings.Count(text, "```")%2 != 0 {
		c.addError("DOCS_QUICKSTART_CODE_FENCE_UNBALANCED", fmt.Sprintf("%s has unbalanced fenced code blocks", filename))
	}

	goBlocks := goFencePattern.FindAllStringSubmatch(text, -1)
	var runnableFacadeBlocks []string
	for _, match := range goBlocks {
		block := match[1]
		if !packageMainPattern.MatchString(block) {
			continue
		}
		if _, ok := collectImports(block)["github.com/imajinyun/knifer-go/"+pkg]; ok {
			runnableFacadeBlocks = append(runnableFacadeBlocks, block)
		}
	}
	if len(goBlocks) > 0 && len(runnableFacadeBlocks) == 0 {
		c.addError("DOCS_QUICKSTART_RUNNABLE_EXAMPLE_MISSING", fmt.Sprintf("%s must include at least one runnable package main example that imports %s", filename, pkg))
	}
	for i, block := range runnableFacadeBlocks {
		if !strings.Contains(block, "func main()") {
			c.addError("DOCS_QUICKSTART_RUNNABLE_EXAMPLE_INVALID", fmt.Sprintf("%s runnable facade example %d must define func main()", filename, i+1))
		}
	}
	if len(runnableFacadeBlocks) > 0 {
		hasObservableOutput := false
		for _, block := range runnableFacadeBlocks {
			if strings.Contains(block, "fmt.Println") || strings.Contains(block, "fmt.Printf") || strings.Contains(block, panicErrToken) {
				hasObservableOutput = true
				break
			}
		}
		if !hasObservableOutput {
			c.addError("DOCS_QUICKSTART_RUNNABLE_EXAMPLE_INVALID", fmt.Sprintf("%s must include at least one runnable facade example with observable output or explicit error handling", filename))
		}
	}
	if indexText != "" && !strings.Contains(indexText, "]("+filename+")") {
		c.addError("DOCS_QUICKSTART_INDEX_LINK_MISSING", fmt.Sprintf("docs/doc/README.md does not link to %s", filename))
	}
}

func collectImports(block string) map[string]struct{} {
	imports := map[string]struct{}{}
	for _, match := range singleImportPattern.FindAllStringSubmatch(block, -1) {
		imports[match[1]] = struct{}{}
	}
	for _, blockMatch := range importBlockPattern.FindAllStringSubmatch(block, -1) {
		for _, line := range strings.Split(blockMatch[1], "\n") {
			match := importLineInBlockPat.FindStringSubmatch(strings.TrimSpace(line))
			if len(match) > 0 {
				imports[match[1]] = struct{}{}
			}
		}
	}
	return imports
}

func (c *checker) requirePhrase(filename, lowerText string, phrases []string, profile, purpose string) {
	if !containsAny(lowerText, phrases) {
		c.addError("DOCS_QUICKSTART_PROFILE_GUIDANCE_MISSING", fmt.Sprintf("%s profile %s must document %s", filename, profile, purpose))
	}
}

func containsAny(text string, terms []string) bool {
	for _, term := range terms {
		if strings.Contains(text, term) {
			return true
		}
	}
	return false
}

func hasProfile(profiles map[string]struct{}, profile string) bool {
	_, ok := profiles[profile]
	return ok
}

func (c *checker) addError(ruleID, message string) {
	c.errors = append(c.errors, ruleError{id: ruleID, message: message})
}

func keys[V any](values map[string]V) map[string]struct{} {
	out := make(map[string]struct{}, len(values))
	for key := range values {
		out[key] = struct{}{}
	}
	return out
}

func difference(left, right map[string]struct{}) []string {
	var out []string
	for value := range left {
		if _, ok := right[value]; !ok {
			out = append(out, value)
		}
	}
	sort.Strings(out)
	return out
}
