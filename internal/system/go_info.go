package system

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type goInfoConfig struct {
	version     func() string
	compiler    func() string
	goRoot      func() string
	goos        func() string
	goarch      func() string
	numCPU      func() int
	numCgoCalls func() int64
}

// GoInfoOption customizes Go runtime metadata collection per call.
type GoInfoOption func(*goInfoConfig)

// WithGoVersionFunc sets the function used to collect the Go version.
func WithGoVersionFunc(fn func() string) GoInfoOption {
	return func(c *goInfoConfig) {
		if fn != nil {
			c.version = fn
		}
	}
}

// WithGoCompilerFunc sets the function used to collect the Go compiler name.
func WithGoCompilerFunc(fn func() string) GoInfoOption {
	return func(c *goInfoConfig) {
		if fn != nil {
			c.compiler = fn
		}
	}
}

// WithGoRootFunc sets the function used to collect GOROOT.
func WithGoRootFunc(fn func() string) GoInfoOption {
	return func(c *goInfoConfig) {
		if fn != nil {
			c.goRoot = fn
		}
	}
}

// WithGoOSFunc sets the function used to collect GOOS.
func WithGoOSFunc(fn func() string) GoInfoOption {
	return func(c *goInfoConfig) {
		if fn != nil {
			c.goos = fn
		}
	}
}

// WithGoArchFunc sets the function used to collect GOARCH.
func WithGoArchFunc(fn func() string) GoInfoOption {
	return func(c *goInfoConfig) {
		if fn != nil {
			c.goarch = fn
		}
	}
}

// WithGoNumCPUFunc sets the function used to collect the CPU count.
func WithGoNumCPUFunc(fn func() int) GoInfoOption {
	return func(c *goInfoConfig) {
		if fn != nil {
			c.numCPU = fn
		}
	}
}

// WithGoNumCgoCallFunc sets the function used to collect the cgo call count.
func WithGoNumCgoCallFunc(fn func() int64) GoInfoOption {
	return func(c *goInfoConfig) {
		if fn != nil {
			c.numCgoCalls = fn
		}
	}
}

func applyGoInfoOptions(opts []GoInfoOption) goInfoConfig {
	cfg := goInfoConfig{
		version:     runtime.Version,
		compiler:    func() string { return runtime.Compiler },
		goRoot:      goRoot,
		goos:        func() string { return runtime.GOOS },
		goarch:      func() string { return runtime.GOARCH },
		numCPU:      runtime.NumCPU,
		numCgoCalls: runtime.NumCgoCall,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.version == nil {
		cfg.version = runtime.Version
	}
	if cfg.compiler == nil {
		cfg.compiler = func() string { return runtime.Compiler }
	}
	if cfg.goRoot == nil {
		cfg.goRoot = goRoot
	}
	if cfg.goos == nil {
		cfg.goos = func() string { return runtime.GOOS }
	}
	if cfg.goarch == nil {
		cfg.goarch = func() string { return runtime.GOARCH }
	}
	if cfg.numCPU == nil {
		cfg.numCPU = runtime.NumCPU
	}
	if cfg.numCgoCalls == nil {
		cfg.numCgoCalls = runtime.NumCgoCall
	}
	return cfg
}

// GoInfo describes Go runtime and compilation metadata.
// It aggregates the metadata that would otherwise be split across Java/JVM info types.
type GoInfo struct {
	Version     string // For example, go1.22.3.
	Compiler    string // gc / gccgo
	GOROOT      string
	GOOS        string
	GOARCH      string
	NumCPU      int
	NumCgoCalls int64
}

// NewGoInfo creates GoInfo.
func NewGoInfo() *GoInfo {
	return NewGoInfoWithOptions()
}

// NewGoInfoWithOptions creates GoInfo using custom metadata providers.
func NewGoInfoWithOptions(opts ...GoInfoOption) *GoInfo {
	cfg := applyGoInfoOptions(opts)
	return &GoInfo{
		Version:     cfg.version(),
		Compiler:    cfg.compiler(),
		GOROOT:      cfg.goRoot(),
		GOOS:        cfg.goos(),
		GOARCH:      cfg.goarch(),
		NumCPU:      cfg.numCPU(),
		NumCgoCalls: cfg.numCgoCalls(),
	}
}

// GetVersion returns the Go version.
func (g *GoInfo) GetVersion() string { return g.Version }

// GetCompiler returns the compiler name.
func (g *GoInfo) GetCompiler() string { return g.Compiler }

// GetGOROOT returns the GOROOT path.
func (g *GoInfo) GetGOROOT() string { return g.GOROOT }

// GetGOOS returns the target OS.
func (g *GoInfo) GetGOOS() string { return g.GOOS }

// GetGOARCH returns the target architecture.
func (g *GoInfo) GetGOARCH() string { return g.GOARCH }

// GetNumCPU returns the number of available CPUs.
func (g *GoInfo) GetNumCPU() int { return g.NumCPU }

// String implements fmt.Stringer.
func (g *GoInfo) String() string {
	var b strings.Builder
	appendLine(&b, "Go Version:    ", g.Version)
	appendLine(&b, "Go Compiler:   ", g.Compiler)
	appendLine(&b, "GOROOT:        ", g.GOROOT)
	appendLine(&b, "GOOS:          ", g.GOOS)
	appendLine(&b, "GOARCH:        ", g.GOARCH)
	appendLine(&b, "NumCPU:        ", g.NumCPU)
	appendLine(&b, "NumCgoCall:    ", g.NumCgoCalls)
	return b.String()
}

func goRoot() string {
	out, err := exec.Command("go", "env", "GOROOT").Output()
	if err == nil {
		if root := strings.TrimSpace(string(out)); root != "" {
			return root
		}
	}
	return os.Getenv("GOROOT")
}
