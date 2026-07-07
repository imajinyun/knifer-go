package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func assertJSONFinding(t *testing.T, output, ruleID string) {
	t.Helper()
	var result struct {
		Status   string `json:"status"`
		Findings []struct {
			RuleID string `json:"rule_id"`
		} `json:"findings"`
	}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, output)
	}
	if result.Status != "failed" {
		t.Fatalf("JSON status = %q, want failed\n%s", result.Status, output)
	}
	for _, finding := range result.Findings {
		if finding.RuleID == ruleID {
			return
		}
	}
	t.Fatalf("JSON findings missing rule id %q:\n%s", ruleID, output)
}

func TestArchImportsCheckAcceptsThinFacadeAndInternalDirection(t *testing.T) {
	fixture := newGovernanceFixture(t)
	fixture.WriteGoMod()
	fixture.WriteJSON("ai-context.json", map[string]any{
		"dependency_tiers": map[string]any{
			"heavy_dependency_allowlist": map[string]any{},
		},
	})
	fixture.WriteFile("vbad/doc.go", `// Package vbad exposes fixture helpers.
package vbad
`)
	fixture.WriteFile("vbad/bad.go", `package vbad

import "github.com/imajinyun/knifer-go/internal/bad"

func Run() string { return bad.Run() }
`)
	fixture.WriteFile("internal/bad/bad.go", `package bad

func Run() string { return "ok" }
`)

	output, err := fixture.RunArchImportsCheck()
	if err != nil {
		t.Fatalf("check_arch_imports.sh failed: %v\n%s", err, output)
	}
	if !strings.Contains(output, "architecture import governance passed") {
		t.Fatalf("arch imports output missing success:\n%s", output)
	}
}

func TestArchImportsCheckRejectsFacadeAndInternalDirectionViolations(t *testing.T) {
	fixture := newGovernanceFixture(t)
	fixture.WriteGoMod()
	fixture.WriteJSON("ai-context.json", map[string]any{
		"dependency_tiers": map[string]any{
			"heavy_dependency_allowlist": map[string]any{},
		},
	})
	fixture.WriteFile("vbad/doc.go", `// Package vbad exposes fixture helpers.
package vbad
`)
	fixture.WriteFile("vbad/bad.go", `package vbad

import (
	"github.com/imajinyun/knifer-go/internal/bad"
	"github.com/imajinyun/knifer-go/vother"
)

func Run() string { return bad.Run() + vother.Name() }
`)
	fixture.WriteFile("vother/doc.go", `// Package vother exposes another fixture facade.
package vother
`)
	fixture.WriteFile("vother/other.go", `package vother

import "github.com/imajinyun/knifer-go/internal/other"

func Name() string { return other.Name() }
`)
	fixture.WriteFile("internal/bad/bad.go", `package bad

import "github.com/imajinyun/knifer-go/vother"

func Run() string { return vother.Name() }
`)
	fixture.WriteFile("internal/other/other.go", `package other

func Name() string { return "other" }
`)

	output, err := fixture.RunArchImportsCheck()
	if err == nil {
		t.Fatalf("check_arch_imports.sh unexpectedly passed:\n%s", output)
	}
	for _, want := range []string{
		"ARCH_IMPORT_FACADE_TO_FACADE",
		"ARCH_IMPORT_INTERNAL_TO_FACADE",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("arch imports output missing rule id %q:\n%s", want, output)
		}
	}
	for _, want := range []string{
		"imports another public package",
		"imports public facade",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("arch imports output missing %q:\n%s", want, output)
		}
	}
}

func TestArchImportsCheckEmitsJSONFindings(t *testing.T) {
	fixture := newGovernanceFixture(t)
	fixture.WriteGoMod()
	fixture.WriteJSON("ai-context.json", map[string]any{
		"dependency_tiers": map[string]any{
			"heavy_dependency_allowlist": map[string]any{},
		},
	})
	fixture.WriteFile("vbad/doc.go", `// Package vbad exposes fixture helpers.
package vbad
`)
	fixture.WriteFile("vbad/bad.go", `package vbad

import "github.com/imajinyun/knifer-go/vother"

func Run() string { return vother.Name() }
`)
	fixture.WriteFile("vother/doc.go", `// Package vother exposes another fixture facade.
package vother
`)
	fixture.WriteFile("vother/other.go", `package vother

import "github.com/imajinyun/knifer-go/internal/other"

func Name() string { return other.Name() }
`)
	fixture.WriteFile("internal/other/other.go", `package other

func Name() string { return "other" }
`)

	output, err := fixture.RunArchImportsCheckJSON()
	if err == nil {
		t.Fatalf("archimportscheck -json unexpectedly passed:\n%s", output)
	}
	assertJSONFinding(t, output, "ARCH_IMPORT_FACADE_TO_FACADE")
}

