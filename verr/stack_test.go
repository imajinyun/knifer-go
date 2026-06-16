package verr_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/imajinyun/go-knifer/verr"
)

func TestStackFacade(t *testing.T) {
	stack := verr.GetStackTrace(0)
	if len(stack) == 0 {
		t.Fatal("GetStackTrace() returned empty stack")
	}
	if formatted := fmt.Sprintf("%+v", stack); !strings.Contains(formatted, "TestStackFacade") {
		t.Fatalf("formatted stack = %q, want current test", formatted)
	}
}

func TestStackTraceWithOptionsFacade(t *testing.T) {
	stack := verr.GetStackTraceWithOptions(verr.WithStackSkip(0), verr.WithStackDepth(4))
	if len(stack) == 0 || len(stack) > 4 {
		t.Fatalf("GetStackTraceWithOptions length = %d, want 1..4", len(stack))
	}
	formatted := fmt.Sprintf("%+v", stack)
	if !strings.Contains(formatted, "TestStackTraceWithOptionsFacade") {
		t.Fatalf("formatted stack = %q, want current test", formatted)
	}
}

type facadeStackError struct{ error }

func (facadeStackError) Stack() string { return "attached facade stack" }

func TestGetStackWithOptionsFacade(t *testing.T) {
	if got := verr.GetStackWithOptions(nil, verr.WithDebugStackFunc(func() []byte { return []byte("unused") })); got != "" {
		t.Fatalf("GetStackWithOptions(nil) = %q, want empty", got)
	}

	stacked := facadeStackError{error: errors.New("stacked")}
	if got := verr.GetStackWithOptions(stacked, verr.WithDebugStackFunc(func() []byte { return []byte("unused") })); got != "attached facade stack" {
		t.Fatalf("GetStackWithOptions(stacked) = %q, want attached stack", got)
	}

	plain := errors.New("plain")
	if got := verr.GetStackWithOptions(plain, verr.WithDebugStackFunc(func() []byte { return []byte("fallback facade stack") })); got != "fallback facade stack" {
		t.Fatalf("GetStackWithOptions(plain) = %q, want fallback stack", got)
	}
}
