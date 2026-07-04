package namingcheck

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRepositoryNamingContracts(t *testing.T) {
	violations, err := Check(CheckConfig{Root: "../.."})
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if len(violations) > 0 {
		var b strings.Builder
		for _, violation := range violations {
			b.WriteString(violation.String())
			b.WriteByte('\n')
		}
		t.Fatalf("naming contract violations:\n%s", b.String())
	}
}

func TestNamingContractsCatchViolations(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "bad/bad.go", `package bad

func BrokenE() int { return 1 }
func MustBroken() int { return 1 }
func WithBroken() int { return 1 }
func LooksSafe() {}
`)
	writeFile(t, root, "internal/file/bad.go", `package file

import "os"

func BadWriteClose() error {
	f, err := os.OpenFile("out.txt", os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	return nil
}
`)
	writeFile(t, root, "bad/bad_test.go", `package bad

func TestSomething(t any) {}
`)

	violations, err := Check(CheckConfig{Root: root})
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	got := strings.Join(violationMessages(violations), "\n")
	for _, want := range []string{
		"BrokenE: functions ending in E must return error",
		"MustBroken: Must functions must panic directly or delegate to another Must function",
		"MustBroken: Must functions must have a test or example referencing the function",
		"WithBroken: With functions must return an option type",
		"LooksSafe: Safe names are reserved for trust-boundary APIs",
		"BadWriteClose: write-path Close errors must be handled",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("violations =\n%s\nwant %q", got, want)
		}
	}
}

func TestNamingContractsAllowCurrentPatterns(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "ok/ok.go", `package http

type RequestOption func(*Request)
type Request struct{}

func GetStringE() (string, error) { return "", nil }
func LogAtE(err error) {}
func MustParse() int { panic("bad") }
func WithHeader() RequestOption { return nil }
func GetSafe() *Request { return nil }
func (r *Request) WithBuilderStyle() *Request { return r }
`)
	writeFile(t, root, "internal/file/ok.go", `package file

import "os"

func GoodWriteClose() (err error) {
	f, err := os.OpenFile("out.txt", os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := f.Close(); err == nil {
			err = closeErr
		}
	}()
	return nil
}
`)
	writeFile(t, root, "ok/ok_test.go", `package http

func TestMustParse(t any) { MustParse() }
`)

	violations, err := Check(CheckConfig{Root: root})
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if len(violations) > 0 {
		t.Fatalf("violations = %v", violations)
	}
}

func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func violationMessages(violations []Violation) []string {
	out := make([]string, 0, len(violations))
	for _, violation := range violations {
		out = append(out, violation.Msg)
	}
	return out
}
