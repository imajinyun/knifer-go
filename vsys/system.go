package vsys

import (
	"io"
	"net"
	"os/user"
	"runtime"

	"github.com/imajinyun/knifer-go/internal/system"
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

// HostInfoOption customizes host information collection per call.
type HostInfoOption = system.HostInfoOption

// GoInfoOption customizes Go runtime metadata collection per call.
type GoInfoOption = system.GoInfoOption

// OsInfoOption customizes OS information collection per call.
type OsInfoOption = system.OsInfoOption

// RuntimeInfoOption customizes runtime information collection per call.
type RuntimeInfoOption = system.RuntimeInfoOption

// ProcessOption customizes process/runtime scalar helpers per call.
type ProcessOption = system.ProcessOption

// EnvOption customizes environment helpers per call.
type EnvOption = system.EnvOption

// DumpOption customizes system information dumping per call.
type DumpOption = system.DumpOption

// UserInfoOption customizes user information collection per call.
type UserInfoOption = system.UserInfoOption

// WithHostNameFunc sets the function used to collect the host name.
func WithHostNameFunc(fn func() (string, error)) HostInfoOption {
	return system.WithHostNameFunc(fn)
}

// WithHostInterfaceAddrsFunc sets the function used to collect local interface addresses.
func WithHostInterfaceAddrsFunc(fn func() ([]net.Addr, error)) HostInfoOption {
	return system.WithHostInterfaceAddrsFunc(fn)
}

// WithHostAddressFunc sets the function used to collect the host address directly.
func WithHostAddressFunc(fn func() string) HostInfoOption {
	return system.WithHostAddressFunc(fn)
}

// WithGoVersionFunc sets the function used to collect the Go version.
func WithGoVersionFunc(fn func() string) GoInfoOption { return system.WithGoVersionFunc(fn) }

// WithGoCompilerFunc sets the function used to collect the Go compiler name.
func WithGoCompilerFunc(fn func() string) GoInfoOption { return system.WithGoCompilerFunc(fn) }

// WithGoRootFunc sets the function used to collect GOROOT.
func WithGoRootFunc(fn func() string) GoInfoOption { return system.WithGoRootFunc(fn) }

// WithGoEnvOutputFunc sets the command runner used by the default GOROOT collector.
func WithGoEnvOutputFunc(fn func(string, ...string) ([]byte, error)) GoInfoOption {
	return system.WithGoEnvOutputFunc(fn)
}

// WithGoRootEnvLookupFunc sets the environment lookup used by the default GOROOT collector.
func WithGoRootEnvLookupFunc(fn func(string) string) GoInfoOption {
	return system.WithGoRootEnvLookupFunc(fn)
}

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

// WithOSEnvLookupFunc sets the environment lookup used by the default OS version collector.
func WithOSEnvLookupFunc(fn func(string) string) OsInfoOption { return system.WithOSEnvLookupFunc(fn) }

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

// WithReadMemStatsFunc sets the function used to collect memory statistics.
func WithReadMemStatsFunc(fn func(*runtime.MemStats)) RuntimeInfoOption {
	return system.WithReadMemStatsFunc(fn)
}

// WithNumGoroutineFunc sets the function used to collect the goroutine count.
func WithNumGoroutineFunc(fn func() int) RuntimeInfoOption {
	return system.WithNumGoroutineFunc(fn)
}

// WithPIDFunc sets the function used to collect the current process id.
func WithPIDFunc(fn func() int) ProcessOption { return system.WithPIDFunc(fn) }

// WithProcessNumGoroutineFunc sets the function used by process scalar helpers to collect goroutine count.
func WithProcessNumGoroutineFunc(fn func() int) ProcessOption {
	return system.WithProcessNumGoroutineFunc(fn)
}

// WithEnvLookupFunc sets the function used to look up environment variables.
func WithEnvLookupFunc(fn func(string) (string, bool)) EnvOption {
	return system.WithEnvLookupFunc(fn)
}

// WithEnvWarningWriter sets the writer used for missing-variable warnings.
func WithEnvWarningWriter(w io.Writer) EnvOption { return system.WithEnvWarningWriter(w) }

// WithEnvIntParser sets the parser used by EnvIntWithOptions.
func WithEnvIntParser(parser func(string) (int, error)) EnvOption {
	return system.WithEnvIntParser(parser)
}

// WithEnvBoolParser sets the parser used by EnvBoolWithOptions.
func WithEnvBoolParser(parser func(string) (bool, error)) EnvOption {
	return system.WithEnvBoolParser(parser)
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

// WithDumpHostOptions sets host information providers used by DumpSystemInfoWithOptions.
func WithDumpHostOptions(opts ...HostInfoOption) DumpOption {
	return system.WithDumpHostOptions(opts...)
}

// WithDumpOsOptions sets OS information providers used by DumpSystemInfoWithOptions.
func WithDumpOsOptions(opts ...OsInfoOption) DumpOption { return system.WithDumpOsOptions(opts...) }

// WithDumpUserOptions sets user information providers used by DumpSystemInfoWithOptions.
func WithDumpUserOptions(opts ...UserInfoOption) DumpOption {
	return system.WithDumpUserOptions(opts...)
}

// WithDumpGoOptions sets Go runtime metadata providers used by DumpSystemInfoWithOptions.
func WithDumpGoOptions(opts ...GoInfoOption) DumpOption { return system.WithDumpGoOptions(opts...) }

// WithDumpRuntimeOptions sets runtime information providers used by DumpSystemInfoWithOptions.
func WithDumpRuntimeOptions(opts ...RuntimeInfoOption) DumpOption {
	return system.WithDumpRuntimeOptions(opts...)
}

// SystemHostInfo returns cached host information.
func SystemHostInfo() *HostInfo { return system.GetHostInfo() }

// ResetInfoCache clears cached singleton system information.
func ResetInfoCache() { system.ResetInfoCache() }

// SysHostInfoWithOptions returns uncached host information collected with per-call options.
func SysHostInfoWithOptions(opts ...HostInfoOption) *HostInfo {
	return system.GetHostInfoWithOptions(opts...)
}

// SystemHostInfoWithOptions returns uncached host information collected with per-call options.
func SystemHostInfoWithOptions(opts ...HostInfoOption) *HostInfo {
	return SysHostInfoWithOptions(opts...)
}

// SystemOsInfo returns cached operating system information.
func SystemOsInfo() *OsInfo { return system.GetOsInfo() }

// SysOsInfoWithOptions returns uncached operating system information collected with per-call options.
func SysOsInfoWithOptions(opts ...OsInfoOption) *OsInfo {
	return system.GetOsInfoWithOptions(opts...)
}

// SystemOsInfoWithOptions returns uncached operating system information collected with per-call options.
func SystemOsInfoWithOptions(opts ...OsInfoOption) *OsInfo {
	return SysOsInfoWithOptions(opts...)
}

// SystemUserInfo returns cached user information.
func SystemUserInfo() *UserInfo { return system.GetUserInfo() }

// SysUserInfoWithOptions returns uncached user information collected with per-call options.
func SysUserInfoWithOptions(opts ...UserInfoOption) *UserInfo {
	return system.GetUserInfoWithOptions(opts...)
}

// SystemUserInfoWithOptions returns uncached user information collected with per-call options.
func SystemUserInfoWithOptions(opts ...UserInfoOption) *UserInfo {
	return SysUserInfoWithOptions(opts...)
}

// SystemGoInfo returns cached Go runtime metadata.
func SystemGoInfo() *GoInfo { return system.GetGoInfo() }

// SysGoInfoWithOptions returns uncached Go runtime metadata collected with per-call options.
func SysGoInfoWithOptions(opts ...GoInfoOption) *GoInfo {
	return system.GetGoInfoWithOptions(opts...)
}

// SystemGoInfoWithOptions returns uncached Go runtime metadata collected with per-call options.
func SystemGoInfoWithOptions(opts ...GoInfoOption) *GoInfo {
	return SysGoInfoWithOptions(opts...)
}

// SystemRuntimeInfo returns refreshed runtime statistics.
func SystemRuntimeInfo() *RuntimeInfo { return system.GetRuntimeInfo() }

// SysRuntimeInfoWithOptions returns uncached runtime statistics collected with per-call options.
func SysRuntimeInfoWithOptions(opts ...RuntimeInfoOption) *RuntimeInfo {
	return system.GetRuntimeInfoWithOptions(opts...)
}

// SystemRuntimeInfoWithOptions returns uncached runtime statistics collected with per-call options.
func SystemRuntimeInfoWithOptions(opts ...RuntimeInfoOption) *RuntimeInfo {
	return SysRuntimeInfoWithOptions(opts...)
}

// CurrentPID returns the current process id.
func CurrentPID() int { return system.GetCurrentPID() }

// CurrentPIDWithOptions returns the current process id using custom providers.
func CurrentPIDWithOptions(opts ...ProcessOption) int {
	return system.GetCurrentPIDWithOptions(opts...)
}

// TotalMemory returns memory allocated from OS by the current Go process.
func TotalMemory() uint64 { return system.GetTotalMemory() }

// TotalMemoryWithOptions returns memory allocated from OS using custom runtime providers.
func TotalMemoryWithOptions(opts ...RuntimeInfoOption) uint64 {
	return system.GetTotalMemoryWithOptions(opts...)
}

// FreeMemory returns idle memory in the current Go process.
func FreeMemory() uint64 { return system.GetFreeMemory() }

// FreeMemoryWithOptions returns idle memory using custom runtime providers.
func FreeMemoryWithOptions(opts ...RuntimeInfoOption) uint64 {
	return system.GetFreeMemoryWithOptions(opts...)
}

// MaxMemory returns the detected memory upper bound.
func MaxMemory() uint64 { return system.GetMaxMemory() }

// MaxMemoryWithOptions returns the detected memory upper bound using custom runtime providers.
func MaxMemoryWithOptions(opts ...RuntimeInfoOption) uint64 {
	return system.GetMaxMemoryWithOptions(opts...)
}

// TotalGoroutineCount returns the current goroutine count.
func TotalGoroutineCount() int { return system.GetTotalThreadCount() }

// TotalGoroutineCountWithOptions returns the current goroutine count using custom providers.
func TotalGoroutineCountWithOptions(opts ...ProcessOption) int {
	return system.GetTotalThreadCountWithOptions(opts...)
}

// Env returns an environment variable value.
func Env(key string) string { return system.Get(key, true) }

// EnvWithOptions returns an environment variable value using custom providers.
func EnvWithOptions(key string, opts ...EnvOption) string {
	return system.GetWithOptions(key, true, opts...)
}

// EnvOrDefault returns an environment variable or def when empty/missing.
func EnvOrDefault(key, def string) string { return system.GetOrDefault(key, def) }

// EnvOrDefaultWithOptions returns an environment variable or def using custom providers.
func EnvOrDefaultWithOptions(key, def string, opts ...EnvOption) string {
	return system.GetOrDefaultWithOptions(key, def, opts...)
}

// EnvInt returns an int environment variable or def when missing/invalid.
func EnvInt(key string, def int) int { return system.GetInt(key, def) }

// EnvIntWithOptions returns an int environment variable or def using custom providers.
func EnvIntWithOptions(key string, def int, opts ...EnvOption) int {
	return system.GetIntWithOptions(key, def, opts...)
}

// EnvBool returns a bool environment variable or def when missing/invalid.
func EnvBool(key string, def bool) bool { return system.GetBool(key, def) }

// EnvBoolWithOptions returns a bool environment variable or def using custom providers.
func EnvBoolWithOptions(key string, def bool, opts ...EnvOption) bool {
	return system.GetBoolWithOptions(key, def, opts...)
}

// DumpSystemInfo writes system information to stdout.
func DumpSystemInfo() { system.DumpSystemInfo() }

// DumpSystemInfoTo writes system information to w.
func DumpSystemInfoTo(w io.Writer) { system.DumpSystemInfoTo(w) }

// DumpSystemInfoWithOptions writes uncached system information to w using per-call providers.
func DumpSystemInfoWithOptions(w io.Writer, opts ...DumpOption) {
	system.DumpSystemInfoWithOptions(w, opts...)
}
