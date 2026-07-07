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

type violation struct {
	ruleID  string
	message string
}

type checker struct {
	root       string
	violations []violation
}

func main() {
	rootFlag := flag.String("root", "", "repository root")
	flag.Parse()

	root := strings.TrimSpace(*rootFlag)
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "PROVIDER CONTRACT VIOLATION: [PROVIDER_CONTRACT_INPUT_ERROR] cannot resolve working directory: %v\n", err)
			os.Exit(1)
		}
	}

	c := &checker{root: root}
	if err := c.run(); err != nil {
		c.addViolation("PROVIDER_CONTRACT_INPUT_ERROR", err.Error())
	}
	if len(c.violations) > 0 {
		for _, violation := range c.violations {
			fmt.Fprintf(os.Stderr, "PROVIDER CONTRACT VIOLATION: [%s] %s\n", violation.ruleID, violation.message)
		}
		os.Exit(1)
	}
}

func (c *checker) run() error {
	data, err := loadJSON(filepath.Join(c.root, "ai-context.json"))
	if err != nil {
		return err
	}
	facadeToInternal := map[string]string{}
	for _, entry := range list(data["public_facades"]) {
		mapping := mapValue(entry)
		pkg := stringValue(mapping["package"])
		internal := strings.TrimRight(stringValue(mapping["internal"]), "/")
		if pkg != "" && internal != "" {
			facadeToInternal[pkg] = internal
		}
	}
	dependencyTiers := mapValue(data["dependency_tiers"])
	providersRaw, ok := dependencyTiers["provider_contract_facades"].([]any)
	if !ok {
		c.addViolation("PROVIDER_CONTRACT_SCHEMA_INVALID", "ai-context.json dependency_tiers.provider_contract_facades must be a list")
		providersRaw = nil
	}
	providers := stringList(providersRaw)

	providerInterfacePattern := regexp.MustCompile(`type\s+\w*Provider\s+interface\s*{`)
	for _, facade := range providers {
		internal := facadeToInternal[facade]
		if internal == "" {
			c.addViolation("PROVIDER_CONTRACT_MISSING_MAPPING", fmt.Sprintf("provider contract facade %s: missing public_facades internal mapping", facade))
			continue
		}
		var paths []string
		for _, dir := range []string{facade, internal} {
			absDir := filepath.Join(c.root, dir)
			entries, err := os.ReadDir(absDir)
			if err != nil {
				c.addViolation("PROVIDER_CONTRACT_MISSING_DIRECTORY", fmt.Sprintf("provider contract facade %s: missing directory %s", facade, filepath.ToSlash(dir)))
				continue
			}
			for _, entry := range entries {
				name := entry.Name()
				if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
					continue
				}
				paths = append(paths, filepath.Join(dir, name))
			}
		}
		sort.Strings(paths)
		combined := ""
		for _, rel := range paths {
			text, err := os.ReadFile(filepath.Join(c.root, rel))
			if err != nil {
				c.addViolation("PROVIDER_CONTRACT_READ_ERROR", fmt.Sprintf("%s: cannot read provider contract source", filepath.ToSlash(rel)))
				continue
			}
			combined += "\n" + string(text)
		}
		if !providerInterfacePattern.MatchString(combined) {
			c.addViolation("PROVIDER_CONTRACT_MISSING_PROVIDER_INTERFACE", fmt.Sprintf("provider contract facade %s: must define a Provider interface contract", facade))
		}
		for _, rel := range paths {
			textBytes, err := os.ReadFile(filepath.Join(c.root, rel))
			if err != nil {
				continue
			}
			text := string(textBytes)
			slashRel := filepath.ToSlash(rel)
			for _, forbiddenImport := range []string{`"net/http"`, `"resty.dev/`, `"google.golang.org/grpc`, `"golang.org/x/oauth2`} {
				if strings.Contains(text, forbiddenImport) {
					c.addViolation("PROVIDER_CONTRACT_FORBIDDEN_IMPORT", fmt.Sprintf("%s: provider contract packages must not import concrete provider/network SDK dependency %s", slashRel, forbiddenImport))
				}
			}
			for _, forbiddenCall := range []string{"os.Getenv", "os.ReadFile", "http.NewRequest", "http.Client", "net.Dial", "grpc.Dial"} {
				if strings.Contains(text, forbiddenCall) {
					c.addViolation("PROVIDER_CONTRACT_FORBIDDEN_SIDE_EFFECT", fmt.Sprintf("%s: provider contract packages must not read credentials, touch local files, or open network connections directly (%s)", slashRel, forbiddenCall))
				}
			}
		}
	}

	if len(c.violations) == 0 {
		fmt.Printf("provider contract governance is valid (%d facades)\n", len(providers))
	}
	return nil
}

func (c *checker) addViolation(ruleID, message string) {
	c.violations = append(c.violations, violation{ruleID: ruleID, message: message})
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

func stringList(values []any) []string {
	var out []string
	for _, value := range values {
		if text, ok := value.(string); ok {
			out = append(out, text)
		}
	}
	return out
}
