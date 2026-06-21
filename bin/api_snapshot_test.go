package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateSnapshotIncludesCompatibilityDetails(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "go.mod", "module github.com/imajinyun/go-knifer\n\ngo 1.25.0\n")
	writeTestFile(t, root, "vcompat/compat.go", `package vcompat

const DefaultName = "demo"

var DefaultResult Result

type Alias = Result

type Result struct {
	Name   string
	Count  int
	hidden bool
}

func Build(input string) (Result, error) {
	return Result{Name: input}, nil
}

func (r Result) String() string { return r.Name }

func (r *Result) SetName(name string) { r.Name = name }

type Validator interface {
	Validate(Result) error
}
`)
	writeTestFile(t, root, "internal/hidden/hidden.go", `package hidden

func Hidden() {}
`)
	writeTestFile(t, root, "notfacade/notfacade.go", `package notfacade

func Hidden() {}
`)

	lines, err := generateSnapshot(root)
	if err != nil {
		t.Fatalf("generateSnapshot() error = %v", err)
	}
	snapshot := strings.Join(lines, "\n")
	for _, want := range []string{
		"github.com/imajinyun/go-knifer/vcompat",
		`const DefaultName untyped string = "demo"`,
		"var DefaultResult Result",
		"func Build(input string) (Result, error)",
		"type Alias = Result",
		"type Result struct{ Count int; Name string }",
		"method (*Result) SetName(name string)",
		"method (Result) String() string",
		"type Validator interface{ Validate(Result) error }",
	} {
		if !strings.Contains(snapshot, want) {
			t.Fatalf("snapshot missing %q:\n%s", want, snapshot)
		}
	}
	for _, unwanted := range []string{"hidden bool", "internal/hidden", "notfacade"} {
		if strings.Contains(snapshot, unwanted) {
			t.Fatalf("snapshot unexpectedly contains %q:\n%s", unwanted, snapshot)
		}
	}
}

func TestGenerateSnapshotIncludesGenericAndEmbeddedAPIDetails(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "go.mod", "module github.com/imajinyun/go-knifer\n\ngo 1.25.0\n")
	writeTestFile(t, root, "vcompat/compat.go", `package vcompat

import "time"

const Enabled = true

type private struct {
	Value string
}

type Number interface {
	~int | ~int64
}

type DurationMap = map[string]time.Duration

type Box[T any] struct {
	Item T
	time.Time
	private
	hidden string
}

type Handler[T Number] interface {
	Handle(Box[T], ...time.Duration) (private, error)
	reset()
}

func Transform[T Number](box Box[T], labels ...string) (out Box[T], err error) {
	return box, nil
}

func (b Box[T]) Value() T { return b.Item }

func (b *Box[T]) Set(value T) *Box[T] {
	b.Item = value
	return b
}
`)

	lines, err := generateSnapshot(root)
	if err != nil {
		t.Fatalf("generateSnapshot() error = %v", err)
	}
	snapshot := strings.Join(lines, "\n")
	for _, want := range []string{
		"const Enabled untyped bool = true",
		"func Transform[T Number](box Box[T], labels ...string) (out Box[T], err error)",
		"method (*Box[T]) Set(value T) *Box[T]",
		"method (Box[T]) Value() T",
		"type Box[T any] struct{ Item T; time.Time time.Time }",
		"type DurationMap = map[string]time.Duration",
		"type Handler[T Number] interface{ Handle(Box[T], ...time.Duration) (private, error) }",
		"type Number interface{}",
	} {
		if !strings.Contains(snapshot, want) {
			t.Fatalf("snapshot missing %q:\n%s", want, snapshot)
		}
	}
	for _, unwanted := range []string{"hidden string", "reset()", "type private"} {
		if strings.Contains(snapshot, unwanted) {
			t.Fatalf("snapshot unexpectedly contains %q:\n%s", unwanted, snapshot)
		}
	}
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

func TestCheckReleaseNotesAcceptsVersionedRelease(t *testing.T) {
	changelog := writeGovernanceFixture(t, `# Changelog

## Unreleased

## 1.2.3

### Governance

- Added a release-note gate.
`)

	output, err := runReleaseNotesCheck(t, changelog, "1.2.3")
	if err != nil {
		t.Fatalf("check_release_notes.sh failed: %v\n%s", err, output)
	}
	if !strings.Contains(output, "release notes are valid for 1.2.3") {
		t.Fatalf("unexpected output:\n%s", output)
	}
}

func TestCheckReleaseNotesRejectsUnmovedUnreleasedEntries(t *testing.T) {
	changelog := writeGovernanceFixture(t, `# Changelog

## Unreleased

### Governance

- This still needs to be moved.

## 1.2.3

### Governance

- Added a release-note gate.
`)

	output, err := runReleaseNotesCheck(t, changelog, "1.2.3")
	if err == nil {
		t.Fatalf("check_release_notes.sh unexpectedly succeeded:\n%s", output)
	}
	if !strings.Contains(output, "still has entries under Unreleased") {
		t.Fatalf("expected unmoved Unreleased error, got:\n%s", output)
	}
}

func TestCheckReleaseNotesAcceptsStructureWithoutReleaseVersion(t *testing.T) {
	changelog := writeGovernanceFixture(t, `# Changelog

## Unreleased

### Governance

- Work in progress is allowed before a release version is selected.
`)

	output, err := runReleaseNotesCheck(t, changelog, "")
	if err != nil {
		t.Fatalf("check_release_notes.sh failed: %v\n%s", err, output)
	}
	if !strings.Contains(output, "release notes structure is valid") {
		t.Fatalf("unexpected output:\n%s", output)
	}
}

func runReleaseNotesCheck(t *testing.T, changelog, version string) (string, error) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	root := filepath.Dir(wd)
	args := []string{filepath.Join(root, "bin/check_release_notes.sh")}
	if version != "" {
		args = append(args, version)
	}
	cmd := exec.Command("bash", args...)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "CHANGELOG_FILE="+changelog)
	combined, err := cmd.CombinedOutput()
	return string(combined), err
}

func writeGovernanceFixture(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "CHANGELOG.md")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return path
}
