package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestGenerateToolsDocIncludesMachineReadableDetails(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "go.mod", "module github.com/imajinyun/go-knifer\n\ngo 1.25.0\n")
	writeTestFile(t, root, "vtool/doc.go", `// Package vtool exposes test facade helpers.
package vtool
`)
	writeTestFile(t, root, "vtool/tool.go", `package vtool

import (
	"context"

	impl "github.com/imajinyun/go-knifer/internal/toolimpl"
)

// Run executes the test tool.
func Run(ctx context.Context, name string, values ...int) (string, error) {
	return name, nil
}

func Double(v int) int { return impl.Double(v) }

func AddToCounter(c *impl.Counter, v int) int { return c.Add(v) }

// Hidden is deliberately unexported from the tool catalog.
func hidden() {}
`)
	writeTestFile(t, root, "vtool/tool_test.go", `package vtool

func ExampleRun() {
	_, _ = Run(nil, "demo", 1)
}

func ExampleRun_withValues() {
	_, _ = Run(nil, "demo", 1, 2)
}
`)
	writeTestFile(t, root, "internal/toolimpl/tool.go", `package toolimpl

// Double doubles v for fallback documentation.
func Double(v int) int { return v * 2 }

type Counter struct { Value int }

// Add adds v to the counter and returns the updated value.
func (c *Counter) Add(v int) int { c.Value += v; return c.Value }
`)
	writeTestFile(t, root, "internal/hidden/hidden.go", `package hidden

func Hidden() {}
`)
	writeTestFile(t, root, "notfacade/notfacade.go", `package notfacade

func Hidden() {}
`)

	doc, err := generateToolsDoc(root)
	if err != nil {
		t.Fatalf("generateToolsDoc() error = %v", err)
	}
	if doc.Schema != schemaVersion || doc.Module != modulePath {
		t.Fatalf("unexpected document identity: %#v", doc)
	}
	wantSummary := SummaryDoc{
		PackageCount:          1,
		FunctionCount:         3,
		FunctionsWithExamples: 1,
		ContextAwareFunctions: 1,
		ReturnsErrorFunctions: 1,
		VariadicFunctions:     1,
		SynopsisSources: map[string]int{
			"empty":    0,
			"facade":   1,
			"internal": 2,
		},
	}
	if !reflect.DeepEqual(doc.Summary, wantSummary) {
		t.Fatalf("summary = %#v, want %#v", doc.Summary, wantSummary)
	}
	if len(doc.Packages) != 1 {
		t.Fatalf("packages len = %d, want 1: %#v", len(doc.Packages), doc.Packages)
	}
	pkg := doc.Packages[0]
	if pkg.ImportPath != modulePath+"/vtool" || pkg.Name != "vtool" {
		t.Fatalf("unexpected package: %#v", pkg)
	}
	wantPackageSummary := PackageSummaryDoc{
		FunctionCount:          3,
		FunctionsWithExamples:  1,
		ExampleCoveragePercent: 33.3,
		SynopsisSources: map[string]int{
			"empty":    0,
			"facade":   1,
			"internal": 2,
		},
	}
	if !reflect.DeepEqual(pkg.Summary, wantPackageSummary) {
		t.Fatalf("package summary = %#v, want %#v", pkg.Summary, wantPackageSummary)
	}
	if !strings.Contains(pkg.Synopsis, "Package vtool exposes test facade helpers.") {
		t.Fatalf("package synopsis = %q", pkg.Synopsis)
	}
	if len(pkg.Functions) != 3 {
		t.Fatalf("functions len = %d, want 3: %#v", len(pkg.Functions), pkg.Functions)
	}
	fns := map[string]FuncDoc{}
	for _, fn := range pkg.Functions {
		fns[fn.Name] = fn
	}
	fn := fns["Run"]
	if fn.Name != "Run" {
		t.Fatalf("function name = %q", fn.Name)
	}
	if fn.Signature != "func Run(ctx context.Context, name string, values ...int) (string, error)" {
		t.Fatalf("signature = %q", fn.Signature)
	}
	if fn.Synopsis != "Run executes the test tool." {
		t.Fatalf("synopsis = %q", fn.Synopsis)
	}
	if fn.SynopsisSource != "facade" {
		t.Fatalf("synopsis source = %q, want facade", fn.SynopsisSource)
	}
	if !fn.ReturnsError || !fn.ContextAware || !fn.Variadic {
		t.Fatalf("flags = returns_error:%v context_aware:%v variadic:%v", fn.ReturnsError, fn.ContextAware, fn.Variadic)
	}
	if got := len(fn.Params); got != 3 {
		t.Fatalf("params len = %d, want 3", got)
	}
	if fn.Params[0] != (Param{Name: "ctx", Type: "context.Context"}) || fn.Params[2] != (Param{Name: "values", Type: "...int"}) {
		t.Fatalf("unexpected params: %#v", fn.Params)
	}
	if strings.Join(fn.Results, ",") != "string,error" {
		t.Fatalf("results = %#v", fn.Results)
	}
	if strings.Join(fn.Examples, ",") != "ExampleRun,ExampleRun_withValues" {
		t.Fatalf("examples = %#v", fn.Examples)
	}

	fallback := fns["Double"]
	if fallback.Synopsis != "Double doubles v for fallback documentation." {
		t.Fatalf("fallback synopsis = %q", fallback.Synopsis)
	}
	if fallback.SynopsisSource != "internal" {
		t.Fatalf("fallback synopsis source = %q, want internal", fallback.SynopsisSource)
	}

	methodFallback := fns["AddToCounter"]
	if methodFallback.Synopsis != "Add adds v to the counter and returns the updated value." {
		t.Fatalf("method fallback synopsis = %q", methodFallback.Synopsis)
	}
	if methodFallback.SynopsisSource != "internal" {
		t.Fatalf("method fallback synopsis source = %q, want internal", methodFallback.SynopsisSource)
	}
}

