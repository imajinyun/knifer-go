package system

import (
	"errors"
	"runtime"
	"strings"
	"testing"
)

func TestGoInfo(t *testing.T) {
	g := GetGoInfo()
	if g == nil {
		t.Fatal("GoInfo 不应为 nil")
	}
	if g.GetVersion() != runtime.Version() {
		t.Errorf("Go Version 不一致: %s vs %s", g.GetVersion(), runtime.Version())
	}
	if g.GetCompiler() != runtime.Compiler {
		t.Errorf("Compiler 不一致")
	}
	if g.GetNumCPU() != runtime.NumCPU() {
		t.Errorf("NumCPU 不一致")
	}
	if !strings.Contains(g.String(), "Go Version:") {
		t.Errorf("GoInfo.String 缺少 caption")
	}
}

func TestGoInfoWithOptions(t *testing.T) {
	g := NewGoInfoWithOptions(
		WithGoVersionFunc(func() string { return "go-option" }),
		WithGoCompilerFunc(func() string { return "compiler-option" }),
		WithGoRootFunc(func() string { return "/go/root" }),
		WithGoOSFunc(func() string { return "plan9" }),
		WithGoArchFunc(func() string { return "wasm" }),
		WithGoNumCPUFunc(func() int { return 9 }),
		WithGoNumCgoCallFunc(func() int64 { return 10 }),
	)
	if g.GetVersion() != "go-option" || g.GetCompiler() != "compiler-option" || g.GetGOROOT() != "/go/root" || g.GetGOOS() != "plan9" || g.GetGOARCH() != "wasm" || g.GetNumCPU() != 9 || g.NumCgoCalls != 10 {
		t.Fatalf("NewGoInfoWithOptions = %#v", g)
	}
}

func TestSystemInfoGettersWithOptions(t *testing.T) {
	g := GetGoInfoWithOptions(WithGoVersionFunc(func() string { return "go-getter" }))
	if g.GetVersion() != "go-getter" {
		t.Fatalf("GetGoInfoWithOptions version = %q", g.GetVersion())
	}

	o := GetOsInfoWithOptions(WithOSNameFunc(func() string { return "linux" }))
	if o.GetName() != "linux" {
		t.Fatalf("GetOsInfoWithOptions name = %q", o.GetName())
	}
}

func TestGoInfoDefaultGOROOTProviderUsesCommandThenEnvFallback(t *testing.T) {
	t.Run("command output wins when non-empty", func(t *testing.T) {
		called := false
		g := NewGoInfoWithOptions(
			WithGoEnvOutputFunc(func(name string, args ...string) ([]byte, error) {
				called = true
				if name != "go" || strings.Join(args, " ") != "env GOROOT" {
					t.Fatalf("go env command = %s %v", name, args)
				}
				return []byte("/deterministic/go/root\n"), nil
			}),
			WithGoRootEnvLookupFunc(func(string) string { return "/env/go/root" }),
		)
		if !called || g.GetGOROOT() != "/deterministic/go/root" {
			t.Fatalf("GOROOT = %q called=%v", g.GetGOROOT(), called)
		}
	})

	t.Run("command error falls back to environment", func(t *testing.T) {
		g := NewGoInfoWithOptions(
			WithGoEnvOutputFunc(func(string, ...string) ([]byte, error) {
				return nil, errors.New("go command unavailable")
			}),
			WithGoRootEnvLookupFunc(func(key string) string {
				if key != "GOROOT" {
					t.Fatalf("env key = %q", key)
				}
				return "/fallback/go/root"
			}),
		)
		if g.GetGOROOT() != "/fallback/go/root" {
			t.Fatalf("GOROOT fallback = %q", g.GetGOROOT())
		}
	})
}

func TestGoInfoNilOptionsFallBackToRuntimeProviders(t *testing.T) {
	g := NewGoInfoWithOptions(
		nil,
		WithGoVersionFunc(nil),
		WithGoCompilerFunc(nil),
		WithGoRootFunc(nil),
		WithGoEnvOutputFunc(func(string, ...string) ([]byte, error) { return []byte("/nil/options/root"), nil }),
		WithGoRootEnvLookupFunc(nil),
		WithGoOSFunc(nil),
		WithGoArchFunc(nil),
		WithGoNumCPUFunc(nil),
		WithGoNumCgoCallFunc(nil),
	)
	if g.GetVersion() != runtime.Version() || g.GetCompiler() != runtime.Compiler || g.GetGOOS() != runtime.GOOS || g.GetGOARCH() != runtime.GOARCH || g.GetNumCPU() != runtime.NumCPU() {
		t.Fatalf("nil option fallback = %#v", g)
	}
	if g.GetGOROOT() != "/nil/options/root" {
		t.Fatalf("nil option GOROOT fallback = %q", g.GetGOROOT())
	}
}
