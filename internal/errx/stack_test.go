package errx

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"testing"
)

func TestGetStackTraceAndFormatting(t *testing.T) {
	stack := GetStackTrace(0)
	if len(stack) == 0 {
		t.Fatal("GetStackTrace() returned an empty stack")
	}
	if neg := GetStackTrace(-1); len(neg) == 0 {
		t.Fatal("GetStackTrace(-1) returned an empty stack")
	}

	short := fmt.Sprintf("%v", stack)
	if !strings.HasPrefix(short, "[") || !strings.HasSuffix(short, "]") {
		t.Fatalf("short stack format = %q, want bracketed slice", short)
	}
	verbose := fmt.Sprintf("%+v", stack)
	if !strings.Contains(verbose, "TestGetStackTraceAndFormatting") {
		t.Fatalf("verbose stack format should include test function, got %q", verbose)
	}
	goSyntax := fmt.Sprintf("%#v", stack)
	if !strings.Contains(goSyntax, "errx.Frame") {
		t.Fatalf("go-syntax stack format = %q, want frame type", goSyntax)
	}
}

func TestGetStackTraceWithOptions(t *testing.T) {
	stack := GetStackTraceWithOptions(WithStackSkip(0), WithStackDepth(2))
	if len(stack) == 0 || len(stack) > 2 {
		t.Fatalf("GetStackTraceWithOptions len = %d", len(stack))
	}
}

func TestStackFrameCacheResetAndDisable(t *testing.T) {
	ResetStackFrameCache()
	var pcs [1]uintptr
	runtime.Callers(0, pcs[:])
	// copy into caller-provided slice while preserving the captured PC.
	callers := func(_ int, out []uintptr) int {
		out[0] = pcs[0]
		return 1
	}
	resolver := func(uintptr) (string, int, string, bool) {
		return "/virtual/custom.go", 123, "virtual.Custom", true
	}

	stack := GetStackTraceWithOptions(WithCallersFunc(callers), WithFuncForPCFunc(resolver))
	if got := fmt.Sprintf("%s:%d:%n", stack[0], stack[0], stack[0]); got != "custom.go:123:Custom" {
		t.Fatalf("cached custom frame = %q", got)
	}

	ResetStackFrameCache()
	if got := fmt.Sprintf("%s:%d:%n", stack[0], stack[0], stack[0]); got == "custom.go:123:Custom" {
		t.Fatalf("ResetStackFrameCache should clear custom metadata, got %q", got)
	}

	stack = GetStackTraceWithOptions(WithCallersFunc(callers), WithFuncForPCFunc(resolver), WithStackFrameCache(false))
	if got := fmt.Sprintf("%s:%d:%n", stack[0], stack[0], stack[0]); got == "custom.go:123:Custom" {
		t.Fatalf("WithStackFrameCache(false) should not store custom metadata, got %q", got)
	}
}

func TestStackFrameCacheConcurrentResetAndCapture(t *testing.T) {
	ResetStackFrameCache()

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				ResetStackFrameCache()
			}
		}()
	}
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				if stack := GetStackTrace(0); len(stack) == 0 {
					t.Error("GetStackTrace returned empty stack")
				}
				if stack := GetStackTraceWithOptions(WithStackFrameCache(false)); len(stack) == 0 {
					t.Error("GetStackTraceWithOptions returned empty stack")
				}
			}
		}()
	}
	wg.Wait()
}

func TestFrameFormatting(t *testing.T) {
	stack := GetStackTrace(0)
	frame := stack[0]

	if got := fmt.Sprintf("%s", frame); got == "" || got == "unknown" {
		t.Fatalf("frame %%s = %q", got)
	}
	if got := fmt.Sprintf("%d", frame); got == "0" || got == "" {
		t.Fatalf("frame %%d = %q", got)
	}
	if got := fmt.Sprintf("%n", frame); got == "" || got == "unknown" {
		t.Fatalf("frame %%n = %q", got)
	}
	if got := fmt.Sprintf("%+s", frame); !strings.Contains(got, "\n\t") {
		t.Fatalf("frame %%+s = %q, want function and file", got)
	}
}
