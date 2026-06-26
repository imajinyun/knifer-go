package vlog_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/imajinyun/knifer-go/vlog"
)

func TestFacadeConsoleLogOptions(t *testing.T) {
	old := vlog.GetLogLevel()
	vlog.SetLogLevel(vlog.LogLevelDebug)
	defer vlog.SetLogLevel(old)

	out := &bytes.Buffer{}
	fixed := time.Date(2024, 4, 5, 6, 7, 8, 0, time.UTC)
	log := vlog.NewConsoleLogWithOptions("facade.options",
		vlog.WithLogClock(func() time.Time { return fixed }),
		vlog.WithLogTimeLayout(time.RFC3339),
		vlog.WithLogOutput(out, &bytes.Buffer{}),
	)
	log.Info("hello")
	if !strings.Contains(out.String(), "2024-04-05T06:07:08Z") || !strings.Contains(out.String(), "hello") {
		t.Fatalf("console log options not applied: %q", out.String())
	}

	colorOut := &bytes.Buffer{}
	colorLog := vlog.NewConsoleColorLogWithOptions("facade.color",
		vlog.WithLogClock(func() time.Time { return fixed }),
		vlog.WithLogTimeLayout("15:04"),
		vlog.WithLogOutput(colorOut, &bytes.Buffer{}),
	)
	colorLog.Info("color")
	if !strings.Contains(colorOut.String(), "06:07") || !strings.Contains(colorOut.String(), "color") {
		t.Fatalf("color log options not applied: %q", colorOut.String())
	}

	customColorOut := &bytes.Buffer{}
	customColorLog := vlog.NewConsoleColorLogWithOptions("facade.color.custom",
		vlog.WithLogOutput(customColorOut, &bytes.Buffer{}),
		vlog.WithLogColorFactory(func(vlog.Level) string { return "\033[36m" }),
	)
	customColorLog.Info("custom-color")
	if !strings.Contains(customColorOut.String(), "\033[36m") || !strings.Contains(customColorOut.String(), "custom-color") {
		t.Fatalf("color factory option not applied: %q", customColorOut.String())
	}
}

func TestFacadeConsoleLogLevelAndWriters(t *testing.T) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	log := vlog.NewConsoleLogWithOptions("facade.level",
		vlog.WithLogOutput(out, errOut),
		vlog.WithLogLevel(vlog.LogLevelWarn),
	)
	if log.IsInfoEnabled() || !log.IsWarnEnabled() || !log.IsErrorEnabled() {
		t.Fatal("instance log level should filter info and enable warn/error")
	}
	log.Info("hidden")
	log.Warn("visible-warn")
	log.LogE(vlog.LogLevelError, errors.New("boom"), "visible {}", "error")
	if out.String() != "" {
		t.Fatalf("info output should be filtered, stdout=%q", out.String())
	}
	if got := errOut.String(); !strings.Contains(got, "visible-warn") || !strings.Contains(got, "visible error") || !strings.Contains(got, "boom") {
		t.Fatalf("stderr output = %q", got)
	}
}

func TestFacadeGlobalColorFactory(t *testing.T) {
	oldLevel := vlog.GetLogLevel()
	vlog.SetLogLevel(vlog.LogLevelInfo)
	t.Cleanup(func() { vlog.SetLogLevel(oldLevel) })

	out := &bytes.Buffer{}
	vlog.SetColorFactory(func(vlog.Level) string { return "\033[35m" })
	vlog.SetLogColorFactory(func(vlog.Level) string { return "\033[34m" })
	log := vlog.NewConsoleColorLogWithOptions("facade.global.color", vlog.WithLogOutput(out, &bytes.Buffer{}))
	log.Info("global-color")
	if got := out.String(); !strings.Contains(got, "\033[34m") || !strings.Contains(got, "global-color") {
		t.Fatalf("global color output = %q", got)
	}
}

func TestFacadeNewConsoleColorLog(t *testing.T) {
	log := vlog.NewConsoleColorLog("facade.color.default")
	if log == nil || log.GetName() != "facade.color.default" {
		t.Fatalf("NewConsoleColorLog = %#v", log)
	}
}