func TestToolsCatalogSnapshotIsCurrent(t *testing.T) {
	root := repositoryRoot(t)
	doc, err := generateToolsDoc(root)
	if err != nil {
		t.Fatalf("generateToolsDoc(%q) error = %v", root, err)
	}
	current, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent() error = %v", err)
	}
	current = append(current, '\n')

	snapshotPath := filepath.Join(root, "docs", "api", "tools.json")
	snapshot, err := os.ReadFile(snapshotPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", snapshotPath, err)
	}
	if string(snapshot) != string(current) {
		t.Fatalf("docs/api/tools.json is stale; run make tools-gen after intentional facade/doc/example changes")
	}
}

func TestToolsMarkdownSnapshotIsCurrent(t *testing.T) {
	root := repositoryRoot(t)
	doc, err := generateToolsDoc(root)
	if err != nil {
		t.Fatalf("generateToolsDoc(%q) error = %v", root, err)
	}
	current := renderToolsMarkdown(doc)

	snapshotPath := filepath.Join(root, "docs", "api", "tools.md")
	snapshot, err := os.ReadFile(snapshotPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", snapshotPath, err)
	}
	if string(snapshot) != string(current) {
		t.Fatalf("docs/api/tools.md is stale; run make docs-gen after intentional facade/doc/example changes")
	}
}

func TestToolsCatalogSynopsisCoverageBudget(t *testing.T) {
	root := repositoryRoot(t)
	doc, err := generateToolsDoc(root)
	if err != nil {
		t.Fatalf("generateToolsDoc(%q) error = %v", root, err)
	}
	const maxEmptySynopses = 0
	got := doc.Summary.SynopsisSources["empty"]
	if got > maxEmptySynopses {
		t.Fatalf("empty synopsis count = %d, want <= %d\n%s", got, maxEmptySynopses, renderToolsQualityReport(doc))
	}
}

func TestToolsCatalogExamplesCoverageBudget(t *testing.T) {
	root := repositoryRoot(t)
	doc, err := generateToolsDoc(root)
	if err != nil {
		t.Fatalf("generateToolsDoc(%q) error = %v", root, err)
	}
	const minFunctionsWithExamples = 270
	got := doc.Summary.FunctionsWithExamples
	if got < minFunctionsWithExamples {
		t.Fatalf("functions with examples = %d, want >= %d\n%s", got, minFunctionsWithExamples, renderToolsQualityReport(doc))
	}
}

