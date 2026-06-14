package verr_test

import (
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
