// Package toolchainpolicy validates the repository's declared Go and lint tools.
package toolchainpolicy

import (
	"encoding/json"
	"fmt"
	"go/version"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"golang.org/x/mod/modfile"
)

// Policy is the canonical tool version configuration in ai-context.json.
type Policy struct {
	Minimum string   `json:"go_minimum"`
	Release string   `json:"go_release"`
	Test    []string `json:"go_test"`
	Lint    string   `json:"golangci_lint"`
}

var patchVersion = regexp.MustCompile(`^1\.[0-9]+\.[0-9]+$`)

// Load reads the policy and checks it against the actual module language level.
func Load(root string) (Policy, error) {
	data, err := os.ReadFile(filepath.Join(root, "ai-context.json"))
	if err != nil {
		return Policy{}, err
	}
	var config struct {
		Project struct {
			GoVersion string `json:"go_version"`
		} `json:"project"`
		CI struct {
			Tools Policy `json:"tool_versions"`
		} `json:"ci_workflows"`
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return Policy{}, err
	}
	data, err = os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return Policy{}, err
	}
	module, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		return Policy{}, err
	}
	if module.Go == nil {
		return Policy{}, fmt.Errorf("go.mod must declare a Go language version")
	}
	p := config.CI.Tools
	if err := p.Validate(module.Go.Version); err != nil {
		return Policy{}, err
	}
	if want := ">=" + strings.TrimPrefix(version.Lang("go"+p.Minimum), "go"); config.Project.GoVersion != want {
		return Policy{}, fmt.Errorf("project.go_version must be %s", want)
	}
	return p, nil
}

// Validate checks fixed versions, minimum coverage and the tested release pin.
func (p Policy) Validate(moduleGo string) error {
	if !patchVersion.MatchString(p.Minimum) || p.Minimum != moduleGo {
		return fmt.Errorf("go_minimum %q must match go.mod %q", p.Minimum, moduleGo)
	}
	if !patchVersion.MatchString(p.Release) {
		return fmt.Errorf("go_release must pin a stable patch version")
	}
	if !regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`).MatchString(p.Lint) {
		return fmt.Errorf("golangci_lint must pin a stable version")
	}
	if len(p.Test) < 2 {
		return fmt.Errorf("go_test must cover the minimum and a newer Go line")
	}
	minimumLang := version.Lang("go" + p.Minimum)
	previous := ""
	for _, v := range p.Test {
		lang := version.Lang("go" + v)
		if !patchVersion.MatchString(v) || version.Compare("go"+v, "go"+p.Minimum) < 0 {
			return fmt.Errorf("go_test version %q must be a fixed patch at or above %s", v, p.Minimum)
		}
		if previous != "" && version.Compare(lang, previous) <= 0 {
			return fmt.Errorf("go_test must contain distinct, ascending Go lines")
		}
		previous = lang
	}
	if version.Lang("go"+p.Test[0]) != minimumLang {
		return fmt.Errorf("go_test must include the minimum Go line %s", minimumLang)
	}
	if !slices.Contains(p.Test, p.Release) {
		return fmt.Errorf("go_release %s must be one of the tested pins", p.Release)
	}
	return nil
}

// ValidateDependency prevents a dependency from silently raising the minimum.
func (p Policy) ValidateDependency(path, required string) error {
	if required == "" {
		return nil
	}
	if !version.IsValid("go"+required) || version.Compare("go"+required, "go"+p.Minimum) > 0 {
		return fmt.Errorf("dependency %s requires Go %s above declared minimum %s", path, required, p.Minimum)
	}
	return nil
}

// ValidateRuntime rejects implicit toolchain selection during validation.
func (p Policy) ValidateRuntime(actual, mode string) error {
	if mode != "local" {
		return fmt.Errorf("validation requires GOTOOLCHAIN=local, got %q", mode)
	}
	if !version.IsValid(actual) || version.Compare(actual, "go"+p.Minimum) < 0 {
		return fmt.Errorf("actual toolchain %q does not support Go %s", actual, p.Minimum)
	}
	return nil
}

// ValidateLinter checks both the pinned tool version and its compiler support.
func (p Policy) ValidateLinter(output, actualGo string) error {
	m := regexp.MustCompile(`version v?([0-9]+\.[0-9]+\.[0-9]+) built with (go[0-9.]+)`).FindStringSubmatch(output)
	if len(m) != 3 || "v"+m[1] != p.Lint {
		return fmt.Errorf("expected golangci-lint %s with identifiable build Go version, got %q", p.Lint, strings.TrimSpace(output))
	}
	if !version.IsValid(m[2]) || version.Compare(version.Lang(m[2]), version.Lang(actualGo)) < 0 {
		return fmt.Errorf("linter built with %s cannot analyze using %s", m[2], actualGo)
	}
	return nil
}