func TestToolsCatalogPerPackageExamplesBudget(t *testing.T) {
	root := repositoryRoot(t)
	doc, err := generateToolsDoc(root)
	if err != nil {
		t.Fatalf("generateToolsDoc(%q) error = %v", root, err)
	}
	const maxMinExamplesPerPackage = 5
	examplesByPackage := toolsExamplesByPackage(doc)
	missing := []string{}
	for _, pkg := range doc.Packages {
		target := min(maxMinExamplesPerPackage, len(pkg.Functions))
		if got := examplesByPackage[pkg.Name]; got < target {
			missing = append(missing, pkg.Name)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("packages below per-package example budget: %s\n%s", strings.Join(missing, ", "), renderToolsQualityReport(doc))
	}
}

func TestToolsCatalogSecuritySensitiveExamplesBudget(t *testing.T) {
	root := repositoryRoot(t)
	doc, err := generateToolsDoc(root)
	if err != nil {
		t.Fatalf("generateToolsDoc(%q) error = %v", root, err)
	}
	const minExamplesPerPackage = 5
	securitySensitivePackages := loadSecuritySensitivePackages(t, root)
	examplesByPackage := toolsExamplesByPackage(doc)
	missing := []string{}
	for _, name := range securitySensitivePackages {
		if got := examplesByPackage[name]; got < minExamplesPerPackage {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("security-sensitive packages below example budget %d: %s\n%s", minExamplesPerPackage, strings.Join(missing, ", "), renderToolsQualityReport(doc))
	}
}

func TestToolsCatalogPackageSummariesMatchFunctions(t *testing.T) {
	root := repositoryRoot(t)
	doc, err := generateToolsDoc(root)
	if err != nil {
		t.Fatalf("generateToolsDoc(%q) error = %v", root, err)
	}
	for _, pkg := range doc.Packages {
		want := summarizePackageDoc(pkg.Functions)
		if !reflect.DeepEqual(pkg.Summary, want) {
			t.Fatalf("%s summary = %#v, want %#v", pkg.Name, pkg.Summary, want)
		}
	}
}

func TestSummarizePackageDocRoundsExampleCoverage(t *testing.T) {
	summary := summarizePackageDoc([]FuncDoc{
		{Name: "One", Examples: []string{"ExampleOne"}, SynopsisSource: "facade"},
		{Name: "Two", SynopsisSource: "internal"},
		{Name: "Three"},
	})
	want := PackageSummaryDoc{
		FunctionCount:          3,
		FunctionsWithExamples:  1,
		ExampleCoveragePercent: 33.3,
		SynopsisSources: map[string]int{
			"empty":    1,
			"facade":   1,
			"internal": 1,
		},
	}
	if !reflect.DeepEqual(summary, want) {
		t.Fatalf("summary = %#v, want %#v", summary, want)
	}
}

func toolsExamplesByPackage(doc ToolsDoc) map[string]int {
	examplesByPackage := map[string]int{}
	for _, pkg := range doc.Packages {
		for _, fn := range pkg.Functions {
			if len(fn.Examples) > 0 {
				examplesByPackage[pkg.Name]++
			}
		}
	}
	return examplesByPackage
}

func TestWriteToolsDocWritesIndentedFile(t *testing.T) {
	outPath := filepath.Join(t.TempDir(), "tools.json")
	doc := ToolsDoc{
		Schema: schemaVersion,
		Module: modulePath,
		Summary: SummaryDoc{
			PackageCount:    0,
			SynopsisSources: map[string]int{"empty": 0, "facade": 0, "internal": 0},
		},
	}

	if err := writeToolsDoc(doc, outPath); err != nil {
		t.Fatalf("writeToolsDoc() error = %v", err)
	}
	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", outPath, err)
	}
	if !strings.HasSuffix(string(got), "\n") {
		t.Fatalf("generated file does not end with newline: %q", got)
	}
	if !strings.Contains(string(got), `"summary": {`) {
		t.Fatalf("generated file missing summary object: %s", got)
	}
}

func TestRenderToolsMarkdownIncludesSummaryAndPackages(t *testing.T) {
	doc := ToolsDoc{
		Schema: schemaVersion,
		Module: modulePath,
		Summary: SummaryDoc{
			PackageCount:          1,
			FunctionCount:         2,
			FunctionsWithExamples: 1,
			ContextAwareFunctions: 1,
			ReturnsErrorFunctions: 1,
			VariadicFunctions:     1,
			SynopsisSources:       map[string]int{"empty": 1, "facade": 1, "internal": 0},
		},
		Packages: []PackageDoc{
			{
				ImportPath: modulePath + "/vtool",
				Name:       "vtool",
				Synopsis:   "Package vtool exposes | test helpers.",
				Summary: PackageSummaryDoc{
					FunctionCount:          2,
					FunctionsWithExamples:  1,
					ExampleCoveragePercent: 50,
					SynopsisSources:        map[string]int{"empty": 1, "facade": 1, "internal": 0},
				},
				Functions: []FuncDoc{
					{
						Name:           "Run",
						Signature:      "func Run(name string) (string, error)",
						Synopsis:       "Run executes | the test helper.",
						SynopsisSource: "facade",
						ReturnsError:   true,
						Examples:       []string{"ExampleRun"},
					},
					{
						Name:      "HiddenDoc",
						Signature: "func HiddenDoc()",
					},
				},
			},
		},
	}

	got := string(renderToolsMarkdown(doc))
	wants := []string{
		"# go-knifer Machine-readable Tool Catalog\n",
		"| Schema | " + schemaVersion + " |",
		"| Module | `" + modulePath + "` |",
		"| Synopsis source: empty | 1 |",
		"### vtool",
		"Import path: `" + modulePath + "/vtool`",
		"Package vtool exposes | test helpers.",
		"Quality: 2 functions · 1 with examples · 50.0% example coverage · synopsis sources: facade=1, internal=0, empty=1",
		"| `Run` | `func Run(name string) (string, error)` | Run executes \\| the test helper. | facade | `ExampleRun` |",
		"| `HiddenDoc` | `func HiddenDoc()` | — | empty | — |",
	}
	for _, want := range wants {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered markdown missing %q:\n%s", want, got)
		}
	}
}

