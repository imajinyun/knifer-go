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
	writeFile(t, root, "internal/httpx/http/bad_safe.go", `package http

func SafeButNotBoundary() {}
`)
	writeFile(t, root, "internal/provider/bad.go", `package provider

type Option func(*config)
type Provider interface{ Provide() }
type config struct {
	provider Provider
	lookup func(string) string
	name string
}

func (c *config) SetProvider(provider Provider) {}

func WithProvider(provider Provider) Option {
	return func(c *config) {
		c.provider = provider
	}
}

func WithLookup(lookup func(string) string) Option {
	return func(c *config) {
		c.lookup = lookup
	}
}

func WithProviderSetter(provider Provider) Option {
	return func(c *config) {
		c.SetProvider(provider)
	}
}

func WithName(name string) Option {
	return func(c *config) {
		c.name = name
	}
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
		"SafeButNotBoundary: Safe names are reserved for trust-boundary APIs",
		"WithProvider: nil provider/function option parameters must not overwrite existing providers",
		"WithLookup: nil provider/function option parameters must not overwrite existing providers",
		"WithProviderSetter: nil provider/function option parameters must not overwrite existing providers",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("violations =\n%s\nwant %q", got, want)
		}
	}
	if strings.Contains(got, "WithName: nil provider/function option parameters") {
		t.Fatalf("violations =\n%s\nWithName must not be treated as a provider/function option", got)
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
	writeFile(t, root, "internal/httpx/http/ok_safe.go", `package http

func GetSafe(rawURL string) any { return nil }
func DownloadFileSafeWithOptions(rawURL, dest string) (int64, error) { return 0, nil }
`)
	writeFile(t, root, "internal/provider/ok.go", `package provider

type Option func(*config)
type Provider interface{ Provide() }
type config struct {
	provider Provider
	lookup func(string) string
	clock func() int
}

func (c *config) SetProvider(provider Provider) {}

func WithProvider(provider Provider) Option {
	return func(c *config) {
		if provider != nil {
			c.provider = provider
		}
	}
}

func WithLookup(lookup func(string) string) Option {
	return func(c *config) {
		if nil != lookup {
			c.lookup = lookup
		}
	}
}

func WithClock(clock func() int) Option {
	return func(c *config) {
		if clock == nil {
			return
	}
		c.clock = clock
	}
}

func WithProviderSetter(provider Provider) Option {
	return func(c *config) {
		if provider != nil {
			c.SetProvider(provider)
		}
	}
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
