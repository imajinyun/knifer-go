package system

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"sync"
)

// Singleton caches avoid repeated collection.
var (
	hostOnce    sync.Once
	hostInfo    *HostInfo
	osOnce      sync.Once
	osInfo      *OsInfo
	userOnce    sync.Once
	userInfo    *UserInfo
	goOnce      sync.Once
	goInfo      *GoInfo
	runtimeOnce sync.Once
	runtimeRef  *RuntimeInfo
)

// GetHostInfo returns cached host information.
func GetHostInfo() *HostInfo {
	hostOnce.Do(func() { hostInfo = NewHostInfo() })
	return hostInfo
}

// GetOsInfo returns cached OS information.
func GetOsInfo() *OsInfo {
	osOnce.Do(func() { osInfo = NewOsInfo() })
	return osInfo
}

// GetUserInfo returns cached user information.
func GetUserInfo() *UserInfo {
	userOnce.Do(func() { userInfo = NewUserInfo() })
	return userInfo
}

// GetUserInfoWithOptions returns uncached user information collected with per-call options.
func GetUserInfoWithOptions(opts ...UserInfoOption) *UserInfo {
	return NewUserInfoWithOptions(opts...)
}

// GetGoInfo returns cached Go runtime metadata.
func GetGoInfo() *GoInfo {
	goOnce.Do(func() { goInfo = NewGoInfo() })
	return goInfo
}

// GetRuntimeInfo returns runtime memory information and refreshes it on each call.
func GetRuntimeInfo() *RuntimeInfo {
	runtimeOnce.Do(func() { runtimeRef = NewRuntimeInfo() })
	return runtimeRef.Refresh()
}

// GetCurrentPID returns the current process PID.
func GetCurrentPID() int {
	return os.Getpid()
}

// GetTotalMemory returns total memory requested from the OS by the current Go program.
func GetTotalMemory() uint64 {
	return GetRuntimeInfo().GetTotalMemory()
}

// GetFreeMemory returns idle memory in the current Go program.
func GetFreeMemory() uint64 {
	return GetRuntimeInfo().GetFreeMemory()
}

// GetMaxMemory returns the detected memory upper bound for the current Go program.
func GetMaxMemory() uint64 {
	return GetRuntimeInfo().GetMaxMemory()
}

// GetTotalThreadCount returns the total goroutine count.
func GetTotalThreadCount() int {
	return runtime.NumGoroutine()
}

// Get returns an environment variable by key.
// If quiet is false and the variable is missing, it prints a warning to stderr.
func Get(key string, quiet bool) string {
	v, ok := os.LookupEnv(key)
	if !ok && !quiet {
		fmt.Fprintf(os.Stderr, "[gksystem] env %q not found\n", key)
	}
	return v
}

// GetOrDefault returns an environment variable, or def when it is missing or empty.
func GetOrDefault(key, def string) string {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def
	}
	return v
}

// GetInt returns an environment variable as an int, or def on conversion failure.
func GetInt(key string, def int) int {
	v, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

// GetBool returns an environment variable as a bool, or def on conversion failure.
func GetBool(key string, def bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

// DumpSystemInfo writes system information to stdout.
func DumpSystemInfo() {
	DumpSystemInfoTo(os.Stdout)
}

// DumpSystemInfoTo writes system information to the specified writer.
func DumpSystemInfoTo(w io.Writer) {
	const sep = "--------------\n"
	_, _ = fmt.Fprint(w, sep)
	_, _ = fmt.Fprint(w, GetGoInfo())
	_, _ = fmt.Fprint(w, sep)
	_, _ = fmt.Fprint(w, GetOsInfo())
	_, _ = fmt.Fprint(w, sep)
	_, _ = fmt.Fprint(w, GetUserInfo())
	_, _ = fmt.Fprint(w, sep)
	_, _ = fmt.Fprint(w, GetHostInfo())
	_, _ = fmt.Fprint(w, sep)
	_, _ = fmt.Fprint(w, GetRuntimeInfo())
	_, _ = fmt.Fprint(w, sep)
}