func TestRenderToolsQualityReportRanksEmptySynopsisPackages(t *testing.T) {
	doc := ToolsDoc{
		Summary: SummaryDoc{
			PackageCount:          2,
			FunctionCount:         5,
			FunctionsWithExamples: 1,
			SynopsisSources:       map[string]int{"empty": 3, "facade": 2, "internal": 0},
		},
		Packages: []PackageDoc{
			{
				Name: "vid",
				Functions: []FuncDoc{
					{Name: "Create", Synopsis: "Create returns an ID."},
					{Name: "FastUUID"},
				},
			},
			{
				Name: "vnet",
				Functions: []FuncDoc{
					{Name: "GetLocalHostName", Examples: []string{"ExampleGetLocalHostName"}},
					{Name: "GetLocalhost"},
					{Name: "ParseIP", Synopsis: "ParseIP parses an IP address."},
				},
			},
		},
	}

	got := string(renderToolsQualityReport(doc))
	wants := []string{
		"# go-knifer Tool Catalog Quality Report\n",
		"| Empty synopses | 3 |",
		"| Package | Functions | Empty synopses | With docs | With examples | Empty functions |",
		"| `vnet` | 3 | 2 | 1 | 1 | `GetLocalHostName`, `GetLocalhost` |",
		"| `vid` | 2 | 1 | 1 | 0 | `FastUUID` |",
	}
	for _, want := range wants {
		if !strings.Contains(got, want) {
			t.Fatalf("quality report missing %q:\n%s", want, got)
		}
	}
	if strings.Index(got, "| `vnet` |") > strings.Index(got, "| `vid` |") {
		t.Fatalf("quality report did not rank packages by empty synopsis count:\n%s", got)
	}
}

func TestWriteMarkdownDocWritesOnlyWhenPathProvided(t *testing.T) {
	doc := ToolsDoc{Schema: schemaVersion, Module: modulePath}
	if err := writeMarkdownDoc(doc, ""); err != nil {
		t.Fatalf("writeMarkdownDoc empty path error = %v", err)
	}
	outPath := filepath.Join(t.TempDir(), "tools.md")
	if err := writeMarkdownDoc(doc, outPath); err != nil {
		t.Fatalf("writeMarkdownDoc() error = %v", err)
	}
	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", outPath, err)
	}
	if !strings.Contains(string(got), "# go-knifer Machine-readable Tool Catalog") {
		t.Fatalf("generated markdown missing heading: %s", got)
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	return filepath.Clean(filepath.Join(wd, "..", ".."))
}

func loadSecuritySensitivePackages(t *testing.T, root string) []string {
	t.Helper()

	type aiContext struct {
		SecuritySensitivePackages []string `json:"security_sensitive_packages"`
	}

	path := filepath.Join(root, "ai-context.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}

	var context aiContext
	if err := json.Unmarshal(data, &context); err != nil {
		t.Fatalf("Unmarshal(%q) error = %v", path, err)
	}
	if len(context.SecuritySensitivePackages) == 0 {
		t.Fatalf("%s has no security_sensitive_packages entries", path)
	}

	return context.SecuritySensitivePackages
}

func writeTestFile(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}
