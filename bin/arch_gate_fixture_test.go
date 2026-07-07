package main

import (
	"strings"
	"testing"
)

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