func TestArchImportsCheckRejectsHeavyDependencyLeak(t *testing.T) {
	fixture := newGovernanceFixture(t)
	fixture.WriteFile("go.mod", `module github.com/imajinyun/knifer-go

go 1.25.0

require example.com/heavy v0.0.0

replace example.com/heavy => ./testheavy
`)
	fixture.WriteFile("testheavy/go.mod", "module example.com/heavy\n\ngo 1.25.0\n")
	fixture.WriteFile("testheavy/heavy.go", "package heavy\n\nfunc Use() {}\n")
	fixture.WriteJSON("ai-context.json", map[string]any{
		"dependency_tiers": map[string]any{
			"heavy_dependency_allowlist": map[string]any{
				"example.com/heavy": []string{"internal/allowed"},
			},
		},
	})
	fixture.WriteFile("vbad/doc.go", `// Package vbad exposes fixture helpers.
package vbad
`)
	fixture.WriteFile("vbad/bad.go", `package vbad

import "example.com/heavy"

func Run() { heavy.Use() }
`)

	output, err := fixture.RunArchImportsCheckJSON()
	if err == nil {
		t.Fatalf("archimportscheck -json unexpectedly passed:\n%s", output)
	}
	assertJSONFinding(t, output, "ARCH_IMPORT_HEAVY_DEPENDENCY_LEAK")
}

func TestPanicPolicyCheckAllowsMustAPIsAndRejectsProductionPanic(t *testing.T) {
	fixture := newGovernanceFixture(t)
	fixture.WriteGoMod()
	fixture.WriteFile("internal/bad/bad.go", `package bad

func MustRun() {
	panic("fixture")
}

func Run() {
	panic("fixture")
}
`)

	output, err := fixture.RunPanicPolicyCheck()
	if err == nil {
		t.Fatalf("check_panic_policy.sh unexpectedly passed:\n%s", output)
	}
	if !strings.Contains(output, "PANIC_POLICY_PRODUCTION_PANIC") {
		t.Fatalf("panic policy output missing PANIC_POLICY_PRODUCTION_PANIC rule id:\n%s", output)
	}
	if !strings.Contains(output, "production panic is not allowed") || !strings.Contains(output, "internal/bad/bad.go") {
		t.Fatalf("panic policy output missing production panic error:\n%s", output)
	}
}

func TestPanicPolicyCheckEmitsJSONFindings(t *testing.T) {
	fixture := newGovernanceFixture(t)
	fixture.WriteGoMod()
	fixture.WriteFile("internal/bad/bad.go", `package bad

func Run() {
	panic("fixture")
}
`)

	output, err := fixture.RunPanicPolicyCheckJSON()
	if err == nil {
		t.Fatalf("panicpolicycheck -json unexpectedly passed:\n%s", output)
	}
	assertJSONFinding(t, output, "PANIC_POLICY_PRODUCTION_PANIC")
}

func TestFacadeBoundaryCheckAcceptsThinFacade(t *testing.T) {
	fixture := newGovernanceFixture(t)
	fixture.WriteGoMod()
	fixture.WriteFile("doc.go", `// Package knifer provides fixture root docs.
package knifer
`)
	fixture.WriteFile("vbad/doc.go", `// Package vbad exposes fixture helpers.
package vbad
`)
	fixture.WriteFile("vbad/bad.go", `package vbad

import "github.com/imajinyun/knifer-go/internal/bad"

func Run() string { return bad.Run() }
`)
	fixture.WriteFile("internal/bad/bad.go", `package bad

func Run() string { return "ok" }
`)

	output, err := fixture.RunFacadeBoundaryCheck()
	if err != nil {
		t.Fatalf("check_facade_boundary.sh failed: %v\n%s", err, output)
	}
	if !strings.Contains(output, "facade boundary governance is valid") {
		t.Fatalf("facade boundary output missing success:\n%s", output)
	}
}

func TestFacadeBoundaryCheckRejectsFacadeControlFlow(t *testing.T) {
	fixture := newGovernanceFixture(t)
	fixture.WriteGoMod()
	fixture.WriteFile("vbad/doc.go", `// Package vbad exposes fixture helpers.
package vbad
`)
	fixture.WriteFile("vbad/bad.go", `package vbad

func Run(ok bool) string {
	if ok {
		return "ok"
	}
	return "bad"
}
`)

	output, err := fixture.RunFacadeBoundaryCheck()
	if err == nil {
		t.Fatalf("check_facade_boundary.sh unexpectedly passed:\n%s", output)
	}
	if !strings.Contains(output, "FACADE_BOUNDARY_THIN_FACADE_VIOLATION") {
		t.Fatalf("facade boundary output missing FACADE_BOUNDARY_THIN_FACADE_VIOLATION rule id:\n%s", output)
	}
	if !strings.Contains(output, "facade packages should not contain implementation control flow") {
		t.Fatalf("facade boundary output missing control-flow error:\n%s", output)
	}
}

func TestFacadeBoundaryCheckEmitsJSONFindings(t *testing.T) {
	fixture := newGovernanceFixture(t)
	fixture.WriteGoMod()
	fixture.WriteFile("vbad/doc.go", `// Package vbad exposes fixture helpers.
package vbad
`)
	fixture.WriteFile("vbad/bad.go", `package vbad

func Run(ok bool) string {
	if ok {
		return "ok"
	}
	return "bad"
}
`)

	output, err := fixture.RunFacadeBoundaryCheckJSON()
	if err == nil {
		t.Fatalf("facadeboundarycheck -json unexpectedly passed:\n%s", output)
	}
	assertJSONFinding(t, output, "FACADE_BOUNDARY_THIN_FACADE_VIOLATION")
}
