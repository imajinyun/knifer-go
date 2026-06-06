package vsys

import (
	"io"
	"os/user"

	"github.com/imajinyun/go-knifer/internal/system"
)

// HostInfo describes current host information.
type HostInfo = system.HostInfo

// OsInfo describes current operating system information.
type OsInfo = system.OsInfo

// UserInfo describes current user information.
type UserInfo = system.UserInfo

// GoInfo describes Go runtime metadata.
type GoInfo = system.GoInfo

// RuntimeInfo describes current process runtime statistics.
type RuntimeInfo = system.RuntimeInfo

// GoInfoOption customizes Go runtime metadata collection per call.
type GoInfoOption = system.GoInfoOption

// OsInfoOption customizes OS information collection per call.
type OsInfoOption = system.OsInfoOption

// UserInfoOption customizes user information collection per call.
type UserInfoOption = system.UserInfoOption

// WithGoVersionFunc sets the function used to collect the Go version.
func WithGoVersionFunc(fn func() string) GoInfoOption { return system.WithGoVersionFunc(fn) }

// WithGoCompilerFunc sets the function used to collect the Go compiler name.
func WithGoCompilerFunc(fn func() string) GoInfoOption { return system.WithGoCompilerFunc(fn) }

// WithGoRootFunc sets the function used to collect GOROOT.
func WithGoRootFunc(fn func() string) GoInfoOption { return system.WithGoRootFunc(fn) }

// WithGoOSFunc sets the function used to collect GOOS.
func WithGoOSFunc(fn func() string) GoInfoOption { return system.WithGoOSFunc(fn) }

// WithGoArchFunc sets the function used to collect GOARCH.
func WithGoArchFunc(fn func() string) GoInfoOption { return system.WithGoArchFunc(fn) }

// WithGoNumCPUFunc sets the function used to collect the CPU count.
func WithGoNumCPUFunc(fn func() int) GoInfoOption { return system.WithGoNumCPUFunc(fn) }

// WithGoNumCgoCallFunc sets the function used to collect the cgo call count.
func WithGoNumCgoCallFunc(fn func() int64) GoInfoOption {
	return system.WithGoNumCgoCallFunc(fn)
}

// WithOSNameFunc sets the function used to collect the OS name.
func WithOSNameFunc(fn func() string) OsInfoOption { return system.WithOSNameFunc(fn) }

// WithOSArchFunc sets the function used to collect the OS architecture.
func WithOSArchFunc(fn func() string) OsInfoOption { return system.WithOSArchFunc(fn) }

// WithOSVersionFunc sets the function used to collect the OS version.
func WithOSVersionFunc(fn func() string) OsInfoOption { return system.WithOSVersionFunc(fn) }

// WithOSFileSeparatorFunc sets the function used to collect the file separator.
func WithOSFileSeparatorFunc(fn func() string) OsInfoOption {
	return system.WithOSFileSeparatorFunc(fn)
}

// WithOSLineSeparatorFunc sets the function used to collect the line separator.
func WithOSLineSeparatorFunc(fn func() string) OsInfoOption {
	return system.WithOSLineSeparatorFunc(fn)
}

// WithOSPathSeparatorFunc sets the function used to collect the path-list separator.
func WithOSPathSeparatorFunc(fn func() string) OsInfoOption {
	return system.WithOSPathSeparatorFunc(fn)
}

// WithCurrentUserFunc sets the function used to discover the current OS user.
func WithCurrentUserFunc(fn func() (*user.User, error)) UserInfoOption {
	return system.WithCurrentUserFunc(fn)
}

// WithUserEnvLookup sets the environment lookup function used by NewUserInfoWithOptions.
func WithUserEnvLookup(lookup func(string) string) UserInfoOption {
	return system.WithUserEnvLookup(lookup)
}

// WithWorkingDirFunc sets the function used to discover the current working directory.
func WithWorkingDirFunc(fn func() (string, error)) UserInfoOption {
	return system.WithWorkingDirFunc(fn)
}

// WithTempDirFunc sets the function used to discover the temporary directory.
func WithTempDirFunc(fn func() string) UserInfoOption { return system.WithTempDirFunc(fn) }

// SystemHostInfo returns cached host information.
func SystemHostInfo() *HostInfo { return system.GetHostInfo() }

// SystemOsInfo returns cached operating system information.
func SystemOsInfo() *OsInfo { return system.GetOsInfo() }

// SystemUserInfo returns cached user information.
func SystemUserInfo() *UserInfo { return system.GetUserInfo() }

// SystemUserInfoWithOptions returns uncached user information collected with per-call options.
func SystemUserInfoWithOptions(opts ...UserInfoOption) *UserInfo {
	return system.GetUserInfoWithOptions(opts...)
}

// SystemGoInfo returns cached Go runtime metadata.
func SystemGoInfo() *GoInfo { return system.GetGoInfo() }

// SystemRuntimeInfo returns refreshed runtime statistics.
func SystemRuntimeInfo() *RuntimeInfo { return system.GetRuntimeInfo() }

// CurrentPID returns the current process id.
func CurrentPID() int { return system.GetCurrentPID() }

// TotalMemory returns memory allocated from OS by the current Go process.
func TotalMemory() uint64 { return system.GetTotalMemory() }

// FreeMemory returns idle memory in the current Go process.
func FreeMemory() uint64 { return system.GetFreeMemory() }

// MaxMemory returns the detected memory upper bound.
func MaxMemory() uint64 { return system.GetMaxMemory() }

// TotalGoroutineCount returns the current goroutine count.
func TotalGoroutineCount() int { return system.GetTotalThreadCount() }

// Env returns an environment variable value.
func Env(key string) string { return system.Get(key, true) }

// EnvOrDefault returns an environment variable or def when empty/missing.
func EnvOrDefault(key, def string) string { return system.GetOrDefault(key, def) }

// EnvInt returns an int environment variable or def when missing/invalid.
func EnvInt(key string, def int) int { return system.GetInt(key, def) }

// EnvBool returns a bool environment variable or def when missing/invalid.
func EnvBool(key string, def bool) bool { return system.GetBool(key, def) }

// DumpSystemInfo writes system information to stdout.
func DumpSystemInfo() { system.DumpSystemInfo() }

// DumpSystemInfoTo writes system information to w.
func DumpSystemInfoTo(w io.Writer) { system.DumpSystemInfoTo(w) }
