package log

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestConsoleLogWithOptions(t *testing.T) {
	prevLevel := GetConsoleLevel()
	SetConsoleLevel(LevelDebug)
	defer SetConsoleLevel(prevLevel)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	fixed := time.Date(2024, 2, 3, 4, 5, 6, 0, time.UTC)
	c := NewConsoleLogWithOptions("test.options",
		WithLogClock(func() time.Time { return fixed }),
		WithLogTimeLayout(time.RFC3339),
		WithLogOutput(out, errOut),
	)
	c.Info("hello")
	c.Warn("careful")

	if !strings.Contains(out.String(), "2024-02-03T04:05:06Z") || !strings.Contains(out.String(), "hello") {
		t.Fatalf("custom clock/layout/output not applied to stdout: %q", out.String())
	}
	if !strings.Contains(errOut.String(), "2024-02-03T04:05:06Z") || !strings.Contains(errOut.String(), "careful") {
		t.Fatalf("custom clock/layout/output not applied to stderr: %q", errOut.String())
	}
}

func TestNilLogOutputOptionDoesNotClearPreviousWriters(t *testing.T) {
	prevLevel := GetConsoleLevel()
	SetConsoleLevel(LevelDebug)
	defer SetConsoleLevel(prevLevel)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	c := NewConsoleLogWithOptions("test.nil.output",
		WithLogOutput(out, errOut),
		WithLogOutput(nil, nil),
	)
	c.Info("hello")
	c.Warn("careful")

	if !strings.Contains(out.String(), "hello") {
		t.Fatalf("nil WithLogOutput cleared stdout writer: %q", out.String())
	}
	if !strings.Contains(errOut.String(), "careful") {
		t.Fatalf("nil WithLogOutput cleared stderr writer: %q", errOut.String())
	}
}
