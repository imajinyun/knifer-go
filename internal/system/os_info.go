package system

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type osInfoConfig struct {
	name          func() string
	arch          func() string
	version       func() string
	fileSeparator func() string
	lineSeparator func() string
	pathSeparator func() string
}

// OsInfoOption customizes OS information collection per call.
type OsInfoOption func(*osInfoConfig)

// WithOSNameFunc sets the function used to collect the OS name.
func WithOSNameFunc(fn func() string) OsInfoOption {
	return func(c *osInfoConfig) {
		if fn != nil {
			c.name = fn
		}
	}
}

// WithOSArchFunc sets the function used to collect the OS architecture.
func WithOSArchFunc(fn func() string) OsInfoOption {
	return func(c *osInfoConfig) {
		if fn != nil {
			c.arch = fn
		}
	}
}

// WithOSVersionFunc sets the function used to collect the OS version.
func WithOSVersionFunc(fn func() string) OsInfoOption {
	return func(c *osInfoConfig) {
		if fn != nil {
			c.version = fn
		}
	}
}

// WithOSFileSeparatorFunc sets the function used to collect the file separator.
func WithOSFileSeparatorFunc(fn func() string) OsInfoOption {
	return func(c *osInfoConfig) {
		if fn != nil {
			c.fileSeparator = fn
		}
	}
}

// WithOSLineSeparatorFunc sets the function used to collect the line separator.
func WithOSLineSeparatorFunc(fn func() string) OsInfoOption {
	return func(c *osInfoConfig) {
		if fn != nil {
			c.lineSeparator = fn
		}
	}
}

// WithOSPathSeparatorFunc sets the function used to collect the path-list separator.
func WithOSPathSeparatorFunc(fn func() string) OsInfoOption {
	return func(c *osInfoConfig) {
		if fn != nil {
			c.pathSeparator = fn
		}
	}
}

func applyOsInfoOptions(opts []OsInfoOption) osInfoConfig {
	cfg := osInfoConfig{
		name:          func() string { return runtime.GOOS },
		arch:          func() string { return runtime.GOARCH },
		version:       readOsVersion,
		fileSeparator: func() string { return string(filepath.Separator) },
		lineSeparator: lineSeparator,
		pathSeparator: func() string { return string(os.PathListSeparator) },
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.name == nil {
		cfg.name = func() string { return runtime.GOOS }
	}
	if cfg.arch == nil {
		cfg.arch = func() string { return runtime.GOARCH }
	}
	if cfg.version == nil {
		cfg.version = readOsVersion
	}
	if cfg.fileSeparator == nil {
		cfg.fileSeparator = func() string { return string(filepath.Separator) }
	}
	if cfg.lineSeparator == nil {
		cfg.lineSeparator = lineSeparator
	}
	if cfg.pathSeparator == nil {
		cfg.pathSeparator = func() string { return string(os.PathListSeparator) }
	}
	return cfg
}

// OsInfo describes current operating system information.
type OsInfo struct {
	Name          string
	Arch          string
	Version       string
	FileSeparator string
	LineSeparator string
	PathSeparator string
}

// NewOsInfo creates current OS information.
func NewOsInfo() *OsInfo {
	return NewOsInfoWithOptions()
}

// NewOsInfoWithOptions creates OS information using custom providers.
func NewOsInfoWithOptions(opts ...OsInfoOption) *OsInfo {
	cfg := applyOsInfoOptions(opts)
	return &OsInfo{
		Name:          cfg.name(),
		Arch:          cfg.arch(),
		Version:       cfg.version(),
		FileSeparator: cfg.fileSeparator(),
		LineSeparator: cfg.lineSeparator(),
		PathSeparator: cfg.pathSeparator(),
	}
}

// GetName returns the OS name (GOOS).
func (o *OsInfo) GetName() string { return o.Name }

// GetArch returns the OS architecture (GOARCH).
func (o *OsInfo) GetArch() string { return o.Arch }

// GetVersion returns the OS version.
func (o *OsInfo) GetVersion() string { return o.Version }

// GetFileSeparator returns the file path separator.
func (o *OsInfo) GetFileSeparator() string { return o.FileSeparator }

// GetLineSeparator returns the line separator.
func (o *OsInfo) GetLineSeparator() string { return o.LineSeparator }

// GetPathSeparator returns the environment path separator.
func (o *OsInfo) GetPathSeparator() string { return o.PathSeparator }

// IsLinux reports whether the OS is Linux.
func (o *OsInfo) IsLinux() bool { return o.Name == "linux" }

// IsMac reports whether the OS is macOS (Darwin).
func (o *OsInfo) IsMac() bool { return o.Name == "darwin" }

// IsMacOsX is equivalent to IsMac.
func (o *OsInfo) IsMacOsX() bool { return o.IsMac() }

// IsWindows reports whether the OS is Windows.
func (o *OsInfo) IsWindows() bool { return o.Name == "windows" }

// IsAix reports whether the OS is AIX.
func (o *OsInfo) IsAix() bool { return o.Name == "aix" }

// IsSolaris reports whether the OS is Solaris.
func (o *OsInfo) IsSolaris() bool { return o.Name == "solaris" }

// IsFreeBSD reports whether the OS is FreeBSD.
func (o *OsInfo) IsFreeBSD() bool { return o.Name == "freebsd" }

// String implements fmt.Stringer.
func (o *OsInfo) String() string {
	var b strings.Builder
	appendLine(&b, "OS Arch:        ", o.Arch)
	appendLine(&b, "OS Name:        ", o.Name)
	appendLine(&b, "OS Version:     ", o.Version)
	appendLine(&b, "File Separator: ", o.FileSeparator)
	appendLine(&b, "Line Separator: ", o.LineSeparator)
	appendLine(&b, "Path Separator: ", o.PathSeparator)
	return b.String()
}

// lineSeparator returns the line separator for the current OS.
func lineSeparator() string {
	if runtime.GOOS == "windows" {
		return "\r\n"
	}
	return "\n"
}

// readOsVersion detects the OS version from environment variables or common fallbacks.
// The Go standard library has no unified API for this, so this is best-effort.
func readOsVersion() string {
	if v := os.Getenv("OSVERSION"); v != "" {
		return v
	}
	if v := os.Getenv("OSTYPE"); v != "" {
		return v
	}
	return strings.TrimSpace(runtime.GOOS)
}
