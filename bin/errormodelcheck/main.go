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

var expectedCodes = setOf(
	"GK_INVALID_INPUT",
	"GK_NOT_FOUND",
	"GK_UNSUPPORTED",
	"GK_UNSAFE_RESOURCE",
	"GK_TIMEOUT",
	"GK_PROVIDER_FAILURE",
	"GK_INTERNAL",
)

var requiredErrorConstants = []string{
	"ErrCodeInvalidInput",
	"ErrCodeNotFound",
	"ErrCodeUnsupported",
	"ErrCodeUnsafeResource",
	"ErrCodeTimeout",
	"ErrCodeProviderFailure",
	"ErrCodeInternal",
}

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
				govreport.Error("ERROR_MODEL_INPUT_ERROR", "", fmt.Sprintf("cannot resolve working directory: %v", err)),
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
			govreport.Error("ERROR_MODEL_INPUT_ERROR", contextPath, err.Error()),
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
	errorModel := mapValue(data["error_model"])
	if errorModel == nil {
		c.addError("ERROR_MODEL_SCHEMA_INVALID", "error_model must be an object")
		return
	}
	taxonomy, ok := errorModel["taxonomy"].([]any)
	if !ok {
		c.addError("ERROR_MODEL_TAXONOMY_SCHEMA", "error_model.taxonomy must be a list")
		return
	}
	codes := map[string]bool{}
	for index, raw := range taxonomy {
		entry := mapValue(raw)
		if entry == nil {
			c.addError("ERROR_MODEL_TAXONOMY_SCHEMA", fmt.Sprintf("error_model.taxonomy[%d] must be an object", index))
			continue
		}
		for _, key := range []string{"category", "code", "use_when"} {
			if strings.TrimSpace(stringValue(entry[key])) == "" {
				c.addError("ERROR_MODEL_TAXONOMY_SCHEMA", fmt.Sprintf("error_model.taxonomy[%d].%s must be non-empty", index, key))
			}
		}
		if code := strings.TrimSpace(stringValue(entry["code"])); code != "" {
			codes[code] = true
		}
	}
	if !sameSet(codes, expectedCodes) {
		c.addError("ERROR_MODEL_TAXONOMY_COVERAGE", "error_model.taxonomy must cover exactly: "+strings.Join(sortedBoolKeys(expectedCodes), ", "))
	}
	c.validateErrorConstants()
	for _, reference := range stringList(errorModel["contract_tests"]) {
		if !c.referenceExists(reference) {
			c.addError("ERROR_MODEL_CONTRACT_TEST_MISSING", "error_model.contract_tests references missing file or function "+reference)
		}
	}
}

func (c *checker) validateErrorConstants() {
	data, err := os.ReadFile(filepath.Join(c.root, "errors.go"))
	if err != nil {
		c.addError("ERROR_MODEL_ERRORS_GO_MISSING", "errors.go must exist and define unified error codes")
		return
	}
	text := string(data)
	for _, constantName := range requiredErrorConstants {
		if !strings.Contains(text, constantName) {
			c.addError("ERROR_MODEL_ERROR_CONSTANT_MISSING", "errors.go must define "+constantName)
		}
	}
}

func (c *checker) addError(ruleID, message string) {
	c.findings = append(c.findings, govreport.Error(ruleID, "ai-context.json", message))
}

func writeReport(jsonOutput bool, report govreport.Envelope) {
	if jsonOutput {
		if err := govreport.WriteJSON(os.Stdout, report); err != nil {
			fmt.Fprintf(os.Stderr, "error model check error: [ERROR_MODEL_INPUT_ERROR] cannot encode JSON output: %v\n", err)
		}
		return
	}
	if report.Status == govreport.StatusFailed {
		for _, finding := range report.Findings {
			fmt.Fprintf(os.Stderr, "error model check error: [%s] %s\n", finding.RuleID, finding.Message)
		}
		return
	}
	fmt.Println("error model governance is valid")
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

func (c *checker) referenceExists(reference string) bool {
	if !strings.Contains(reference, ":") {
		stat, err := os.Stat(filepath.Join(c.root, filepath.FromSlash(reference)))
		return err == nil && !stat.IsDir()
	}
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

func mapValue(value any) map[string]any {
	mapping, _ := value.(map[string]any)
	return mapping
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

func sameSet(left, right map[string]bool) bool {
	if len(left) != len(right) {
		return false
	}
	for key := range left {
		if !right[key] {
			return false
		}
	}
	return true
}

func sortedBoolKeys(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
