package errx

import (
	"fmt"
	"io"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

type stackTraceConfig struct {
	skip      int
	depth     int
	callers   CallersFunc
	funcForPC FuncForPCFunc
	cache     bool
}

// CallersFunc captures call stack PCs.
type CallersFunc func(skip int, pc []uintptr) int

// FuncForPCFunc resolves a PC into file, line, and function name.
type FuncForPCFunc func(pc uintptr) (file string, line int, name string, ok bool)

// StackTraceOption customizes stack trace capture.
type StackTraceOption func(*stackTraceConfig)

// WithStackSkip sets the number of caller frames to skip.
func WithStackSkip(skip int) StackTraceOption {
	return func(c *stackTraceConfig) { c.skip = skip }
}

// WithStackDepth sets the maximum number of stack frames to capture.
func WithStackDepth(depth int) StackTraceOption {
	return func(c *stackTraceConfig) { c.depth = depth }
}

// WithCallersFunc sets the function used to capture stack PCs.
func WithCallersFunc(callers CallersFunc) StackTraceOption {
	return func(c *stackTraceConfig) {
		if callers != nil {
			c.callers = callers
		}
	}
}

// WithFuncForPCFunc sets the resolver used to format captured stack frames.
func WithFuncForPCFunc(fn FuncForPCFunc) StackTraceOption {
	return func(c *stackTraceConfig) {
		if fn != nil {
			c.funcForPC = fn
		}
	}
}

// WithStackFrameCache controls whether captured frame metadata is stored in the package-level cache.
func WithStackFrameCache(enabled bool) StackTraceOption {
	return func(c *stackTraceConfig) { c.cache = enabled }
}

func applyStackTraceOptions(skip int, opts []StackTraceOption) stackTraceConfig {
	cfg := stackTraceConfig{skip: skip, depth: 32, callers: runtime.Callers, funcForPC: defaultFuncForPC, cache: true}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.skip < 0 {
		cfg.skip = 0
	}
	if cfg.depth <= 0 {
		cfg.depth = 32
	}
	if cfg.callers == nil {
		cfg.callers = runtime.Callers
	}
	if cfg.funcForPC == nil {
		cfg.funcForPC = defaultFuncForPC
	}
	return cfg
}

type frameInfo struct {
	file string
	line int
	name string
	ok   bool
}

var capturedFrameInfo sync.Map

// ResetStackFrameCache clears cached stack frame metadata captured by GetStackTraceWithOptions.
func ResetStackFrameCache() { capturedFrameInfo.Clear() }

func defaultFuncForPC(pc uintptr) (string, int, string, bool) {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown", 0, "unknown", false
	}
	file, line := fn.FileLine(pc)
	return file, line, fn.Name(), true
}

func resolveFrame(pc uintptr) frameInfo {
	if info, ok := capturedFrameInfo.Load(pc); ok {
		return info.(frameInfo)
	}
	file, line, name, ok := defaultFuncForPC(pc)
	return frameInfo{file: file, line: line, name: name, ok: ok}
}

// WithStackTrace is implemented by errors that expose structured stack frames.
type WithStackTrace interface {
	StackTrace() StackTrace
}

// Frame represents a program counter inside a stack frame.
// For historical reasons, if Frame is interpreted as a uintptr, its value is
// the program counter plus one.
type Frame uintptr

func (f Frame) pc() uintptr { return uintptr(f) - 1 }

func (f Frame) file() string {
	info := resolveFrame(f.pc())
	if !info.ok {
		return "unknown"
	}
	return info.file
}

func (f Frame) line() int {
	info := resolveFrame(f.pc())
	if !info.ok {
		return 0
	}
	return info.line
}

func (f Frame) name() string {
	info := resolveFrame(f.pc())
	if !info.ok {
		return "unknown"
	}
	return info.name
}

// Format formats the frame according to fmt.Formatter.
func (f Frame) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		if s.Flag('+') {
			_, _ = io.WriteString(s, f.name())
			_, _ = io.WriteString(s, "\n\t")
			_, _ = io.WriteString(s, f.file())
			return
		}
		_, _ = io.WriteString(s, path.Base(f.file()))
	case 'd':
		_, _ = io.WriteString(s, strconv.Itoa(f.line()))
	case 'n':
		_, _ = io.WriteString(s, funcname(f.name()))
	case 'v':
		f.Format(s, 's')
		_, _ = io.WriteString(s, ":")
		f.Format(s, 'd')
	}
}

// StackTrace is a stack of frames from innermost to outermost.
type StackTrace []Frame

// Format formats the stack of frames according to fmt.Formatter.
func (st StackTrace) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			for _, f := range st {
				_, _ = io.WriteString(s, "\n")
				f.Format(s, verb)
			}
			return
		}
		if s.Flag('#') {
			_, _ = fmt.Fprintf(s, "%#v", []Frame(st))
			return
		}
		st.formatSlice(s, verb)
	case 's':
		st.formatSlice(s, verb)
	}
}

func (st StackTrace) formatSlice(s fmt.State, verb rune) {
	_, _ = io.WriteString(s, "[")
	for i, f := range st {
		if i > 0 {
			_, _ = io.WriteString(s, " ")
		}
		f.Format(s, verb)
	}
	_, _ = io.WriteString(s, "]")
}

// GetStackTrace captures the current goroutine stack trace.
func GetStackTrace(skip int) StackTrace {
	return GetStackTraceWithOptions(WithStackSkip(skip))
}

// GetStackTraceWithOptions captures the current goroutine stack trace with custom options.
func GetStackTraceWithOptions(opts ...StackTraceOption) StackTrace {
	cfg := applyStackTraceOptions(0, opts)
	pcs := make([]uintptr, cfg.depth)
	n := cfg.callers(cfg.skip, pcs)
	stack := make(StackTrace, n)
	for idx, frame := range pcs[:n] {
		stack[idx] = Frame(frame)
		pc := Frame(frame).pc()
		file, line, name, ok := cfg.funcForPC(pc)
		if cfg.cache {
			capturedFrameInfo.Store(pc, frameInfo{file: file, line: line, name: name, ok: ok})
		}
	}
	return stack
}

func funcname(name string) string {
	i := strings.LastIndex(name, "/")
	name = name[i+1:]
	i = strings.Index(name, ".")
	if i < 0 {
		return name
	}
	return name[i+1:]
}
